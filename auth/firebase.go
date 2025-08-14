package auth

import (
	"context"
	"flag"
	"log"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

var FirebaseAuthClient *auth.Client

func isTestMode() bool {
	return flag.Lookup("test.v") != nil || os.Getenv("GO_ENV") == "test"
}

func init() {
	if !isTestMode() {
		InitFirebase()
	}
}

func InitFirebase() {
	opt := option.WithCredentialsFile("auth/firebase_credentials.json")

	ctx := context.Background()
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("🔥 Error al inicializar Firebase: %v", err)
	}

	client, err := app.Auth(ctx)
	if err != nil {
		log.Fatalf("🔥 Error al obtener cliente de Auth: %v", err)
	}

	FirebaseAuthClient = client
	log.Println("✅ Firebase Auth inicializado")
}
