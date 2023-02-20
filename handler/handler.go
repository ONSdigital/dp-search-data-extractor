package handler

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	kafka "github.com/ONSdigital/dp-kafka/v3"
	"github.com/ONSdigital/dp-search-data-extractor/clients"
	"github.com/ONSdigital/dp-search-data-extractor/config"
	"github.com/ONSdigital/dp-search-data-extractor/models"
	"github.com/ONSdigital/dp-search-data-extractor/schema"
	"github.com/ONSdigital/log.go/v2/log"
)

const (
	OnsSearchIndex  = "ONS"
	ZebedeeDataType = "legacy"
	DatasetDataType = "datasets"
)

// ContentPublished struct to hold handle for config with zebedee, datasetAPI client and the producer
type ContentPublished struct {
	Cfg        *config.Config
	ZebedeeCli clients.ZebedeeClient
	DatasetCli clients.DatasetClient
	Producer   kafka.IProducer
}

// Handle takes a single event and triages it according to its data type, which can be 'legacy' (zebedee) or 'datasets'
// If the type is not correct, the message is ignored with just a log.
func (h *ContentPublished) Handle(ctx context.Context, workerID int, msg kafka.Message) error {
	e := &models.ContentPublished{}
	s := schema.ContentPublishedEvent

	if err := s.Unmarshal(msg.GetData(), e); err != nil {
		return &Error{
			err: fmt.Errorf("failed to unmarshal event: %w", err),
			logData: map[string]interface{}{
				"msg_data": string(msg.GetData()),
			},
		}
	}

	logData := log.Data{"event": e}
	log.Info(ctx, "event received", logData)

	switch e.DataType {
	case ZebedeeDataType:
		if err := h.handleZebedeeType(ctx, e); err != nil {
			return err
		}
	case DatasetDataType:
		if err := h.handleDatasetDataType(ctx, e); err != nil {
			return err
		}
	default:
		log.Warn(ctx,
			"data type not handles by data extractor",
			log.FormatErrors([]error{fmt.Errorf("unrecognised data type received")}),
			log.Data{"data_type": e.DataType},
		)
		return nil
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
