package handler

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-kafka/v2/kafkatest"
	"github.com/ONSdigital/dp-search-data-extractor/config"
	"github.com/ONSdigital/dp-search-data-extractor/event"
	"github.com/ONSdigital/dp-search-data-extractor/event/mock"
	"github.com/ONSdigital/dp-search-data-extractor/models"
	"github.com/ONSdigital/dp-search-data-extractor/schema"

	dpkafka "github.com/ONSdigital/dp-kafka/v2"
	clientMock "github.com/ONSdigital/dp-search-data-extractor/clients/mock"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	testTimeout = time.Second * 5
	ctx         = context.Background()

	testZebedeeEvent = models.ContentPublished{
		URI:          "testZebdeeUri",
		DataType:     "legacy",
		CollectionID: "testZebdeeCollectionID",
	}

	testZebedeeEventWithDatasetURI = models.ContentPublished{
		URI:          "/datasets/cphi01/testedition2018",
		DataType:     "legacy",
		CollectionID: "testZebdeeCollectionID",
	}

	testDatasetEvent = models.ContentPublished{
		URI:          "/datasets/cphi01/editions/timeseries/versions/version/metadata",
		DataType:     "datasets",
		CollectionID: "testDatasetApiCollectionID",
	}

	testInvalidEvent = models.ContentPublished{
		URI:          "/datasets/invalidEvent/metadata",
		DataType:     "Unknown-uris",
		CollectionID: "invalidDatasetApiCollectionID",
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

	mockDatasetAPIJSONResponse = setupMetadata()
	getVersionMetadataFunc     = func(ctx context.Context, userAuthToken, serviceAuthToken, collectionId, datasetId, edition, version string) (dataset.Metadata, error) {
		data := mockDatasetAPIJSONResponse
		return data, nil
	}

	errDatasetAPI                 = errors.New("dataset api version metadata test error")
	getVersionMetadataFuncInError = func(ctx context.Context, userAuthToken, serviceAuthToken, collectionId, datasetId, edition, version string) (dataset.Metadata, error) {
		return dataset.Metadata{}, errDatasetAPI
	}

	pChannels = &dpkafka.ProducerChannels{
		Output: make(chan []byte, 1),
	}
	getChannelFunc = func() *dpkafka.ProducerChannels {
		return pChannels
	}

	cfg, _ = config.Get()
)

