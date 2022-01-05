package models

// CMDData provides model for datasetAPI metadata response
type CMDData struct {
	DataId         string         `json:"dataid"`
	VersionDetails VersionDetails `json:"version"`
	DatasetDetails DatasetDetails `json:"datasetdetails"`
}

// Version represents a version for an edition within a dataset
type VersionDetails struct {
	ReleaseDate   string   `json:"release_date,omitempty"`
	LatestChanges []string `json:"latestchanges,omitempty"`
}

// DatasetDetails represents a DatasetDetails for an edition within a dataset
type DatasetDetails struct {
	Title             string   `json:"title"`
	Description       string   `json:"description"`
	Keywords          []string `json:"keywords,omitempty"`
	ReleaseFrequency  string   `json:"releasefrequency,omitempty"`
	NextRelease       string   `json:"nextrelease,omitempty"`
	UnitOfMeasure     string   `json:"Unitofmeasure,omitempty"`
	License           string   `json:"license,omitempty"`
	NationalStatistic bool     `json:"nationalstatistic,omitempty"`
}
