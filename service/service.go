package service

import (
	"context"
	"fmt"
	"strings"
	"sync"

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

	// If start/stop on health updates is disabled, start consuming as soon as possible
	if !svc.Cfg.StopConsumingOnUnhealthy {
		if err := svc.ContentPublishedConsumer.Start(); err != nil {
			return fmt.Errorf("content-publish consumer failed to start: %w", err)
		}
		if err := svc.SearchContentConsumer.Start(); err != nil {
			return fmt.Errorf("search-content consumer failed to start: %w", err)
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

func (svc *Service) Close(ctx context.Context) error {
	log.Info(ctx, "closing service gracefully")

	shutdownCtx, cancel := context.WithTimeout(ctx, svc.Cfg.GracefulShutdownTimeout)
	defer cancel()

	var shutdownErrors []error
	var wg sync.WaitGroup

	// Define helper for shutting down a component
	shutdownComponent := func(name string, shutdownFunc func() error) {
		defer wg.Done()
		if err := shutdownFunc(); err != nil {
			log.Error(ctx, fmt.Sprintf("failed to close %s", name), err)
			shutdownErrors = append(shutdownErrors, fmt.Errorf("%s: %w", name, err))
		} else {
			log.Info(ctx, fmt.Sprintf("closed %s", name))
		}
	}

	// Shutdown consumers
	wg.Add(2)
	go shutdownComponent("content-published consumer", svc.ContentPublishedConsumer.StopAndWait)
	go shutdownComponent("search-content consumer", svc.SearchContentConsumer.StopAndWait)

	// Shutdown producer (wrap to provide context)
	wg.Add(1)
	go shutdownComponent("producer", func() error { return svc.Producer.Close(shutdownCtx) })

	// Stop healthcheck
	wg.Add(1)
	go func() {
		defer wg.Done()
		svc.HealthCheck.Stop()
		log.Info(ctx, "stopped healthcheck")
	}()

	// Shutdown HTTP server (wrap to provide context)
	wg.Add(1)
	go shutdownComponent("HTTP server", func() error { return svc.Server.Shutdown(shutdownCtx) })

	// Wait for all components to shutdown
	wg.Wait()

	// Handle timeout
	if ctx.Err() == context.DeadlineExceeded {
		log.Error(ctx, "shutdown timed out", ctx.Err())
		return fmt.Errorf("shutdown timed out: %w", ctx.Err())
	}

	// Aggregate errors
	if len(shutdownErrors) > 0 {
		for _, err := range shutdownErrors {
			log.Error(ctx, "shutdown error", err)
		}
		return fmt.Errorf("failed to shutdown gracefully: %v", joinErrors(shutdownErrors))
	}

	log.Info(ctx, "graceful shutdown was successful")
	return nil
}

// Helper function to join multiple errors into a single error message
func joinErrors(errs []error) string {
	errorMessages := make([]string, len(errs))
	for _, err := range errs {
		errorMessages = append(errorMessages, err.Error())
	}
	return strings.Join(errorMessages, "; ")
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
