@Generic
Feature:Data extractor should listen to the relevant topic and publish extracted data for Generic Upstream Services

  Background: Service setup
    Given dp-dataset-api is healthy
    And zebedee is healthy
    And the service starts

  Scenario: When a valid `search-content-updated` of type release event is consumed, a `search-data-import` event is published
    Given this "search-content-updated" event is queued, to be consumed
    """
    {
      "URI":           "/some/uri",
      "Title":         "title",
      "Edition":       "something",
      "ContentType":   "release",
      "SearchIndex":   "ons",
      "CDID":          "123",
      "DatasetID":     "456"
    }
    """
    Then this search-data-import event is sent
    """
    {
      "UID":         "/some/uri",
      "URI":         "/some/uri",
      "Title":       "title",
      "Edition":     "something",
      "DataType":    "release",
      "SearchIndex": "ons",
      "CDID":        "123",
      "DatasetID":   "456",
      "Keywords":    [],
      "Topics":      []
    }
    """


  Scenario: When a valid `search-content-updated` that is not a release event is consumed, a `search-data-import` event is published
    Given this "search-content-updated" event is queued, to be consumed
    """
    {
      "URI":           "/some/uri",
      "Title":         "title",
      "Edition":       "something",
      "ContentType":   "article",
      "SearchIndex":   "ons",
      "CDID":          "123",
      "DatasetID":     "456"
    }
    """
    Then this search-data-import event is sent
    """
    {
      "UID":         "/some/uri",
      "URI":         "/some/uri",
      "Title":       "title",
      "Edition":     "something",
      "DataType":    "article",
      "SearchIndex": "ons",
      "CDID":        "123",
      "DatasetID":   "456",
      "Keywords":    [],
      "Topics":      []
    }
    """
    Then no search-content-deleted events are produced


  Scenario: When a valid `search-content-updated` with uri_old in event is consumed, a `search-data-import` event is published
    Given this "search-content-updated" event is queued, to be consumed
    """
    {
      "URI":           "/some/uri",
      "URIOld":        "/my/old/uri",
      "Title":         "title",
      "Edition":       "something",
      "ContentType":   "article",
      "CDID":          "123",
      "DatasetID":     "456",
      "SearchIndex":   "ons",
      "TraceID":       "trace1234"
    }
    """
    Then this search-data-import event is sent
    """
    {
      "UID":         "/some/uri",
      "URI":         "/some/uri",
      "Title":       "title",
      "TraceID":     "trace1234",
      "SearchIndex": "ons",
      "Edition":     "something",
      "DataType":    "article",
      "CDID":        "123",
      "DatasetID":   "456",
      "Keywords":    [],
      "Topics":      []
    }
    """
    And this search-content-deleted event is sent
    """
    {
      "URI":         "/my/old/uri",
      "SearchIndex": "ons",
      "TraceID":     "trace1234"
    }
    """
