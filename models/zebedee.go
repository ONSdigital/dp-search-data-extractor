package models

// ZebedeeData provides model for zebedee publisheddata response
type ZebedeeData struct {
	UID             string              `json:"uid"`
	URI             string              `json:"uri"`
	DataType        string              `json:"type"`
	Description     Description         `json:"description"`
	DateChanges     []ReleaseDateChange `json:"date_changes,omitempty"`
	Cancelled       bool                `json:"cancelled,omitempty"`
	Finalised       bool                `json:"finalised,omitempty"`
	ProvisionalDate string              `json:"provisional_date,omitempty"`
	Published       bool                `json:"published,omitempty"`
}

type Description struct {
	CDID            string   `json:"cdid"`
	DatasetID       string   `json:"datasetId"`
	Edition         string   `json:"edition"`
	Keywords        []string `json:"keywords,omitempty"`
	MetaDescription string   `json:"metaDescription"`
	ReleaseDate     string   `json:"releaseDate"`
	Summary         string   `json:"summary"`
	Title           string   `json:"title"`
	Topics          []string `json:"topics,omitempty"`
}

// ReleaseDateChange represent a date change of a release
type ReleaseDateChange struct {
	ChangeNotice string `json:"change_notice"`
	Date         string `json:"previous_date"`
}
