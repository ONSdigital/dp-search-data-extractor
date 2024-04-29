dp-search-data-extractor
================

Service to retrieve published data to be used to update a search index
This service calls /publisheddata endpoint on [zebedee](https://github.com/ONSdigital/zebedee) and metadata endpoint on [dataset API](https://github.com/ONSdigital/dp-dataset-api).

This service listens to the `content-updated` kafka topic for events of type contentUpdatedEvent e.g.
see [schemas](schema) package.

This service takes the uri, from the consumed event, and either calls ...

1. ... /publisheddata endpoint on zebedee. It passes in the URI as a path parameter e.g.
  http://localhost:8082/publisheddata?uri=businessindustryandtrade
1. ... /datasets/<id>/editions/<edition>/versions/<version>/metadata endpoint on dataset API, e.g.
  http://localhost:22000/datasets/CPIH01/editions/timeseries/versions/1/metadata

See [search service architecture docs here](https://github.com/ONSdigital/dp-search-api/tree/develop/architecture#search-service-architecture)

### Getting started

* Run `make debug`
* Run `make help` to see full list of make targets

The service runs in the background consuming messages from Kafka.
An example event can be created using the helper script, `make produce`.

### Dependencies

* golang 1.20.x
* Running instance of zebedee
* Requires running…
  * [kafka](https://github.com/ONSdigital/dp/blob/main/guides/INSTALLING.md#prerequisites)
* No further dependencies other than those defined in `go.mod`

### Configuration

| Environment variable           | Default                  | Description
| ----------------------------   | -------------------------| -----------
| BIND_ADDR                      | localhost:25800          | The host and port to bind to
| DATASET_API_URL                | `http://localhost:22000` | The URL for the DatasetAPI
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
| KAFKA_PRODUCER_TOPIC           | search-data-import       | The name of the topic to produce messages to
| KEYWORDS_LIMITS                | -1                       | The keywords allowed, default no limit
| SERVICE_AUTH_TOKEN             | _unset_                  | The user auth token for the DatasetAPI
| STOP_CONSUMING_ON_UNHEALTHY    | true                     | Application stops consuming kafka messages if application is in unhealthy state
| ZEBEDEE_URL                    | `http://localhost:8082`  | The URL for the Zebedee

**Notes:**

1. <a name="notes_1">For more info, see the [kafka TLS examples documentation](https://github.com/ONSdigital/dp-kafka/tree/main/examples#tls)</a>

### Healthcheck

 The `/health` endpoint returns the current status of the service. Dependent services are health checked on an interval defined by the `HEALTHCHECK_INTERVAL` environment variable.

 On a development machine a request to the health check endpoint can be made by:

 `curl localhost:25800/health`

### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Copyright © 2024, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
