package handler

import (
	"context"
	"encoding/json"

	"github.com/ONSdigital/dp-net/request"
	"github.com/ONSdigital/dp-search-data-extractor/clients"
	"github.com/ONSdigital/dp-search-data-extractor/event"
	"github.com/ONSdigital/dp-search-data-extractor/models"
	"github.com/ONSdigital/log.go/v2/log"
)

const (
	ZEBEDEE_DATATYPE    = "Reviewed-uris"
	DATASETAPI_DATATYPE = "Dataset-uris"
)

// ContentPublishedHandler struct to hold handle for zebedee client and the producer
type ContentPublishedHandler struct {
	ZebedeeCli clients.ZebedeeClient
	DatasetCli clients.DatasetClient
	Producer   event.SearchDataImportProducer
}

// Handle takes a single event.
func (h *ContentPublishedHandler) Handle(ctx context.Context, event *models.ContentPublished, keywordsLimit int) (err error) {

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
			log.Fatal(ctx, "error while attempting to send SearchDataImport event to producer", err)
			return err
		}

	} else if event.DataType == DATASETAPI_DATATYPE {

		// Make a call to DatasetAPI
		datasetContentPublished, err := h.DatasetCli.GetEdition(
			ctx, "userAuthToken", "serviceAuthToken", "collectionID", "datasetID", "edition")
		if err != nil {
			log.Info(ctx, "cannot get dataset published contents from api")
			return err
		}

		logData = log.Data{
			"datasetContentPublished": datasetContentPublished,
		}
		log.Info(ctx, "datasetAPI response ", logData)

		//Mapping Json to Avro
		editionData := models.Edition{
			Edition: datasetContentPublished.Edition,
			ID:      datasetContentPublished.ID,
			// Links: datasetContentPublished.Links,
			State: datasetContentPublished.State,
		}
		datasetApiEditionData := models.MapDatasetApiToSearchDataImport(editionData)
		logData = log.Data{
			"datasetApiEditionData": datasetApiEditionData,
		}
		log.Info(ctx, "datasetApiEditionData ", logData)

		// Marshall Avro and sending message
		if err := h.Producer.DatasetAPIImport(ctx, datasetApiEditionData); err != nil {
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
