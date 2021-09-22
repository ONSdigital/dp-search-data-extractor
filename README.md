dp-search-data-extractor
================
Service to retrieve published data to be used to update a search index
This service calls /publisheddata endpoint on zebedee.

This service listens to the "content-published" kafka topic for events of type contentPublishedEvent e.g. 
"data": {
  "contentPublishedEvent": {
  "CollectionID": "",
  "DataType": "",
  "URL": "businessindustryandtrade"
  }
}

This service takes the URL, from the consumed event, and calls /publisheddata endpoint on zebedee. It passes in the URL as a path parameter e.g.
http://localhost:8082/publisheddata?uri=businessindustryandtrade
This service then logs whether Zebedee has retrieved the data successfully.

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

| Environment variable         | Default                           | Description
| ---------------------------- | --------------------------------- | -----------
| BIND_ADDR                    | localhost:25800                   | The host and port to bind to
| GRACEFUL_SHUTDOWN_TIMEOUT    | 5s                                | The graceful shutdown timeout in seconds (`time.Duration` format)
| HEALTHCHECK_INTERVAL         | 30s                               | Time between self-healthchecks (`time.Duration` format)
| HEALTHCHECK_CRITICAL_TIMEOUT | 90s                               | Time to wait until an unhealthy dependent propagates its state to make this app unhealthy (`time.Duration` format)
| KAFKA_ADDR                   | "localhost:9092"                  | The address of Kafka (accepts list)
| KAFKA_OFFSET_OLDEST          | true                              | Start processing Kafka messages in order from the oldest in the queue
| KAFKA_NUM_WORKERS            | 1                                 | The maximum number of parallel kafka consumers
| KAFKA_CONTENT_PUBLISHED_GROUP  | dp-search-data-extractor          | The consumer group this application to consume ImageUploaded messages
| KAFKA_CONTENT_PUBLISHED_TOPIC  | content-published                 | The name of the topic to consume messages from

### Healthcheck

 The `/health` endpoint returns the current status of the service. Dependent services are health checked on an interval defined by the `HEALTHCHECK_INTERVAL` environment variable.

 On a development machine a request to the health check endpoint can be made by:

 `curl localhost:25800/health`


### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Copyright © 2021, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.

