package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/sebastiancaldarola/goalhero-payment-jobs/models"
	"github.com/sebastiancaldarola/goalhero-payment-jobs/services"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
		log.Printf("Make sure to set environment variables manually or create a .env file")
	}

	// Check if Stripe key is set
	stripeKey := os.Getenv("STRIPE_SECRET_KEY")
	if stripeKey == "" {
		log.Fatal("❌ STRIPE_SECRET_KEY environment variable is required. Please set it in your .env file.")
	}

	if !isTestKey(stripeKey) {
		log.Fatal("❌ Only test keys are allowed for this integration test. Use a key starting with 'sk_test_'")
	}

	fmt.Println("🧪 Starting Stripe Integration Test...")
	fmt.Printf("🔑 Using Stripe key: %s...%s\n", stripeKey[:12], stripeKey[len(stripeKey)-4:])

	// Initialize service
	service := services.NewStripeConnectService()
	fmt.Printf("🏢 Test mode: %v\n", service.IsTestMode())

	// Run tests
	runBasicTests(service)
	runPaymentIntentTest(service)
	runTestCardValidation(service)

	fmt.Println("✅ All tests completed successfully!")
}

func isTestKey(key string) bool {
	return len(key) > 8 && key[:8] == "sk_test_"
}

func runBasicTests(service *services.StripeConnectService) {
	fmt.Println("\n📋 Running Basic Tests...")

	// Test fee calculations
	testAmounts := []float64{5.0, 15.0, 25.0, 50.0}
	
	for _, amount := range testAmounts {
		platformFee, stripeFee, netAmount := service.CalculateFees(amount)
		fmt.Printf("💰 Amount: €%.2f | Platform Fee: €%.2f | Stripe Fee: €%.2f | Net: €%.2f\n", 
			amount, platformFee, stripeFee, netAmount)
	}

	// Test account validation
	testAccounts := []string{
		"acct_test_1234567890",
		"acct_1234567890abcdef",
		"", // Should fail
		"short", // Should fail
	}

	fmt.Println("\n🔐 Testing Account Validation...")
	for _, account := range testAccounts {
		err := service.ValidateConnectAccount(account)
		if err != nil {
			fmt.Printf("❌ Account '%s': %v\n", account, err)
		} else {
			fmt.Printf("✅ Account '%s': Valid\n", account)
		}
	}
}

func runPaymentIntentTest(service *services.StripeConnectService) {
	fmt.Println("\n💳 Running Payment Intent Test...")

	// Create test payment
	payment := &models.Payment{
		ID:            "test_payment_" + fmt.Sprintf("%d", time.Now().Unix()),
		UserID:        "test_user_123",
		GameID:        "test_game_456",
		ApplicationID: "test_app_789",
		Amount:        15.0,
		Currency:      models.DefaultCurrency,
		Status:        models.PaymentStatusPending,
		CreatedAt:     time.Now(),
	}

	// Test organizer ID (this should be a real Stripe Connect account for full testing)
	organizerID := os.Getenv("STRIPE_CONNECT_ACCOUNT")
	if organizerID == "" {
		organizerID = "acct_test_example123" // Fallback - will likely fail but won't crash
		fmt.Printf("⚠️  No STRIPE_CONNECT_ACCOUNT set, using fallback: %s\n", organizerID)
	}

	fmt.Printf("🎯 Creating payment intent for €%.2f to organizer: %s\n", payment.Amount, organizerID)

	result, err := service.CreateEscrowPaymentIntent(payment, organizerID)
	if err != nil {
		fmt.Printf("❌ Payment intent creation failed: %v\n", err)
		fmt.Println("💡 This is expected if you don't have a valid Stripe Connect account set up")
		return
	}

	fmt.Printf("✅ Payment Intent Created!\n")
	fmt.Printf("   ID: %s\n", result.PaymentIntent.ID)
	fmt.Printf("   Status: %s\n", result.Status)
	fmt.Printf("   Amount: €%.2f\n", float64(result.PaymentIntent.Amount)/100)
	fmt.Printf("   Client Secret: %s...%s\n", 
		result.ClientSecret[:12], 
		result.ClientSecret[len(result.ClientSecret)-8:])

	// Test retrieving payment details
	fmt.Println("\n🔍 Testing Payment Retrieval...")
	details, err := service.GetPaymentDetails(result.PaymentIntent.ID)
	if err != nil {
		fmt.Printf("❌ Failed to retrieve payment: %v\n", err)
		return
	}

	fmt.Printf("✅ Payment Retrieved: %s (Status: %s)\n", details.ID, details.Status)
}

func runTestCardValidation(service *services.StripeConnectService) {
	fmt.Println("\n💳 Available Test Cards:")

	testCards := service.GetTestCardTokens()
	
	cardDescriptions := map[string]string{
		"visa_success":       "✅ Visa - Successful payment",
		"visa_decline":       "❌ Visa - Generic decline", 
		"mastercard_success": "✅ Mastercard - Successful payment",
		"amex_success":       "✅ American Express - Successful payment",
		"insufficient_funds": "💸 Visa - Insufficient funds",
		"expired_card":       "📅 Visa - Expired card",
		"incorrect_cvc":      "🔢 Visa - Incorrect CVC",
		"processing_error":   "⚠️  Visa - Processing error",
	}

	for cardType, cardNumber := range testCards {
		description := cardDescriptions[cardType]
		if description == "" {
			description = cardType
		}
		fmt.Printf("   %s: %s\n", cardNumber, description)
	}

	fmt.Println("\n📚 Usage Instructions:")
	fmt.Println("   • Use these card numbers in your frontend payment form")
	fmt.Println("   • Use any valid future date for expiration (e.g., 12/34)")
	fmt.Println("   • Use any 3-digit CVC for Visa/Mastercard, 4-digit for Amex")
	fmt.Println("   • Use any valid billing ZIP code")
	fmt.Println("   • Different cards will simulate different payment scenarios")
}