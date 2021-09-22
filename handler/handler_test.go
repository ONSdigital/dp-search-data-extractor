package handler_test

import (
	"context"
	"errors"
	"os"
	"testing"

	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/dp-kafka/v2/kafkatest"
	clientMock "github.com/ONSdigital/dp-search-data-extractor/clients/mock"
	"github.com/ONSdigital/dp-search-data-extractor/config"
	"github.com/ONSdigital/dp-search-data-extractor/event"
	"github.com/ONSdigital/dp-search-data-extractor/event/mock"
	"github.com/ONSdigital/dp-search-data-extractor/handler"
	"github.com/ONSdigital/dp-search-data-extractor/models"
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

var contentPublishedTest = `{"cdid": "testCDID","data_type": "testDataType","description": "testDescription"}`
var getPublishDataFunc = func(ctx context.Context, uriString string) ([]byte, error) {
	data := []byte(contentPublishedTest)
	return data, nil
}

var pChannels = &kafka.ProducerChannels{
	Output: make(chan []byte, 1),
}
var getChannelFunc = func() *kafka.ProducerChannels {
	return pChannels
}

func TestContentPublishedHandler_Handle(t *testing.T) {

	var zebedeeMock = &clientMock.ZebedeeClientMock{GetPublishedDataFunc: getPublishDataFunc}
	kafkaProducerMock := &kafkatest.IProducerMock{
		ChannelsFunc: getChannelFunc,
	}

	marshallerMock := &mock.MarshallerMock{
		MarshalFunc: func(s interface{}) ([]byte, error) {
			return nil, errorMock
		},
	}
	producerMock := &event.SearchDataImportProducer{
		Producer:   kafkaProducerMock,
		Marshaller: marshallerMock,
	}

	Convey("Given a successful event handler, when Handle is triggered", t, func() {

		// eventHandler := &event.ContentPublishedHandler{zebedeeMock}
		eventHandler := &handler.ContentPublishedHandler{zebedeeMock, *producerMock}
		filePath := "/tmp/dpSearchDataExtractor.txt"
		os.Remove(filePath)
		err := eventHandler.Handle(context.Background(), &config.Config{OutputFilePath: filePath}, &testEvent)
		So(err, ShouldBeNil)
		So(zebedeeMock.GetPublishedDataCalls(), ShouldNotBeEmpty)
		So(zebedeeMock.GetPublishedDataCalls(), ShouldHaveLength, 1)
		So(zebedeeMock.GetPublishedDataCalls()[0].UriString, ShouldEqual, testEvent.URL)
	})

	Convey("Given a un-successful event handler, when Handle is triggered", t, func() {
		var zebedeeMockInError = &clientMock.ZebedeeClientMock{GetPublishedDataFunc: getPublishDataFuncInError}
		eventHandler := &handler.ContentPublishedHandler{zebedeeMockInError, *producerMock}
		filePath := ""
		err := eventHandler.Handle(context.Background(), &config.Config{OutputFilePath: filePath}, &testEvent)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, errZebedee.Error())
		So(zebedeeMockInError.GetPublishedDataCalls(), ShouldNotBeEmpty)
		So(zebedeeMockInError.GetPublishedDataCalls(), ShouldHaveLength, 1)
		So(zebedeeMockInError.GetPublishedDataCalls()[0].UriString, ShouldEqual, testEvent.URL)
	})

	Convey("handler returns an error when cannot write to file", t, func() {
		var zebedeeMock = &clientMock.ZebedeeClientMock{GetPublishedDataFunc: getPublishDataFunc}

		eventHandler := &handler.ContentPublishedHandler{zebedeeMock, *producerMock}
		filePath := ""
		err := eventHandler.Handle(context.Background(), &config.Config{OutputFilePath: filePath}, &testEvent)
		So(err, ShouldNotBeNil)
		So(zebedeeMock.GetPublishedDataCalls(), ShouldNotBeEmpty)
		So(zebedeeMock.GetPublishedDataCalls(), ShouldHaveLength, 1)
		So(zebedeeMock.GetPublishedDataCalls()[0].UriString, ShouldEqual, testEvent.URL)
	})
}

// func TestSearchDataImportEventHandler_Handle(t *testing.T) {

// 	Convey("Given the handler has been configured", t, func() {
// 		// Set up mocks
// 		var zebedeeMock = &eventMock.ZebedeeClientMock{GetPublishedDataFunc: getPublishDataFunc}
// 		eventHandler := &event.ContentPublishedHandler{zebedeeMock}
// 		filePath := "/tmp/dpSearchDataExtractor.txt"
// 		os.Remove(filePath)

// 		Convey("When given a valid event", func() {
// 			err := eventHandler.Handle(testCtx,
// 				&config.Config{OutputFilePath: filePath},
// 				&testEvent)

// 			Convey("Then Producer is called 1 time with the expected parameters", func() {
// 				So(err, ShouldBeNil)
// 				So(zebedeeMock.GetPublishedDataCalls(), ShouldNotBeEmpty)
// 				So(zebedeeMock.GetPublishedDataCalls(), ShouldHaveLength, 1)
// 				So(zebedeeMock.GetPublishedDataCalls()[0].UriString, ShouldEqual, testEvent.URL)
// 				So(json.Unmarshal(contentPublishedTest, event.ZebedeeData), ShouldNotBeEmpty)
// 			})
// 		})
// 	})
// }
