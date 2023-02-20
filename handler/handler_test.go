package handler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-kafka/v3/avro"
	"github.com/ONSdigital/dp-kafka/v3/kafkatest"
	"github.com/ONSdigital/dp-search-data-extractor/config"
	"github.com/ONSdigital/dp-search-data-extractor/models"
	"github.com/ONSdigital/dp-search-data-extractor/schema"

	dpkafka "github.com/ONSdigital/dp-kafka/v3"
	clientMock "github.com/ONSdigital/dp-search-data-extractor/clients/mock"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	testTimeout  = time.Second * 5
	testWorkerID = 1
	ctx          = context.Background()

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

func TestHandle(t *testing.T) {
	expectedZebedeeProducedEvent := models.SearchDataImport{
		UID:         "testedition",
		Edition:     "testedition",
		DataType:    "testDataType",
		SearchIndex: "ONS",
		CDID:        "testCDID",
		Keywords:    []string{},
	}

	expectedDatasetProducedEvent := models.SearchDataImport{
		UID:            "cphi01-timeseries",
		Edition:        "timeseries",
		DataType:       "dataset_landing_page",
		SearchIndex:    "ONS",
		DatasetID:      "cphi01",
		Keywords:       []string{"keyword_1", "keyword_2"},
		ReleaseDate:    "release date",
		Summary:        "description",
		Title:          "title",
		Topics:         []string{"subTopic1"},
		CanonicalTopic: "testTopic",
	}

	Convey("Given an event handler with a working zebedee client and kafka producer", t, func() {
		var zebedeeMock = &clientMock.ZebedeeClientMock{
			GetPublishedDataFunc: getPublishDataFunc,
		}
		var producerMock = &kafkatest.IProducerMock{
			SendFunc: func(schema *avro.Schema, event interface{}) error {
				return nil
			},
		}
		h := &ContentPublished{cfg, zebedeeMock, nil, producerMock}

		Convey("When a legacy event containing a valid URI is handled", func() {
			msg := createMessage(testZebedeeEvent)
			err := h.Handle(ctx, testWorkerID, msg)

			Convey("Then no error is reported", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then published data is obtained from zebedee", func() {
				So(zebedeeMock.GetPublishedDataCalls(), ShouldHaveLength, 1)
				So(zebedeeMock.GetPublishedDataCalls()[0].UriString, ShouldEqual, testZebedeeEvent.URI)
			})

			Convey("Then the expected event search data import event is producer", func() {
				So(producerMock.SendCalls(), ShouldHaveLength, 1)
				So(producerMock.SendCalls()[0].Schema, ShouldEqual, schema.SearchDataImportEvent)
				So(producerMock.SendCalls()[0].Event, ShouldResemble, expectedZebedeeProducedEvent)
			})
		})
	})

	Convey("Given an event handler with a working dataset api client and kafka producer", t, func() {
		var datasetMock = &clientMock.DatasetClientMock{
			GetVersionMetadataFunc: getVersionMetadataFunc,
		}
		var producerMock = &kafkatest.IProducerMock{
			SendFunc: func(schema *avro.Schema, event interface{}) error {
				return nil
			},
		}
		h := &ContentPublished{cfg, nil, datasetMock, producerMock}

		Convey("When a valid cmd dataset event is handled", func() {
			msg := createMessage(testDatasetEvent)
			err := h.Handle(ctx, testWorkerID, msg)

			Convey("Then no error is reported", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then version metadata details are obtained from dataset API", func() {
				So(datasetMock.GetVersionMetadataCalls(), ShouldHaveLength, 1)
				So(datasetMock.GetVersionMetadataCalls()[0].DatasetID, ShouldEqual, "cphi01")
				So(datasetMock.GetVersionMetadataCalls()[0].Edition, ShouldEqual, "timeseries")
				So(datasetMock.GetVersionMetadataCalls()[0].Version, ShouldEqual, "version")
			})

			Convey("Then the expected event search data import event is producer", func() {
				So(producerMock.SendCalls(), ShouldHaveLength, 1)
				So(producerMock.SendCalls()[0].Schema, ShouldEqual, schema.SearchDataImportEvent)
				So(producerMock.SendCalls()[0].Event, ShouldResemble, &expectedDatasetProducedEvent)
			})
		})
	})

	Convey("Given an event handler without clients or producer", t, func() {
		h := &ContentPublished{cfg, nil, nil, nil}

		Convey("When an event with an unsupported type is handled", func() {
			msg := createMessage(testInvalidEvent)
			err := h.Handle(ctx, testWorkerID, msg)

			Convey("Then no error is reported", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestHandleErrors(t *testing.T) {
	Convey("Given an event handler working successfully", t, func() {
		var producerMock = &kafkatest.IProducerMock{
			SendFunc: func(schema *avro.Schema, event interface{}) error {
				return nil
			},
		}
		h := &ContentPublished{cfg, nil, nil, producerMock}

		Convey("When a malformed event is handled", func() {
			msg, err := kafkatest.NewMessage([]byte{1, 2, 3}, 0)
			So(err, ShouldBeNil)
			err = h.Handle(ctx, testWorkerID, msg)

			Convey("Then the expected error is reported", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "failed to unmarshal event: Invalid string length")
			})
		})
	})
}

// func TestHandlerForDatasetVersionMetadata(t *testing.T) {
// 	t.Parallel()
// 	expectedVersionMetadataEvent := models.SearchDataImport{
// 		UID:             "cphi01-timeseries",
// 		DataType:        "dataset_landing_page",
// 		JobID:           "",
// 		SearchIndex:     "ONS",
// 		CDID:            "",
// 		DatasetID:       "",
// 		Keywords:        []string{"somekeyword0", "somekeyword1"},
// 		MetaDescription: "someDescription",
// 		Summary:         "",
// 		ReleaseDate:     "2020-11-07T00:00:00.000Z",
// 		Title:           "someTitle",
// 		Topics:          []string{"testtopic1", "testtopic2"},
// 		CanonicalTopic:  "something",
// 	}

// 	producerMock := &kafkatest.IProducerMock{
// 		ChannelsFunc: getChannelFunc,
// 	}

// // for mock marshaller
// expectedVersionMetadataSearchDataImport := marshalSearchDataImport(t, expectedVersionMetadataEvent)
// marshallerMock := &mock.MarshallerMock{
// 	MarshalFunc: func(s interface{}) ([]byte, error) {
// 		return expectedVersionMetadataSearchDataImport, nil
// 	},
// }

// producerMock := &event.SearchDataImportProducer{
// 	Producer:   kafkaProducerMock,
// 	Marshaller: marshallerMock,
// }

// 	Convey("Given an event handler working successfully, and an event containing a valid dataType", t, func() {
// 		var zebedeeMock = &clientMock.ZebedeeClientMock{GetPublishedDataFunc: getPublishDataFunc}
// 		var datasetMock = &clientMock.DatasetClientMock{GetVersionMetadataFunc: getVersionMetadataFunc}

// 		eventHandler := &ContentPublished{cfg, zebedeeMock, datasetMock, producerMock}

// 		Convey("When given a valid event for cmd dataset", func() {
// 			msg := createMessage(testDatasetEvent)
// 			err := eventHandler.Handle(ctx, testWorkerID, msg)

// 			var avroBytes []byte
// 			delay := time.NewTimer(testTimeout)
// 			select {
// 			case avroBytes = <-pChannels.Output:
// 				// Ensure timer is stopped and its resources are freed
// 				if !delay.Stop() {
// 					// if the timer has been stopped then read from the channel
// 					<-delay.C
// 				}
// 				t.Log("avro byte sent to producer output")
// 			case <-delay.C:
// 				t.FailNow()
// 			}

// 			Convey("Then no error is reported", func() {
// 				So(err, ShouldBeNil)
// 			})
// 			Convey("And then get content published from datasetAPI but no action for zebedee", func() {
// 				So(len(zebedeeMock.GetPublishedDataCalls()), ShouldEqual, 0)
// 				So(len(datasetMock.GetVersionMetadataCalls()), ShouldEqual, 1)
// 			})
// 			Convey("And then the expected bytes are sent to producer.output", func() {
// 				var actual models.SearchDataImport
// 				err = schema.SearchDataImportEvent.Unmarshal(avroBytes, &actual)
// 				So(err, ShouldBeNil)
// 				So(actual.UID, ShouldEqual, expectedVersionMetadataEvent.UID)
// 				So(actual.ReleaseDate, ShouldEqual, expectedVersionMetadataEvent.ReleaseDate)
// 				So(actual.Title, ShouldEqual, expectedVersionMetadataEvent.Title)
// 				So(actual.MetaDescription, ShouldEqual, expectedVersionMetadataEvent.MetaDescription)
// 				So(actual.Keywords, ShouldHaveLength, 2)
// 				So(actual.Keywords[0], ShouldEqual, "somekeyword0")
// 				So(actual.Keywords[1], ShouldEqual, "somekeyword1")
// 				So(actual.DataType, ShouldResemble, expectedVersionMetadataEvent.DataType)
// 				So(actual.CanonicalTopic, ShouldEqual, expectedVersionMetadataEvent.CanonicalTopic)
// 				So(actual.Topics[0], ShouldEqual, expectedVersionMetadataEvent.Topics[0])
// 				So(actual.Topics[1], ShouldEqual, expectedVersionMetadataEvent.Topics[1])
// 				So(actual.URI, ShouldEqual, expectedVersionMetadataEvent.URI)
// 				So(actual.Edition, ShouldEqual, expectedVersionMetadataEvent.Edition)
// 				So(actual.DatasetID, ShouldEqual, expectedVersionMetadataEvent.DatasetID)
// 			})
// 		})
// 	})
// 	Convey("Given an event handler not working successfully, and an event containing a valid dataType", t, func() {
// 		var zebedeeMockInError = &clientMock.ZebedeeClientMock{GetPublishedDataFunc: getPublishDataFuncInError}
// 		var datasetMockInError = &clientMock.DatasetClientMock{GetVersionMetadataFunc: getVersionMetadataFuncInError}
// 		eventHandler := &ContentPublished{cfg, zebedeeMockInError, datasetMockInError, producerMock}

// 		Convey("When given a valid event for a valid dataset", func() {
// 			msg := createMessage(testDatasetEvent)
// 			err := eventHandler.Handle(ctx, testWorkerID, msg)

// 			Convey("Then Zebedee is called 0 time and Dataset called 1 time with the expected error ", func() {
// 				So(err, ShouldNotBeNil)
// 				So(err.Error(), ShouldEqual, errDatasetAPI.Error())
// 				So(zebedeeMockInError.GetPublishedDataCalls(), ShouldHaveLength, 0)

// 				So(len(datasetMockInError.GetVersionMetadataCalls()), ShouldEqual, 1)
// 				So(datasetMockInError.GetVersionMetadataCalls(), ShouldNotBeEmpty)
// 				So(datasetMockInError.GetVersionMetadataCalls(), ShouldHaveLength, 1)
// 			})
// 		})
// 	})
// }

// func TestHandlerForInvalidDataType(t *testing.T) {
// 	t.Parallel()
// 	// expectedInvalidSearchDataImportEvent := models.SearchDataImport{}

// 	producerMock := &kafkatest.IProducerMock{
// 		ChannelsFunc: getChannelFunc,
// 	}

// 	// // for mock marshaller
// 	// expectedVersionMetadataSearchDataImport := marshalSearchDataImport(t, expectedInvalidSearchDataImportEvent)
// 	// marshallerMock := &mock.MarshallerMock{
// 	// 	MarshalFunc: func(s interface{}) ([]byte, error) {
// 	// 		return expectedVersionMetadataSearchDataImport, nil
// 	// 	},
// 	// }

// 	// producerMock := &event.SearchDataImportProducer{
// 	// 	Producer:   kafkaProducerMock,
// 	// 	Marshaller: marshallerMock,
// 	// }

// 	Convey("Given an event handler working successfully, and an event containing a invalid dataType", t, func() {
// 		var zebedeeMockInError = &clientMock.ZebedeeClientMock{GetPublishedDataFunc: getPublishDataFuncInError}
// 		var datasetMockInError = &clientMock.DatasetClientMock{GetVersionMetadataFunc: getVersionMetadataFuncInError}
// 		eventHandler := &ContentPublished{cfg, zebedeeMockInError, datasetMockInError, producerMock}

// 		Convey("When given a invalid event", func() {
// 			msg := createMessage(testInvalidEvent)
// 			err := eventHandler.Handle(ctx, testWorkerID, msg)

// 			Convey("Then both Zebedee and Dataset is called 0 time with no expected error ", func() {
// 				So(err, ShouldBeNil)

// 				So(zebedeeMockInError.GetPublishedDataCalls(), ShouldHaveLength, 0)
// 				So(zebedeeMockInError.GetPublishedDataCalls(), ShouldBeEmpty)

// 				So(datasetMockInError.GetVersionMetadataCalls(), ShouldHaveLength, 0)
// 				So(datasetMockInError.GetVersionMetadataCalls(), ShouldBeEmpty)
// 			})
// 		})
// 	})
// }

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

func createMessage(s interface{}) dpkafka.Message {
	e, err := schema.ContentPublishedEvent.Marshal(s)
	So(err, ShouldBeNil)
	msg, err := kafkatest.NewMessage(e, 0)
	So(err, ShouldBeNil)
	return msg
}

func setupMetadata() dataset.Metadata {
	m := dataset.Metadata{
		Version: dataset.Version{
			ReleaseDate: "release date",
		},
		DatasetDetails: dataset.DatasetDetails{
			Title:          "title",
			Description:    "description",
			Keywords:       &[]string{"keyword_1", "keyword_2"},
			CanonicalTopic: "testTopic",
			Subtopics:      []string{"subTopic1"},
		},
	}
	return m
}
