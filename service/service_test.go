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
		"Zebedee client":    {},
		"DatasetAPI client": {},
		"Kafka producer":    {},
	}

	errKafkaConsumer = errors.New("Kafka consumer error")
	errKafkaProducer = errors.New("Kafka producer error")
	errHealthcheck   = errors.New("healthCheck error")
	errServer        = errors.New("HTTP Server error")
	errAddCheck      = fmt.Errorf("healthcheck add check error")

	funcDoGetKafkaConsumerErr = func(ctx context.Context, cfg *config.Kafka) (kafka.IConsumerGroup, error) {
		return nil, errKafkaConsumer
	}

	funcDoGetKafkaProducerErr = func(ctx context.Context, cfg *config.Kafka) (kafka.IProducer, error) {
		return nil, errKafkaProducer
	}

	funcDoGetHealthcheckErr = func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
		return nil, errHealthcheck
	}

	funcDoGetHTTPServerNil = func(bindAddr string, router http.Handler) service.HTTPServer {
		return nil
	}
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

		consumerMock := &kafkatest.IConsumerGroupMock{
			RegisterHandlerFunc: func(ctx context.Context, h kafka.Handler) error {
				return nil
			},
		}
		service.GetKafkaConsumer = func(ctx context.Context, cfg *config.Kafka) (kafka.IConsumerGroup, error) {
			return consumerMock, nil
		}

		producerMock := &kafkatest.IProducerMock{}
		service.GetKafkaProducer = func(ctx context.Context, cfg *config.Kafka) (kafka.IProducer, error) {
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

		datasetApiMock := &clientMock.DatasetClientMock{
			CheckerFunc: func(context.Context, *healthcheck.CheckState) error { return nil },
		}
		service.GetDatasetClient = func(cfg *config.Config) clients.DatasetClient {
			return datasetApiMock
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
			service.GetKafkaConsumer = func(ctx context.Context, cfg *config.Kafka) (kafka.IConsumerGroup, error) {
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

		Convey("Given that Kafka consumer fails to register a handler", func() {
			consumerMock.RegisterHandlerFunc = func(ctx context.Context, h kafka.Handler) error {
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
			service.GetKafkaProducer = func(ctx context.Context, cfg *config.Kafka) (kafka.IProducer, error) {
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
				So(svc.Consumer, ShouldResemble, consumerMock)

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
				So(svc.Consumer, ShouldResemble, consumerMock)

				Convey("And all other checkers try to register", func() {
					So(hcMock.AddAndGetCheckCalls(), ShouldHaveLength, 4)
				})
			})
		})

		Convey("Given that all dependencies are successfully initialised", func() {

			Convey("Then service Init succeeds and all dependencies are initialised", func() {
				err := svc.Init(ctx, cfg, testBuildTime, testGitCommit, testVersion)
				So(err, ShouldBeNil)
				So(svc.Cfg, ShouldResemble, cfg)
				So(svc.Server, ShouldEqual, serverMock)
				So(svc.HealthCheck, ShouldResemble, hcMock)
				So(svc.Consumer, ShouldResemble, consumerMock)
				So(svc.Producer, ShouldResemble, producerMock)
				So(svc.ZebedeeCli, ShouldResemble, zebedeeMock)
				So(svc.DatasetCli, ShouldResemble, datasetApiMock)

				Convey("Then all checks are registered", func() {
					So(hcMock.AddAndGetCheckCalls(), ShouldHaveLength, 4)
					So(hcMock.AddAndGetCheckCalls()[0].Name, ShouldResemble, "Zebedee client")
					So(hcMock.AddAndGetCheckCalls()[1].Name, ShouldResemble, "Kafka consumer")
					So(hcMock.AddAndGetCheckCalls()[2].Name, ShouldResemble, "Kafka producer")
					So(hcMock.AddAndGetCheckCalls()[3].Name, ShouldResemble, "DatasetAPI client")
				})

				Convey("Then kafka consumer subscribes to the expected healthcheck checks", func() {
					So(subscribedTo, ShouldHaveLength, 3)
					So(hcMock.SubscribeCalls(), ShouldHaveLength, 1)
					So(hcMock.SubscribeCalls()[0].Checks, ShouldContain, testChecks["Zebedee client"])
					So(hcMock.SubscribeCalls()[0].Checks, ShouldContain, testChecks["Kafka producer"])
					So(hcMock.SubscribeCalls()[0].Checks, ShouldContain, testChecks["DatasetAPI client"])
				})

				Convey("Then the handler is registered to the kafka consumer", func() {
					So(consumerMock.RegisterHandlerCalls(), ShouldHaveLength, 1)
				})
			})
		})
	})
}

