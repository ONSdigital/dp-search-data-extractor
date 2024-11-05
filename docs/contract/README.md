# Search Contract

We expect upstream services to behave in a predictable way in order to reduce the development effort of adding them to the search service.

To do this, they must provide the following services which are detailed in our specifications:

- [get resources endpoint endpoint](https://github.com/ONSdigital/dis-search-upstream-stub/blob/develop/specification.yml)
- [content-updated kafka event](../../specification.yml)
- [content-deleted kafka event](https://github.com/ONSdigital/dp-search-data-importer/blob/develop/specification.yml)

The endpoints specified will be paths that we will store for re-use. For example:

- /publishedindex (get all resources)

To add your service to the search index, you will need to contact the development team and provide them with:

- get resources endpoint
- service name
- service id

Endpoints can either be provided as a path of dp-api-router or as a fully qualified URL.

## Upstream actions

Here is a list of actions we expect from upstream systems when certain events occur. These are all based on when that
action is 'published' to the public.

### Resource has been updated

When a resource has been updated, issue an event to the topic `search-content-updated` with your full metadata.

### Resource has been deleted

When a resource has been updated, issue an event to the topic `search-content-deleted` with your resource's uri.

### Resource has been created

When a resource has been created, issue an event to the topic `search-content-updated` with your full metadata.

### Resource has been renamed

When a resource has been renamed (the `uri` has changed), issue an event to the topic `search-content-updated` using the
`uri_old` property to indicate the old uri to be deleted.
