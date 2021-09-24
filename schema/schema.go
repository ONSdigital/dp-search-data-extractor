package schema

import (
	"github.com/ONSdigital/go-ns/avro"
)

var contentPublished = `{
  "type": "record",
  "name": "content-published",
  "fields": [
    {"name": "url", "type": "string", "default": ""},
    {"name": "data_type", "type": "string", "default": ""},
    {"name": "collection_id", "type": "string", "default": ""}
  ]
}`

// ContentPublishedSchema is the Avro schema for Content Published messages.
var ContentPublishedSchema = &avro.Schema{
	Definition: contentPublished,
}

var searchDataImport = `{
  "type": "record",
  "name": "search-data-import",
  "fields": [
    {"name": "type", "type": "string", "default": ""},
    {"name": "job_id", "type": "string", "default": ""},
    {"name": "search_index", "type": "string", "default": ""},
    {"name": "cdid", "type": "string", "default": ""},
    {"name": "dataset_id", "type": "string", "default": ""},
    {"name": "keywords", "type": "string", "default": ""},
    {"name": "meta_description", "type": "string", "default": ""},
    {"name": "release_date", "type": "string", "default": ""},
    {"name": "summary", "type": "string", "default": ""},
    {"name": "title", "type": "string", "default": ""},
    {"name": "trace_id", "type": "string", "default": ""}
  ]
}`

// SearchDataImportEvent the Avro schema for search-data-import messages.
var SearchDataImportEvent = &avro.Schema{
	Definition: searchDataImport,
}
