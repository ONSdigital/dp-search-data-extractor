Feature: Data extractor should listen to the relevant topic and publish extracted data
    Scenario: When searching for the extracted data I get the expected result
        Given I send a kafka event to content published topic
         | URI                | DataType             |   CollectionID  |
         | some_uri           | some_datatype        |    123          |
        When The kafka event is processed
        Then I should receive the published data