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
