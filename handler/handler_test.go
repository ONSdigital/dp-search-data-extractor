package handler_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/dp-kafka/v2/kafkatest"
	clientMock "github.com/ONSdigital/dp-search-data-extractor/clients/mock"
	"github.com/ONSdigital/dp-search-data-extractor/config"
	"github.com/ONSdigital/dp-search-data-extractor/event"
	"github.com/ONSdigital/dp-search-data-extractor/event/mock"
	"github.com/ONSdigital/dp-search-data-extractor/handler"
	"github.com/ONSdigital/dp-search-data-extractor/models"
	"github.com/ONSdigital/dp-search-data-extractor/schema"
	. "github.com/smartystreets/goconvey/convey"
)

var testEvent = models.ContentPublished{
	URL:          "moo.com",
	DataType:     "Thing",
	CollectionID: "Col123",
}

var errorMock = errors.New("mock error")

var errZebedee = errors.New("zebedee test error")
var getPublishDataFuncInError = func(ctx context.Context, uriString string) ([]byte, error) {
	return nil, errZebedee
}

var contentPublishedTestData = `{"description":{"cdid": "testCDID","edition": "testedition"},"type": "testDataType"}`
var getPublishDataFunc = func(ctx context.Context, uriString string) ([]byte, error) {
	data := []byte(contentPublishedTestData)
	return data, nil
}

var pChannels = &kafka.ProducerChannels{
	Output: make(chan []byte, 1),
}
var getChannelFunc = func() *kafka.ProducerChannels {
	return pChannels
}

func TestContentPublishedHandler_Handle(t *testing.T) {

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

	kafkaProducerMock := &kafkatest.IProducerMock{
		ChannelsFunc: getChannelFunc,
	}

	marshallerMock := &mock.MarshallerMock{
		MarshalFunc: func(s interface{}) ([]byte, error) {
			return expectedSearchDataImport, nil
		},
	}

	producerMock := &event.SearchDataImportProducer{
		Producer:   kafkaProducerMock,
		Marshaller: marshallerMock,
	}

	Convey("Given an event handler working successfully, and an event containing a URL", t, func() {
		var zebedeeMock = &clientMock.ZebedeeClientMock{GetPublishedDataFunc: getPublishDataFunc}
		eventHandler := &handler.ContentPublishedHandler{zebedeeMock, *producerMock}

		Convey("When given a valid event", func() {
			err := eventHandler.Handle(context.Background(), &config.Config{}, &testEvent)

			var avroBytes []byte
			select {
			case avroBytes = <-pChannels.Output:
				fmt.Printf("avroBytes: %v\n", avroBytes)
			case <-time.After(time.Second * 5):
				t.FailNow()
			}

			Convey("Then no error is reported", func() {
				So(err, ShouldBeNil)
			})
			Convey("And then get content published from zebedee", func() {
				So(zebedeeMock.GetPublishedDataCalls(), ShouldNotBeEmpty)
				So(zebedeeMock.GetPublishedDataCalls(), ShouldHaveLength, 1)
				So(zebedeeMock.GetPublishedDataCalls()[0].UriString, ShouldEqual, testEvent.URL)
			})
			Convey("And then the expected bytes are sent to producer.output", func() {
				var actual models.SearchDataImport
				err = schema.SearchDataImportSchema.Unmarshal(avroBytes, &actual)
				So(err, ShouldBeNil)
				So(searchDataImportEvent1, ShouldResemble, actual)
			})
		})
	})
	Convey("Given an event handler not working successfully, and an event containing a URL", t, func() {
		var zebedeeMockInError = &clientMock.ZebedeeClientMock{GetPublishedDataFunc: getPublishDataFuncInError}
		eventHandler := &handler.ContentPublishedHandler{zebedeeMockInError, *producerMock}

		Convey("When given a valid event", func() {
			err := eventHandler.Handle(context.Background(), &config.Config{}, &testEvent)

			Convey("Then Zebedee is called 1 time with the expected error ", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, errZebedee.Error())
				So(zebedeeMockInError.GetPublishedDataCalls(), ShouldNotBeEmpty)
				So(zebedeeMockInError.GetPublishedDataCalls(), ShouldHaveLength, 1)
				So(zebedeeMockInError.GetPublishedDataCalls()[0].UriString, ShouldEqual, testEvent.URL)
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

// func TestContentPublishedHandler_Handle(t *testing.T) {

// 	var zebedeeMock = &clientMock.ZebedeeClientMock{GetPublishedDataFunc: getPublishDataFunc}
// 	kafkaProducerMock := &kafkatest.IProducerMock{
// 		ChannelsFunc: getChannelFunc,
// 	}

// 	marshallerMock := &mock.MarshallerMock{
// 		MarshalFunc: func(s interface{}) ([]byte, error) {
// 			return nil, nil
// 		},
// 	}
// 	producerMock := &event.SearchDataImportProducer{
// 		Producer:   kafkaProducerMock,
// 		Marshaller: marshallerMock,
// 	}

// 	Convey("Given a successful event handler, when Handle is triggered", t, func() {

// 		// eventHandler := &event.ContentPublishedHandler{zebedeeMock}
// 		eventHandler := &handler.ContentPublishedHandler{zebedeeMock, *producerMock}
// 		filePath := "/tmp/dpSearchDataExtractor.txt"
// 		os.Remove(filePath)
// 		err := eventHandler.Handle(context.Background(), &config.Config{OutputFilePath: filePath}, &testEvent)
// 		So(err, ShouldBeNil)
// 		So(zebedeeMock.GetPublishedDataCalls(), ShouldNotBeEmpty)
// 		So(zebedeeMock.GetPublishedDataCalls(), ShouldHaveLength, 1)
// 		So(zebedeeMock.GetPublishedDataCalls()[0].UriString, ShouldEqual, testEvent.URL)
// 	})

// 	Convey("Given a un-successful event handler, when Handle is triggered", t, func() {
// 		var zebedeeMockInError = &clientMock.ZebedeeClientMock{GetPublishedDataFunc: getPublishDataFuncInError}
// 		eventHandler := &handler.ContentPublishedHandler{zebedeeMockInError, *producerMock}
// 		filePath := ""
// 		err := eventHandler.Handle(context.Background(), &config.Config{OutputFilePath: filePath}, &testEvent)
// 		So(err, ShouldNotBeNil)
// 		So(err.Error(), ShouldEqual, errZebedee.Error())
// 		So(zebedeeMockInError.GetPublishedDataCalls(), ShouldNotBeEmpty)
// 		So(zebedeeMockInError.GetPublishedDataCalls(), ShouldHaveLength, 1)
// 		So(zebedeeMockInError.GetPublishedDataCalls()[0].UriString, ShouldEqual, testEvent.URL)
// 	})

// 	Convey("handler returns an error when cannot write to file", t, func() {
// 		var zebedeeMock = &clientMock.ZebedeeClientMock{GetPublishedDataFunc: getPublishDataFunc}

// 		eventHandler := &handler.ContentPublishedHandler{zebedeeMock, *producerMock}
// 		filePath := ""
// 		err := eventHandler.Handle(context.Background(), &config.Config{OutputFilePath: filePath}, &testEvent)
// 		So(err, ShouldNotBeNil)
// 		So(zebedeeMock.GetPublishedDataCalls(), ShouldNotBeEmpty)
// 		So(zebedeeMock.GetPublishedDataCalls(), ShouldHaveLength, 1)
// 		So(zebedeeMock.GetPublishedDataCalls()[0].UriString, ShouldEqual, testEvent.URL)
// 	})
// }
