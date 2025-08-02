package service_test

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	kafka "github.com/ONSdigital/dp-kafka/v3"
	"github.com/ONSdigital/dp-kafka/v3/kafkatest"
	"github.com/ONSdigital/dp-search-data-extractor/clients"
	clientMock "github.com/ONSdigital/dp-search-data-extractor/clients/mock"
	"github.com/ONSdigital/dp-search-data-extractor/config"
	service "github.com/ONSdigital/dp-search-data-extractor/service"
	serviceMock "github.com/ONSdigital/dp-search-data-extractor/service/mock"

	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	ctx           = context.Background()
	testBuildTime = "BuildTime"
	testGitCommit = "GitCommit"
	testVersion   = "Version"
	testChecks    = map[string]*healthcheck.Check{
		"Zebedee client":                  {},
		"DatasetAPI client":               {},
		"Kafka producer":                  {},
		"ContentPublished Kafka consumer": {},
		"SearchContent Kafka consumer":    {},
	}
	errKafkaConsumer = errors.New("Kafka consumer error")
	errKafkaProducer = errors.New("Kafka producer error")
	errHealthcheck   = errors.New("healthCheck error")
	errServer        = errors.New("HTTP Server error")
	errAddCheck      = fmt.Errorf("healthcheck add check error")
)

func TestNew(t *testing.T) {
	Convey("service.New returns a new empty service struct", t, func() {
		srv := service.New()
		So(*srv, ShouldResemble, service.Service{})
	})
}

