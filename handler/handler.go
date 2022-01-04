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

// ContentPublishedHandler struct to hold handle for zebedee client and the producer
type ContentPublishedHandler struct {
	ZebedeeCli clients.ZebedeeClient
	Producer   event.SearchDataImportProducer
}

// Handle takes a single event.
func (h *ContentPublishedHandler) Handle(ctx context.Context, cpEvent *models.ContentPublished, keywordsLimit int) (err error) {

	traceID := request.NewRequestID(16)

	logData := log.Data{
		"event": cpEvent,
	}
	log.Info(ctx, "event handler called", logData)

	contentPublished, err := h.ZebedeeCli.GetPublishedData(ctx, cpEvent.URI)
	if err != nil {
		return err
	}

	logData = log.Data{
		"contentPublished": string(contentPublished),
	}
	log.Info(ctx, "zebedee response ", logData)

	//byte slice to Json & unMarshall Json
	var zebedeeData models.ZebedeeData
	err = json.Unmarshal(contentPublished, &zebedeeData)
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

	log.Info(ctx, "event successfully handled", logData)
	return nil
}