func TestStart(t *testing.T) {
	Convey("Having a correctly initialised Service with mocked dependencies", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)

		consumerMock := &kafkatest.IConsumerGroupMock{
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
			Cfg:         cfg,
			Server:      serverMock,
			HealthCheck: hcMock,
			Consumer:    consumerMock,
			Producer:    producerMock,
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
			consumerMock.StartFunc = func() error { return nil }
			serverMock.ListenAndServeFunc = func() error { return nil }
			err := svc.Start(ctx, make(chan error, 1))
			So(err, ShouldBeNil)

			Convey("Then the kafka consumer is manually started", func() {
				So(consumerMock.StartCalls(), ShouldHaveLength, 1)
			})
		})

		Convey("When a service is started with StopConsumingOnUnhealthy disabled and the Start func returns an error", func() {
			cfg.StopConsumingOnUnhealthy = false
			consumerMock.StartFunc = func() error { return errKafkaConsumer }
			serverMock.ListenAndServeFunc = func() error { return nil }
			err := svc.Start(ctx, make(chan error, 1))

			Convey("Then the expected error is returned", func() {
				So(consumerMock.StartCalls(), ShouldHaveLength, 1)
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

	Convey("Having a service with a kafka consumer that takes more time to stop listening than the graceful shutdown timeout", t, func() {
		cfg := &config.Config{
			GracefulShutdownTimeout: time.Millisecond,
		}
		consumerMock := &kafkatest.IConsumerGroupMock{
			StopAndWaitFunc: func() error {
				time.Sleep(100 * time.Millisecond)
				return nil
			},
		}

		svc := service.Service{
			Cfg:      cfg,
			Consumer: consumerMock,
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
		consumerMock := &kafkatest.IConsumerGroupMock{}
		producerMock := &kafkatest.IProducerMock{}
		hcMock := &serviceMock.HealthCheckerMock{}
		serverMock := &serviceMock.HTTPServerMock{}

		svc := &service.Service{
			Cfg:         cfg,
			Server:      serverMock,
			HealthCheck: hcMock,
			Consumer:    consumerMock,
			Producer:    producerMock,
		}

		Convey("And all mocks can successfully close, if done in the right order", func() {
			hcStopped := false

			consumerMock.StopAndWaitFunc = func() error { return nil }
			consumerMock.CloseFunc = func(ctx context.Context, optFuncs ...kafka.OptFunc) error { return nil }
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
					So(consumerMock.StopAndWaitCalls(), ShouldHaveLength, 1)
					So(hcMock.StopCalls(), ShouldHaveLength, 1)
					So(consumerMock.CloseCalls(), ShouldHaveLength, 1)
					So(serverMock.ShutdownCalls(), ShouldHaveLength, 1)
				})
			})
		})

		Convey("And all mocks fail to close", func() {
			consumerMock.StopAndWaitFunc = func() error { return errKafkaConsumer }
			consumerMock.CloseFunc = func(ctx context.Context, optFuncs ...kafka.OptFunc) error { return errKafkaConsumer }
			producerMock.CloseFunc = func(ctx context.Context) error { return errKafkaProducer }
			hcMock.StopFunc = func() {}
			serverMock.ShutdownFunc = func(ctx context.Context) error { return errServer }

			Convey("Then the service returns the expected error when closed", func() {
				err := svc.Close(context.Background())
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "failed to shutdown gracefully")

				Convey("And all the dependencies are closed", func() {
					So(consumerMock.StopAndWaitCalls(), ShouldHaveLength, 1)
					So(hcMock.StopCalls(), ShouldHaveLength, 1)
					So(consumerMock.CloseCalls(), ShouldHaveLength, 1)
					So(serverMock.ShutdownCalls(), ShouldHaveLength, 1)
				})
			})
		})
	})
}

// func TestRun(t *testing.T) {
// 	Convey("Having a set of mocked dependencies", t, func() {
// 		consumerMock := &kafkatest.IConsumerGroupMock{
// 			CheckerFunc:   func(ctx context.Context, state *healthcheck.CheckState) error { return nil },
// 			ChannelsFunc:  func() *kafka.ConsumerGroupChannels { return &kafka.ConsumerGroupChannels{} },
// 			LogErrorsFunc: func(ctx context.Context) {},
// 		}

// 		producerMock := &kafkatest.IProducerMock{
// 			CheckerFunc:   func(ctx context.Context, state *healthcheck.CheckState) error { return nil },
// 			LogErrorsFunc: func(ctx context.Context) {},
// 		}

