package handler_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/dp-kafka/v2/kafkatest"
	clientMock "github.com/ONSdigital/dp-search-data-extractor/clients/mock"
	"github.com/ONSdigital/dp-search-data-extractor/event"
	"github.com/ONSdigital/dp-search-data-extractor/event/mock"
	"github.com/ONSdigital/dp-search-data-extractor/handler"
	"github.com/ONSdigital/dp-search-data-extractor/models"
	"github.com/ONSdigital/dp-search-data-extractor/schema"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	testTimeout = time.Second * 5

	testZebedeeEvent = models.ContentPublished{
		URI:          "testZebdeeUri",
		DataType:     "Reviewed-uris",
		CollectionID: "testZebdeeCollectionID",
	}

	errZebedee                = errors.New("zebedee test error")
	getPublishDataFuncInError = func(ctx context.Context, uriString string) ([]byte, error) {
		return nil, errZebedee
	}

	mockZebedeePublishedResponse = `{"description":{"cdid": "testCDID","edition": "testedition"},"type": "testDataType"}`
	getPublishDataFunc           = func(ctx context.Context, uriString string) ([]byte, error) {
		data := []byte(mockZebedeePublishedResponse)
		return data, nil
	}

	mockedEditionResponse = dataset.Edition{
		Edition: "edition-1",
		ID:      "id",
		// Links: "links",
		State: "State",
	}
	getDatasetEditionFunc = func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, datasetID string, editionID string) (dataset.Edition, error) {
		data := mockedEditionResponse
		return data, nil
	}

	errDatasetAPI                = errors.New("dataset api test error")
	getDatasetEditionFuncInError = func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, datasetID string, editionID string) (dataset.Edition, error) {
		return dataset.Edition{}, errDatasetAPI
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
		var datasetMock = &clientMock.DatasetClientMock{GetEditionFunc: getDatasetEditionFunc}
		eventHandler := &handler.ContentPublishedHandler{zebedeeMock, datasetMock, *producerMock}

		Convey("When given a valid event", func() {
			err := eventHandler.Handle(context.Background(), &testZebedeeEvent, -1)

			var avroBytes []byte
			select {
			case avroBytes = <-pChannels.Output:
				t.Log("avro byte sent to producer output")
			case <-time.After(testTimeout):
				t.FailNow()
			}

			Convey("Then no error is reported", func() {
				So(err, ShouldBeNil)
			})
			Convey("And then get content published from zebedee", func() {
				So(zebedeeMock.GetPublishedDataCalls(), ShouldNotBeEmpty)
				So(zebedeeMock.GetPublishedDataCalls(), ShouldHaveLength, 1)
				So(zebedeeMock.GetPublishedDataCalls()[0].UriString, ShouldEqual, testZebedeeEvent.URI)
				So(len(datasetMock.GetEditionCalls()), ShouldEqual, 0)
			})
			Convey("And then the expected bytes are sent to producer.output", func() {
				var actual models.SearchDataImport
				err = schema.SearchDataImportEvent.Unmarshal(avroBytes, &actual)
				So(err, ShouldBeNil)
				So(actual.DataType, ShouldEqual, expectedSearchDataImportEvent.DataType)
				So(actual.JobID, ShouldEqual, expectedSearchDataImportEvent.JobID)
				So(actual.Keywords, ShouldHaveLength, 2)
				So(actual.Keywords[0], ShouldEqual, expectedSearchDataImportEvent.Keywords[0])
				So(actual.Keywords[1], ShouldEqual, expectedSearchDataImportEvent.Keywords[1])
				So(actual.MetaDescription, ShouldEqual, expectedSearchDataImportEvent.MetaDescription)
				So(actual.Summary, ShouldEqual, expectedSearchDataImportEvent.Summary)
				So(actual.ReleaseDate, ShouldEqual, expectedSearchDataImportEvent.ReleaseDate)
				So(actual.Title, ShouldEqual, expectedSearchDataImportEvent.Title)
				So(actual.TraceID, ShouldNotBeNil)
			})
		})
	})
	Convey("Given an event handler not working successfully, and an event containing a URI", t, func() {
		var zebedeeMockInError = &clientMock.ZebedeeClientMock{GetPublishedDataFunc: getPublishDataFuncInError}
		var datasetMockInError = &clientMock.DatasetClientMock{GetEditionFunc: getDatasetEditionFuncInError}
		eventHandler := &handler.ContentPublishedHandler{zebedeeMockInError, datasetMockInError, *producerMock}

		Convey("When given a valid event", func() {
			err := eventHandler.Handle(context.Background(), &testZebedeeEvent, 1)

			Convey("Then Zebedee is called 1 time with the expected error ", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, errZebedee.Error())
				So(zebedeeMockInError.GetPublishedDataCalls(), ShouldNotBeEmpty)
				So(zebedeeMockInError.GetPublishedDataCalls(), ShouldHaveLength, 1)
				So(zebedeeMockInError.GetPublishedDataCalls()[0].UriString, ShouldEqual, testZebedeeEvent.URI)
				So(len(datasetMockInError.GetEditionCalls()), ShouldEqual, 0)
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
		var datasetMock = &clientMock.DatasetClientMock{GetEditionFunc: getDatasetEditionFunc}
		eventHandler := &handler.ContentPublishedHandler{zebedeeMock, datasetMock, *producerMock}

		Convey("When given a valid event with default keywords limit", func() {

			err := eventHandler.Handle(context.Background(), &testZebedeeEvent, -1)

			var avroBytes []byte
			select {
			case avroBytes = <-pChannels.Output:
				t.Log("avro byte sent to producer output")
			case <-time.After(testTimeout):
				t.FailNow()
			}

			Convey("Then no error is reported", func() {
				So(err, ShouldBeNil)
			})
			Convey("And then get content published from zebedee", func() {
				So(zebedeeMock.GetPublishedDataCalls(), ShouldNotBeEmpty)
				So(zebedeeMock.GetPublishedDataCalls(), ShouldHaveLength, 1)
				So(zebedeeMock.GetPublishedDataCalls()[0].UriString, ShouldEqual, testZebedeeEvent.URI)
				So(len(datasetMock.GetEditionCalls()), ShouldEqual, 0)
			})
			Convey("And then the expected bytes are sent to producer.output", func() {
				var actual models.SearchDataImport
				err = schema.SearchDataImportEvent.Unmarshal(avroBytes, &actual)
				So(err, ShouldBeNil)
				So(actual.DataType, ShouldEqual, expectedFullSearchDataImportEvent.DataType)
				So(actual.JobID, ShouldEqual, expectedFullSearchDataImportEvent.JobID)
				So(actual.Keywords, ShouldHaveLength, 1)
				So(actual.Keywords[0], ShouldEqual, expectedFullSearchDataImportEvent.Keywords[0])
				So(actual.SearchIndex, ShouldEqual, expectedFullSearchDataImportEvent.SearchIndex)
				So(actual.CDID, ShouldEqual, expectedFullSearchDataImportEvent.CDID)
				So(actual.DatasetID, ShouldEqual, expectedFullSearchDataImportEvent.DatasetID)
				So(actual.MetaDescription, ShouldEqual, expectedFullSearchDataImportEvent.MetaDescription)
				So(actual.Summary, ShouldEqual, expectedFullSearchDataImportEvent.Summary)
				So(actual.ReleaseDate, ShouldEqual, expectedFullSearchDataImportEvent.ReleaseDate)
				So(actual.Title, ShouldEqual, expectedFullSearchDataImportEvent.Title)
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
