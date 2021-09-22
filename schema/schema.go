package schema

import (
	"github.com/ONSdigital/go-ns/avro"
)

var contentPublishedEvent = `{
  "type": "record",
  "name": "content-published",
  "fields": [
    {"name": "url", "type": "string", "default": ""},
    {"name": "data_type", "type": "string", "default": ""},
    {"name": "collection_id", "type": "string", "default": ""}
  ]
}`

// ContentPublishedEvent is the Avro schema for Content Published messages.
var ContentPublishedEvent = &avro.Schema{
	Definition: contentPublishedEvent,
}

var searchDataImportEvent = `{
  "type": "record",
  "name": "search-data-import",
  "fields": [
    {"name": "data_type", "type": "string", "default": ""},
    {"name": "job_id", "type": "string", "default": ""},
    {"name": "search_index", "type": "string", "default": ""},
    {"name": "cdid", "type": "string", "default": ""},
    {"name": "dataset_id", "type": "string", "default": ""},
    {"name": "keywords", "type": "string", "default": ""},
    {"name": "meta_description", "type": "int", "default": ""},
    {"name": "release_date", "type": "string", "default": ""},
    {"name": "summary", "type": "string", "default": ""},
    {"name": "title", "type": "string", "default": ""}
    {"name": "trace_id", "type": "string", "default": ""}
  ]
}`

// SearchDataImportEvent the Avro schema for search-data-import messages.
var SearchDataImportEvent = &avro.Schema{
	Definition: searchDataImportEvent,
}