// 		hcMock := &serviceMock.HealthCheckerMock{
// 			AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
// 			StartFunc:    func(ctx context.Context) {},
// 		}
// 		serverWg := &sync.WaitGroup{}
// 		serverMock := &serviceMock.HTTPServerMock{
// 			ListenAndServeFunc: func() error {
// 				serverWg.Done()
// 				return nil
// 			},
// 		}

// 		zebedeeMock := &clientMock.ZebedeeClientMock{
// 			CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error { return nil },
// 		}

// 		datasetMock := &clientMock.DatasetClientMock{
// 			CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error { return nil },
// 		}

// 		funcDoGetKafkaConsumerOk := func(ctx context.Context, cfg *config.Kafka) (kafka.IConsumerGroup, error) {
// 			return consumerMock, nil
// 		}

// 		funcDoGetKafkaProducerOk := func(ctx context.Context, cfg *config.Kafka) (kafka.IProducer, error) {
// 			return producerMock, nil
// 		}

// 		funcDoGetHealthcheckOk := func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
// 			return hcMock, nil
// 		}
// 		funcDoGetHTTPServer := func(bindAddr string, router http.Handler) service.HTTPServer {
// 			return serverMock
// 		}

// 		funcDoGetZebedeeOk := func(cfg *config.Config) clients.ZebedeeClient {
// 			return zebedeeMock
// 		}

// 		funcDoGetDatasetOk := func(cfg *config.Config) clients.DatasetClient {
// 			return datasetMock
// 		}

// 		Convey("Given that initialising Kafka consumer returns an error", func() {
// 			initMock := &serviceMock.InitialiserMock{
// 				DoGetHTTPServerFunc:    funcDoGetHTTPServerNil,
// 				DoGetKafkaConsumerFunc: funcDoGetKafkaConsumerErr,
// 				DoGetKafkaProducerFunc: funcDoGetKafkaProducerErr,
// 				DoGetZebedeeClientFunc: funcDoGetZebedeeOk,
// 				DoGetDatasetClientFunc: funcDoGetDatasetOk,
// 			}
// 			svcErrors := make(chan error, 1)
// 			svcList := service.NewServiceList(initMock)
// 			_, err := service.Run(ctx, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)
// 			Convey("Then service Run fails with the same error and the flag is not set", func() {
// 				So(err, ShouldResemble, errKafkaConsumer)
// 				So(svcList.KafkaConsumer, ShouldBeFalse)
// 				So(svcList.KafkaProducer, ShouldBeFalse)
// 				So(svcList.ZebedeeClient, ShouldBeTrue)
// 				So(svcList.HealthCheck, ShouldBeFalse)
// 				So(svcList.DatasetClient, ShouldBeTrue)
// 			})
// 		})
// 		Convey("Given that initialising healthcheck returns an error", func() {
// 			initMock := &serviceMock.InitialiserMock{
// 				DoGetHTTPServerFunc:    funcDoGetHTTPServerNil,
// 				DoGetHealthCheckFunc:   funcDoGetHealthcheckErr,
// 				DoGetKafkaConsumerFunc: funcDoGetKafkaConsumerOk,
// 				DoGetKafkaProducerFunc: funcDoGetKafkaProducerOk,
// 				DoGetZebedeeClientFunc: funcDoGetZebedeeOk,
// 				DoGetDatasetClientFunc: funcDoGetDatasetOk,
// 			}
// 			svcErrors := make(chan error, 1)
// 			svcList := service.NewServiceList(initMock)
// 			_, err := service.Run(ctx, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)
// 			Convey("Then service Run fails with the same error and the flag is not set", func() {
// 				So(err, ShouldResemble, errHealthcheck)
// 				So(svcList.KafkaConsumer, ShouldBeTrue)
// 				So(svcList.KafkaProducer, ShouldBeTrue)
// 				So(svcList.ZebedeeClient, ShouldBeTrue)
// 				So(svcList.HealthCheck, ShouldBeFalse)
// 				So(svcList.DatasetClient, ShouldBeTrue)
// 			})
// 		})
// 		Convey("Given that all dependencies are successfully initialised", func() {
// 			initMock := &serviceMock.InitialiserMock{
// 				DoGetHTTPServerFunc:    funcDoGetHTTPServer,
// 				DoGetHealthCheckFunc:   funcDoGetHealthcheckOk,
// 				DoGetKafkaConsumerFunc: funcDoGetKafkaConsumerOk,
// 				DoGetKafkaProducerFunc: funcDoGetKafkaProducerOk,
// 				DoGetZebedeeClientFunc: funcDoGetZebedeeOk,
// 				DoGetDatasetClientFunc: funcDoGetDatasetOk,
// 			}
// 			svcErrors := make(chan error, 1)
// 			svcList := service.NewServiceList(initMock)
// 			serverWg.Add(1)
// 			_, err := service.Run(ctx, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)
// 			Convey("Then service Run succeeds and all the flags are set", func() {
// 				So(err, ShouldBeNil)
// 				So(svcList.ZebedeeClient, ShouldBeTrue)
// 				So(svcList.KafkaConsumer, ShouldBeTrue)
// 				So(svcList.HealthCheck, ShouldBeTrue)
// 				So(svcList.DatasetClient, ShouldBeTrue)
// 			})

