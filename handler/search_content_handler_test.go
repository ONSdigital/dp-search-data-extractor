package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"testing"

	kafka "github.com/ONSdigital/dp-kafka/v4"
	"github.com/ONSdigital/dp-kafka/v4/avro"
	"github.com/ONSdigital/dp-kafka/v4/kafkatest"
	"github.com/ONSdigital/dp-search-data-extractor/models"
	"github.com/ONSdigital/dp-search-data-extractor/schema"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSearchContentHandler_Handle(t *testing.T) {
	Convey("Given a SearchContentHandler with a producer and invalid input", t, func() {
		var producerMock = &kafkatest.IProducerMock{
			SendFunc: func(ctx context.Context, schema *avro.Schema, event interface{}) error {
				return nil
			},
			SendJSONFunc: func(ctx context.Context, event interface{}) error {
				return nil
			},
		}

		handler := &SearchContentHandler{
			ImportProducer: producerMock,
			DeleteProducer: producerMock,
		}

		Convey("When an event with invalid data is handled", func() {
			ctx := context.Background()
			msg, err := kafkatest.NewMessage([]byte("invalid data"), 0)
			So(err, ShouldBeNil)

			err = handler.Handle(ctx, 0, msg)

			Convey("Then an unmarshaling error is reported", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "failed to unmarshal event")
			})
		})

		Convey("Given a SearchContentHandler with a working producer", func() {
			ctx := context.Background()

			var producerMock = &kafkatest.IProducerMock{
				SendFunc: func(ctx context.Context, schema *avro.Schema, event interface{}) error {
					return nil
				},
				SendJSONFunc: func(ctx context.Context, event interface{}) error {
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
					TraceID:     "trace1234",
				}

				msg := createJSONMessage(expectedEvent)
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
					TraceID: "trace1234",
				}

				msg := createJSONMessage(expectedEvent)
				err := handler.Handle(ctx, 0, msg)

				Convey("Then no error is reported", func() {
					So(err, ShouldBeNil)
				})

				Convey("And the expected search-data-imported events are sent to the producer", func() {
					// 1 Avro send (search-data-import)
					So(producerMock.SendCalls(), ShouldHaveLength, 1)
					So(producerMock.SendCalls()[0].Schema, ShouldEqual, schema.SearchDataImportEvent)
					importEv, ok := producerMock.SendCalls()[0].Event.(models.SearchDataImport)
					So(ok, ShouldBeTrue)
					So(importEv.URI, ShouldEqual, expectedEvent.URI)
					So(importEv.Title, ShouldEqual, expectedEvent.Title)
					So(importEv.DataType, ShouldEqual, expectedEvent.ContentType)
					So(importEv.Topics, ShouldResemble, expectedEvent.Topics)
					So(importEv.DateChanges, ShouldResemble, expectedEvent.DateChanges)

					// 1 JSON send (search-content-deleted)
					So(producerMock.SendJSONCalls(), ShouldHaveLength, 1)
					delEv, ok := producerMock.SendJSONCalls()[0].Event.(models.SearchContentDeleted)
					So(ok, ShouldBeTrue)
					So(delEv.URI, ShouldEqual, expectedEvent.URIOld)
				})
			})
		})

		Convey("When an event with nil slices is handled", func() {
			ctx := context.Background()

			expectedEvent := &models.SearchContentUpdate{
				URI:         "/uri/with/nil",
				Title:       "Nil Slices",
				ContentType: "article",
				Topics:      nil, // Intentionally nil
				DateChanges: nil, // Intentionally nil
			}

			msg := createJSONMessage(expectedEvent)
			err := handler.Handle(ctx, 0, msg)

			Convey("Then no error is reported", func() {
				So(err, ShouldBeNil)
			})

			Convey("And only the Avro import event is sent (no delete)", func() {
				So(producerMock.SendCalls(), ShouldHaveLength, 1)     // Avro import
				So(producerMock.SendJSONCalls(), ShouldHaveLength, 0) // no delete

				sent, ok := producerMock.SendCalls()[0].Event.(models.SearchDataImport)
				So(ok, ShouldBeTrue)
				So(err, ShouldBeNil)
				So(sent.Topics, ShouldResemble, []string{})
				So(sent.DateChanges, ShouldBeNil)
			})
		})

		Convey("When the producer fails to send the event", func() {
			producerMock.SendFunc = func(ctx context.Context, schema *avro.Schema, event interface{}) error {
				return errors.New("producer error")
			}

			expectedEvent := &models.SearchContentUpdate{
				URI: "/some/uri",
			}
			msg := createJSONMessage(expectedEvent)
			err := handler.Handle(ctx, 0, msg)

			Convey("Then the expected error is reported", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "failed to send search data import event")
			})
		})
	})
}

func createJSONMessage(s interface{}) kafka.Message {
	b, err := json.Marshal(s)
	if err != nil {
		log.Fatalf("Error marshaling JSON message: %v", err)
	}
	msg, err := kafkatest.NewMessage(b, 0)
	if err != nil {
		log.Fatalf("Error creating Kafka message: %v", err)
	}
	return msg
}
