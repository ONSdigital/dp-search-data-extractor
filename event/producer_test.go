package event_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	dpkafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/dp-kafka/v2/kafkatest"
	"github.com/ONSdigital/dp-search-data-extractor/event"
	"github.com/ONSdigital/dp-search-data-extractor/event/mock"
	"github.com/ONSdigital/dp-search-data-extractor/models"
	"github.com/ONSdigital/dp-search-data-extractor/schema"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	somekeyword0 = "keyword0"
	somekeyword1 = "keyword1"

	someMetaDescription = "meta desc"
	someReleaseDate     = "2021-12-13"
	someTitle           = "Some-Incredible-Title"
)

var (
	ctx = context.Background()

	expectedSearchDataImportEvent = models.SearchDataImport{
		DataType:        "testDataType",
		JobID:           "",
		SearchIndex:     "ONS",
		CDID:            "",
		DatasetID:       "",
		Keywords:        []string{""},
		MetaDescription: "",
		Summary:         "",
		ReleaseDate:     "",
		Title:           "",
		Topics:          []string{""},
		DateChanges:     []models.ReleaseDateDetails{{}},
		Cancelled:       false,
		Finalised:       false,
		ProvisionalDate: "",
		Published:       false,
	}

	expectedVersionMetadataEvent = models.SearchDataImport{
		DataType:        "testDataType",
		JobID:           "",
		SearchIndex:     "ONS",
		CDID:            "",
		DatasetID:       "",
		Keywords:        []string{somekeyword0, somekeyword1},
		MetaDescription: someMetaDescription,
		Summary:         "",
		ReleaseDate:     someReleaseDate,
		Title:           someTitle,
		Topics:          []string{"testtopic1", "testtopic2"},
	}
)

func TestProducer_SearchDataImport(t *testing.T) {
	Convey("Given SearchDataImportProducer has been configured correctly", t, func() {
		pChannels := &dpkafka.ProducerChannels{
			Output: make(chan []byte, 1),
		}

		kafkaProducerMock := &kafkatest.IProducerMock{
			ChannelsFunc: func() *dpkafka.ProducerChannels {
				return pChannels
			},
		}

		marshallerMock := &mock.MarshallerMock{
			MarshalFunc: func(s interface{}) ([]byte, error) {
				return schema.SearchDataImportEvent.Marshal(s)
			},
		}

		// event is message
		searchDataImportProducer := event.SearchDataImportProducer{
			Producer:   kafkaProducerMock,
			Marshaller: marshallerMock,
		}
		Convey("When SearchDataImport is called on the event producer", func() {
			err := searchDataImportProducer.SearchDataImport(ctx, expectedSearchDataImportEvent)
			So(err, ShouldBeNil)

			var avroBytes []byte
			var testTimeout = time.Second * 5
			select {
			case avroBytes = <-pChannels.Output:
				t.Log("avro byte sent to producer output")
			case <-time.After(testTimeout):
				t.Fatalf("failing test due to timing out after %v seconds", testTimeout)
				t.FailNow()
			}

			Convey("Then the expected bytes are sent to producer.output", func() {
				var actual models.SearchDataImport
				err = schema.SearchDataImportEvent.Unmarshal(avroBytes, &actual)
				So(err, ShouldBeNil)
				So(expectedSearchDataImportEvent, ShouldResemble, actual)
			})
		})
	})
}

func TestProducer_SearchDataImport_MarshalErr(t *testing.T) {
	Convey("Given InstanceCompletedProducer has been configured correctly", t, func() {
		pChannels := &dpkafka.ProducerChannels{
			Output: make(chan []byte, 1),
		}

		kafkaProducerMock := &kafkatest.IProducerMock{
			ChannelsFunc: func() *dpkafka.ProducerChannels {
				return pChannels
			},
		}

		marshallerMock := &mock.MarshallerMock{
			MarshalFunc: func(s interface{}) ([]byte, error) {
				return nil, errors.New("mock error")
			},
		}

		// event is message
		searchDataImportProducer := event.SearchDataImportProducer{
			Producer:   kafkaProducerMock,
			Marshaller: marshallerMock,
		}

		Convey("When marshaller.Marshal returns an error", func() {
			err := searchDataImportProducer.SearchDataImport(ctx, expectedSearchDataImportEvent)

			Convey("Then the expected error is returned", func() {
				expectedError := fmt.Errorf(fmt.Sprintf("Marshaller.Marshal returned an error: event=%v: %%w", expectedSearchDataImportEvent), errors.New("mock error"))
				So(err.Error(), ShouldEqual, expectedError.Error())
			})

			Convey("And producer.Output is never called", func() {
				So(len(kafkaProducerMock.ChannelsCalls()), ShouldEqual, 0)
			})
		})
	})
}

func TestProducer_SearchDatasetVersionMetadataImport(t *testing.T) {
	t.Parallel()
	Convey("Given SearchDatasetVersionMetadataImport has been configured correctly", t, func() {
		pChannels := &dpkafka.ProducerChannels{
			Output: make(chan []byte, 1),
		}

		kafkaProducerMock := &kafkatest.IProducerMock{
			ChannelsFunc: func() *dpkafka.ProducerChannels {
				return pChannels
			},
		}

		marshallerMock := &mock.MarshallerMock{
			MarshalFunc: func(s interface{}) ([]byte, error) {
				return schema.SearchDataImportEvent.Marshal(s)
			},
		}

		// event is message
		searchDataVersionImportProducer := event.SearchDataImportProducer{
			Producer:   kafkaProducerMock,
			Marshaller: marshallerMock,
		}
		Convey("When SearchDatasetVersionMetadataImport is called on the event producer", func() {
			err := searchDataVersionImportProducer.SearchDataImport(ctx, expectedVersionMetadataEvent)
			So(err, ShouldBeNil)

			var avroBytes []byte
			var testTimeout = time.Second * 5
			select {
			case avroBytes = <-pChannels.Output:
				t.Log("avro byte sent to producer output")
			case <-time.After(testTimeout):
				t.Fatalf("failing test due to timing out after %v seconds", testTimeout)
				t.FailNow()
			}

			Convey("Then the expected bytes are sent to producer.output", func() {
				var actual models.SearchDataImport
				err = schema.SearchDataImportEvent.Unmarshal(avroBytes, &actual)
				So(err, ShouldBeNil)
				So(expectedVersionMetadataEvent, ShouldResemble, actual)
			})
		})
	})
}
