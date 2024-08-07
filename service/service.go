package service

import (
	"context"
	"fmt"

	kafka "github.com/ONSdigital/dp-kafka/v3"
	"github.com/ONSdigital/dp-search-data-extractor/cache"
	cachePrivate "github.com/ONSdigital/dp-search-data-extractor/cache/private"
	"github.com/ONSdigital/dp-search-data-extractor/clients"
	"github.com/ONSdigital/dp-search-data-extractor/config"
	"github.com/ONSdigital/dp-search-data-extractor/handler"
	topicCli "github.com/ONSdigital/dp-topic-api/sdk"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// Service contains all the configs, server and clients to run the event handler service
type Service struct {
	Cfg         *config.Config
	Cache       cache.List
	Server      HTTPServer
	HealthCheck HealthChecker
	Consumer    kafka.IConsumerGroup
	Producer    kafka.IProducer
	ZebedeeCli  clients.ZebedeeClient
	DatasetCli  clients.DatasetClient
	TopicCli    topicCli.Clienter
}

func New() *Service {
	return &Service{}
}

func (svc *Service) Init(ctx context.Context, cfg *config.Config, buildTime, gitCommit, version string) error {
	var err error

	if cfg == nil {
		return errors.New("nil config passed to service init")
	}

	svc.Cfg = cfg

	if svc.Consumer, err = GetKafkaConsumer(ctx, cfg.Kafka); err != nil {
		return fmt.Errorf("failed to create kafka consumer: %w", err)
	}

	if svc.Producer, err = GetKafkaProducer(ctx, cfg.Kafka); err != nil {
		return fmt.Errorf("failed to create kafka producer: %w", err)
	}

	svc.ZebedeeCli = GetZebedee(cfg)
	svc.DatasetCli = GetDatasetClient(cfg)
	svc.TopicCli = GetTopicClient(cfg)

	if svc.Cfg.EnableTopicTagging {
		// Initialise caching
		svc.Cache.Topic, err = cache.NewTopicCache(ctx, &svc.Cfg.TopicCacheUpdateInterval)
		if err != nil {
			log.Error(ctx, "failed to create topic cache", err)
			return err
		}

		// Load cache with topics on startup
		svc.Cache.Topic.AddUpdateFunc(svc.Cache.Topic.GetTopicCacheKey(), cachePrivate.UpdateTopicCache(ctx, svc.Cfg.ServiceAuthToken, svc.TopicCli))
	}

	h := handler.ContentPublished{
		Cfg:        svc.Cfg,
		ZebedeeCli: svc.ZebedeeCli,
		DatasetCli: svc.DatasetCli,
		Producer:   svc.Producer,
		Cache:      svc.Cache,
	}

	err = svc.Consumer.RegisterHandler(ctx, h.Handle)
	if err != nil {
		return fmt.Errorf("could not register kafka handler: %w", err)
	}

	// Get HealthCheck
	if svc.HealthCheck, err = GetHealthCheck(cfg, buildTime, gitCommit, version); err != nil {
		return fmt.Errorf("could not instantiate healthcheck: %w", err)
	}

	if err := svc.registerCheckers(ctx); err != nil {
		return fmt.Errorf("unable to register checkers: %w", err)
	}

	r := mux.NewRouter()
	r.StrictSlash(true).Path("/health").HandlerFunc(svc.HealthCheck.Handler)
	svc.Server = GetHTTPServer(cfg.BindAddr, r)

	return nil
}

// Start the service
func (svc *Service) Start(ctx context.Context, svcErrors chan error) error {
	log.Info(ctx, "starting service")

	// Kafka error logging go-routine
	svc.Consumer.LogErrors(ctx)
	svc.Producer.LogErrors(ctx)

	// If start/stop on health updates is disabled, start consuming as soon as possible
	if !svc.Cfg.StopConsumingOnUnhealthy {
		if err := svc.Consumer.Start(); err != nil {
			return fmt.Errorf("consumer failed to start: %w", err)
		}
	}

	// Start cache updates
	if svc.Cfg.EnableTopicTagging {
		go svc.Cache.Topic.StartUpdates(ctx, svcErrors)
	}

	// Always start healthcheck.
	// If start/stop on health updates is enabled,
	// the consumer will start consuming on the first healthy update
	svc.HealthCheck.Start(ctx)

	// Run the http server in a new go-routine
	go func() {
		if err := svc.Server.ListenAndServe(); err != nil {
			svcErrors <- fmt.Errorf("failure in http listen and serve: %w", err)
		}
	}()

	return nil
}

// Close gracefully shuts the service down in the required order, with timeout
func (svc *Service) Close(ctx context.Context) error {
	timeout := svc.Cfg.GracefulShutdownTimeout
	log.Info(ctx, "commencing graceful shutdown", log.Data{"graceful_shutdown_timeout": timeout})
	ctx, cancel := context.WithTimeout(ctx, timeout)
	hasShutdownError := false

	go func() {
		defer cancel()

		// stop healthcheck, as it depends on everything else
		if svc.HealthCheck != nil {
			log.Info(ctx, "stopping health checker...")
			svc.HealthCheck.Stop()
			log.Info(ctx, "stopped health checker")
		}

		// stop cache updates
		if svc.Cfg.EnableTopicTagging {
			svc.Cache.Topic.Close()
		}

		// If kafka consumer exists, stop listening to it.
		// This will automatically stop the event consumer loops and no more messages will be processed.
		// The kafka consumer will be closed after the service shuts down.
		//nolint
		if svc.Consumer != nil {
			log.Info(ctx, "stopping kafka consumer listener...")
			if err := svc.Consumer.StopAndWait(); err != nil {
				log.Error(ctx, "error stopping kafka consumer listener", err)
				hasShutdownError = true
			} else {
				log.Info(ctx, "stopped kafka consumer listener")
			}
		}

		// Shutdown the HTTP server
		if svc.Server != nil {
			log.Info(ctx, "shutting http server down...")
			if err := svc.Server.Shutdown(ctx); err != nil {
				log.Error(ctx, "failed to shutdown http server", err)
				hasShutdownError = true
			} else {
				log.Info(ctx, "shut down http server")
			}
		}

		// If kafka producer exists, close it.
		//nolint
		if svc.Producer != nil {
			log.Info(ctx, "closing kafka producer")
			if err := svc.Producer.Close(ctx); err != nil {
				log.Error(ctx, "failed to close kafka producer", err)
				hasShutdownError = true
			} else {
				log.Info(ctx, "closed kafka producer")
			}
		}

		// If kafka consumer exists, close it.
		//nolint
		if svc.Consumer != nil {
			log.Info(ctx, "closing kafka consumer")
			if err := svc.Consumer.Close(ctx); err != nil {
				log.Error(ctx, "failed to close kafka consumer", err)
				hasShutdownError = true
			} else {
				log.Info(ctx, "closed kafka consumer")
			}
		}
	}()

	// wait for shutdown success (via cancel) or failure (timeout)
	<-ctx.Done()

	// timeout expired
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("shutdown timed out: %w", ctx.Err())
	}

	// other error
	if hasShutdownError {
		return errors.New("failed to shutdown gracefully")
	}

	log.Info(ctx, "graceful shutdown was successful")
	return nil
}