func TestInit(t *testing.T) {
	Convey("Having a set of mocked dependencies", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)

		cfg.EnableZebedeeCallbacks = true
		cfg.EnableDatasetAPICallbacks = true
		cfg.EnableSearchContentUpdatedHandler = true

		// Mocking the Kafka consumers
		consumerMock1 := &kafkatest.IConsumerGroupMock{
			RegisterHandlerFunc: func(ctx context.Context, h kafka.Handler) error {
				return nil
			},
		}
		consumerMock2 := &kafkatest.IConsumerGroupMock{
			RegisterHandlerFunc: func(ctx context.Context, h kafka.Handler) error {
				return nil
			},
		}

		// Mocking the Kafka consumer retrieval function to return both consumers
		service.GetKafkaConsumer = func(ctx context.Context, cfg *config.Kafka, topic string) (kafka.IConsumerGroup, error) {
			switch topic {
			case "content-updated":
				return consumerMock1, nil
			case "search-content-updated":
				return consumerMock2, nil
			default:
				return nil, fmt.Errorf("unknown topic: %s", topic)
			}
		}

		producerMock := &kafkatest.IProducerMock{}
		service.GetKafkaProducer = func(ctx context.Context, cfg *config.Kafka, topic string) (kafka.IProducer, error) {
			return producerMock, nil
		}

		subscribedTo := []*healthcheck.Check{}
		hcMock := &serviceMock.HealthCheckerMock{
			AddAndGetCheckFunc: func(name string, checker healthcheck.Checker) (*healthcheck.Check, error) {
				return testChecks[name], nil
			},
			SubscribeFunc: func(s healthcheck.Subscriber, checks ...*healthcheck.Check) {
				subscribedTo = append(subscribedTo, checks...)
			},
		}
		service.GetHealthCheck = func(cfg *config.Config, buildTime, gitCommit, version string) (service.HealthChecker, error) {
			return hcMock, nil
		}

		serverMock := &serviceMock.HTTPServerMock{}
		service.GetHTTPServer = func(bindAddr string, router http.Handler) service.HTTPServer {
			return serverMock
		}

		datasetAPIMock := &clientMock.DatasetClientMock{
			CheckerFunc: func(context.Context, *healthcheck.CheckState) error { return nil },
		}
		service.GetDatasetClient = func(cfg *config.Config) clients.DatasetClient {
			return datasetAPIMock
		}

		zebedeeMock := &clientMock.ZebedeeClientMock{
			CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error { return nil },
		}
		service.GetZebedee = func(cfg *config.Config) clients.ZebedeeClient {
			return zebedeeMock
		}

		svc := &service.Service{}

		Convey("Tying to initialise a service without a config returns the expected error", func() {
			err := svc.Init(ctx, nil, testBuildTime, testGitCommit, testVersion)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "nil config passed to service init")
		})

		Convey("Given that initialising Kafka consumer returns an error", func() {
			service.GetKafkaConsumer = func(ctx context.Context, cfg *config.Kafka, _ string) (kafka.IConsumerGroup, error) {
				return nil, errKafkaConsumer
			}

			Convey("Then service Init fails with the same error and no further initialisations are attempted", func() {
				err := svc.Init(ctx, cfg, testBuildTime, testGitCommit, testVersion)
				So(errors.Unwrap(err), ShouldResemble, errKafkaConsumer)
				So(svc.Cfg, ShouldResemble, cfg)

				Convey("And no checkers are registered ", func() {
					So(hcMock.AddAndGetCheckCalls(), ShouldHaveLength, 0)
				})
			})
		})

		Convey("Given that Kafka consumers fail to register a handler", func() {
			consumerMock1.RegisterHandlerFunc = func(ctx context.Context, h kafka.Handler) error {
				return errKafkaConsumer
			}

			Convey("Then service Init fails with the same error and no further initialisations are attempted", func() {
				err := svc.Init(ctx, cfg, testBuildTime, testGitCommit, testVersion)
				So(errors.Unwrap(err), ShouldResemble, errKafkaConsumer)
				So(svc.Cfg, ShouldResemble, cfg)

				Convey("And no checkers are registered ", func() {
					So(hcMock.AddAndGetCheckCalls(), ShouldHaveLength, 0)
				})
			})
		})

		Convey("Given that initialising Kafka producer returns an error", func() {
			service.GetKafkaProducer = func(ctx context.Context, cfg *config.Kafka, topic string) (kafka.IProducer, error) {
				return nil, errKafkaProducer
			}

			Convey("Then service Init fails with the same error and no further initialisations are attempted", func() {
				err := svc.Init(ctx, cfg, testBuildTime, testGitCommit, testVersion)
				So(errors.Unwrap(err), ShouldResemble, errKafkaProducer)
				So(svc.Cfg, ShouldResemble, cfg)

				Convey("And no checkers are registered ", func() {
					So(hcMock.AddAndGetCheckCalls(), ShouldHaveLength, 0)
				})
			})
		})

		Convey("Given that initialising healthcheck returns an error", func() {
			service.GetHealthCheck = func(cfg *config.Config, buildTime, gitCommit, version string) (service.HealthChecker, error) {
				return nil, errHealthcheck
			}

			Convey("Then service Init fails with the same error and no further initialisations are attempted", func() {
				err := svc.Init(ctx, cfg, testBuildTime, testGitCommit, testVersion)
				So(errors.Unwrap(err), ShouldResemble, errHealthcheck)
				So(svc.Cfg, ShouldResemble, cfg)
				So(svc.ContentPublishedConsumer, ShouldResemble, consumerMock1)
				So(svc.SearchContentConsumer, ShouldResemble, consumerMock2)

				Convey("And no checkers are registered ", func() {
					So(hcMock.AddAndGetCheckCalls(), ShouldHaveLength, 0)
				})
			})
		})

		Convey("Given that Checkers cannot be registered", func() {
			hcMock.AddAndGetCheckFunc = func(name string, checker healthcheck.Checker) (*healthcheck.Check, error) { return nil, errAddCheck }

			Convey("Then service Init fails with the expected error", func() {
				err := svc.Init(ctx, cfg, testBuildTime, testGitCommit, testVersion)
				So(err, ShouldNotBeNil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "unable to register checkers: Error(s) registering checkers for healthcheck")
				So(svc.Cfg, ShouldResemble, cfg)
				So(svc.ContentPublishedConsumer, ShouldResemble, consumerMock1)
				So(svc.SearchContentConsumer, ShouldResemble, consumerMock2)

				Convey("And all other checkers try to register", func() {
					So(hcMock.AddAndGetCheckCalls(), ShouldHaveLength, 6)
				})
			})
		})

		Convey("Given one of the feature flags for content-published is disabled and the flag for search-content-updated is enabled", func() {
			cfg.EnableZebedeeCallbacks = true
			cfg.EnableDatasetAPICallbacks = false
			cfg.EnableSearchContentUpdatedHandler = true

			err := svc.Init(ctx, cfg, testBuildTime, testGitCommit, testVersion)
			So(err, ShouldBeNil)
			So(svc.Cfg, ShouldResemble, cfg)
			So(svc.ImportProducer, ShouldResemble, producerMock)
			So(svc.DeleteProducer, ShouldResemble, producerMock)
			So(svc.ZebedeeCli, ShouldNotBeNil)
			So(svc.DatasetCli, ShouldBeNil)
			So(svc.ContentPublishedConsumer, ShouldNotBeNil)
			So(svc.SearchContentConsumer, ShouldNotBeNil)

			Convey("Then only necessary checks are registered based on feature flags", func() {
				subscribed := hcMock.AddAndGetCheckCalls()
				So(subscribed, ShouldHaveLength, 5)
			})
		})

		Convey("Given both of feature flags for content-published are disabled and the flag for search-content-updated is enabled", func() {
			cfg.EnableZebedeeCallbacks = false
			cfg.EnableDatasetAPICallbacks = false
			cfg.EnableSearchContentUpdatedHandler = true

			err := svc.Init(ctx, cfg, testBuildTime, testGitCommit, testVersion)
			So(err, ShouldBeNil)
			So(svc.Cfg, ShouldResemble, cfg)
			So(svc.ImportProducer, ShouldResemble, producerMock)
			So(svc.DeleteProducer, ShouldResemble, producerMock)
			So(svc.ZebedeeCli, ShouldBeNil)
			So(svc.DatasetCli, ShouldBeNil)
			So(svc.ContentPublishedConsumer, ShouldBeNil)
			So(svc.SearchContentConsumer, ShouldNotBeNil)

			Convey("Then only necessary checks are registered based on feature flags", func() {
				subscribed := hcMock.AddAndGetCheckCalls()
				So(subscribed, ShouldHaveLength, 3)
			})
		})

		Convey("Given only the feature flag for search-content-updated is disabled and both flags for content-published are enabled", func() {
			cfg.EnableZebedeeCallbacks = true
			cfg.EnableDatasetAPICallbacks = true
			cfg.EnableSearchContentUpdatedHandler = false

			err := svc.Init(ctx, cfg, testBuildTime, testGitCommit, testVersion)
			So(err, ShouldBeNil)
			So(svc.Cfg, ShouldResemble, cfg)
			So(svc.ImportProducer, ShouldResemble, producerMock)
			So(svc.DeleteProducer, ShouldResemble, producerMock)
			So(svc.ZebedeeCli, ShouldNotBeNil)
			So(svc.DatasetCli, ShouldNotBeNil)
			So(svc.ContentPublishedConsumer, ShouldNotBeNil)
			So(svc.SearchContentConsumer, ShouldBeNil)

			Convey("Then only necessary checks are registered based on feature flags", func() {
				subscribed := hcMock.AddAndGetCheckCalls()
				So(subscribed, ShouldHaveLength, 5)
			})
		})

		Convey("Given all consumer feature flags are disabled", func() {
			cfg.EnableZebedeeCallbacks = false
			cfg.EnableDatasetAPICallbacks = false
			cfg.EnableSearchContentUpdatedHandler = false

			service.GetHealthCheck = func(cfg *config.Config, buildTime, gitCommit, version string) (service.HealthChecker, error) {
				return nil, errHealthcheck
			}

			err := svc.Init(ctx, cfg, testBuildTime, testGitCommit, testVersion)
			So(err, ShouldNotBeNil)
			So(svc.Cfg, ShouldResemble, cfg)
			So(svc.ImportProducer, ShouldResemble, producerMock)
			So(svc.DeleteProducer, ShouldResemble, producerMock)
			So(svc.ZebedeeCli, ShouldBeNil)
			So(svc.DatasetCli, ShouldBeNil)
			So(svc.ContentPublishedConsumer, ShouldBeNil)
			So(svc.SearchContentConsumer, ShouldBeNil)

			Convey("Then no subscribers are added when all feature flags are disabled", func() {
				subscribed := hcMock.SubscribeCalls()
				So(subscribed, ShouldHaveLength, 0)
			})
		})

		Convey("Given that all dependencies are successfully initialised", func() {
			Convey("Then service Init succeeds and all dependencies are initialised", func() {
				err := svc.Init(ctx, cfg, testBuildTime, testGitCommit, testVersion)
				So(err, ShouldBeNil)
				So(svc.Cfg, ShouldResemble, cfg)
				So(svc.Server, ShouldEqual, serverMock)
				So(svc.HealthCheck, ShouldResemble, hcMock)
				So(svc.ContentPublishedConsumer, ShouldResemble, consumerMock1)
				So(svc.SearchContentConsumer, ShouldResemble, consumerMock2)
				So(svc.ImportProducer, ShouldResemble, producerMock)
				So(svc.DeleteProducer, ShouldResemble, producerMock)
				So(svc.ZebedeeCli, ShouldResemble, zebedeeMock)
				So(svc.DatasetCli, ShouldResemble, datasetAPIMock)

				Convey("Then only necessary checks are registered based on feature flags", func() {
					registeredChecks := make(map[string]bool)
					for _, check := range hcMock.AddAndGetCheckCalls() {
						registeredChecks[check.Name] = true
					}
					if cfg.EnableZebedeeCallbacks {
						So(registeredChecks["Zebedee client"], ShouldBeTrue)
					}
					if cfg.EnableDatasetAPICallbacks {
						So(registeredChecks["DatasetAPI client"], ShouldBeTrue)
					}
					So(registeredChecks["ContentImport Kafka producer"], ShouldBeTrue)
					So(registeredChecks["ContentDelete Kafka producer"], ShouldBeTrue)
					So(registeredChecks["ContentPublished Kafka consumer"], ShouldBeTrue)
					So(registeredChecks["SearchContent Kafka consumer"], ShouldBeTrue)
				})

				Convey("Then Kafka consumers subscribe to the correct health checks", func() {
					So(subscribedTo, ShouldHaveLength, 8)
					if cfg.EnableZebedeeCallbacks {
						So(hcMock.SubscribeCalls()[0].Checks, ShouldContain, testChecks["Zebedee client"])
					}
					So(hcMock.SubscribeCalls()[0].Checks, ShouldContain, testChecks["ContentImport Kafka producer"])
					So(hcMock.SubscribeCalls()[0].Checks, ShouldContain, testChecks["ContentDelete Kafka producer"])
					if cfg.EnableDatasetAPICallbacks {
						So(hcMock.SubscribeCalls()[0].Checks, ShouldContain, testChecks["DatasetAPI client"])
					}
				})
			})
		})
	})
}

