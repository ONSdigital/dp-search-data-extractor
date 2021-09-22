package event_test

import (
	"context"
	"testing"

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

var pChannels = &kafka.ProducerChannels{
	Output: make(chan []byte, 1),
}
var getChannelFunc = func() *kafka.ProducerChannels {
	return pChannels
}

func TestProducer_SearchDataImport(t *testing.T) {

	Convey("Given an a mock message producer", t, func() {

		// channel to capture messages sent.
		outputChannel := make(chan []byte, 1)

		avroBytes := []byte("Search Data Import")

		kafkaProducerMock := &kafkatest.IProducerMock{
			ChannelsFunc: getChannelFunc,
		}

		marshallerMock := &mock.MarshallerMock{
			MarshalFunc: func(s interface{}) ([]byte, error) {
				return avroBytes, nil
			},
		}

		eventProducer := event.SearchDataImportProducer{
			Producer:   kafkaProducerMock,
			Marshaller: marshallerMock,
		}

		Convey("when SearchDataImport is called with a nil event", func() {
			err := eventProducer.SearchDataImport(ctx, nil)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldEqual, "event required but was nil")
			})

			Convey("and marshaller is never called", func() {
				So(len(marshallerMock.MarshalCalls()), ShouldEqual, 0)
			})
		})

		Convey("When SearchDataImport is called on the event producer", func() {
			event := &models.SearchDataImport{
				DataType:        "",
				JobID:           "",
				SearchIndex:     "",
				CDID:            "",
				DatasetID:       "",
				Keywords:        "",
				MetaDescription: "",
				Summary:         "",
				ReleaseDate:     "",
				Title:           "",
				TraceID:         "",
			}
			err := eventProducer.SearchDataImport(ctx, event)

			Convey("The expected event is available on the output channel", func() {
				log.Event(ctx, "error is:", log.INFO, log.Data{"error": err})
				So(err, ShouldBeNil)

				messageBytes := <-outputChannel
				close(outputChannel)
				So(messageBytes, ShouldResemble, avroBytes)
			})
		})
	})
}

// Unmarshal converts observation events to []byte.
func unmarshal(bytes []byte) *models.SearchDataImport {
	event := &models.SearchDataImport{}
	err := schema.SearchDataImportEvent.Unmarshal(bytes, event)
	So(err, ShouldBeNil)
	return event
}
