package config

import (
	"context"
	"log"
	"os"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
)

var firestoreClient *firestore.Client

// InitConfig initializes Firestore connection
func InitConfig() {
	log.Println("🔧 Initializing Firestore...")
	
	ctx := context.Background()
	
	// Get credentials path
	credentialsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credentialsPath == "" {
		credentialsPath = "./auth/firebase_credentials.json"
	}

	if _, err := os.Stat(credentialsPath); os.IsNotExist(err) {
		log.Printf("⚠️ Firestore credentials not found at %s", credentialsPath)
		return
	}

	// Get project ID from environment or default
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		projectID = "goalhero-dev" // Default project ID
	}

	// Initialize Firestore client
	client, err := firestore.NewClient(ctx, projectID, option.WithCredentialsFile(credentialsPath))
	if err != nil {
		log.Fatalf("Failed to initialize Firestore: %v", err)
	}

	firestoreClient = client
	log.Println("✅ Firestore initialized successfully")
}

// FirestoreClient returns the Firestore client
func FirestoreClient() *firestore.Client {
	if firestoreClient == nil {
		InitConfig()
	}
	return firestoreClient
}