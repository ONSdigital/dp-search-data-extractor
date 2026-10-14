package clients

import (
	"context"

	//nolint:staticcheck // work ongoing to replace datasetclient with a new implementation
	datasetclient "github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
)

//go:generate moq -out mock/datasetapi.go -pkg mock . DatasetClient

// DatasetApiClient defines the zebedee client
type DatasetClient interface {
	Checker(context.Context, *healthcheck.CheckState) error
	GetVersionMetadata(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, datasetID, edition, version string) (metadata datasetclient.Metadata, err error)
}