func TestStart(t *testing.T) {
	Convey("Having a correctly initialised Service with mocked dependencies", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)

		consumerMock1 := &kafkatest.IConsumerGroupMock{
			LogErrorsFunc: func(ctx context.Context) {},
		}
		consumerMock2 := &kafkatest.IConsumerGroupMock{
			LogErrorsFunc: func(ctx context.Context) {},
		}

		producerMock := &kafkatest.IProducerMock{
			LogErrorsFunc: func(ctx context.Context) {},
		}

		hcMock := &serviceMock.HealthCheckerMock{
			StartFunc: func(ctx context.Context) {},
		}

		serverWg := &sync.WaitGroup{}
		serverMock := &serviceMock.HTTPServerMock{}

		svc := &service.Service{
			Cfg:                      cfg,
			Server:                   serverMock,
			HealthCheck:              hcMock,
			ContentPublishedConsumer: consumerMock1,
			SearchContentConsumer:    consumerMock2,
			ImportProducer:           producerMock,
			DeleteProducer:           producerMock,
		}

		Convey("When a service with a successful HTTP server is started", func() {
			cfg.StopConsumingOnUnhealthy = true
			serverMock.ListenAndServeFunc = func() error {
				serverWg.Done()
				return nil
			}
			serverWg.Add(1)
			err := svc.Start(ctx, make(chan error, 1))
			So(err, ShouldBeNil)

			Convey("Then healthcheck is started and HTTP server starts listening", func() {
				So(len(hcMock.StartCalls()), ShouldEqual, 1)
				serverWg.Wait() // Wait for HTTP server go-routine to finish
				So(len(serverMock.ListenAndServeCalls()), ShouldEqual, 1)
			})
		})

		Convey("When a service is started with StopConsumingOnUnhealthy disabled", func() {
			cfg.StopConsumingOnUnhealthy = false
			consumerMock1.StartFunc = func() error { return nil }
			consumerMock2.StartFunc = func() error { return nil }
			serverMock.ListenAndServeFunc = func() error { return nil }
			err := svc.Start(ctx, make(chan error, 1))
			So(err, ShouldBeNil)

			Convey("Then the kafka consumer is manually started", func() {
				So(consumerMock1.StartCalls(), ShouldHaveLength, 1)
				So(consumerMock2.StartCalls(), ShouldHaveLength, 1)
			})
		})

		Convey("When a service is started with StopConsumingOnUnhealthy disabled and the Start func returns an error", func() {
			cfg.StopConsumingOnUnhealthy = false
			consumerMock1.StartFunc = func() error { return errKafkaConsumer }
			consumerMock2.StartFunc = func() error { return nil }
			serverMock.ListenAndServeFunc = func() error { return nil }
			err := svc.Start(ctx, make(chan error, 1))

			Convey("Then the expected error is returned", func() {
				So(consumerMock1.StartCalls(), ShouldHaveLength, 1)
				So(consumerMock2.StartCalls(), ShouldHaveLength, 0)
				So(err, ShouldNotBeNil)
				So(errors.Unwrap(err), ShouldResemble, errKafkaConsumer)
			})
		})

		Convey("When a service with a failing HTTP server is started", func() {
			cfg.StopConsumingOnUnhealthy = true
			serverMock.ListenAndServeFunc = func() error {
				serverWg.Done()
				return errServer
			}
			errChan := make(chan error, 1)
			serverWg.Add(1)
			err := svc.Start(ctx, errChan)
			So(err, ShouldBeNil)

			Convey("Then HTTP server errors are reported to the provided errors channel", func() {
				rxErr := <-errChan
				So(rxErr.Error(), ShouldResemble, fmt.Sprintf("failure in http listen and serve: %s", errServer.Error()))
			})
		})
	})
}

