package handler_test

import (
	"context"
	"errors"
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

var (
	ctx = context.Background()

	testTimeout = time.Second * 5

	testEvent = models.ContentPublished{
		URI:          "testUri",
		DataType:     "Thing",
		CollectionID: "Col123",
	}

	errZebedee                = errors.New("zebedee test error")
	getPublishDataFuncInError = func(ctx context.Context, uriString string) ([]byte, error) {
		return nil, errZebedee
	}

	contentPublishedTestData = `{"description":{"cdid": "testCDID","edition": "testedition"},"type": "testDataType"}`
	getPublishDataFunc       = func(ctx context.Context, uriString string) ([]byte, error) {
		data := []byte(contentPublishedTestData)
		return data, nil
	}

	pChannels = &kafka.ProducerChannels{
		Output: make(chan []byte, 1),
	}
	getChannelFunc = func() *kafka.ProducerChannels {
		return pChannels
	}
)

func TestHandlerForZebedeeReturningMandatoryFields(t *testing.T) {

	expectedSearchDataImportEvent := models.SearchDataImport{
		DataType:        "testDataType",
		JobID:           "",
		SearchIndex:     "ONS",
		CDID:            "",
		DatasetID:       "",
		Keywords:        []string{"testkeyword1", "testkeyword2"},
		MetaDescription: "",
		Summary:         "",
		ReleaseDate:     "",
		Title:           "testTitle",
	}

	kafkaProducerMock := &kafkatest.IProducerMock{
		ChannelsFunc: getChannelFunc,
	}

	//for mock marshaller
	expectedSearchDataImport := marshalSearchDataImport(t, expectedSearchDataImportEvent)
	marshallerMock := &mock.MarshallerMock{
		MarshalFunc: func(s interface{}) ([]byte, error) {
			return expectedSearchDataImport, nil
		},
	}

	producerMock := &event.SearchDataImportProducer{
		Producer:   kafkaProducerMock,
		Marshaller: marshallerMock,
	}

	Convey("Given an event handler working successfully, and an event containing a URI", t, func() {
		var zebedeeMock = &clientMock.ZebedeeClientMock{GetPublishedDataFunc: getPublishDataFunc}
		eventHandler := &handler.ContentPublishedHandler{zebedeeMock, *producerMock}

		Convey("When given a valid event", func() {
			err := eventHandler.Handle(context.Background(), &config.Config{}, &testEvent)

			var avroBytes []byte
			select {
			case avroBytes = <-pChannels.Output:
				t.Log("avro byte sent to producer output")
			case <-time.After(testTimeout):
				t.Fatalf("failing test due to timing out after %v seconds", testTimeout)
				t.FailNow()
			}

			Convey("Then no error is reported", func() {
				So(err, ShouldBeNil)
			})
			Convey("And then get content published from zebedee", func() {
				So(zebedeeMock.GetPublishedDataCalls(), ShouldNotBeEmpty)
				So(zebedeeMock.GetPublishedDataCalls(), ShouldHaveLength, 1)
				So(zebedeeMock.GetPublishedDataCalls()[0].UriString, ShouldEqual, testEvent.URI)
			})
			Convey("And then the expected bytes are sent to producer.output", func() {
				var actual models.SearchDataImport
				err = schema.SearchDataImportEvent.Unmarshal(avroBytes, &actual)
				So(err, ShouldBeNil)
				So(expectedSearchDataImportEvent.DataType, ShouldEqual, actual.DataType)
				So(expectedSearchDataImportEvent.JobID, ShouldEqual, actual.JobID)
				So(expectedSearchDataImportEvent.Keywords[0], ShouldEqual, actual.Keywords[0])
				So(expectedSearchDataImportEvent.Keywords[1], ShouldEqual, actual.Keywords[1])
				So(expectedSearchDataImportEvent.Keywords, ShouldHaveLength, 2)
				So(actual.TraceID, ShouldNotBeNil)
				So(expectedSearchDataImportEvent.MetaDescription, ShouldEqual, actual.MetaDescription)
				So(expectedSearchDataImportEvent.Summary, ShouldEqual, actual.Summary)
				So(expectedSearchDataImportEvent.ReleaseDate, ShouldEqual, actual.ReleaseDate)
			})
		})
	})
	Convey("Given an event handler not working successfully, and an event containing a URI", t, func() {
		var zebedeeMockInError = &clientMock.ZebedeeClientMock{GetPublishedDataFunc: getPublishDataFuncInError}
		eventHandler := &handler.ContentPublishedHandler{zebedeeMockInError, *producerMock}

		Convey("When given a valid event", func() {
			err := eventHandler.Handle(context.Background(), &config.Config{}, &testEvent)

			Convey("Then Zebedee is called 1 time with the expected error ", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, errZebedee.Error())
				So(zebedeeMockInError.GetPublishedDataCalls(), ShouldNotBeEmpty)
				So(zebedeeMockInError.GetPublishedDataCalls(), ShouldHaveLength, 1)
				So(zebedeeMockInError.GetPublishedDataCalls()[0].UriString, ShouldEqual, testEvent.URI)
			})
		})
	})
}

