package service

import (
	"context"
	"fmt"
	"strings"
	"time"

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
	Cfg                      *config.Config
	Cache                    cache.List
	Server                   HTTPServer
	HealthCheck              HealthChecker
	SearchContentConsumer    kafka.IConsumerGroup
	ContentPublishedConsumer kafka.IConsumerGroup
	Producer                 kafka.IProducer
	ZebedeeCli               clients.ZebedeeClient
	DatasetCli               clients.DatasetClient
	TopicCli                 topicCli.Clienter
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

	if svc.Producer, err = GetKafkaProducer(ctx, cfg.Kafka); err != nil {
		return fmt.Errorf("failed to create kafka producer: %w", err)
	}

	// Initialize Consumer for "content-published"
	if svc.ContentPublishedConsumer, err = GetKafkaConsumer(ctx, cfg.Kafka, cfg.Kafka.ContentUpdatedTopic); err != nil {
		return fmt.Errorf("failed to create content-published consumer: %w", err)
	}

	contentHandler := &handler.ContentPublished{
		Cfg:        svc.Cfg,
		ZebedeeCli: GetZebedee(cfg),
		DatasetCli: GetDatasetClient(cfg),
		Producer:   svc.Producer,
		Cache:      svc.Cache,
	}
	if err = svc.ContentPublishedConsumer.RegisterHandler(ctx, contentHandler.Handle); err != nil {
		return fmt.Errorf("could not register content-published handler: %w", err)
	}

	// Initialize Consumer for "search-content-updated"
	if svc.SearchContentConsumer, err = GetKafkaConsumer(ctx, cfg.Kafka, cfg.Kafka.SearchContentTopic); err != nil {
		return fmt.Errorf("failed to create search-content-updated consumer: %w", err)
	}

	searchContentHandler := &handler.SearchContentHandler{
		Cfg:      svc.Cfg,
		Producer: svc.Producer,
	}
	if err = svc.SearchContentConsumer.RegisterHandler(ctx, searchContentHandler.Handle); err != nil {
		return fmt.Errorf("could not register search-content-updated handler: %w", err)
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
	svc.ContentPublishedConsumer.LogErrors(ctx)
	svc.SearchContentConsumer.LogErrors(ctx)
	svc.Producer.LogErrors(ctx)

	// Start cache updates
	if svc.Cfg.EnableTopicTagging {
		// FIXME This code starts a goroutine with no control to stop it gracefully later. Also it does the initial
		// cache population concurrently so it causes a race condition for the handler which later relies on it.
		// To fix it will require a change to the underlying library but in the meantime we can work around the race
		// condition by synchronously updating the cache content first.
		err := svc.Cache.Topic.UpdateContent(ctx)
		if err != nil {
			return fmt.Errorf("topic cache failed to initialise: %w", err)
		}

		go func() {
			time.Sleep(svc.Cfg.TopicCacheUpdateInterval) // to prevent immediate re-update of cache content
			svc.Cache.Topic.StartUpdates(ctx, svcErrors)
		}()
	}

	// If start/stop on health updates is disabled, start consuming as soon as possible
	if !svc.Cfg.StopConsumingOnUnhealthy {
		if err := svc.ContentPublishedConsumer.Start(); err != nil {
			return fmt.Errorf("content-publish consumer failed to start: %w", err)
		}
		if err := svc.SearchContentConsumer.Start(); err != nil {
			return fmt.Errorf("search-content consumer failed to start: %w", err)
		}
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

func (svc *Service) Close(ctx context.Context) error {
	timeout := svc.Cfg.GracefulShutdownTimeout
	log.Info(ctx, "commencing graceful shutdown", log.Data{"graceful_shutdown_timeout": timeout})
	ctx, cancel := context.WithTimeout(ctx, timeout)
	hasShutdownError := false

	go func() {
		defer cancel()

		// Stop health check, as it depends on everything else
		if svc.HealthCheck != nil {
			log.Info(ctx, "stopping health checker...")
			svc.HealthCheck.Stop()
			log.Info(ctx, "stopped health checker")
		}

		// Stop cache updates
		if svc.Cfg.EnableTopicTagging {
			svc.Cache.Topic.Close()
		}

		// Shutdown consumers and producer
		if err := svc.shutdownConsumers(ctx); err != nil {
			log.Error(ctx, "consumer shutdown error", err)
			hasShutdownError = true
		}

		if err := svc.closeProducer(ctx); err != nil {
			log.Error(ctx, "producer shutdown error", err)
			hasShutdownError = true
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
	}()

	// Wait for shutdown success (via cancel) or failure (timeout)
	<-ctx.Done()

	// Timeout expired
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("shutdown timed out: %w", ctx.Err())
	}

	// Other error
	if hasShutdownError {
		return errors.New("failed to shutdown gracefully")
	}

	log.Info(ctx, "graceful shutdown was successful")
	return nil
}

// shutdownConsumers stops and closes Kafka consumers
func (svc *Service) shutdownConsumers(ctx context.Context) error {
	var errMessages []string

	stopAndCloseConsumer := func(consumer kafka.IConsumerGroup, name string) {
		if consumer == nil {
			return
		}

		log.Info(ctx, "stopping kafka consumer listener...", log.Data{"Consumer": name})
		if err := consumer.StopAndWait(); err != nil {
			log.Error(ctx, "error stopping kafka consumer listener", err, log.Data{"Consumer": name})
			errMessages = append(errMessages, fmt.Sprintf("%s consumer stop error: %v", name, err))
		} else {
			log.Info(ctx, "stopped kafka consumer listener", log.Data{"Consumer": name})
		}

		log.Info(ctx, "closing kafka consumer...", log.Data{"Consumer": name})
		if err := consumer.Close(ctx); err != nil {
			log.Error(ctx, "error closing kafka consumer", err, log.Data{"Consumer": name})
			errMessages = append(errMessages, fmt.Sprintf("%s consumer close error: %v", name, err))
		} else {
			log.Info(ctx, "closed kafka consumer", log.Data{"Consumer": name})
		}
	}

	// Handle both consumers
	stopAndCloseConsumer(svc.ContentPublishedConsumer, "content-published")
	stopAndCloseConsumer(svc.SearchContentConsumer, "search-content")

	// Aggregate errors, if any
	if len(errMessages) > 0 {
		return fmt.Errorf("consumer shutdown error: %s", strings.Join(errMessages, "; "))
	}
	return nil
}

// closeProducer closes the Kafka producer
func (svc *Service) closeProducer(ctx context.Context) error {
	if svc.Producer == nil {
		return nil
	}

	log.Info(ctx, "closing kafka producer...")
	if err := svc.Producer.Close(ctx); err != nil {
		return fmt.Errorf("failed to close kafka producer: %w", err)
	}
	log.Info(ctx, "closed kafka producer")
	return nil
}

func (svc *Service) registerCheckers(ctx context.Context) (err error) {
	hasErrors := false

	chkZebedee, err := svc.HealthCheck.AddAndGetCheck("Zebedee client", svc.ZebedeeCli.Checker)
	if err != nil {
		hasErrors = true
		log.Error(ctx, "error adding check for ZebedeeClient", err)
	}

	_, err = svc.HealthCheck.AddAndGetCheck("ContentPublished Kafka consumer", svc.ContentPublishedConsumer.Checker)
	if err != nil {
		hasErrors = true
		log.Error(ctx, "error adding check for Kafka", err)
	}

	_, err = svc.HealthCheck.AddAndGetCheck("SearchContent Kafka consumer", svc.SearchContentConsumer.Checker)
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
		svc.HealthCheck.Subscribe(svc.ContentPublishedConsumer, chkZebedee, chkDataset, chkProducer)
		svc.HealthCheck.Subscribe(svc.SearchContentConsumer, chkZebedee, chkDataset, chkProducer)
	}

	return nil
}
