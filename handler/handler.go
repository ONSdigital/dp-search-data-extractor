package handler

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"github.com/ONSdigital/dp-search-data-extractor/clients"
	"github.com/ONSdigital/dp-search-data-extractor/config"
	"github.com/ONSdigital/dp-search-data-extractor/event"
	"github.com/ONSdigital/dp-search-data-extractor/models"
	"github.com/ONSdigital/log.go/v2/log"
)

const (
	OnsSearchIndex  = "ONS"
	ZebedeeDataType = "legacy"
	DatasetDataType = "datasets"
)

// ContentPublishedHandler struct to hold handle for config with zebedee, datasetAPI client and the producer
type ContentPublishedHandler struct {
	ZebedeeCli clients.ZebedeeClient
	DatasetCli clients.DatasetClient
	Producer   event.SearchDataImportProducer
}

// Handle takes a single event.
func (h *ContentPublishedHandler) Handle(ctx context.Context, cpEvent *models.ContentPublished, cfg config.Config) error {
	logData := log.Data{
		"event": cpEvent,
	}
	log.Info(ctx, "event handler called with event", logData)

	switch cpEvent.DataType {
	case ZebedeeDataType:
		if err := h.handleZebedeeType(ctx, cpEvent, cfg); err != nil {
			return err
		}
	case DatasetDataType:
		if err := h.handleDatasetDataType(ctx, cpEvent, cfg); err != nil {
			return err
		}
	default:
		log.Info(ctx, "Invalid content data type received, no action")
		return nil // TODO would this need to be an error?
	}

	log.Info(ctx, "event successfully handled", logData)
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

func getIndexName(indexName string) string {
	if indexName != "" {
		return indexName
	}

	return OnsSearchIndex
}
