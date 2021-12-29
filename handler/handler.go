package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/ONSdigital/dp-net/request"
	"github.com/ONSdigital/dp-search-data-extractor/clients"
	"github.com/ONSdigital/dp-search-data-extractor/config"
	"github.com/ONSdigital/dp-search-data-extractor/event"
	"github.com/ONSdigital/dp-search-data-extractor/models"
	"github.com/ONSdigital/log.go/v2/log"
)

const (
	ZEBEDEE_DATATYPE    = "Reviewed-uris"
	DATASETAPI_DATATYPE = "Dataset-uris"
)

// ContentPublishedHandler struct to hold handle for config with zebedee, datasetAPI client and the producer
type ContentPublishedHandler struct {
	ZebedeeCli clients.ZebedeeClient
	DatasetCli clients.DatasetClient
	Producer   event.SearchDataImportProducer
}

// Handle takes a single event.
func (h *ContentPublishedHandler) Handle(ctx context.Context, event *models.ContentPublished, keywordsLimit int, cfg config.Config) (err error) {

	traceID := request.NewRequestID(16)

	logData := log.Data{
		"event":    event,
		"datatype": event.DataType,
	}
	log.Info(ctx, "event handler called with datatype", logData)

	if event.DataType == ZEBEDEE_DATATYPE {

		// Make a call to Zebedee
		zebedeeContentPublished, err := h.ZebedeeCli.GetPublishedData(ctx, event.URI)
		if err != nil {
			return err
		}

		logData = log.Data{
			"contentPublished": string(zebedeeContentPublished),
		}
		log.Info(ctx, "zebedee response ", logData)

		//byte slice to Json & unMarshall Json
		var zebedeeData models.ZebedeeData
		err = json.Unmarshal(zebedeeContentPublished, &zebedeeData)
		if err != nil {
			log.Fatal(ctx, "error while attempting to unmarshal zebedee response into zebedeeData", err)
			return err
		}

		//keywords validation
		logData = log.Data{
			"keywords":      zebedeeData.Description.Keywords,
			"keywordsLimit": keywordsLimit,
		}
		//Mapping Json to Avro
		searchData := models.MapZebedeeDataToSearchDataImport(zebedeeData, keywordsLimit)
		searchData.TraceID = traceID
		searchData.JobID = ""
		searchData.SearchIndex = "ONS"

		//Marshall Avro and sending message
		if err := h.Producer.SearchDataImport(ctx, searchData); err != nil {
			log.Error(ctx, "error while attempting to send SearchDataImport event to producer", err)
			return err
		}

	} else if event.DataType == DATASETAPI_DATATYPE {

		datasetId, edition, version, err := getIDsFromUri(event.URI)
		if err != nil {
			log.Error(ctx, "error while attempting to get Ids for dataset, edition and version", err)
			return err
		}

		//ID is be a combination of the dataset id and the edition like so: <datasets_id>-<edition> e.g. cpih-timeseries
		generatedID := fmt.Sprintf("%s-%s", datasetId, edition)

		// Make a call to DatasetAPI
		datasetMetadataPublished, err := h.DatasetCli.GetVersionMetadata(ctx, cfg.UserAuthToken, cfg.ServiceAuthToken, event.CollectionID, datasetId, edition, version)
		if err != nil {
			log.Error(ctx, "cannot get dataset published contents version %s from api", err)
			return err
		}

		logData = log.Data{
			"ID generated":            generatedID,
			"datasetContentPublished": datasetMetadataPublished,
		}
		log.Info(ctx, "datasetAPI response ", logData)

		//Mapping Json to Avro
		versionData := models.VersionMetadata{
			CollectionId: event.CollectionID,
			Edition:      edition,
			ID:           generatedID,
			DatasetId:    datasetId,
			ReleaseDate:  datasetMetadataPublished.ReleaseDate,
			Version:      version,
		}
		datasetVersionData := models.MapDatasetVersionToSearchDataImport(versionData)
		logData = log.Data{
			"datasetVersionData": datasetVersionData,
		}
		log.Info(ctx, "datasetApiEditionData ", logData)

		// Marshall Avro and sending message
		if err := h.Producer.SearchDatasetVersionMetadataImport(ctx, datasetVersionData); err != nil {
			log.Fatal(ctx, "error while attempting to send DatasetAPIImport event to producer", err)
			return err
		}
	} else {
		log.Info(ctx, "Invalid data type received, no action")
		return
	}

	log.Info(ctx, "event successfully handled", logData)
	return nil
}

func getIDsFromUri(uri string) (datasetID, editionID, versionID string, err error) {
	parsedURL, err := url.Parse(uri)
	if err != nil {
		return "", "", "", err
	}

	s := strings.Split(parsedURL.Path, "/")
	if len(s) < 7 {
		return "", "", "", errors.New("not enough arguments in path")
	}
	datasetID = s[1]
	editionID = s[3]
	versionID = s[5]
	return datasetID, editionID, versionID, nil
}
