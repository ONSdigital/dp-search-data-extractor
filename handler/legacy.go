package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

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
	log.Info(ctx, "zebedee data ", log.Data{
		"uid":           zebedeeData.UID,
		"keywords":      zebedeeData.Description.Keywords,
		"keywordsLimit": h.Cfg.KeywordsLimit,
	})

	if zebedeeData.Description.Title == "" {
		log.Info(ctx, "not processing content as no title present", log.Data{
			"UID": zebedeeData.UID,
			"URI": zebedeeData.URI,
		})
		return nil
	}

	// Map data returned by Zebedee to the kafka Event structure
	searchData := models.MapZebedeeDataToSearchDataImport(zebedeeData, h.Cfg.KeywordsLimit)
	searchData.TraceID = cpEvent.TraceID
	searchData.JobID = cpEvent.JobID
	searchData.SearchIndex = getIndexName(cpEvent.SearchIndex)

	if err := h.Producer.Send(schema.SearchDataImportEvent, &searchData); err != nil {
		log.Error(ctx, "error while attempting to send SearchDataImport event to producer", err)
		return fmt.Errorf("failed to send search data import event: %w", err)
	}

	return nil
}

func retrieveCorrectURI(uri string) (correctURI string, err error) {
	correctURI = uri

	// Remove edition segment of path from Zebedee dataset uri to
	// enable retrieval of the dataset resource for edition
	if strings.Contains(uri, DatasetDataType) {
		correctURI, err = extractDatasetURI(uri)
		if err != nil {
			return "", err
		}
	}

	return correctURI, nil
}

func extractDatasetURI(editionURI string) (string, error) {
	parsedURI, err := url.Parse(editionURI)
	if err != nil {
		return "", err
	}

	slicedURI := strings.Split(parsedURI.Path, "/")
	slicedURI = slicedURI[:len(slicedURI)-1]
	datasetURI := strings.Join(slicedURI, "/")

	return datasetURI, nil
}
