package models

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
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

func (s *SearchDataImport) MapDatasetMetadataValues(ctx context.Context, metadata *dataset.Metadata) error {
	if metadata == nil {
		return fmt.Errorf("nil metadata cannot be mapped")
	}

	uri := GetURI(metadata)
	parsedURI, err := url.Parse(uri)
	if err != nil {
		return fmt.Errorf("error parsing the metadata uri: %w", err)
	}
	s.URI = parsedURI.Path

	s.ReleaseDate = metadata.ReleaseDate
	s.Title = metadata.Title
	s.Summary = metadata.Description
	s.CanonicalTopic = metadata.CanonicalTopic
	s.Topics = metadata.Subtopics

	if metadata.Keywords != nil {
		s.Keywords = *metadata.Keywords
	}

	s.PopulateCantabularFields(ctx, metadata)

	return nil
}

// PopulateCantabularFields checks if the provided dataset metadata corresponds to a Cantabular Data type,
// if it does, it populates the dimensions array of SearchDataImport with the dimension names, labels and processed labels,
// and assigns the population type corresponding to the 'IsBasedOn' id value.
func (s *SearchDataImport) PopulateCantabularFields(ctx context.Context, metadata *dataset.Metadata) {
	if metadata.DatasetDetails.IsBasedOn == nil {
		return // is_based_on not present in Dataset
	}

	t := metadata.DatasetDetails.IsBasedOn.Type
	if _, isCantabular := CantabularTypes[t]; !isCantabular {
		return // Dataset type is not Cantabular
	}

	s.DataType = t
	log.Info(ctx, "identified dataset with cantabular type", log.Data{
		"type":           t,
		"num_dimensions": len(metadata.Dimensions)},
	)

	s.Dimensions = []Dimension{}
	for i := range metadata.Dimensions {
		// Using pointers to prevent copying lots of data.
		// TODO consider changing type to []*VersionDimension in dp-api-clients-go
		dim := &metadata.Dimensions[i]
		if dim.IsAreaType != nil && *dim.IsAreaType {
			continue
		}
		s.Dimensions = append(s.Dimensions, Dimension{
			Name:     dim.ID,
			RawLabel: dim.Label,
			Label:    cleanDimensionLabel(dim.Label),
		})
	}

	popTypeLabel, ok := PopulationTypes[metadata.DatasetDetails.IsBasedOn.ID]
	if !ok {
		log.Warn(ctx, "population type not identified",
			log.Data{
				"pop_type":    metadata.DatasetDetails.IsBasedOn.ID,
				"valid_types": PopulationTypes,
			},
		)
	}
	s.PopulationType = PopulationType{
		Name:  metadata.DatasetDetails.IsBasedOn.ID,
		Label: popTypeLabel,
	}
}

// cleanDimensionLabel is a helper function that parses dimension labels from cantabular into display text
func cleanDimensionLabel(label string) string {
	matcher := regexp.MustCompile(regexCleanDimensionLabel)
	result := matcher.ReplaceAllString(label, "")
	return strings.TrimSpace(result)
}

// GetURI obtains the URI from the provided metadata struct
func GetURI(metadata *dataset.Metadata) string {
	if metadata == nil {
		return ""
	}
	if len(metadata.DatasetLinks.LatestVersion.URL) > 0 {
		return metadata.DatasetLinks.LatestVersion.URL
	}
	if len(metadata.DatasetDetails.Links.Version.URL) > 0 {
		return metadata.DatasetDetails.Links.Version.URL
	}
	return metadata.Version.Links.Version.URL
}