func TestClose(t *testing.T) {
	Convey("Having a service without initialised dependencies", t, func() {
		cfg := &config.Config{
			GracefulShutdownTimeout: 5 * time.Second,
		}
		svc := service.Service{
			Cfg: cfg,
		}

		Convey("Then the service can be closed without any issue (noop)", func() {
			err := svc.Close(context.Background())
			So(err, ShouldBeNil)
		})
	})

	Convey("Having a service with a kafka consumers that takes more time to stop listening than the graceful shutdown timeout", t, func() {
		cfg := &config.Config{
			GracefulShutdownTimeout: time.Millisecond,
		}
		consumerMock1 := &kafkatest.IConsumerGroupMock{
			StopAndWaitFunc: func() error {
				time.Sleep(100 * time.Millisecond)
				return nil
			},
			CloseFunc: func(ctx context.Context, optFuncs ...kafka.OptFunc) error {
				return nil
			},
		}
		consumerMock2 := &kafkatest.IConsumerGroupMock{
			StopAndWaitFunc: func() error {
				time.Sleep(100 * time.Millisecond)
				return nil
			},
			CloseFunc: func(ctx context.Context, optFuncs ...kafka.OptFunc) error {
				return nil
			},
		}

		svc := service.Service{
			Cfg:                      cfg,
			ContentPublishedConsumer: consumerMock1,
			SearchContentConsumer:    consumerMock2,
		}

		Convey("Then the service fails to close due to a timeout error", func() {
			err := svc.Close(context.Background())
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "shutdown timed out: context deadline exceeded")
		})
	})

	Convey("Having a fully initialised service", t, func() {
		cfg := &config.Config{
			GracefulShutdownTimeout: 5 * time.Second,
		}
		consumerMock1 := &kafkatest.IConsumerGroupMock{}
		consumerMock2 := &kafkatest.IConsumerGroupMock{}
		producerMock := &kafkatest.IProducerMock{}
		hcMock := &serviceMock.HealthCheckerMock{}
		serverMock := &serviceMock.HTTPServerMock{}

		svc := &service.Service{
			Cfg:                      cfg,
			Server:                   serverMock,
			HealthCheck:              hcMock,
			ContentPublishedConsumer: consumerMock1,
			SearchContentConsumer:    consumerMock2,
			ImportProducer:           producerMock,
			DeleteProducer:           producerMock,
		}

		Convey("And all mocks can successfully close, if done in the right order", func() {
			hcStopped := false

			consumerMock1.StopAndWaitFunc = func() error { return nil }
			consumerMock1.CloseFunc = func(ctx context.Context, optFuncs ...kafka.OptFunc) error { return nil }
			consumerMock2.StopAndWaitFunc = func() error { return nil }
			consumerMock2.CloseFunc = func(ctx context.Context, optFuncs ...kafka.OptFunc) error { return nil }
			producerMock.CloseFunc = func(ctx context.Context) error { return nil }
			hcMock.StopFunc = func() { hcStopped = true }
			serverMock.ShutdownFunc = func(ctx context.Context) error {
				if !hcStopped {
					return fmt.Errorf("server stopped before healthcheck")
				}
				return nil
			}

			Convey("Then the service can be successfully closed", func() {
				err := svc.Close(context.Background())
				So(err, ShouldBeNil)

				Convey("And all the dependencies are closed", func() {
					So(consumerMock1.StopAndWaitCalls(), ShouldHaveLength, 1)
					So(consumerMock2.StopAndWaitCalls(), ShouldHaveLength, 1)
					So(hcMock.StopCalls(), ShouldHaveLength, 1)
					So(consumerMock1.CloseCalls(), ShouldHaveLength, 1)
					So(consumerMock2.CloseCalls(), ShouldHaveLength, 1)
					So(serverMock.ShutdownCalls(), ShouldHaveLength, 1)
				})
			})
		})

		Convey("And all mocks fail to close", func() {
			consumerMock1.StopAndWaitFunc = func() error { return errKafkaConsumer }
			consumerMock1.CloseFunc = func(ctx context.Context, optFuncs ...kafka.OptFunc) error { return errKafkaConsumer }
			consumerMock2.StopAndWaitFunc = func() error { return errKafkaConsumer }
			consumerMock2.CloseFunc = func(ctx context.Context, optFuncs ...kafka.OptFunc) error { return errKafkaConsumer }
			producerMock.CloseFunc = func(ctx context.Context) error { return errKafkaProducer }
			hcMock.StopFunc = func() {}
			serverMock.ShutdownFunc = func(ctx context.Context) error { return errServer }

			Convey("Then the service returns the expected error when closed", func() {
				err := svc.Close(context.Background())
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "failed to shutdown gracefully")

				Convey("And all the dependencies are closed", func() {
					So(consumerMock1.StopAndWaitCalls(), ShouldHaveLength, 1)
					So(consumerMock2.StopAndWaitCalls(), ShouldHaveLength, 1)
					So(hcMock.StopCalls(), ShouldHaveLength, 1)
					So(consumerMock1.CloseCalls(), ShouldHaveLength, 1)
					So(consumerMock2.CloseCalls(), ShouldHaveLength, 1)
					So(serverMock.ShutdownCalls(), ShouldHaveLength, 1)
				})
			})
		})
	})
}
