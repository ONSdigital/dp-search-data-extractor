package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	kafka "github.com/ONSdigital/dp-kafka/v4"
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
	ImportProducer           kafka.IProducer
	DeleteProducer           kafka.IProducer
	ZebedeeCli               clients.ZebedeeClient
	DatasetCli               clients.DatasetClient
	TopicCli                 topicCli.Clienter
}

func New() *Service {
	return &Service{}
}

// Init initializes the Service by setting up essential components like Kafka producers, consumers, health checks, and caching.
// It validates the provided configuration, initializes clients and consumers, registers health checks, and sets up an HTTP server for health endpoints.
func (svc *Service) Init(ctx context.Context, cfg *config.Config, buildTime, gitCommit, version string) error {
	var err error

	if cfg == nil {
		return errors.New("nil config passed to service init")
	}

	svc.Cfg = cfg

	svc.initClients(ctx)

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

	if svc.ImportProducer, err = GetKafkaProducer(ctx, cfg.Kafka, cfg.Kafka.SearchDataImportedTopic); err != nil {
		return fmt.Errorf("failed to create import kafka producer: %w", err)
	}

	if svc.DeleteProducer, err = GetKafkaProducer(ctx, cfg.Kafka, cfg.Kafka.SearchContentDeletedTopic); err != nil {
		return fmt.Errorf("failed to create delete kafka producer: %w", err)
	}

	if err := svc.initConsumers(ctx); err != nil {
		return err
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

// initClients initializes external service clients based on the service configuration.
// It sets up clients for Zebedee, DatasetAPI, and Topic services if their corresponding flags are enabled.
func (svc *Service) initClients(ctx context.Context) {
	if svc.Cfg != nil && svc.Cfg.EnableZebedeeCallbacks {
		svc.ZebedeeCli = GetZebedee(svc.Cfg)
	} else {
		log.Info(ctx, "zebedee client not initialised as callbacks disabled")
	}

	if svc.Cfg != nil && svc.Cfg.EnableDatasetAPICallbacks {
		svc.DatasetCli = GetDatasetClient(svc.Cfg)
	} else {
		log.Info(ctx, "Dataset client not initialised as callbacks disabled")
	}

	if svc.Cfg != nil && svc.Cfg.EnableTopicTagging {
		svc.TopicCli = GetTopicClient(svc.Cfg)
	} else {
		log.Info(ctx, "Topic client not initialised as tagging disabled")
	}
}

func (svc *Service) initConsumers(ctx context.Context) error {
	var err error

	if svc.Cfg != nil && (svc.Cfg.EnableZebedeeCallbacks || svc.Cfg.EnableDatasetAPICallbacks) {
		if svc.ContentPublishedConsumer, err = GetKafkaConsumer(ctx, svc.Cfg.Kafka, svc.Cfg.Kafka.ContentUpdatedTopic); err != nil {
			return fmt.Errorf("failed to create content-published consumer: %w", err)
		}
		contentHandler := &handler.ContentPublished{
			Cfg:            svc.Cfg,
			ZebedeeCli:     svc.ZebedeeCli,
			DatasetCli:     svc.DatasetCli,
			ImportProducer: svc.ImportProducer,
			DeleteProducer: svc.DeleteProducer,
			Cache:          svc.Cache,
		}
		if err = svc.ContentPublishedConsumer.RegisterHandler(ctx, contentHandler.Handle); err != nil {
			return fmt.Errorf("could not register content-published handler: %w", err)
		}
		log.Info(ctx, "content-published consumer and handler registered")
	}

	if svc.Cfg != nil && svc.Cfg.EnableSearchContentUpdatedHandler {
		if svc.SearchContentConsumer, err = GetKafkaConsumer(ctx, svc.Cfg.Kafka, svc.Cfg.Kafka.SearchContentTopic); err != nil {
			return fmt.Errorf("failed to create search-content-updated consumer: %w", err)
		}
		searchContentHandler := &handler.SearchContentHandler{
			Cfg:            svc.Cfg,
			ImportProducer: svc.ImportProducer,
			DeleteProducer: svc.DeleteProducer,
		}
		if err = svc.SearchContentConsumer.RegisterHandler(ctx, searchContentHandler.Handle); err != nil {
			return fmt.Errorf("could not register search-content-updated handler: %w", err)
		}
		log.Info(ctx, "search-content-updated consumer and handler registered")
	}
	return nil
}

// Start the service
func (svc *Service) Start(ctx context.Context, svcErrors chan error) error {
	log.Info(ctx, "starting service")

	// Kafka error logging go-routine based on feature-flags
	if svc.ImportProducer != nil {
		svc.ImportProducer.LogErrors(ctx)
	}
	if svc.DeleteProducer != nil {
		svc.DeleteProducer.LogErrors(ctx)
	}

	if svc.Cfg != nil && (svc.Cfg.EnableZebedeeCallbacks || svc.Cfg.EnableDatasetAPICallbacks) {
		svc.ContentPublishedConsumer.LogErrors(ctx)
	}
	if svc.Cfg != nil && svc.Cfg.EnableSearchContentUpdatedHandler {
		svc.SearchContentConsumer.LogErrors(ctx)
	}

	// Start cache updates
	if svc.Cfg.EnableTopicTagging {
		// Start periodic updates using StartAndManageUpdates
		go func() {
			errorChannel := make(chan error)

			// Start and manage updates (handles both synchronous and periodic updates)
			svc.Cache.Topic.StartAndManageUpdates(ctx, errorChannel)

			// Handle any errors from periodic updates
			for err := range errorChannel {
				log.Error(ctx, "periodic cache update failed", err)
			}
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

// Close gracefully shuts the service down in the required order, with timeout
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

		if err := svc.closeKafkaProducers(ctx); err != nil {
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

// closeKafkaProducers closes the Kafka producers
func (svc *Service) closeKafkaProducers(ctx context.Context) error {
	var errs []error
	closeOne := func(p kafka.IProducer, name string) {
		if p == nil {
			return
		}
		log.Info(ctx, "closing kafka producer...", log.Data{"producer": name})
		if err := p.Close(ctx); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", name, err))
		}
		log.Info(ctx, "closed kafka producer", log.Data{"producer": name})
	}
	closeOne(svc.ImportProducer, "import")
	closeOne(svc.DeleteProducer, "delete")
	if len(errs) > 0 {
		return fmt.Errorf("failed to close producers: %v", errs)
	}
	return nil
}

// registerCheckers registers health checks for various service components such as Kafka consumers and clients.
// It divides the registration process into three distinct steps:
// 1. Client Checkers: Registers health checks for Zebedee and Dataset clients, as well as the Kafka producer.
// 2. Consumer Checkers: Registers health checks for Kafka consumers (ContentPublished and SearchContent).
// 3. Subscription: Subscribes consumers to health checks based on feature flags and component availability.
// If any of the health checks fail during registration, the method returns an error.
func (svc *Service) registerCheckers(ctx context.Context) error {
	var hasErrors bool

	// Step 1: Register client and producer checkers
	chkZebedee, chkDataset, chkImportProd, chkDeleteProd := svc.registerClientAndProducerCheckers(ctx, &hasErrors)

	// Step 2: Register consumer checkers
	svc.registerConsumerCheckers(ctx, &hasErrors)

	if hasErrors {
		return errors.New("Error(s) registering checkers for healthcheck")
	}

	// Step 3: Subscribe consumers to health checks
	svc.subscribeToHealthCheck(chkZebedee, chkDataset, chkImportProd, chkDeleteProd)
	return nil
}

// registerClientAndProducerCheckers registers health checks for Zebedee, Dataset API clients, and the Kafka producers.
// It returns the respective health check references and updates the hasErrors flag if any registration fails.
func (svc *Service) registerClientAndProducerCheckers(ctx context.Context, hasErrors *bool) (zebedeeCheck, datasetCheck, importProducerCheck, deleteProducerCheck *healthcheck.Check) {
	if svc.ZebedeeCli != nil {
		zebedeeCheck, *hasErrors = svc.addHealthCheck(ctx, svc.Cfg.EnableZebedeeCallbacks, "Zebedee client", svc.ZebedeeCli.Checker, *hasErrors)
	}
	if svc.DatasetCli != nil {
		datasetCheck, *hasErrors = svc.addHealthCheck(ctx, svc.Cfg.EnableDatasetAPICallbacks, "DatasetAPI client", svc.DatasetCli.Checker, *hasErrors)
	}

	if svc.ImportProducer != nil {
		importProducerCheck, *hasErrors = svc.addHealthCheck(ctx, true, "ContentImport Kafka producer", svc.ImportProducer.Checker, *hasErrors)
	}
	if svc.DeleteProducer != nil {
		deleteProducerCheck, *hasErrors = svc.addHealthCheck(ctx, true, "ContentDelete Kafka producer", svc.DeleteProducer.Checker, *hasErrors)
	}
	return zebedeeCheck, datasetCheck, importProducerCheck, deleteProducerCheck
}

// registerConsumerCheckers registers health checks for Kafka consumers, such as ContentPublished and SearchContent.
// It updates the hasErrors flag if any health check registration fails.
func (svc *Service) registerConsumerCheckers(ctx context.Context, hasErrors *bool) {
	if svc.ContentPublishedConsumer != nil {
		if _, err := svc.HealthCheck.AddAndGetCheck("ContentPublished Kafka consumer", svc.ContentPublishedConsumer.Checker); err != nil {
			*hasErrors = true
			log.Error(ctx, "error adding check for ContentPublished Kafka consumer", err)
		}
	}

	if svc.SearchContentConsumer != nil {
		if _, err := svc.HealthCheck.AddAndGetCheck("SearchContent Kafka consumer", svc.SearchContentConsumer.Checker); err != nil {
			*hasErrors = true
			log.Error(ctx, "error adding check for SearchContent Kafka consumer", err)
		}
	}
}

// subscribeToHealthCheck handles the subscription process of Kafka consumers to relevant health checks.
// It evaluates the configuration settings and component availability to ensure that only healthy
// components participate in message consumption. If a component becomes unhealthy and stop-consume
// behavior is enabled, the corresponding consumer is halted.
func (svc *Service) subscribeToHealthCheck(chkZebedee, chkDataset, chkImportProd, chkDeleteProd *healthcheck.Check) {
	if svc.Cfg.StopConsumingOnUnhealthy {
		subscribers := []healthcheck.Subscriber{}
		checks := []*healthcheck.Check{}

		if svc.Cfg.EnableZebedeeCallbacks && chkZebedee != nil {
			checks = append(checks, chkZebedee)
		}
		if svc.Cfg.EnableDatasetAPICallbacks && chkDataset != nil {
			checks = append(checks, chkDataset)
		}
		if svc.ImportProducer != nil {
			checks = append(checks, chkImportProd)
		}
		if svc.DeleteProducer != nil {
			checks = append(checks, chkDeleteProd)
		}
		if svc.ContentPublishedConsumer != nil {
			subscribers = append(subscribers, svc.ContentPublishedConsumer)
		}
		if svc.SearchContentConsumer != nil {
			subscribers = append(subscribers, svc.SearchContentConsumer)
		}

		for _, sub := range subscribers {
			svc.HealthCheck.Subscribe(sub, checks...)
		}
	}
}

func (svc *Service) addHealthCheck(ctx context.Context, enabled bool, name string, checker healthcheck.Checker, hasErrors bool) (*healthcheck.Check, bool) {
	if enabled {
		check, err := svc.HealthCheck.AddAndGetCheck(name, checker)
		if err != nil {
			log.Error(ctx, "error adding check for "+name, err)
			hasErrors = true
			return nil, hasErrors
		}
		return check, hasErrors
	}
	log.Info(ctx, "skipping health check registration as configuration is disabled", log.Data{
		"name":    name,
		"enabled": enabled,
	})
	return nil, hasErrors
}
