package handler

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	dpkafka "github.com/ONSdigital/dp-kafka/v3"
	"github.com/ONSdigital/dp-kafka/v3/avro"
	"github.com/ONSdigital/dp-kafka/v3/kafkatest"
	"github.com/ONSdigital/dp-search-data-extractor/cache"
	clientMock "github.com/ONSdigital/dp-search-data-extractor/clients/mock"
	"github.com/ONSdigital/dp-search-data-extractor/config"
	"github.com/ONSdigital/dp-search-data-extractor/models"
	"github.com/ONSdigital/dp-search-data-extractor/schema"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	testWorkerID   = 1
	ctx            = context.Background()
	testZebedeeURI = "testZebedeeURI"
	testEdition    = "testEdition"
	testDataType   = "testDataType"
	testTitle      = "testTitle"

	testZebedeeEvent = models.ContentPublished{
		URI:          testZebedeeURI,
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

	mockZebedeePublishedResponse = fmt.Sprintf(`{"description":{"cdid": "testCDID","edition": %q, "title": %q},"type": %q, "URI": %q}`, testEdition, testTitle, testDataType, testZebedeeURI)
	getPublishDataFunc           = func(ctx context.Context, uriString string) ([]byte, error) {
		data := []byte(mockZebedeePublishedResponse)
		return data, nil
	}

	mockDatasetAPIJSONResponse = setupMetadata()
	getVersionMetadataFunc     = func(ctx context.Context, userAuthToken, serviceAuthToken, collectionId, datasetId, edition, version string) (dataset.Metadata, error) {
		data := mockDatasetAPIJSONResponse
		return data, nil
	}

	cfg, _ = config.Get()
)

