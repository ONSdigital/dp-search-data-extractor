dp-search-data-extractor
================
Service to retrieve published data to be used to update a search index
This service calls /publisheddata endpoint on zebedee.

This service listens to the "content-published" kafka topic for events of type contentPublishedEvent e.g. 
"data": {
  "contentPublishedEvent": {
  "CollectionID": "",
  "DataType": "",
  "URI": "businessindustryandtrade"
  }
}

This service takes the URI, from the consumed event, and calls /publisheddata endpoint on zebedee. It passes in the URI as a path parameter e.g.
http://localhost:8082/publisheddata?uri=businessindustryandtrade
This service then logs whether Zebedee has retrieved the data successfully.

###  healthcheck

        make debug and then 
        check http://localhost:25800/health

**TODO** : we are simply writing the retrieved content to a file for now but in future this will be passed on to another service via an as yet unwritten kafka producer.

### Getting started

* Run `make debug`

The service runs in the background consuming messages from Kafka.
An example event can be created using the helper script, `make produce`.

### Dependencies

* golang 1.16.x
* Running instance of zebedee
* Requires running…
  * [kafka](https://github.com/ONSdigital/dp/blob/main/guides/INSTALLING.md#prerequisites)
* No further dependencies other than those defined in `go.mod`

### Configuration

| Environment variable           | Default                  | Description
| ----------------------------   | -------------------------| -----------
| BIND_ADDR                      | localhost:25800          | The host and port to bind to
| GRACEFUL_SHUTDOWN_TIMEOUT      | 5s                       | The graceful shutdown timeout in seconds (`time.Duration` format)
| HEALTHCHECK_INTERVAL           | 30s                      | Time between self-healthchecks (`time.Duration` format)
| HEALTHCHECK_CRITICAL_TIMEOUT   | 90s                      | Time to wait until an unhealthy dependent propagates its state to make this app unhealthy (`time.Duration` format)
| KAFKA_ADDR                     | "localhost:9092"         | The address of Kafka (accepts list)
| KAFKA_OFFSET_OLDEST            | true                     | Start processing Kafka messages in order from the oldest in the queue
| KAFKA_VERSION                  | `1.0.2`                  | The version of Kafka
| KAFKA_NUM_WORKERS              | 1                        | The maximum number of parallel kafka consumers
| KAFKA_SEC_PROTO                | _unset_   (only `TLS`)   | if set to `TLS`, kafka connections will use TLS
| KAFKA_SEC_CLIENT_KEY           | _unset_                  | PEM [2] for the client key (optional, used for client auth) [[1]](#notes_1)
| KAFKA_SEC_CLIENT_CERT          | _unset_                  | PEM [2] for the client certificate (optional, used for client auth) [[1]](#notes_1)
| KAFKA_SEC_CA_CERTS             | _unset_                  | PEM [2] of CA cert chain if using private CA for the server cert [[1]](#notes_1)
| KAFKA_SEC_SKIP_VERIFY          | false                    | ignore server certificate issues if set to `true` [[1]](#notes_1)
| KAFKA_CONTENT_UPDATED_GROUP    | dp-search-data-extractor | The consumer group this application to consume content-updated messages
| KAFKA_CONTENT_UPDATED_TOPIC    | content-updated          | The name of the topic to consume messages from
| ZEBEDEE_URL                    | `http://localhost:8082`  | The URL for the Zebedee
| KEYWORDS_LIMITS                | -1                       | The keywords allowed, default no limit
| DATASET_API_URL                | `http://localhost:22000` | The URL for the DatasetAPI
| SERVICE_AUTH_TOKEN             | ""                       | The user auth token for the DatasetAPI


**Notes:**

1. <a name="notes_1">For more info, see the [kafka TLS examples documentation](https://github.com/ONSdigital/dp-kafka/tree/main/examples#tls)</a>

### Healthcheck

 The `/health` endpoint returns the current status of the service. Dependent services are health checked on an interval defined by the `HEALTHCHECK_INTERVAL` environment variable.

 On a development machine a request to the health check endpoint can be made by:

 `curl localhost:25800/health`

### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Copyright © 2022, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.