func TestHandlerForZebedeeReturningAllFields(t *testing.T) {

	expectedFullSearchDataImportEvent := models.SearchDataImport{
		DataType:        "testDataType",
		JobID:           "",
		SearchIndex:     "ONS",
		CDID:            "testCDID",
		DatasetID:       "testDaetasetId",
		Keywords:        []string{"testkeyword"},
		MetaDescription: "testMetaDescription",
		Summary:         "testSummary",
		ReleaseDate:     "testReleaseDate",
		Title:           "testTitle",
	}

	kafkaProducerMock := &kafkatest.IProducerMock{
		ChannelsFunc: getChannelFunc,
	}

	marshallerMock := &mock.MarshallerMock{
		MarshalFunc: func(s interface{}) ([]byte, error) {
			return schema.SearchDataImportEvent.Marshal(s)
		},
	}

	producerMock := &event.SearchDataImportProducer{
		Producer:   kafkaProducerMock,
		Marshaller: marshallerMock,
	}

	Convey("Given an event handler working successfully, and an event containing a URI", t, func() {

		//used by zebedee mock
		fullContentPublishedTestData := `{"description":{"cdid": "testCDID","datasetId": "testDaetasetId","edition": "testedition","keywords": ["testkeyword"],"metaDescription": "testMetaDescription","releaseDate": "testReleaseDate","summary": "testSummary","title": "testTitle"},"type": "testDataType"}`
		getFullPublishDataFunc := func(ctx context.Context, uriString string) ([]byte, error) {
			data := []byte(fullContentPublishedTestData)
			return data, nil
		}

		var zebedeeMock = &clientMock.ZebedeeClientMock{GetPublishedDataFunc: getFullPublishDataFunc}
		eventHandler := &handler.ContentPublishedHandler{zebedeeMock, *producerMock}

		Convey("When given a valid event", func() {

			err := eventHandler.Handle(context.Background(), &config.Config{}, &testEvent)

			var avroBytes []byte
			select {
			case avroBytes = <-pChannels.Output:
				t.Log("avro byte sent to producer output")
			case <-time.After(testTimeout):
				t.Fatalf("failing test due to timing out after %v seconds", testTimeout)
				t.FailNow()
			}

			Convey("Then no error is reported", func() {
				So(err, ShouldBeNil)
			})
			Convey("And then get content published from zebedee", func() {
				So(zebedeeMock.GetPublishedDataCalls(), ShouldNotBeEmpty)
				So(zebedeeMock.GetPublishedDataCalls(), ShouldHaveLength, 1)
				So(zebedeeMock.GetPublishedDataCalls()[0].UriString, ShouldEqual, testEvent.URI)
			})
			Convey("And then the expected bytes are sent to producer.output", func() {
				var actual models.SearchDataImport
				err = schema.SearchDataImportEvent.Unmarshal(avroBytes, &actual)
				So(err, ShouldBeNil)
				So(expectedFullSearchDataImportEvent.DataType, ShouldEqual, actual.DataType)
				So(expectedFullSearchDataImportEvent.JobID, ShouldEqual, actual.JobID)
				So(expectedFullSearchDataImportEvent.Keywords[0], ShouldResemble, actual.Keywords[0])
				So(expectedFullSearchDataImportEvent.Keywords, ShouldHaveLength, 1)
				So(actual.TraceID, ShouldNotBeNil)
			})
		})
	})
}

// marshalSearchDataImport helper method to marshal a event into a []byte
func marshalSearchDataImport(t *testing.T, event models.SearchDataImport) []byte {
	bytes, err := schema.SearchDataImportEvent.Marshal(event)
	if err != nil {
		t.Fatalf("avro mashalling failed with error : %v", err)
	}
	return bytes
}
