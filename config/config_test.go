package config

import (
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConfig(t *testing.T) {
	Convey("Given an environment with no environment variables set", t, func() {
		os.Clearenv()
		cfg, err := Get()
		Convey("When the config values are retrieved", func() {
			Convey("Then there should be no error returned", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the values should be set to the expected defaults", func() {
				So(cfg.BindAddr, ShouldEqual, "localhost:25800")
				So(cfg.GracefulShutdownTimeout, ShouldEqual, 5*time.Second)
				So(cfg.HealthCheckInterval, ShouldEqual, 30*time.Second)
				So(cfg.HealthCheckCriticalTimeout, ShouldEqual, 90*time.Second)
				So(cfg.KafkaAddr, ShouldResemble, []string{"localhost:9092", "localhost:9093", "localhost:9094"})
				So(cfg.KafkaVersion, ShouldEqual, "1.0.2")
				So(cfg.KafkaNumWorkers, ShouldEqual, 1)
				So(cfg.KafkaSecProtocol, ShouldEqual, "")
				So(cfg.KafkaSecCACerts, ShouldEqual, "")
				So(cfg.KafkaSecClientCert, ShouldEqual, "")
				So(cfg.KafkaSecClientKey, ShouldEqual, "")
				So(cfg.KafkaSecSkipVerify, ShouldBeFalse)
				So(cfg.ContentUpdatedGroup, ShouldEqual, "dp-search-data-extractor")
				So(cfg.ContentUpdatedTopic, ShouldEqual, "content-published")
				So(cfg.KafkaProducerTopic, ShouldEqual, "search-data-import")
				So(cfg.ZebedeeURL, ShouldEqual, "http://localhost:8082")
				So(cfg.DatasetAPIURL, ShouldEqual, "http://localhost:22000")
				So(cfg.ServiceAuthToken, ShouldEqual, "")
			})
			Convey("Then a second call to config should return the same config", func() {
				newCfg, newErr := Get()
				So(newErr, ShouldBeNil)
				So(newCfg, ShouldResemble, cfg)
			})
		})
	})
}
