package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	var zebedeeContentPublished []byte
	var err error
	if cpEvent.DataType == ZebedeeDataType {
		// obtain correct uri to callback to Zebedee to retrieve content metadata
		uri, InvalidURIErr := retrieveCorrectURI(cpEvent.URI)
		if InvalidURIErr != nil {
			return InvalidURIErr
		}

		zebedeeContentPublished, err = h.ZebedeeCli.GetPublishedData(ctx, uri)
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
		logData = log.Data{
			"uid":           zebedeeData.UID,
			"keywords":      zebedeeData.Description.Keywords,
			"keywordsLimit": cfg.KeywordsLimit,
		}
		log.Info(ctx, "zebedee data ", logData)
		// Mapping Json to Avro
		searchData := models.MapZebedeeDataToSearchDataImport(zebedeeData, cfg.KeywordsLimit)
		searchData.TraceID = cpEvent.TraceID
		searchData.JobID = cpEvent.JobID
		searchData.SearchIndex = getIndexName(cpEvent.SearchIndex)

		// Marshall Avro and sending message
		if sdImportErr := h.Producer.SearchDataImport(ctx, searchData); sdImportErr != nil {
			log.Error(ctx, "error while attempting to send SearchDataImport event to producer", sdImportErr)
			return sdImportErr
		}
	} else if cpEvent.DataType == DatasetDataType {
		datasetID, edition, version, getIDErr := getIDsFromURI(cpEvent.URI)
		if getIDErr != nil {
			log.Error(ctx, "error while attempting to get Ids for dataset, edition and version", getIDErr)
			return getIDErr
		}

		// ID is be a combination of the dataset id and the edition like so: <datasets_id>-<edition>
		generatedID := fmt.Sprintf("%s-%s", datasetID, edition)

		// Make a call to DatasetAPI
		datasetMetadataPublished, metadataErr := h.DatasetCli.GetVersionMetadata(ctx, "", cfg.ServiceAuthToken, cpEvent.CollectionID, datasetID, edition, version)
		if metadataErr != nil {
			log.Error(ctx, "cannot get dataset published contents version %s from api", metadataErr)
			return metadataErr
		}
		logData = log.Data{
			"uid generated":    generatedID,
			"contentPublished": datasetMetadataPublished,
		}
		log.Info(ctx, "datasetAPI response ", logData)

		var uri string
		if len(datasetMetadataPublished.DatasetLinks.LatestVersion.URL) > 0 {
			uri = datasetMetadataPublished.DatasetLinks.LatestVersion.URL
		} else if len(datasetMetadataPublished.DatasetDetails.Links.Version.URL) > 0 {
			uri = datasetMetadataPublished.DatasetDetails.Links.Version.URL
		} else {
			uri = datasetMetadataPublished.Version.Links.Version.URL
		}

		parsedURI, parseErr := url.Parse(uri)
		if err != nil {
			log.Error(ctx, "error parsing the metadata uri", parseErr)
			return parseErr
		}

		// Mapping Json to Avro
		versionDetails := models.VersionDetails{
			ReleaseDate: datasetMetadataPublished.ReleaseDate,
		}

		// Storing Dimensions if cantabular && is not area type.
		if datasetMetadataPublished.Version.IsBasedOn.Type == "cantabular_multivariate_table" ||
			datasetMetadataPublished.Version.IsBasedOn.Type == "cantabular_flexible_table" {
			for _, value := range datasetMetadataPublished.Dimensions {
				if !*value.IsAreaType {
					versionDetails.Dimensions = append(versionDetails.Dimensions, value)
				}
			}
		}

		datasetDetailsData := models.DatasetDetails{
			Title:          datasetMetadataPublished.Title,
			Summary:        datasetMetadataPublished.Description,
			CanonicalTopic: datasetMetadataPublished.CanonicalTopic,
			Subtopics:      datasetMetadataPublished.Subtopics,
			Edition:        edition,
			DatasetID:      datasetID,
			Type:           "dataset_landing_page",
		}

		if datasetMetadataPublished.Keywords != nil {
			datasetDetailsData.Keywords = *datasetMetadataPublished.Keywords
		}

		versionMetadata := models.CMDData{
			UID:            generatedID,
			URI:            parsedURI.Path,
			VersionDetails: versionDetails,
			DatasetDetails: datasetDetailsData,
		}

		datasetVersionMetadata := models.MapVersionMetadataToSearchDataImport(versionMetadata)
		logData = log.Data{
			"datasetVersionData": datasetVersionMetadata,
		}
		log.Info(ctx, "datasetVersionMetadata ", logData)

		datasetVersionMetadata.TraceID = cpEvent.TraceID
		datasetVersionMetadata.JobID = cpEvent.JobID
		datasetVersionMetadata.SearchIndex = getIndexName(cpEvent.SearchIndex)
		datasetVersionMetadata.DataType = "dataset_landing_page"

		// Marshall Avro and sending message
		if sdImportErr := h.Producer.SearchDataImport(ctx, datasetVersionMetadata); sdImportErr != nil {
			log.Fatal(ctx, "error while attempting to send DatasetAPIImport event to producer", sdImportErr)
			return sdImportErr
		}
	} else {
		log.Info(ctx, "Invalid content data type received, no action")
		return err
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
