package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ONSdigital/dp-search-data-extractor/models"
	"github.com/ONSdigital/dp-search-data-extractor/schema"
	"github.com/ONSdigital/log.go/v2/log"
)

// handleZebedeeType handles a kafka message corresponding to Zebedee event type (legacy)
func (h *ContentPublished) handleZebedeeType(ctx context.Context, cpEvent *models.ContentPublished) error {
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
		"keywordsLimit": h.Cfg.KeywordsLimit,
	}
	log.Info(ctx, "zebedee data ", logData)

	// Mapping Json to Avro
	searchData := models.MapZebedeeDataToSearchDataImport(zebedeeData, h.Cfg.KeywordsLimit)
	searchData.TraceID = cpEvent.TraceID
	searchData.JobID = cpEvent.JobID
	searchData.SearchIndex = getIndexName(cpEvent.SearchIndex)

	if err := h.Producer.Send(schema.SearchDataImportEvent, searchData); err != nil {
		log.Error(ctx, "error while attempting to send SearchDataImport event to producer", err)
		return fmt.Errorf("failed to send search data import event: %w", err)
	}

	return nil
}
