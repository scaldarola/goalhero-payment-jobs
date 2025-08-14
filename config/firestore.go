package config

import (
	"cloud.google.com/go/firestore"
)

// GetFirestoreClient returns the initialized Firestore client
// This is set by handlers during initialization to avoid circular imports
var GetFirestoreClient func() *firestore.Client

// SetFirestoreClient allows handlers to set the client reference
func SetFirestoreClient(client *firestore.Client) {
	GetFirestoreClient = func() *firestore.Client {
		return client
	}
}

// FirestoreClient is a convenience accessor
func FirestoreClient() *firestore.Client {
	if GetFirestoreClient != nil {
		return GetFirestoreClient()
	}
	return nil
}