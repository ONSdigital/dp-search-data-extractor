Feature: Search data extractor unhealthy


  Scenario: Not consuming events, because a dataset API is not healthy
    Given dp-dataset-api is unhealthy
    And zebedee is healthy

    When the service starts
    And this content-updated event is queued, to be consumed
      | URI      | DataType | CollectionID |
      | some_uri | legacy   | 123          |

    Then no search-data-import events are produced


  Scenario: Not consuming events, because a zebedee is not healthy
    Given dp-dataset-api is healthy
    And zebedee is unhealthy

    When the service starts
    And this content-updated event is queued, to be consumed
      | URI      | DataType | CollectionID |
      | some_uri | legacy   | 123          |

    Then no search-data-import events are produced
