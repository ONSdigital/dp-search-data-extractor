package handler

import (
	"context"
	"encoding/json"

	"github.com/ONSdigital/dp-search-data-extractor/config"
	"github.com/ONSdigital/dp-search-data-extractor/models"
	"github.com/ONSdigital/log.go/v2/log"
)

// handleZebedeeType handles a kafka message corresponding to Zebedee event type (legacy)
func (h *ContentPublishedHandler) handleZebedeeType(ctx context.Context, cpEvent *models.ContentPublished, cfg config.Config) error {
	// obtain correct uri to callback to Zebedee to retrieve content metadata
	uri, InvalidURIErr := retrieveCorrectURI(cpEvent.URI)
	if InvalidURIErr != nil {
		return InvalidURIErr
	}

	zebedeeContentPublished, err := h.ZebedeeCli.GetPublishedData(ctx, uri)
	if err != nil {
		log.Error(ctx, "failed to retrieve published data from zebedee", err)
		return err
	}

	// byte slice to Json & unMarshall Json
	var zebedeeData models.ZebedeeData
	err = json.Unmarshal(zebedeeContentPublished, &zebedeeData)
	if err != nil {
		log.Fatal(ctx, "error while attempting to unmarshal zebedee response into zebedeeData", err)
		return err
	}

	// keywords validation
	logData := log.Data{
		"uid":           zebedeeData.UID,
		"keywords":      zebedeeData.Description.Keywords,
		"keywordsLimit": cfg.KeywordsLimit,
	}
	log.Info(ctx, "zebedee data ", logData)

	// Map data returned by Zebedee to the kafka Event structure
	searchData := models.MapZebedeeDataToSearchDataImport(zebedeeData, cfg.KeywordsLimit)
	searchData.TraceID = cpEvent.TraceID
	searchData.JobID = cpEvent.JobID
	searchData.SearchIndex = getIndexName(cpEvent.SearchIndex)

	// Marshall Avro and sending message
	if sdImportErr := h.Producer.SearchDataImport(ctx, searchData); sdImportErr != nil {
		log.Error(ctx, "error while attempting to send SearchDataImport event to producer", sdImportErr)
		return sdImportErr
	}
	return nil
}
