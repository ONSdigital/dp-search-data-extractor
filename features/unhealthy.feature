Feature: Search data extractor unhealthy

  # This file validates that events are not consumed by the service if a dependency is not healthy

  Scenario: Not consuming events, because a dataset API is not healthy
    Given dp-dataset-api is unhealthy
    And zebedee is healthy

    When the service starts
    And I send a kafka event to content published topic
      | URI      | DataType | CollectionID |
      | some_uri | legacy   | 123          |

    Then no search-data-import events are produced