// 			Convey("The checkers are registered and the healthcheck and http server started", func() {
// 				So(len(hcMock.AddCheckCalls()), ShouldEqual, 4)
// 				So(hcMock.AddCheckCalls()[0].Name, ShouldResemble, "Zebedee client")
// 				So(hcMock.AddCheckCalls()[1].Name, ShouldResemble, "Kafka consumer")
// 				So(hcMock.AddCheckCalls()[2].Name, ShouldResemble, "Kafka producer")
// 				So(hcMock.AddCheckCalls()[3].Name, ShouldResemble, "DatasetAPI client")

// 				So(len(initMock.DoGetHTTPServerCalls()), ShouldEqual, 1)
// 				So(initMock.DoGetHTTPServerCalls()[0].BindAddr, ShouldEqual, "localhost:25800")
// 				So(len(hcMock.StartCalls()), ShouldEqual, 1)
// 				serverWg.Wait() // Wait for HTTP server go-routine to finish
// 				So(len(serverMock.ListenAndServeCalls()), ShouldEqual, 1)
// 			})
// 		})
// 		Convey("Given that Checkers cannot be registered", func() {
// 			errAddheckFail := errors.New("Error(s) registering checkers for healthcheck")
// 			hcMockAddFail := &serviceMock.HealthCheckerMock{
// 				AddCheckFunc: func(name string, checker healthcheck.Checker) error { return errAddheckFail },
// 				StartFunc:    func(ctx context.Context) {},
// 			}
// 			initMock := &serviceMock.InitialiserMock{
// 				DoGetHTTPServerFunc: funcDoGetHTTPServerNil,
// 				DoGetHealthCheckFunc: func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
// 					return hcMockAddFail, nil
// 				},
// 				DoGetKafkaConsumerFunc: funcDoGetKafkaConsumerOk,
// 				DoGetKafkaProducerFunc: funcDoGetKafkaProducerOk,
// 				DoGetZebedeeClientFunc: funcDoGetZebedeeOk,
// 				DoGetDatasetClientFunc: funcDoGetDatasetOk,
// 			}
// 			svcErrors := make(chan error, 1)
// 			svcList := service.NewServiceList(initMock)
// 			_, err := service.Run(ctx, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)
// 			Convey("Then service Run fails, but all checks try to register", func() {
// 				So(err, ShouldNotBeNil)
// 				So(err.Error(), ShouldResemble, fmt.Sprintf("unable to register checkers: %s", errAddheckFail.Error()))
// 				So(svcList.HealthCheck, ShouldBeTrue)
// 				So(svcList.ZebedeeClient, ShouldBeTrue)
// 				So(svcList.KafkaConsumer, ShouldBeTrue)
// 				So(svcList.KafkaProducer, ShouldBeTrue)
// 				So(len(hcMockAddFail.AddCheckCalls()), ShouldEqual, 4)
// 				So(hcMockAddFail.AddCheckCalls()[0].Name, ShouldResemble, "Zebedee client")
// 				So(hcMockAddFail.AddCheckCalls()[1].Name, ShouldResemble, "Kafka consumer")
// 				So(hcMockAddFail.AddCheckCalls()[2].Name, ShouldResemble, "Kafka producer")
// 				So(hcMockAddFail.AddCheckCalls()[3].Name, ShouldResemble, "DatasetAPI client")
// 			})
// 		})
// 	})
// }

// func TestClose(t *testing.T) {
// 	Convey("Having a correctly initialised service", t, func() {
// 		hcStopped := false

// 		zebedeeMock := &clientMock.ZebedeeClientMock{
// 			CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error { return nil },
// 		}

// 		datasetMock := &clientMock.DatasetClientMock{
// 			CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error { return nil },
// 		}

// 		consumerMock := &kafkatest.IConsumerGroupMock{
// 			LogErrorsFunc:   func(ctx context.Context) {},
// 			StopAndWaitFunc: func() error { return nil },
// 			CloseFunc:       func(ctx context.Context, optFuncs ...kafka.OptFunc) error { return nil },
// 			CheckerFunc:     func(ctx context.Context, state *healthcheck.CheckState) error { return nil },
// 			ChannelsFunc:    func() *kafka.ConsumerGroupChannels { return &kafka.ConsumerGroupChannels{} },
// 		}