func (svc *Service) registerCheckers(ctx context.Context) (err error) {
	hasErrors := false

	chkZebedee, err := svc.HealthCheck.AddAndGetCheck("Zebedee client", svc.ZebedeeCli.Checker)
	if err != nil {
		hasErrors = true
		log.Error(ctx, "error adding check for ZebedeeClient", err)
	}

	_, err = svc.HealthCheck.AddAndGetCheck("Kafka consumer", svc.Consumer.Checker)
	if err != nil {
		hasErrors = true
		log.Error(ctx, "error adding check for Kafka", err)
	}

	chkProducer, err := svc.HealthCheck.AddAndGetCheck("Kafka producer", svc.Producer.Checker)
	if err != nil {
		hasErrors = true
		log.Error(ctx, "error adding check for Kafka producer", err)
	}

	chkDataset, err := svc.HealthCheck.AddAndGetCheck("DatasetAPI client", svc.DatasetCli.Checker)
	if err != nil {
		hasErrors = true
		log.Error(ctx, "error adding check for DatasetClient", err)
	}

	if hasErrors {
		return errors.New("Error(s) registering checkers for healthcheck")
	}

	if svc.Cfg.StopConsumingOnUnhealthy {
		svc.HealthCheck.Subscribe(svc.Consumer, chkZebedee, chkDataset, chkProducer)
	}

	return nil
}
