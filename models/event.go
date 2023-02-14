package models

// ContentPublished provides an avro structure for a Content Published event
type ContentPublished struct {
	URI          string `avro:"uri"`
	DataType     string `avro:"data_type"`
	CollectionID string `avro:"collection_id"`
	JobID        string `avro:"job_id"`
	SearchIndex  string `avro:"search_index"`
	TraceID      string `avro:"trace_id"`
}

// SearchDataImport provides event data for a search data import
type SearchDataImport struct {
	UID             string               `avro:"uid"`
	URI             string               `avro:"uri"`
	Edition         string               `avro:"edition"`
	DataType        string               `avro:"data_type"`
	JobID           string               `avro:"job_id"`
	SearchIndex     string               `avro:"search_index"`
	CDID            string               `avro:"cdid"`
	DatasetID       string               `avro:"dataset_id"`
	Keywords        []string             `avro:"keywords"`
	MetaDescription string               `avro:"meta_description"`
	ReleaseDate     string               `avro:"release_date"`
	Summary         string               `avro:"summary"`
	Title           string               `avro:"title"`
	Topics          []string             `avro:"topics"`
	TraceID         string               `avro:"trace_id"`
	DateChanges     []ReleaseDateDetails `avro:"date_changes"`
	Cancelled       bool                 `avro:"cancelled"`
	Finalised       bool                 `avro:"finalised"`
	ProvisionalDate string               `avro:"provisional_date"`
	CanonicalTopic  string               `avro:"canonical_topic"`
	Published       bool                 `avro:"published"`
	Language        string               `avro:"language"`
	Survey          string               `avro:"survey"`
	Dimensions      []VersionDimension   `avro:"dimensions"`
}

type VersionDimension struct {
	ID              string `avro:"id"`
	Name            string `avro:"name"`
	Links           Links  `avro:"links"`
	Description     string `avro:"description"`
	Label           string `avro:"label"`
	URL             string `avro:"href"`
	Variable        string `avro:"variable"`
	NumberOfOptions int    `avro:"number_of_options"`
	//IsAreaType      *bool  `avro:"is_area_type"`
}

type Links struct {
	AccessRights  Link `avro:"access_rights"`
	Dataset       Link `avro:"dataset"`
	Dimensions    Link `avro:"dimensions"`
	Edition       Link `avro:"edition"`
	Editions      Link `avro:"editions"`
	LatestVersion Link `avro:"latest_version"`
	Versions      Link `avro:"versions"`
	Self          Link `avro:"self"`
	CodeList      Link `avro:"code_list"`
	Options       Link `avro:"options"`
	Version       Link `avro:"version"`
	Code          Link `avro:"code"`
	Taxonomy      Link `avro:"taxonomy"`
	Job           Link `avro:"job"`
}

type Link struct {
	URL string `avro:"href"`
	ID  string `avro:"id"`
}

// ReleaseDateChange represent a date change of a release
type ReleaseDateDetails struct {
	ChangeNotice string `avro:"change_notice"`
	Date         string `avro:"previous_date"`
}
