package models

// ZebedeeData provides model for zebedee publisheddata response
type ZebedeeData struct {
	UID         string              `json:"uid"`
	URI         string              `json:"uri"`
	DataType    string              `json:"type"`
	Description Description         `json:"description"`
	DateChanges []ReleaseDateChange `json:"dateChanges,omitempty"`
}

type Description struct {
	Cancelled       bool     `json:"cancelled,omitempty"`
	CDID            string   `json:"cdid"`
	DatasetID       string   `json:"datasetId"`
	Edition         string   `json:"edition"`
	Finalised       bool     `json:"finalised,omitempty"`
	Keywords        []string `json:"keywords,omitempty"`
	MetaDescription string   `json:"metaDescription"`
	MigrationLink   string   `json:"migrationLink,omitempty"`
	ProvisionalDate string   `json:"provisionalDate,omitempty"`
	CanonicalTopic  string   `json:"canonicalTopic,omitempty"`
	Published       bool     `json:"published,omitempty"`
	ReleaseDate     string   `json:"releaseDate"`
	Summary         string   `json:"summary"`
	Title           string   `json:"title"`
	Topics          []string `json:"secondaryTopics,omitempty"`
	Language        string   `json:"language,omitempty"`
	Survey          string   `json:"survey,omitempty"`
}

// ReleaseDateChange represent a date change of a release
type ReleaseDateChange struct {
	ChangeNotice string `json:"changeNotice"`
	Date         string `json:"previousDate"`
}
