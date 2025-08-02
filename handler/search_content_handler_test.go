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
			ImportProducer: producerMock,
			DeleteProducer: producerMock,
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

		Convey("Given a SearchContentHandler with a working producer", func() {
			var producerMock = &kafkatest.IProducerMock{
				SendFunc: func(schema *avro.Schema, event interface{}) error {
					return nil
				},
			}

			handler := &SearchContentHandler{
				ImportProducer: producerMock,
				DeleteProducer: producerMock,
			}

			Convey("When an event without uri_old is handled", func() {
				expectedEvent := &models.SearchContentUpdate{
					URI:         "/uri/without/old",
					Title:       "No Old URI",
					ContentType: "article",
					SearchIndex: "ons",
					TraceID:     "trace1234",
				}

				msg := createSearchContentMessage(expectedEvent)
				err := handler.Handle(ctx, 0, msg)

				Convey("Then no error is reported", func() {
					So(err, ShouldBeNil)
				})

				Convey("And only the search-data-imported event is sent", func() {
					So(producerMock.SendCalls(), ShouldHaveLength, 1)
					So(producerMock.SendCalls()[0].Schema, ShouldEqual, schema.SearchDataImportEvent)
				})
			})

			Convey("When an event with uri_old is handled", func() {
				expectedEvent := &models.SearchContentUpdate{
					URI:             "/some/uri",
					URIOld:          "/some/old/uri",
					Title:           "Test Title",
					ContentType:     "release",
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
					Cancelled:       false,
					Finalised:       true,
					Published:       true,
					ProvisionalDate: "2023-01-02",
					DateChanges: []models.ReleaseDateDetails{
						{
							ChangeNotice: "Notice 1",
							Date:         "2023-01-01",
						},
					},
					SearchIndex: "ons",
					TraceID:     "trace1234",
				}

				msg := createSearchContentMessage(expectedEvent)
				err := handler.Handle(ctx, 0, msg)

				Convey("Then no error is reported", func() {
					So(err, ShouldBeNil)
				})

				Convey("And the expected search-data-imported events are sent to the producer", func() {
					So(producerMock.SendCalls(), ShouldHaveLength, 2) // Expecting two events: one for import and one for delete
					So(producerMock.SendCalls()[0].Schema, ShouldEqual, schema.SearchDataImportEvent)

					// Check the first event: search-data-imported
					sentEvent, ok := producerMock.SendCalls()[0].Event.(models.SearchDataImport)
					So(ok, ShouldBeTrue) // Assert the event is of the correct type
					So(sentEvent.URI, ShouldEqual, expectedEvent.URI)
					So(sentEvent.Title, ShouldEqual, expectedEvent.Title)
					So(sentEvent.DataType, ShouldEqual, expectedEvent.ContentType)
					So(sentEvent.Topics, ShouldResemble, expectedEvent.Topics)
					So(sentEvent.DateChanges, ShouldResemble, expectedEvent.DateChanges)

					// Check the second event: search-content-deleted
					So(producerMock.SendCalls()[1].Schema, ShouldEqual, schema.SearchContentDeletedEvent)

					deleteEvent, ok := producerMock.SendCalls()[1].Event.(models.SearchContentDeleted)
					So(ok, ShouldBeTrue) // Assert the delete event is of the correct type
					So(deleteEvent.URI, ShouldEqual, expectedEvent.URIOld)
				})
			})
		})

		Convey("When an event with nil slices is handled", func() {
			expectedEvent := &models.SearchContentUpdate{
				URI:         "/uri/with/nil",
				Title:       "Nil Slices",
				ContentType: "article",
				Topics:      nil, // Intentionally nil
				DateChanges: nil, // Intentionally nil
			}

			msg := createSearchContentMessage(expectedEvent)
			err := handler.Handle(ctx, 0, msg)

			Convey("Then no error is reported", func() {
				So(err, ShouldBeNil)
			})

			Convey("And the empty slices are properly handled by the producer", func() {
				sentEvent, ok := producerMock.SendCalls()[0].Event.(models.SearchDataImport)
				So(ok, ShouldBeTrue) // Assert the event is of the correct type

				So(err, ShouldBeNil)
				So(sentEvent.Topics, ShouldResemble, []string{})
				So(sentEvent.DateChanges, ShouldBeNil) // Remains nil
			})
		})

		Convey("When the producer fails to send the event", func() {
			producerMock.SendFunc = func(schema *avro.Schema, event interface{}) error {
				return errors.New("producer error")
			}

			expectedEvent := &models.SearchContentUpdate{
				URI: "/some/uri",
			}
			msg := createSearchContentMessage(expectedEvent)
			err := handler.Handle(ctx, 0, msg)

			Convey("Then the expected error is reported", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "failed to send search data import event")
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
