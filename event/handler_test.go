package event_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/ONSdigital/dp-search-data-extractor/config"
	"github.com/ONSdigital/dp-search-data-extractor/event"
	eventMock "github.com/ONSdigital/dp-search-data-extractor/event/mock"
	. "github.com/smartystreets/goconvey/convey"
)

var errZebedee = errors.New("zebedee test error")
var getPublishDataFuncInError = func(ctx context.Context, uriString string) ([]byte, error) {
	return nil, errZebedee
}

var getPublishDataFunc = func(ctx context.Context, uriString string) ([]byte, error) {
	data := []byte("test data")
	return data, nil
}

func TestContentPublishedHandler_Handle(t *testing.T) {

	Convey("Given an event handler working successfully, and an event containing a URL", t, func() {
		var zebedeeMock = &eventMock.ZebedeeClientMock{GetPublishedDataFunc: getPublishDataFunc}
		eventHandler := &event.ContentPublishedHandler{zebedeeMock}
		filePath := "/tmp/dpSearchDataExtractor.txt"
		os.Remove(filePath)
		Convey("When given a valid event", func() {
			err := eventHandler.Handle(testCtx, &config.Config{OutputFilePath: filePath}, &testEvent)
			Convey("Then Zebedee is called 1 time with no error", func() {
				So(err, ShouldBeNil)
				So(zebedeeMock.GetPublishedDataCalls(), ShouldNotBeEmpty)
				So(zebedeeMock.GetPublishedDataCalls(), ShouldHaveLength, 1)
				So(zebedeeMock.GetPublishedDataCalls()[0].UriString, ShouldEqual, testEvent.URL)
			})
		})
	})

	Convey("Given an event handler not working successfully, and an event containing a URL", t, func() {
		var zebedeeMockInError = &eventMock.ZebedeeClientMock{GetPublishedDataFunc: getPublishDataFuncInError}
		eventHandler := &event.ContentPublishedHandler{zebedeeMockInError}
		filePath := ""
		Convey("When given a valid event", func() {
			err := eventHandler.Handle(testCtx, &config.Config{OutputFilePath: filePath}, &testEvent)
			Convey("Then Zebedee is called 1 time with the expected error ", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, errZebedee.Error())
				So(zebedeeMockInError.GetPublishedDataCalls(), ShouldNotBeEmpty)
				So(zebedeeMockInError.GetPublishedDataCalls(), ShouldHaveLength, 1)
				So(zebedeeMockInError.GetPublishedDataCalls()[0].UriString, ShouldEqual, testEvent.URL)
			})
		})
	})

	Convey("handler returns an error when cannot write to file", t, func() {
		var zebedeeMock = &eventMock.ZebedeeClientMock{GetPublishedDataFunc: getPublishDataFunc}
		eventHandler := &event.ContentPublishedHandler{zebedeeMock}
		filePath := ""
		Convey("When given a valid event ", func() {
			err := eventHandler.Handle(testCtx, &config.Config{OutputFilePath: filePath}, &testEvent)
			Convey("Then Zebedee is called 1 time with the expected error ", func() {
				So(err, ShouldNotBeNil)
				So(zebedeeMock.GetPublishedDataCalls(), ShouldNotBeEmpty)
				So(zebedeeMock.GetPublishedDataCalls(), ShouldHaveLength, 1)
				So(zebedeeMock.GetPublishedDataCalls()[0].UriString, ShouldEqual, testEvent.URL)
			})
		})
	})
}
