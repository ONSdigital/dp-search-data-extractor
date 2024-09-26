# Search Contract

We expect upstream services to behave in a predictable way in order to reduce the development effort of adding them to the search service.

To do this, they must provide the following services which are detailed in our specifications:

- [get resource endpoint and get resource uris endpoint](./upstream.yml)
- [content-updated kafka event](../../specification.yml)
- [content-deleted kafka event](https://github.com/ONSdigital/dp-search-data-importer/blob/develop/specification.yml)

The endpoints specified will be paths that we will store for re-use. For example:

- /data (get one resource)
- /publishedindex (get all resource uris)

`content-updated` messages will need to use the `callback` data type, alongside providing a `service_id` (matching below).

To add your service to the search index, you will need to contact the development team and provide them with:

- get resource endpoint
- get resource uris endpoint
- service name
- service id

Endpoints can either be provided as a path of dp-api-router or as a fully qualified URL.
