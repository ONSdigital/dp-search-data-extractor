# Feature: Data extractor should listen to the relevant topic and publish extracted data for Dataset API datasets


#   Scenario: When searching for the extracted dataset generic metadata I get the expected result
#     Given dp-dataset-api is healthy
#     And zebedee is healthy
#     And the following metadata with dataset-id "cphi01", edition "timeseries" and version "version" is available in dp-dataset-api
#     """
#     {
#         "release_date": "releasedate",
#         "title": "title",
#         "description": "description",
#         "keywords": ["keyword1", "keyword2"]
#     }
#     """

#     When the service starts
#     And this content-updated event is queued, to be consumed
#       | URI                                                            | DataType             |   CollectionID  |
#       | /datasets/cphi01/editions/timeseries/versions/version/metadata | datasets             |    123          |

#     Then I should receive a kafka event to search-data-import topic with the following fields
#       | UID               | Edition    | DataType             | SearchIndex | CDID | DatasetID | Keywords            | ReleaseDate | Summary     | Title | Topics |
#       | cphi01-timeseries | timeseries | dataset_landing_page | ONS         |      | cphi01    | [keyword1,keyword2] | releasedate | description | title |        |
