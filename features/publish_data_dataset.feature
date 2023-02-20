# Feature: Data extractor should listen to the relevant topic and publish extracted data for Dataset API

#   Scenario: When searching for dataset-api extracted data I get the expected result
#     Given I send a kafka event to content published topic
#       | URI                                                            | DataType             |   CollectionID  |
#       | /datasets/cphi01/editions/timeseries/versions/version/metadata | datasets             |    123          |
#     When The kafka event is processed
#     Then I should receive a kafka event to search-data-import topic with the following fields
#       | UID               | Edition    | DataType             | SearchIndex | CDID | DatasetID | Keywords            | ReleaseDate | Summary     | Title | Topics |
#       | cphi01-timeseries | timeseries | dataset_landing_page | ONS         |      | cphi01    | [keyword1,keyword2] | releasedate | description | title |        |
