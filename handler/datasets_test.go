package handler

import (
	"context"
	"errors"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-kafka/v5/avro"
	"github.com/ONSdigital/dp-kafka/v5/kafkatest"
	clientMock "github.com/ONSdigital/dp-search-data-extractor/clients/mock"
	"github.com/ONSdigital/dp-search-data-extractor/config"
	"github.com/ONSdigital/dp-search-data-extractor/models"
	. "github.com/smartystreets/goconvey/convey"
)

func TestHandleDatasetDataTypeErrors(t *testing.T) {
	Convey("Given an empty handler and a ContentPbulished event with a malformed URI", t, func() {
		h := &ContentPublished{}
		cpEvent := models.ContentPublished{
			URI: "wrong%%uri",
		}

		Convey("Then the dataset handler returns the expected error", func() {
			err := h.handleDatasetDataType(ctx, &cpEvent)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "parse \"wrong%%uri\": invalid URL escape \"%%u\"")
		})
	})

	Convey("Given a handler with a dataset api mock that fails to return metadata", t, func() {
		datasetMock := &clientMock.DatasetClientMock{
			GetVersionMetadataFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, datasetID string, edition string, version string) (dataset.Metadata, error) {
				return dataset.Metadata{}, errors.New("dataset api error")
			},
		}
		h := &ContentPublished{
			DatasetCli: datasetMock,
			Cfg: &config.Config{
				ServiceAuthToken: "testToken",
			},
		}

		Convey("Then the dataset handler fails with the expected error when a valid event is handled", func() {
			err := h.handleDatasetDataType(ctx, &testDatasetEvent)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "dataset api error")
		})
	})

	Convey("Given a handler with a dataset api mock that returns metadata with a malformed latest version link", t, func() {
		datasetMock := &clientMock.DatasetClientMock{
			GetVersionMetadataFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, datasetID string, edition string, version string) (dataset.Metadata, error) {
				return dataset.Metadata{
					DatasetLinks: dataset.Links{
						LatestVersion: dataset.Link{
							URL: "wrong%%url",
						},
					},
				}, nil
			},
		}
		h := &ContentPublished{
			DatasetCli: datasetMock,
			Cfg: &config.Config{
				ServiceAuthToken: "testToken",
			},
		}

		Convey("Then the dataset handler fails with the expected error when a valid event is handled", func() {
			err := h.handleDatasetDataType(ctx, &testDatasetEvent)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "failed to map dataset metadata values :error parsing the metadata uri: parse \"wrong%%url\": invalid URL escape \"%%u\"")
		})
	})

	Convey("Given a handler with a valid dataset api mock and a producer that fails to send a message", t, func() {
		datasetMock := &clientMock.DatasetClientMock{
			GetVersionMetadataFunc: getVersionMetadataFunc,
		}
		producerMock := &kafkatest.IProducerMock{
			SendFunc: func(ctx context.Context, schema *avro.Schema, event interface{}) error {
				return errors.New("failed to send kafka message")
			},
		}
		h := &ContentPublished{
			DatasetCli:     datasetMock,
			ImportProducer: producerMock,
			DeleteProducer: producerMock,
			Cfg: &config.Config{
				ServiceAuthToken: "testToken",
			},
		}

		Convey("Then the dataset handler fails with the expected error when a valid event is handled", func() {
			err := h.handleDatasetDataType(ctx, &testDatasetEvent)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "failed to send search data import event: failed to send kafka message")
		})
	})
}

func TestGetIDsFromURI(t *testing.T) {
	Convey("getIDsFromURI returns the expected id, edition and version for a valid URL", t, func() {
		uri := "/datasets/cphi01/editions/timeseries/versions/version/metadata"
		datasetID, editionID, versionID, err := getIDsFromURI(uri)
		So(err, ShouldBeNil)
		So(datasetID, ShouldEqual, "cphi01")
		So(editionID, ShouldEqual, "timeseries")
		So(versionID, ShouldEqual, "version")
	})

	Convey("getIDsFromURI returns the expected error when given a malformed URI", t, func() {
		uri := "malformed%%uri"
		_, _, _, err := getIDsFromURI(uri)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldResemble, "parse \"malformed%%uri\": invalid URL escape \"%%u\"")
	})

	Convey("getIDsFromURI returns the expected error when given an URI with a path that doesn't have at least the expected depth", t, func() {
		uri := "/datasets/cphi01/editions/timeseries"
		_, _, _, err := getIDsFromURI(uri)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldResemble, "not enough arguments in path for version metadata endpoint")
	})
}
