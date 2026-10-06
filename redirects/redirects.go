package redirects

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"

	redirectAPI "github.com/ONSdigital/dis-redirect-api/sdk/go"
)

// GetPreviousURIs returns sorted, unique redirect source paths to the specified target URI.
func GetPreviousURIs(ctx context.Context, client redirectAPI.Clienter, targetURI, serviceAuthToken string) ([]string, error) {
	if client == nil {
		return nil, errors.New("redirect client is not configured")
	}
	query := url.Values{"to": {targetURI}}
	headers := http.Header{}
	if serviceAuthToken != "" {
		headers.Set(redirectAPI.Authorization, serviceAuthToken)
	}

	unique := make(map[string]struct{})
	for {
		redirects, err := client.GetRedirects(ctx, redirectAPI.Options{Headers: headers, Query: query})
		if err != nil {
			return nil, fmt.Errorf("get redirects for target URI %s: %w", targetURI, err)
		}
		for _, redirect := range redirects.RedirectList {
			if redirect.From != "" {
				unique[redirect.From] = struct{}{}
			}
		}
		if redirects.NextCursor == "0" {
			break
		}
		query.Set("cursor", redirects.NextCursor)
	}

	previousURIs := make([]string, 0, len(unique))
	for uri := range unique {
		previousURIs = append(previousURIs, uri)
	}
	sort.Strings(previousURIs)
	return previousURIs, nil
}
