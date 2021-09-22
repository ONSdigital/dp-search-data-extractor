package clients

import (
	"context"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
)

//go:generate moq -out mock/zebedee.go -pkg mock . ZebedeeClient

// ZebedeeClient defines the zebedee client
type ZebedeeClient interface {
	Checker(context.Context, *healthcheck.CheckState) error
	GetPublishedData(ctx context.Context, uriString string) ([]byte, error)
}
