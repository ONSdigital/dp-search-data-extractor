package handler

import (
	"errors"
	"log"
	"testing"

	kafka "github.com/ONSdigital/dp-kafka/v3"
	"github.com/ONSdigital/dp-kafka/v3/avro"
	"github.com/ONSdigital/dp-kafka/v3/kafkatest"
	"github.com/ONSdigital/dp-search-data-extractor/models"
	"github.com/ONSdigital/dp-search-data-extractor/schema"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSearchContentHandler_Handle(t *testing.T) {
	Convey("Given a SearchContentHandler with a producer and invalid input", t, func() {
		var producerMock = &kafkatest.IProducerMock{
			SendFunc: func(schema *avro.Schema, event interface{}) error {
				return nil
			},
		}

		handler := &SearchContentHandler{
			Producer: producerMock,
		}

		Convey("When an event with invalid data is handled", func() {
			msg, err := kafkatest.NewMessage([]byte("invalid data"), 0)
			So(err, ShouldBeNil)

			err = handler.Handle(ctx, 0, msg)

			Convey("Then an unmarshaling error is reported", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "failed to unmarshal event")
			})
		})
	})

	Convey("Given a SearchContentHandler with a working producer", t, func() {
		expectedEvent := &models.SearchContentUpdate{
			URI:             "/some/uri",
			URIOld:          "/some/old/uri",
			Title:           "Test Title",
			ContentType:     "article",
			Summary:         "Test Summary",
			Survey:          "Test Survey",
			MetaDescription: "Test Meta Description",
			Topics:          []string{"topic1", "topic2"},
			ReleaseDate:     "2023-01-01",
			Language:        "en",
			Edition:         "Test Edition",
			DatasetID:       "dataset123",
			CDID:            "CDID456",
			CanonicalTopic:  "Canonical Topic",
			Release: models.Release{
				Cancelled:       false,
				Finalised:       true,
				Published:       true,
				ProvisionalDate: "2023-01-02",
			},
		}

		var producerMock = &kafkatest.IProducerMock{
			SendFunc: func(schema *avro.Schema, event interface{}) error {
				return nil
			},
		}

		handler := &SearchContentHandler{
			Producer: producerMock,
		}

		Convey("When a valid search-content-updated event is handled", func() {
			msg := createSearchContentMessage(expectedEvent)
			err := handler.Handle(ctx, 0, msg)

			Convey("Then no error is reported", func() {
				So(err, ShouldBeNil)
			})

			Convey("And the expected event is sent to the producer", func() {
				So(producerMock.SendCalls(), ShouldHaveLength, 1)
				So(producerMock.SendCalls()[0].Schema, ShouldEqual, schema.SearchDataImportEvent)

				// Ensure Event is a []byte before unmarshalling
				eventBytes, ok := producerMock.SendCalls()[0].Event.([]byte)
				So(ok, ShouldBeTrue) // Assert the type conversion is successful

				var sentEvent models.SearchDataImport
				err := schema.SearchDataImportEvent.Unmarshal(eventBytes, &sentEvent)
				So(err, ShouldBeNil)
				So(sentEvent.URI, ShouldEqual, expectedEvent.URI)
				So(sentEvent.Title, ShouldEqual, expectedEvent.Title)
				So(sentEvent.DataType, ShouldEqual, expectedEvent.ContentType)
			})

			Convey("When the producer fails to send the event", func() {
				producerMock.SendFunc = func(schema *avro.Schema, event interface{}) error {
					return errors.New("producer error")
				}

				err := handler.Handle(ctx, 0, msg)

				Convey("Then the expected error is reported", func() {
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "failed to send search data import event")
				})
			})
		})
	})
}

func createSearchContentMessage(s interface{}) kafka.Message {
	e, err := schema.SearchContentUpdateEvent.Marshal(s)
	if err != nil {
		log.Fatalf("Error marshaling SearchContentUpdateEvent: %v", err)
	}
	msg, err := kafkatest.NewMessage(e, 0)
	if err != nil {
		log.Fatalf("Error creating Kafka message: %v", err)
	}
	return msg
}
