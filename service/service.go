package service

import (
	"context"
	"time"

	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/dp-search-data-extractor/clients"
	"github.com/ONSdigital/dp-search-data-extractor/config"
	"github.com/ONSdigital/dp-search-data-extractor/event"
	"github.com/ONSdigital/dp-search-data-extractor/handler"
	"github.com/ONSdigital/dp-search-data-extractor/schema"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// Service contains all the configs, server and clients to run the event handler service
type Service struct {
	server          HTTPServer
	router          *mux.Router
	serviceList     *ExternalServiceList
	healthCheck     HealthChecker
	consumer        kafka.IConsumerGroup
	producer        kafka.IProducer
	shutdownTimeout time.Duration
}

// Run the service
func Run(ctx context.Context, serviceList *ExternalServiceList, buildTime, gitCommit, version string, svcErrors chan error) (*Service, error) {
	log.Info(ctx, "running service")
	// Read config
	cfg, err := config.Get()
	if err != nil {
		return nil, errors.Wrap(err, "unable to retrieve service configuration")
	}
	log.Info(ctx, "got service configuration", log.Data{"config": cfg})
	// Get HTTP Server with collectionID checkHeader middleware
	r := mux.NewRouter()
	s := serviceList.GetHTTPServer(cfg.BindAddr, r)

	// Get the zebedee client
	zebedeeClient := serviceList.GetZebedee(cfg)

	// Get the DatasetApi Client
	datasetapiClient := serviceList.GetDatasetClient(cfg)

	// Get Kafka consumer
	consumer, err := serviceList.GetKafkaConsumer(ctx, cfg)
	if err != nil {
		log.Fatal(ctx, "failed to initialise kafka consumer", err)
		return nil, err
	}

	// Get Kafka producer
	producer, err := serviceList.GetKafkaProducer(ctx, cfg)
	if err != nil {
		log.Fatal(ctx, "failed to initialise kafka producer", err)
		return nil, err
	}

	searchDataImportProducer := event.SearchDataImportProducer{
		Marshaller: schema.SearchDataImportEvent,
		Producer:   producer,
	}

	// Kafka error logging go-routine
	consumer.Channels().LogErrors(ctx, "kafka consumer")

	// Kafka error logging go-routine
	producer.Channels().LogErrors(ctx, "kafka producer channels error")

	// Event Handler for Kafka Consumer
	eventHandler := &handler.ContentPublishedHandler{
		ZebedeeCli: zebedeeClient,
		DatasetCli: datasetapiClient,
		Producer:   searchDataImportProducer,
	}
	event.Consume(ctx, consumer, eventHandler, cfg)

	// Get HealthCheck
	hc, err := serviceList.GetHealthCheck(cfg, buildTime, gitCommit, version)
	if err != nil {
		log.Fatal(ctx, "could not instantiate healthcheck", err)
		return nil, err
	}

	if err := registerCheckers(ctx, hc, zebedeeClient, consumer, producer, datasetapiClient); err != nil {
		return nil, errors.Wrap(err, "unable to register checkers")
	}

	r.StrictSlash(true).Path("/health").HandlerFunc(hc.Handler)
	hc.Start(ctx)
	// Run the http server in a new go-routine
	go func() {
		if err := s.ListenAndServe(); err != nil {
			svcErrors <- errors.Wrap(err, "failure in http listen and serve")
		}
	}()
	return &Service{
		server:          s,
		router:          r,
		serviceList:     serviceList,
		healthCheck:     hc,
		consumer:        consumer,
		producer:        producer,
		shutdownTimeout: cfg.GracefulShutdownTimeout,
	}, nil
}

// Close gracefully shuts the service down in the required order, with timeout
func (svc *Service) Close(ctx context.Context) error {
	timeout := svc.shutdownTimeout
	log.Info(ctx, "commencing graceful shutdown", log.Data{"graceful_shutdown_timeout": timeout})
	ctx, cancel := context.WithTimeout(ctx, timeout)
	// track shutdown gracefully closes up
	var gracefulShutdown bool
	go func() {
		defer cancel()
		var hasShutdownError bool
		// stop healthcheck, as it depends on everything else
		if svc.serviceList.HealthCheck {
			svc.healthCheck.Stop()
		}

		// If kafka consumer exists, stop listening to it.
		// This will automatically stop the event consumer loops and no more messages will be processed.
		// The kafka consumer will be closed after the service shuts down.
		if svc.serviceList.KafkaConsumer {
			log.Info(ctx, "stopping kafka consumer listener")
			if err := svc.consumer.StopListeningToConsumer(ctx); err != nil {
				log.Error(ctx, "error stopping kafka consumer listener", err)
				hasShutdownError = true
			}
			log.Info(ctx, "stopped kafka consumer listener")
		}
		// stop any incoming requests before closing any outbound connections
		if err := svc.server.Shutdown(ctx); err != nil {
			log.Error(ctx, "failed to shutdown http server", err)
			hasShutdownError = true
		}

		// If kafka producer exists, close it.
		if svc.serviceList.KafkaProducer {
			log.Info(ctx, "closing kafka producer")
			if err := svc.producer.Close(ctx); err != nil {
				log.Error(ctx, "error closing kafka producer", err)
				hasShutdownError = true
			}
			log.Info(ctx, "closed kafka producer")
		}

		// If kafka consumer exists, close it.
		if svc.serviceList.KafkaConsumer {
			log.Info(ctx, "closing kafka consumer")
			if err := svc.consumer.Close(ctx); err != nil {
				log.Error(ctx, "error closing kafka consumer", err)
				hasShutdownError = true
			}
			log.Info(ctx, "closed kafka consumer")
		}

		if !hasShutdownError {
			gracefulShutdown = true
		}
	}()
	// wait for shutdown success (via cancel) or failure (timeout)
	<-ctx.Done()
	if !gracefulShutdown {
		err := errors.New("failed to shutdown gracefully")
		log.Error(ctx, "failed to shutdown gracefully ", err)
		return err
	}
	log.Info(ctx, "graceful shutdown was successful")
	return nil
}

func registerCheckers(ctx context.Context, hc HealthChecker, zebedeeClient clients.ZebedeeClient, consumer kafka.IConsumerGroup, producer kafka.IProducer, datasetClient clients.DatasetClient) (err error) {
	hasErrors := false

	if err := hc.AddCheck("Zebedee client", zebedeeClient.Checker); err != nil {
		hasErrors = true
		log.Error(ctx, "error adding check for ZebedeeClient", err)
	}

	if err := hc.AddCheck("Kafka consumer", consumer.Checker); err != nil {
		hasErrors = true
		log.Error(ctx, "error adding check for Kafka", err)
	}

	if err := hc.AddCheck("Kafka producer", producer.Checker); err != nil {
		hasErrors = true
		log.Error(ctx, "error adding check for Kafka producer", err)
	}

	if err := hc.AddCheck("DatasetAPI client", datasetClient.Checker); err != nil {
		hasErrors = true
		log.Error(ctx, "error adding check for DatasetClient", err)
	}

	if hasErrors {
		return errors.New("Error(s) registering checkers for healthcheck")
	}
	return nil
}