// 		producerMock := &kafkatest.IProducerMock{
// 			LogErrorsFunc: func(ctx context.Context) {},
// 			CloseFunc:     func(ctx context.Context) error { return nil },
// 			CheckerFunc:   func(ctx context.Context, state *healthcheck.CheckState) error { return nil },
// 			ChannelsFunc:  func() *kafka.ProducerChannels { return &kafka.ProducerChannels{} },
// 		}

// 		// healthcheck Stop does not depend on any other service being closed/stopped
// 		hcMock := &serviceMock.HealthCheckerMock{
// 			AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
// 			StartFunc:    func(ctx context.Context) {},
// 			StopFunc:     func() { hcStopped = true },
// 		}
// 		// server Shutdown will fail if healthcheck is not stopped
// 		serverMock := &serviceMock.HTTPServerMock{
// 			ListenAndServeFunc: func() error { return nil },
// 			ShutdownFunc: func(ctx context.Context) error {
// 				if !hcStopped {
// 					return errors.New("Server stopped before healthcheck")
// 				}
// 				return nil
// 			},
// 		}
// 		Convey("Closing the service results in all the dependencies being closed in the expected order", func() {
// 			initMock := &serviceMock.InitialiserMock{
// 				DoGetHTTPServerFunc: func(bindAddr string, router http.Handler) service.HTTPServer { return serverMock },
// 				DoGetHealthCheckFunc: func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
// 					return hcMock, nil
// 				},
// 				DoGetKafkaConsumerFunc: func(ctx context.Context, cfg *config.Kafka) (kafka.IConsumerGroup, error) {
// 					return consumerMock, nil
// 				},
// 				DoGetKafkaProducerFunc: func(ctx context.Context, cfg *config.Kafka) (kafka.IProducer, error) { return producerMock, nil },
// 				DoGetZebedeeClientFunc: func(cfg *config.Config) clients.ZebedeeClient { return zebedeeMock },
// 				DoGetDatasetClientFunc: func(cfg *config.Config) clients.DatasetClient { return datasetMock },
// 			}

// 			svcErrors := make(chan error, 1)
// 			svcList := service.NewServiceList(initMock)
// 			svc, err := service.Run(ctx, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)
// 			So(err, ShouldBeNil)
// 			err = svc.Close(context.Background())
// 			So(err, ShouldBeNil)
// 			So(len(consumerMock.StopAndWaitCalls()), ShouldEqual, 1)
// 			So(len(hcMock.StopCalls()), ShouldEqual, 1)
// 			So(len(consumerMock.CloseCalls()), ShouldEqual, 1)
// 			So(len(serverMock.ShutdownCalls()), ShouldEqual, 1)
// 		})
// 		Convey("If services fail to stop, the Close operation tries to close all dependencies and returns an error", func() {
// 			failingserverMock := &serviceMock.HTTPServerMock{
// 				ListenAndServeFunc: func() error { return nil },
// 				ShutdownFunc: func(ctx context.Context) error {
// 					return errors.New("Failed to stop http server")
// 				},
// 			}
// 			initMock := &serviceMock.InitialiserMock{
// 				DoGetHTTPServerFunc: func(bindAddr string, router http.Handler) service.HTTPServer { return failingserverMock },
// 				DoGetHealthCheckFunc: func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
// 					return hcMock, nil
// 				},
// 				DoGetKafkaConsumerFunc: func(ctx context.Context, cfg *config.Kafka) (kafka.IConsumerGroup, error) {
// 					return consumerMock, nil
// 				},
// 				DoGetKafkaProducerFunc: func(ctx context.Context, cfg *config.Kafka) (kafka.IProducer, error) { return producerMock, nil },
// 				DoGetZebedeeClientFunc: func(cfg *config.Config) clients.ZebedeeClient { return zebedeeMock },
// 				DoGetDatasetClientFunc: func(cfg *config.Config) clients.DatasetClient { return datasetMock },
// 			}
// 			svcErrors := make(chan error, 1)
// 			svcList := service.NewServiceList(initMock)
// 			svc, err := service.Run(ctx, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)
// 			So(err, ShouldBeNil)
// 			err = svc.Close(context.Background())
// 			So(err, ShouldNotBeNil)
// 			So(len(hcMock.StopCalls()), ShouldEqual, 1)
// 			So(len(failingserverMock.ShutdownCalls()), ShouldEqual, 1)
// 		})
// 	})
// }
