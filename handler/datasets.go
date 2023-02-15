package handler

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-search-data-extractor/config"
	"github.com/ONSdigital/dp-search-data-extractor/models"
	"github.com/ONSdigital/log.go/v2/log"
)

const regexCleanDimensionLabel = `(\(\d+ (([Cc])ategories|([Cc])ategory)\))`

// CantabularTypes are dataset types corresponding to Cantabular datasets
var CantabularTypes = map[string]struct{}{
	"cantabular_flexible_table":     {},
	"cantabular_multivariate_table": {},
}

// PopulationTypes is a mapping between dataset is_based_ok @ID values and population type labels
// Note: this can be also obtained by calling population-api: GET /population-types
var PopulationTypes = map[string]string{
	"atc-ts-demmig-hh-ct-oa":     "All households",
	"atc-ts-demmig-str-ct-oa":    "All non-UK born short-term residents",
	"atc-ts-demmig-ur-ct-oa":     "All usual residents",
	"atc-ts-demmig-ur-pd-oa":     "All usual residents",
	"atc-ts-ed-ftetta-ct-oa":     "All schoolchildren and full-time students aged 5 years and over at their term-time address",
	"atc-ts-eilr-ur-ct-ltla":     "All usual residents aged 3 years and over",
	"atc-ts-eilr-ur-ct-msoa":     "All usual residents",
	"atc-ts-hduc-ur-asp-ltla":    "All usual residents",
	"atc-ts-hous-ur-ct-oa":       "All usual residents",
	"atc-ts-hous-urce-ct-msoa":   "All usual residents in communal establishments",
	"atc-ts-lmttw-ur-ct-oa":      "All usual residents",
	"atc-ts-sogi-ur16o-ct-ltla":  "All usual residents aged 16 years and over",
	"atc-ts-sogi-ur16o-ct-msoa":  "All usual residents aged 16 years and over",
	"atc-ts-vets-vetsur-ct-msoa": "All usual residents who have previously served in the UK armed forces",
	"HH":                         "All Households",
	"UR_HH":                      "All usual residents in households",
	"UR":                         "All usual residents",
}

// handleDatasetDataType handles a kafka message corresponding to Dataset event type
func (h *ContentPublishedHandler) handleDatasetDataType(ctx context.Context, cpEvent *models.ContentPublished, cfg config.Config) error {
	datasetID, edition, version, err := getIDsFromURI(cpEvent.URI)
	if err != nil {
		log.Error(ctx, "error while attempting to get Ids for dataset, edition and version", err)
		return err
	}

	// ID is be a combination of the dataset id and the edition like so: <datasets_id>-<edition>
	generatedID := fmt.Sprintf("%s-%s", datasetID, edition)
	logData := log.Data{
		"uid_generated": generatedID,
	}

	// Make a call to DatasetAPI
	datasetMetadataPublished, err := h.DatasetCli.GetVersionMetadata(ctx, "", cfg.ServiceAuthToken, cpEvent.CollectionID, datasetID, edition, version)
	if err != nil {
		log.Error(ctx, "cannot get dataset published contents version %s from api", err)
		return err
	}

	logData["content_published"] = datasetMetadataPublished
	log.Info(ctx, "datasetAPI response ", logData)

	var uri string
	if len(datasetMetadataPublished.DatasetLinks.LatestVersion.URL) > 0 {
		uri = datasetMetadataPublished.DatasetLinks.LatestVersion.URL
	} else if len(datasetMetadataPublished.DatasetDetails.Links.Version.URL) > 0 {
		uri = datasetMetadataPublished.DatasetDetails.Links.Version.URL
	} else {
		uri = datasetMetadataPublished.Version.Links.Version.URL
	}

	parsedURI, err := url.Parse(uri)
	if err != nil {
		log.Error(ctx, "error parsing the metadata uri", err)
		return err
	}

	// Mapping Json to Avro
	versionDetails := models.VersionDetails{
		ReleaseDate: datasetMetadataPublished.ReleaseDate,
	}

	datasetDetailsData := models.DatasetDetails{
		Title:          datasetMetadataPublished.Title,
		Summary:        datasetMetadataPublished.Description,
		CanonicalTopic: datasetMetadataPublished.CanonicalTopic,
		Subtopics:      datasetMetadataPublished.Subtopics,
		Edition:        edition,
		DatasetID:      datasetID,
		Type:           "dataset_landing_page",
		PopulationType: models.PopulationType{},
		Dimensions:     []models.Dimension{},
	}

	if err := populateCantabularFields(ctx, datasetMetadataPublished, &datasetDetailsData); err != nil {
		return fmt.Errorf("failed to populate cantabular fields: %w", err)
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
	return nil
}

// populateCantabularFields sets the dimensions, type and population type for Cantabular datasets, according to the type in the provided metadata.
// The provided DatasetDetails model will be mutated, adding type, dimensions and population type
func populateCantabularFields(ctx context.Context, metadata dataset.Metadata, dd *models.DatasetDetails) error {
	if dd == nil {
		return errors.New("nil DatasetDetails provided")
	}

	if metadata.DatasetDetails.IsBasedOn == nil {
		return nil // is_based_on not present in Dataset
	}

	t := metadata.DatasetDetails.IsBasedOn.Type
	if _, isCantabular := CantabularTypes[t]; !isCantabular {
		return nil // Dataset type is not Cantabular
	}

	dd.Type = t
	log.Info(ctx, "identified dataset with cantabular type", log.Data{
		"type":           t,
		"num_dimensions": len(metadata.Dimensions)},
	)

	dd.Dimensions = []models.Dimension{}
	for i := range metadata.Dimensions {
		// Using pointers to prevent copying lots of data.
		// TODO consider changing type to []*VersionDimension in dp-api-clients-go
		dim := &metadata.Dimensions[i]
		if dim.IsAreaType != nil && *dim.IsAreaType {
			continue
		}
		dd.Dimensions = append(dd.Dimensions, models.Dimension{
			Name:     dim.ID,
			RawLabel: dim.Label,
			Label:    cleanDimensionLabel(dim.Label),
		})
	}

	popTypeLabel, ok := PopulationTypes[metadata.DatasetDetails.IsBasedOn.ID]
	if !ok {
		log.Warn(ctx, "population type not identified", log.Data{
			"pop_type":    metadata.DatasetDetails.IsBasedOn.ID,
			"valid_types": PopulationTypes},
		)
	}
	dd.PopulationType = models.PopulationType{
		Name:  metadata.DatasetDetails.IsBasedOn.ID,
		Label: popTypeLabel,
	}
	return nil
}

// cleanDimensionLabel is a helper function that parses dimension labels from cantabular into display text
func cleanDimensionLabel(label string) string {
	matcher := regexp.MustCompile(regexCleanDimensionLabel)
	result := matcher.ReplaceAllString(label, "")
	return strings.TrimSpace(result)
}