func TestHandle(t *testing.T) {
	expectedZebedeeProducedEvent := &models.SearchDataImport{
		UID:         testZebedeeURI,
		URI:         testZebedeeURI,
		Title:       testTitle,
		Edition:     testEdition,
		DataType:    testDataType,
		SearchIndex: "ons",
		CDID:        "testCDID",
		Keywords:    []string{},
	}

	expectedDatasetProducedEvent := &models.SearchDataImport{
		UID:            "cphi01-timeseries",
		Edition:        "timeseries",
		DataType:       "dataset_landing_page",
		SearchIndex:    "ons",
		DatasetID:      "cphi01",
		Keywords:       []string{"keyword_1", "keyword_2"},
		ReleaseDate:    "release date",
		Summary:        "description",
		Title:          "title",
		Topics:         []string{"subTopic1"},
		CanonicalTopic: "testTopic",
	}

	Convey("Given an event handler with a working zebedee client and kafka producer", t, func() {
		cacheList, err := cache.GetMockCacheList(ctx)
		if err != nil {
			t.Fatalf("Failed to get mock cache list: %v", err)
		}
		var zebedeeMock = &clientMock.ZebedeeClientMock{
			GetPublishedDataFunc: getPublishDataFunc,
		}
		var producerMock = &kafkatest.IProducerMock{
			SendFunc: func(schema *avro.Schema, event interface{}) error {
				return nil
			},
		}
		h := &ContentPublished{cfg, *cacheList, zebedeeMock, nil, producerMock}

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

			Convey("Then the expected event search data import event is produced", func() {
				So(producerMock.SendCalls(), ShouldHaveLength, 1)
				So(producerMock.SendCalls()[0].Schema, ShouldEqual, schema.SearchDataImportEvent)
				So(producerMock.SendCalls()[0].Event, ShouldResemble, expectedZebedeeProducedEvent)
			})
		})
	})

	Convey("Given an event handler and kafka producer", t, func() {
		cacheList, err := cache.GetMockCacheList(ctx)
		if err != nil {
			t.Fatalf("Failed to get mock cache list: %v", err)
		}
		var zebedeeMock = &clientMock.ZebedeeClientMock{}
		var producerMock = &kafkatest.IProducerMock{
			SendFunc: func(schema *avro.Schema, event interface{}) error {
				return nil
			},
		}
		h := &ContentPublished{cfg, *cacheList, zebedeeMock, nil, producerMock}

		Convey("When a legacy event containing an item we don't want to index", func() {

			nonIndexedEvent := models.ContentPublished{
				URI:          "/timeseries/current/previous/cx15",
				DataType:     "legacy",
				CollectionID: "testZebdeeCollectionID",
			}

			msg := createMessage(nonIndexedEvent)
			err := h.Handle(ctx, testWorkerID, msg)

			Convey("Then no error is reported", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then there is no request to zebedee", func() {
				So(zebedeeMock.GetPublishedDataCalls(), ShouldHaveLength, 0)
			})

			Convey("Then no search data import event is produced", func() {
				So(producerMock.SendCalls(), ShouldHaveLength, 0)
			})
		})
	})

	Convey("Given an event handler with a failing zebedee client", t, func() {
		cacheList, err := cache.GetMockCacheList(ctx)
		if err != nil {
			t.Fatalf("Failed to get mock cache list: %v", err)
		}
		var zebedeeMock = &clientMock.ZebedeeClientMock{
			GetPublishedDataFunc: func(ctx context.Context, uriString string) ([]byte, error) {
				return nil, errors.New("zebedee error")
			},
		}
		h := &ContentPublished{cfg, *cacheList, zebedeeMock, nil, nil}

		Convey("When a legacy event containing a valid URI is handled", func() {
			msg := createMessage(testZebedeeEvent)
			err := h.Handle(ctx, testWorkerID, msg)

			Convey("Then the expected error is reported", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "zebedee error")
			})
		})
	})

	Convey("Given an event handler with a working dataset api client and kafka producer", t, func() {
		cacheList, err := cache.GetMockCacheList(ctx)
		if err != nil {
			t.Fatalf("Failed to get mock cache list: %v", err)
		}
		var datasetMock = &clientMock.DatasetClientMock{
			GetVersionMetadataFunc: getVersionMetadataFunc,
		}
		var producerMock = &kafkatest.IProducerMock{
			SendFunc: func(schema *avro.Schema, event interface{}) error {
				return nil
			},
		}
		h := &ContentPublished{cfg, *cacheList, nil, datasetMock, producerMock}

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
				So(producerMock.SendCalls()[0].Event, ShouldResemble, expectedDatasetProducedEvent)
			})
		})
	})

	Convey("Given an event handler with a failing dataset api client", t, func() {
		cacheList, err := cache.GetMockCacheList(ctx)
		if err != nil {
			t.Fatalf("Failed to get mock cache list: %v", err)
		}
		var datasetMock = &clientMock.DatasetClientMock{
			GetVersionMetadataFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, datasetID string, edition string, version string) (dataset.Metadata, error) {
				return dataset.Metadata{}, errors.New("dataset api error")
			},
		}
		h := &ContentPublished{cfg, *cacheList, nil, datasetMock, nil}

		Convey("When a valid cmd dataset event is handled", func() {
			msg := createMessage(testDatasetEvent)
			err := h.Handle(ctx, testWorkerID, msg)

			Convey("Then the expected error is reported", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "dataset api error")
			})
		})
	})

	Convey("Given an event handler without clients or producer", t, func() {
		cacheList, err := cache.GetMockCacheList(ctx)
		if err != nil {
			t.Fatalf("Failed to get mock cache list: %v", err)
		}
		h := &ContentPublished{cfg, *cacheList, nil, nil, nil}

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
		cacheList, err := cache.GetMockCacheList(ctx)
		if err != nil {
			t.Fatalf("Failed to get mock cache list: %v", err)
		}
		var producerMock = &kafkatest.IProducerMock{
			SendFunc: func(schema *avro.Schema, event interface{}) error {
				return nil
			},
		}
		h := &ContentPublished{cfg, *cacheList, nil, nil, producerMock}

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
