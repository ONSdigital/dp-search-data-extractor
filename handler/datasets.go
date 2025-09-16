package handler

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/ONSdigital/dp-search-data-extractor/models"
	"github.com/ONSdigital/dp-search-data-extractor/schema"
	"github.com/ONSdigital/log.go/v2/log"
)

const DefaultType = "dataset_landing_page"

// handleDatasetDataType handles a kafka message corresponding to Dataset event type
func (h *ContentPublished) handleDatasetDataType(ctx context.Context, cpEvent *models.ContentPublished) error {
	datasetID, edition, version, err := getIDsFromURI(cpEvent.URI)
	if err != nil {
		log.Error(ctx, "error while attempting to get Ids for dataset, edition and version", err)
		return err
	}

	// Create a new SearchDataImport event with the available data from cpEvent
	// A new unique identifier is generated with the combination of the dataset id and the edition,
	// like so: <datasets_id>-<edition>
	// A default DataType is set, but it might be overwritten later if a Cantabular type is identified
	searchDataImport := &models.SearchDataImport{
		UID:         fmt.Sprintf("%s-%s", datasetID, edition),
		Edition:     edition,
		DatasetID:   datasetID,
		TraceID:     cpEvent.TraceID,
		JobID:       cpEvent.JobID,
		SearchIndex: getIndexName(cpEvent.SearchIndex),
		DataType:    DefaultType,
	}

	// Make a call to DatasetAPI
	datasetMetadataPublished, err := h.DatasetCli.GetVersionMetadata(ctx, "", h.Cfg.ServiceAuthToken, cpEvent.CollectionID, datasetID, edition, version)
	if err != nil {
		log.Error(ctx, "cannot get dataset published metadata from api", err)
		return err
	}
	log.Info(ctx, "successfully obtained metadata from dataset api", log.Data{
		"collection_id": cpEvent.CollectionID,
		"dataset_id":    datasetID,
		"edition":       edition,
		"version":       version,
	})

	// Map data returned by Dataset to the kafka Event structure, including Cantabular fields
	if err := searchDataImport.MapDatasetMetadataValues(ctx, &datasetMetadataPublished); err != nil {
		return fmt.Errorf("failed to map dataset metadata values :%w", err)
	}

	// Marshall Avro and sending message
	if err := h.ImportProducer.Send(schema.SearchDataImportEvent, searchDataImport); err != nil {
		log.Error(ctx, "error while attempting to send DatasetAPIImport event to producer", err)
		return fmt.Errorf("failed to send search data import event: %w", err)
	}
	return nil
}

func getIDsFromURI(uri string) (datasetID, editionID, versionID string, err error) {
	parsedURL, err := url.Parse(uri)
	if err != nil {
		return "", "", "", err
	}

	s := strings.Split(parsedURL.Path, "/")
	if len(s) < 7 {
		return "", "", "", errors.New("not enough arguments in path for version metadata endpoint")
	}
	datasetID = s[2]
	editionID = s[4]
	versionID = s[6]
	return
}
