@Generic
Feature:Data extractor should listen to the relevant topic and publish extracted data for Generic Upstream Services

  Background: Service setup
    Given dp-dataset-api is healthy
    And zebedee is healthy
    And the service starts

  Scenario: When a valid `search-content-updated` of type release event is consumed, a `search-data-import` event is published
    Given this "search-content-updated" json event is queued, to be consumed
    """
    {
      "uri":           "/some/uri",
      "title":         "title",
      "edition":       "something",
      "content_type":   "release",
      "search_index":   "ons",
      "cdid":          "123",
      "dataset_id":     "456"
    }
    """
    Then this search-data-import event is sent
    """
    {
      "uid":         "/some/uri",
      "uri":         "/some/uri",
      "title":       "title",
      "edition":     "something",
      "data_type":    "release",
      "search_index": "ons",
      "cdid":        "123",
      "dataset_id":   "456",
      "keywords":    [],
      "topics":      []
    }
    """


  Scenario: When a valid `search-content-updated` that is not a release event is consumed, a `search-data-import` event is published
    Given this "search-content-updated" json event is queued, to be consumed
    """
    {
      "uri":           "/some/uri",
      "title":         "title",
      "edition":       "something",
      "content_type":   "article",
      "search_index":   "ons",
      "cdid":          "123",
      "dataset_id":     "456"
    }
    """
    Then this search-data-import event is sent
    """
    {
      "uid":         "/some/uri",
      "uri":         "/some/uri",
      "title":       "title",
      "edition":     "something",
      "data_type":    "article",
      "search_index": "ons",
      "cdid":        "123",
      "dataset_id":   "456",
      "keywords":    [],
      "topics":      []
    }
    """
    Then no search-content-deleted events are produced


  Scenario: When a valid `search-content-updated` with uri_old in event is consumed, a `search-data-import` event is published
    Given this "search-content-updated" json event is queued, to be consumed
    """
    {
      "uri":           "/some/uri",
      "uri_old":       "/my/old/uri",
      "title":         "title",
      "edition":       "something",
      "content_type":   "article",
      "cdid":          "123",
      "dataset_id":    "456",
      "trace_id":      "trace1234"
    }
    """
    Then this search-data-import event is sent
    """
    {
      "uid":         "/some/uri",
      "uri":         "/some/uri",
      "title":       "title",
      "trace_id":     "trace1234",
      "search_index": "ons",
      "edition":     "something",
      "data_type":   "article",
      "cdid":        "123",
      "dataset_id":   "456",
      "keywords":    [],
      "topics":      []
    }
    """
    And this search-content-deleted event is sent
    """
    {
      "uri": "/my/old/uri",
      "collection_id": "123",
      "search_index": "ons",
      "trace_id": "trace1234"
    }
    """