func TestHandlerForZebedeeReturningMandatoryFields(t *testing.T) {
	t.Parallel()
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
		Topics:          []string{"testtopic1", "testtopic2"},
	}

	kafkaProducerMock := &kafkatest.IProducerMock{
		ChannelsFunc: getChannelFunc,
	}

	// for mock marshaller
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
		var datasetMock = &clientMock.DatasetClientMock{GetVersionMetadataFunc: getVersionMetadataFunc}

		eventHandler := &ContentPublishedHandler{zebedeeMock, datasetMock, *producerMock}

		Convey("When given a valid event", func() {
			err := eventHandler.Handle(ctx, &testZebedeeEvent, *cfg)
			Convey("Then no error is reported", func() {
				So(err, ShouldBeNil)
			})

			Convey("And then get content published from zebedee", func() {
				So(zebedeeMock.GetPublishedDataCalls(), ShouldNotBeEmpty)
				So(zebedeeMock.GetPublishedDataCalls(), ShouldHaveLength, 1)
				So(zebedeeMock.GetPublishedDataCalls()[0].UriString, ShouldEqual, testZebedeeEvent.URI)

				So(len(datasetMock.GetVersionMetadataCalls()), ShouldEqual, 0)
			})

			var avroBytes []byte
			delay := time.NewTimer(testTimeout)
			select {
			case avroBytes = <-pChannels.Output:
				// Ensure timer is stopped and its resources are freed
				if !delay.Stop() {
					// if the timer has been stopped then read from the channel
					<-delay.C
				}
				t.Log("avro byte sent to producer output")
			case <-delay.C:
				t.FailNow()
			}
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

		Convey("When given a valid event with dataset uri", func() {
			expectedParsedURI := strings.Split(testZebedeeEventWithDatasetURI.URI, "/")
			expectedParsedURI = expectedParsedURI[:len(expectedParsedURI)-1]
			err := eventHandler.Handle(ctx, &testZebedeeEventWithDatasetURI, *cfg)
			Convey("Then no error is reported", func() {
				So(err, ShouldBeNil)
			})

			Convey("And then get content published from zebedee", func() {
				So(zebedeeMock.GetPublishedDataCalls(), ShouldNotBeEmpty)
				So(zebedeeMock.GetPublishedDataCalls(), ShouldHaveLength, 1)
				So(zebedeeMock.GetPublishedDataCalls()[0].UriString, ShouldEqual, strings.Join(expectedParsedURI, "/"))

				So(len(datasetMock.GetVersionMetadataCalls()), ShouldEqual, 0)
			})
			delay := time.NewTimer(testTimeout)
			select {
			case <-pChannels.Output:
				// Ensure timer is stopped and its resources are freed
				if !delay.Stop() {
					// if the timer has been stopped then read from the channel
					<-delay.C
				}
				t.Log("avro byte sent to producer output")
			case <-delay.C:
				t.FailNow()
			}
		})
	})
	Convey("Given an event handler not working successfully, and an event containing a URI", t, func() {
		var zebedeeMockInError = &clientMock.ZebedeeClientMock{GetPublishedDataFunc: getPublishDataFuncInError}
		var datasetMockInError = &clientMock.DatasetClientMock{GetVersionMetadataFunc: getVersionMetadataFuncInError}
		eventHandler := &ContentPublishedHandler{zebedeeMockInError, datasetMockInError, *producerMock}

		Convey("When given a valid event", func() {
			err := eventHandler.Handle(ctx, &testZebedeeEvent, *cfg)

			Convey("Then Zebedee is called 1 time with the expected error ", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, errZebedee.Error())
				So(zebedeeMockInError.GetPublishedDataCalls(), ShouldNotBeEmpty)
				So(zebedeeMockInError.GetPublishedDataCalls(), ShouldHaveLength, 1)
				So(zebedeeMockInError.GetPublishedDataCalls()[0].UriString, ShouldEqual, testZebedeeEvent.URI)

				So(len(datasetMockInError.GetVersionMetadataCalls()), ShouldEqual, 0)
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
		Topics:          []string{"testtopic"},
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
		// used by zebedee mock
		fullContentPublishedTestData := `{"description":{"cdid": "testCDID","datasetId": "testDaetasetId","edition": "testedition","keywords": ["testkeyword"],"metaDescription": "testMetaDescription","releaseDate": "testReleaseDate","summary": "testSummary","title": "testTitle"},"type": "testDataType"}`
		getFullPublishDataFunc := func(ctx context.Context, uriString string) ([]byte, error) {
			data := []byte(fullContentPublishedTestData)
			return data, nil
		}

		var zebedeeMock = &clientMock.ZebedeeClientMock{GetPublishedDataFunc: getFullPublishDataFunc}
		var datasetMock = &clientMock.DatasetClientMock{GetVersionMetadataFunc: getVersionMetadataFunc}
		eventHandler := &ContentPublishedHandler{zebedeeMock, datasetMock, *producerMock}

		Convey("When given a valid event with default keywords limit", func() {
			err := eventHandler.Handle(ctx, &testZebedeeEvent, *cfg)

			var avroBytes []byte
			delay := time.NewTimer(testTimeout)
			select {
			case avroBytes = <-pChannels.Output:
				// Ensure timer is stopped and its resources are freed
				if !delay.Stop() {
					// if the timer has been stopped then read from the channel
					<-delay.C
				}
				t.Log("avro byte sent to producer output")
			case <-delay.C:
				t.FailNow()
			}
			Convey("Then no error is reported", func() {
				So(err, ShouldBeNil)
			})
			Convey("And then get content published from zebedee", func() {
				So(zebedeeMock.GetPublishedDataCalls(), ShouldNotBeEmpty)
				So(zebedeeMock.GetPublishedDataCalls(), ShouldHaveLength, 1)
				So(zebedeeMock.GetPublishedDataCalls()[0].UriString, ShouldEqual, testZebedeeEvent.URI)

				So(len(datasetMock.GetVersionMetadataCalls()), ShouldEqual, 0)
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

func TestHandlerForDatasetVersionMetadata(t *testing.T) {
	t.Parallel()
	expectedVersionMetadataEvent := models.SearchDataImport{
		UID:             "cphi01-timeseries",
		DataType:        "dataset_landing_page",
		JobID:           "",
		SearchIndex:     "ONS",
		CDID:            "",
		DatasetID:       "",
		Keywords:        []string{"somekeyword0", "somekeyword1"},
		MetaDescription: "someDescription",
		Summary:         "",
		ReleaseDate:     "2020-11-07T00:00:00.000Z",
		Title:           "someTitle",
		Topics:          []string{"testtopic1", "testtopic2"},
	}

	kafkaProducerMock := &kafkatest.IProducerMock{
		ChannelsFunc: getChannelFunc,
	}

	// for mock marshaller
	expectedVersionMetadataSearchDataImport := marshalSearchDataImport(t, expectedVersionMetadataEvent)
	marshallerMock := &mock.MarshallerMock{
		MarshalFunc: func(s interface{}) ([]byte, error) {
			return expectedVersionMetadataSearchDataImport, nil
		},
	}

	producerMock := &event.SearchDataImportProducer{
		Producer:   kafkaProducerMock,
		Marshaller: marshallerMock,
	}

	Convey("Given an event handler working successfully, and an event containing a valid dataType", t, func() {
		var zebedeeMock = &clientMock.ZebedeeClientMock{GetPublishedDataFunc: getPublishDataFunc}
		var datasetMock = &clientMock.DatasetClientMock{GetVersionMetadataFunc: getVersionMetadataFunc}

		eventHandler := &ContentPublishedHandler{zebedeeMock, datasetMock, *producerMock}

		Convey("When given a valid event for cmd dataset", func() {
			err := eventHandler.Handle(ctx, &testDatasetEvent, *cfg)

			var avroBytes []byte
			delay := time.NewTimer(testTimeout)
			select {
			case avroBytes = <-pChannels.Output:
				// Ensure timer is stopped and its resources are freed
				if !delay.Stop() {
					// if the timer has been stopped then read from the channel
					<-delay.C
				}
				t.Log("avro byte sent to producer output")
			case <-delay.C:
				t.FailNow()
			}

			Convey("Then no error is reported", func() {
				So(err, ShouldBeNil)
			})
			Convey("And then get content published from datasetAPI but no action for zebedee", func() {
				So(len(zebedeeMock.GetPublishedDataCalls()), ShouldEqual, 0)
				So(len(datasetMock.GetVersionMetadataCalls()), ShouldEqual, 1)
			})
			Convey("And then the expected bytes are sent to producer.output", func() {
				var actual models.SearchDataImport
				err = schema.SearchDataImportEvent.Unmarshal(avroBytes, &actual)
				So(err, ShouldBeNil)
				So(actual.UID, ShouldEqual, expectedVersionMetadataEvent.UID)
				So(actual.ReleaseDate, ShouldEqual, expectedVersionMetadataEvent.ReleaseDate)
				So(actual.Title, ShouldEqual, expectedVersionMetadataEvent.Title)
				So(actual.MetaDescription, ShouldEqual, expectedVersionMetadataEvent.MetaDescription)
				So(actual.Keywords, ShouldHaveLength, 2)
				So(actual.Keywords[0], ShouldEqual, "somekeyword0")
				So(actual.Keywords[1], ShouldEqual, "somekeyword1")
				So(actual.DataType, ShouldResemble, expectedVersionMetadataEvent.DataType)
			})
		})
	})
	Convey("Given an event handler not working successfully, and an event containing a valid dataType", t, func() {
		var zebedeeMockInError = &clientMock.ZebedeeClientMock{GetPublishedDataFunc: getPublishDataFuncInError}
		var datasetMockInError = &clientMock.DatasetClientMock{GetVersionMetadataFunc: getVersionMetadataFuncInError}
		eventHandler := &ContentPublishedHandler{zebedeeMockInError, datasetMockInError, *producerMock}

		Convey("When given a valid event for a valid dataset", func() {
			err := eventHandler.Handle(ctx, &testDatasetEvent, *cfg)

			Convey("Then Zebedee is called 0 time and Dataset called 1 time with the expected error ", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, errDatasetAPI.Error())
				So(zebedeeMockInError.GetPublishedDataCalls(), ShouldHaveLength, 0)

				So(len(datasetMockInError.GetVersionMetadataCalls()), ShouldEqual, 1)
				So(datasetMockInError.GetVersionMetadataCalls(), ShouldNotBeEmpty)
				So(datasetMockInError.GetVersionMetadataCalls(), ShouldHaveLength, 1)
			})
		})
	})
}

func TestHandlerForInvalidDataType(t *testing.T) {
	t.Parallel()
	expectedInvalidSearchDataImportEvent := models.SearchDataImport{}

	kafkaProducerMock := &kafkatest.IProducerMock{
		ChannelsFunc: getChannelFunc,
	}

	// for mock marshaller
	expectedVersionMetadataSearchDataImport := marshalSearchDataImport(t, expectedInvalidSearchDataImportEvent)
	marshallerMock := &mock.MarshallerMock{
		MarshalFunc: func(s interface{}) ([]byte, error) {
			return expectedVersionMetadataSearchDataImport, nil
		},
	}

	producerMock := &event.SearchDataImportProducer{
		Producer:   kafkaProducerMock,
		Marshaller: marshallerMock,
	}

	Convey("Given an event handler working successfully, and an event containing a invalid dataType", t, func() {
		var zebedeeMockInError = &clientMock.ZebedeeClientMock{GetPublishedDataFunc: getPublishDataFuncInError}
		var datasetMockInError = &clientMock.DatasetClientMock{GetVersionMetadataFunc: getVersionMetadataFuncInError}
		eventHandler := &ContentPublishedHandler{zebedeeMockInError, datasetMockInError, *producerMock}

		Convey("When given a invalid event", func() {
			err := eventHandler.Handle(ctx, &testInvalidEvent, *cfg)

			Convey("Then both Zebedee and Dataset is called 0 time with no expected error ", func() {
				So(err, ShouldBeNil)

				So(zebedeeMockInError.GetPublishedDataCalls(), ShouldHaveLength, 0)
				So(zebedeeMockInError.GetPublishedDataCalls(), ShouldBeEmpty)

				So(datasetMockInError.GetVersionMetadataCalls(), ShouldHaveLength, 0)
				So(datasetMockInError.GetVersionMetadataCalls(), ShouldBeEmpty)
			})
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
	})
}

func TestRetrieveCorrectURI(t *testing.T) {
	t.Parallel()

	Convey("Given a valid edition uri", t, func() {
		editionURI := "/datasets/uk-economy/2016"
		expectedURI := "/datasets/uk-economy"

		Convey("When calling retrieveCorrectURI function", func() {
			datasetURI, err := retrieveCorrectURI(editionURI)

			Convey("Then successfully return a dataset uri and no errors", func() {
				So(err, ShouldBeNil)
				So(datasetURI, ShouldEqual, expectedURI)
			})
		})
	})

	Convey("Given a valid uri which does not contain \"datasets\"", t, func() {
		bulletinURI := "/bulletins/uk-economy/2016"
		expectedURI := "/bulletins/uk-economy/2016"

		Convey("When calling retrieveCorrectURI function", func() {
			datasetURI, err := retrieveCorrectURI(bulletinURI)

			Convey("Then successfully return the original uri and no errors", func() {
				So(err, ShouldBeNil)
				So(datasetURI, ShouldEqual, expectedURI)
			})
		})
	})
}

func TestGetIndexName(t *testing.T) {
	t.Parallel()

	Convey("Given index name is not empty", t, func() {
		index := "ons123456789"
		expectedIndex := index

		Convey("When calling getIndexName function", func() {
			indexName := getIndexName(index)

			Convey("Then successfully return the original index name", func() {
				So(indexName, ShouldEqual, expectedIndex)
			})
		})
	})

	Convey("Given index name is empty", t, func() {
		index := ""
		expectedIndex := OnsSearchIndex

		Convey("When calling getIndexName function", func() {
			indexName := getIndexName(index)

			Convey("Then successfully return the default index name", func() {
				So(indexName, ShouldEqual, expectedIndex)
			})
		})
	})
}

// marshalSearchDataImport helper method to marshal a event into a []byte
func marshalSearchDataImport(t *testing.T, sdEvent models.SearchDataImport) []byte {
	bytes, err := schema.SearchDataImportEvent.Marshal(sdEvent)
	if err != nil {
		t.Fatalf("avro mashalling failed with error : %v", err)
	}
	return bytes
}

func setupMetadata() dataset.Metadata {
	m := dataset.Metadata{
		Version: dataset.Version{
			ReleaseDate: "release date",
		},
		DatasetDetails: dataset.DatasetDetails{
			Title:       "title",
			Description: "description",
			Keywords:    &[]string{"keyword_1", "keyword_2"},
		},
	}
	return m
}
