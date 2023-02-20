package service

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
)

//go:generate moq -out mock/server.go -pkg mock . HTTPServer
//go:generate moq -out mock/healthCheck.go -pkg mock . HealthChecker

// HTTPServer defines the required methods from the HTTP server
type HTTPServer interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}

// HealthChecker defines the required methods from Healthcheck
type HealthChecker interface {
	Handler(w http.ResponseWriter, req *http.Request)
	Start(ctx context.Context)
	Stop()
	AddAndGetCheck(name string, checker healthcheck.Checker) (check *healthcheck.Check, err error)
	Subscribe(s healthcheck.Subscriber, checks ...*healthcheck.Check)
}
