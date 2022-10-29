package util

import (
	"context"
	"fmt"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
)

func GetProjectID(ctx context.Context) (string, error) {
	credentials, err := google.FindDefaultCredentials(ctx, compute.ComputeScope)
	if err != nil {
		return "", fmt.Errorf("google.FindDefaultCredentials failed; %w", err)
	}
	return credentials.ProjectID, nil
}
