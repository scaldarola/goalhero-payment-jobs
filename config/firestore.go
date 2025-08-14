package config

import (
	"cloud.google.com/go/firestore"
)

// Global Firestore client reference
var globalFirestoreClient *firestore.Client

// SetFirestoreClient allows setting the global client reference
func SetFirestoreClient(client *firestore.Client) {
	globalFirestoreClient = client
}

// FirestoreClient returns the global Firestore client
func FirestoreClient() *firestore.Client {
	return globalFirestoreClient
}