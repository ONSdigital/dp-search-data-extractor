package models

// ContentPublished provides an avro structure for a Content Published event
type ContentPublished struct {
	URI          string `avro:"uri"`
	DataType     string `avro:"data_type"`
	CollectionID string `avro:"collection_id"`
	JobID        string `avro:"job_id"`
	SearchIndex  string `avro:"search_index"`
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
	DateChanges     []ReleaseDateDetails `avro:"date_changes"`
	Cancelled       bool                 `avro:"cancelled"`
	Finalised       bool                 `avro:"finalised"`
	ProvisionalDate string               `avro:"provisional_date"`
	CanonicalTopic  string               `avro:"canonical_topic"`
	Published       bool                 `avro:"published"`
	Language        string               `avro:"language"`
	Survey          string               `avro:"survey"`
}

// ReleaseDateChange represent a date change of a release
type ReleaseDateDetails struct {
	ChangeNotice string `avro:"change_notice"`
	Date         string `avro:"previous_date"`
}
