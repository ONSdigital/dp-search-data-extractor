package event_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/dp-kafka/v2/kafkatest"
	"github.com/ONSdigital/dp-search-data-extractor/event"
	"github.com/ONSdigital/dp-search-data-extractor/event/mock"
	"github.com/ONSdigital/dp-search-data-extractor/models"
	"github.com/ONSdigital/dp-search-data-extractor/schema"
	"github.com/ONSdigital/log.go/log"

	. "github.com/smartystreets/goconvey/convey"
)

var ctx = context.Background()

func TestProducer_SearchDataImport(t *testing.T) {
	Convey("Given SearchDataImportProducer has been configured correctly", t, func() {

		searchDataImportEvent1 := models.SearchDataImport{
			DataType:        "testDataType",
			JobID:           "",
			SearchIndex:     "ONS",
			CDID:            "",
			DatasetID:       "",
			Keywords:        "",
			MetaDescription: "",
			Summary:         "",
			ReleaseDate:     "",
			Title:           "",
			TraceID:         "testTraceID",
		}
		expectedSearchDataImport := marshalSearchDataImport(t, searchDataImportEvent1)

		pChannels := &kafka.ProducerChannels{
			Output: make(chan []byte, 1),
		}

		kafkaProducerMock := &kafkatest.IProducerMock{
			ChannelsFunc: func() *kafka.ProducerChannels {
				return pChannels
			},
		}

		marshallerMock := &mock.MarshallerMock{
			MarshalFunc: func(s interface{}) ([]byte, error) {
				return expectedSearchDataImport, nil
			},
		}

		//event is message
		searchDataImportProducer := event.SearchDataImportProducer{
			Producer:   kafkaProducerMock,
			Marshaller: marshallerMock,
		}
		Convey("When SearchDataImport is called on the event producer", func() {

			err := searchDataImportProducer.SearchDataImport(ctx, searchDataImportEvent1)
			So(err, ShouldBeNil)

			var avroBytes []byte
			select {
			case avroBytes = <-pChannels.Output:
				fmt.Printf("avroBytes: %v\n", avroBytes)
				log.Event(ctx, "avro byte sent to producer output", log.INFO)
			case <-time.After(time.Second * 5):
				log.Event(ctx, "failing test due to timed out", log.INFO)
				t.FailNow()
			}

			Convey("Then the expected bytes are sent to producer.output", func() {
				var actual models.SearchDataImport
				err = schema.SearchDataImportSchema.Unmarshal(avroBytes, &actual)
				So(err, ShouldBeNil)
				So(searchDataImportEvent1, ShouldResemble, actual)
			})
		})
	})
}

func TestProducer_SearchDataImport_MarshalErr(t *testing.T) {
	Convey("Given InstanceCompletedProducer has been configured correctly", t, func() {

		searchDataImportEvent := models.SearchDataImport{
			DataType:        "testDataType",
			JobID:           "",
			SearchIndex:     "ONS",
			CDID:            "",
			DatasetID:       "",
			Keywords:        "",
			MetaDescription: "",
			Summary:         "",
			ReleaseDate:     "",
			Title:           "",
			TraceID:         "testTraceID",
		}

		pChannels := &kafka.ProducerChannels{
			Output: make(chan []byte, 1),
		}

		kafkaProducerMock := &kafkatest.IProducerMock{
			ChannelsFunc: func() *kafka.ProducerChannels {
				return pChannels
			},
		}

		marshallerMock := &mock.MarshallerMock{
			MarshalFunc: func(s interface{}) ([]byte, error) {
				return nil, errors.New("mock error")
			},
		}

		//event is message
		searchDataImportProducer := event.SearchDataImportProducer{
			Producer:   kafkaProducerMock,
			Marshaller: marshallerMock,
		}

		Convey("When marshaller.Marshal returns an error", func() {
			err := searchDataImportProducer.SearchDataImport(ctx, searchDataImportEvent)

			Convey("Then the expected error is returned", func() {
				expectedError := fmt.Errorf(fmt.Sprintf("Marshaller.Marshal returned an error: event=%v: %%w", searchDataImportEvent), errors.New("mock error"))
				So(err.Error(), ShouldEqual, expectedError.Error())
			})

			Convey("And producer.Output is never called", func() {
				So(len(kafkaProducerMock.ChannelsCalls()), ShouldEqual, 0)
			})
		})
	})
}

// marshalSearchDataImport helper method to marshal a event into a []byte
func marshalSearchDataImport(t *testing.T, event models.SearchDataImport) []byte {
	bytes, err := schema.SearchDataImportSchema.Marshal(event)
	if err != nil {
		t.Fatalf("avro mashalling failed with error : %v", err)
	}
	return bytes
}
