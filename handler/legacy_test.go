package handler

import (
	"context"
	"errors"
	"testing"

	"github.com/ONSdigital/dp-kafka/v3/avro"
	"github.com/ONSdigital/dp-kafka/v3/kafkatest"
	clientMock "github.com/ONSdigital/dp-search-data-extractor/clients/mock"
	"github.com/ONSdigital/dp-search-data-extractor/config"
	"github.com/ONSdigital/dp-search-data-extractor/models"
	. "github.com/smartystreets/goconvey/convey"
)

func TestHandleZebedeeTypeErrors(t *testing.T) {
	Convey("Given an empty handler and a ContentPublished event with a malformed URI that contains 'datasets' substring", t, func() {
		h := &ContentPublished{}
		cpEvent := models.ContentPublished{
			URI: "wrong%%datasets",
		}

		Convey("Then the zebedee handler returns the expected error", func() {
			err := h.handleZebedeeType(ctx, &cpEvent)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "parse \"wrong%%datasets\": invalid URL escape \"%%d\"")
		})
	})

	Convey("Given a handler with a zebedee mock that fails to return published data", t, func() {
		zebedeeMock := &clientMock.ZebedeeClientMock{
			GetPublishedDataFunc: func(ctx context.Context, uriString string) ([]byte, error) {
				return nil, errors.New("zebedee error")
			},
		}
		h := &ContentPublished{ZebedeeCli: zebedeeMock}

		Convey("Then the zebedee handler fails with the expected error when a valid event is handled", func() {
			err := h.handleZebedeeType(ctx, &testZebedeeEvent)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "zebedee error")
		})
	})

	Convey("Given a handler with a zebedee mock that returns malformed data", t, func() {
		zebedeeMock := &clientMock.ZebedeeClientMock{
			GetPublishedDataFunc: func(ctx context.Context, uriString string) ([]byte, error) {
				return []byte{1, 2, 3, 4, 5, 6, 7, 8, 9}, nil
			},
		}
		h := &ContentPublished{ZebedeeCli: zebedeeMock}

		Convey("Then the zebedee handler fails with the expected error when a valid event is handled", func() {
			err := h.handleZebedeeType(ctx, &testZebedeeEvent)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "invalid character '\\x01' looking for beginning of value")
		})
	})

	Convey("Given a handler with a valid zebedee mock and a producer that fails to send a message", t, func() {
		zebedeeMock := &clientMock.ZebedeeClientMock{
			GetPublishedDataFunc: getPublishDataFunc,
		}
		producerMock := &kafkatest.IProducerMock{
			SendFunc: func(schema *avro.Schema, event interface{}) error {
				return errors.New("failed to send kafka message")
			},
		}
		h := &ContentPublished{
			ZebedeeCli: zebedeeMock,
			Producer:   producerMock,
			Cfg: &config.Config{
				KeywordsLimit: 10,
			},
		}

		Convey("Then the zebedee handler fails with the expected error when a valid event is handled", func() {
			err := h.handleZebedeeType(ctx, &testZebedeeEvent)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "failed to send search data import event: failed to send kafka message")
		})
	})

	Convey("Given a handler with a zebedee mock that returns a content item with no title", t, func() {
		mockZebedeePublishedResponse = `{"description":{"cdid": "testCDID","edition": "testEdition"},"type": "test", "URI": "test"}`
		getPublishDataFunc = func(ctx context.Context, uriString string) ([]byte, error) {
			data := []byte(mockZebedeePublishedResponse)
			return data, nil
		}

		zebedeeMock := &clientMock.ZebedeeClientMock{
			GetPublishedDataFunc: getPublishDataFunc,
		}

		producerMock := &kafkatest.IProducerMock{
			SendFunc: func(schema *avro.Schema, event interface{}) error {
				return nil
			},
		}

		h := &ContentPublished{
			ZebedeeCli: zebedeeMock,
			Producer:   producerMock,
			Cfg: &config.Config{
				KeywordsLimit: 10,
			},
		}

		Convey("Then the zebedee handler doesn't create an import event", func() {
			err := h.handleZebedeeType(ctx, &testZebedeeEvent)
			So(err, ShouldBeNil)
			So(len(producerMock.SendCalls()), ShouldEqual, 0)
		})
	})
}

func TestExtractDatasetURIFromEditionURI(t *testing.T) {
	t.Parallel()

	Convey("Given a valid edition uri", t, func() {
		editionURI := "/datasets/uk-economy/2016"
		expectedURI := "/datasets/uk-economy"
		Convey("When calling extractDatasetURI function", func() {
			datasetURI, err := extractDatasetURI(editionURI)

			Convey("Then successfully return a dataset uri and no errors", func() {
				So(err, ShouldBeNil)
				So(datasetURI, ShouldEqual, expectedURI)
			})
		})
		Convey("When calling retrieveCorrectURI function", func() {
			datasetURI, err := retrieveCorrectURI(editionURI)

			Convey("Then successfully return a dataset uri and no errors", func() {
				So(err, ShouldBeNil)
				So(datasetURI, ShouldEqual, expectedURI)
			})
		})
	})
}

func TestRetrieveCorrectURI(t *testing.T) {
	t.Parallel()

	Convey("Given a valid uri which does not contain \"datasets\"", t, func() {
		expectedURI := "/bulletins/uk-economy/2016"

		Convey("When calling retrieveCorrectURI function", func() {
			datasetURI, err := retrieveCorrectURI(expectedURI)

			Convey("Then successfully return the original uri and no errors", func() {
				So(err, ShouldBeNil)
				So(datasetURI, ShouldEqual, expectedURI)
			})
		})
	})
}
