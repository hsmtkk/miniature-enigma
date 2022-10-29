package main

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
)

func firestoreSave(ctx context.Context, projectID, collection string, data map[string]interface{}) error {
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("firestore.NewClient failed; %w", err)
	}
	defer client.Close()

	if _, _, err := client.Collection(collection).Add(ctx, data); err != nil {
		return fmt.Errorf("firestore.CollectionRef.Add failed; %w", err)
	}
	return nil
}
