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
	PopulationType  PopulationType       `avro:"population_type"`
	Dimensions      []Dimension          `avro:"dimensions"`
}

// ReleaseDateChange represent a date change of a release
type ReleaseDateDetails struct {
	ChangeNotice string `avro:"change_notice"`
	Date         string `avro:"previous_date"`
}

// Dimension represents the required information for each dataset dimension: name (unique ID) and label
// and an aggregation key which combines name and label
type Dimension struct {
	Key      string `avro:"key"`
	AggKey   string `avro:"agg_key"`
	Name     string `avro:"name"`
	Label    string `avro:"label"`
	RawLabel string `avro:"raw_label"`
}

// PopulationType represents the population type name (unique ID) and label
// and an aggregation key which combines name and label
type PopulationType struct {
	Key    string `avro:"key"`
	AggKey string `avro:"agg_key"`
	Name   string `avro:"name"`
	Label  string `avro:"label"`
}

// SearchContentUpdate - this represents a standard resource metadata model and json representation for API
type SearchContentUpdate struct {
	CanonicalTopic  string   `avro:"canonical_topic"`
	CDID            string   `avro:"cdid" json:"cdid"`
	ContentType     string   `avro:"content_type"`
	DatasetID       string   `avro:"dataset_id"`
	Edition         string   `avro:"edition"`
	Language        string   `avro:"language"`
	MetaDescription string   `avro:"meta_description"`
	ReleaseDate     string   `avro:"release_date"`
	Summary         string   `avro:"summary"`
	Survey          string   `avro:"survey"`
	SearchIndex     string   `avro:"search_index"`
	Title           string   `avro:"title"`
	Topics          []string `avro:"topics"`
	TraceID         string   `avro:"trace_id"`
	URI             string   `avro:"uri"`
	URIOld          string   `avro:"uri_old"`
	// These fields are only used for content_type=release
	Cancelled       bool                 `avro:"cancelled"`
	DateChanges     []ReleaseDateDetails `avro:"date_changes"`
	Finalised       bool                 `avro:"finalised"`
	ProvisionalDate string               `avro:"provisional_date"`
	Published       bool                 `avro:"published"`
}

// SearchContentDeleted represents event data for search-content-deleted
type SearchContentDeleted struct {
	URI         string `avro:"uri"`
	SearchIndex string `avro:"search_index"`
	TraceID     string `avro:"trace_id"`
}
