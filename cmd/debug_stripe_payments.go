package main

import (
	"fmt"
	"log"
	"os"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/account"
	"github.com/stripe/stripe-go/v76/paymentintent"
	"github.com/stripe/stripe-go/v76/transfer"
)

func main() {
	// Load Stripe key
	stripeKey := os.Getenv("STRIPE_SECRET_KEY")
	if stripeKey == "" {
		log.Fatal("❌ STRIPE_SECRET_KEY environment variable is required")
	}
	stripe.Key = stripeKey

	connectAccountID := os.Getenv("STRIPE_CONNECT_ACCOUNT")
	if connectAccountID == "" {
		log.Fatal("❌ STRIPE_CONNECT_ACCOUNT environment variable is required")
	}

	fmt.Println("🔍 Stripe Connect Payment Debugging")
	fmt.Printf("🔑 Using key: %s...%s\n", stripeKey[:12], stripeKey[len(stripeKey)-4:])
	fmt.Printf("🏢 Connect Account: %s\n", connectAccountID)

	// Check Connect account status
	checkConnectAccount(connectAccountID)

	// List recent payment intents
	listRecentPaymentIntents()

	// List transfers to the Connect account
	listTransfersToAccount(connectAccountID)

	// Provide troubleshooting guidance
	provideTroubleshootingGuidance()
}

func checkConnectAccount(accountID string) {
	fmt.Println("\n🏢 Checking Connect Account...")
	
	acct, err := account.GetByID(accountID, nil)
	if err != nil {
		fmt.Printf("❌ Failed to retrieve account: %v\n", err)
		fmt.Println("💡 Make sure your Connect account ID is correct and belongs to your platform")
		return
	}

	fmt.Printf("✅ Account found: %s\n", acct.ID)
	fmt.Printf("   Type: %s\n", acct.Type)
	fmt.Printf("   Country: %s\n", acct.Country)
	fmt.Printf("   Created: %d\n", acct.Created)
	fmt.Printf("   Charges enabled: %v\n", acct.ChargesEnabled)
	fmt.Printf("   Payouts enabled: %v\n", acct.PayoutsEnabled)
	fmt.Printf("   Details submitted: %v\n", acct.DetailsSubmitted)

	if !acct.ChargesEnabled {
		fmt.Println("⚠️  WARNING: Charges not enabled on this account!")
		fmt.Println("   This account cannot receive payments yet.")
	}

	if len(acct.Requirements.CurrentlyDue) > 0 {
		fmt.Println("📋 Requirements currently due:")
		for _, req := range acct.Requirements.CurrentlyDue {
			fmt.Printf("   - %s\n", req)
		}
	}
}

func listRecentPaymentIntents() {
	fmt.Println("\n💳 Recent Payment Intents (last 10)...")
	
	params := &stripe.PaymentIntentListParams{}
	params.Limit = stripe.Int64(10)
	
	iter := paymentintent.List(params)
	
	count := 0
	for iter.Next() {
		pi := iter.PaymentIntent()
		count++
		
		fmt.Printf("📄 %s - %s - €%.2f - %s\n", 
			pi.ID, 
			pi.Status, 
			float64(pi.Amount)/100,
			pi.Description)
		
		// Show transfer data if present
		if pi.TransferData != nil && pi.TransferData.Destination != nil {
			fmt.Printf("   🎯 Destination: %s\n", pi.TransferData.Destination.ID)
		}
		
		// Show application fee
		if pi.ApplicationFeeAmount > 0 {
			fmt.Printf("   💰 Platform fee: €%.2f\n", float64(pi.ApplicationFeeAmount)/100)
		}
		
		// Show metadata
		if len(pi.Metadata) > 0 {
			fmt.Printf("   📋 Metadata: ")
			for key, value := range pi.Metadata {
				fmt.Printf("%s=%s ", key, value)
			}
			fmt.Println()
		}
		
		fmt.Println()
	}
	
	if err := iter.Err(); err != nil {
		fmt.Printf("❌ Error listing payment intents: %v\n", err)
		return
	}
	
	if count == 0 {
		fmt.Println("📭 No payment intents found")
		fmt.Println("💡 Make sure you've created some payment intents first")
	}
}

func listTransfersToAccount(accountID string) {
	fmt.Printf("\n💸 Transfers to account %s...\n", accountID)
	
	params := &stripe.TransferListParams{}
	params.Destination = stripe.String(accountID)
	params.Limit = stripe.Int64(10)
	
	iter := transfer.List(params)
	
	count := 0
	for iter.Next() {
		t := iter.Transfer()
		count++
		
		fmt.Printf("💸 %s - €%.2f - %s - %s\n", 
			t.ID, 
			float64(t.Amount)/100,
			t.Description,
			t.Created)
		
		if t.SourceTransaction != nil {
			fmt.Printf("   🔗 Source: %s\n", t.SourceTransaction.ID)
		}
		
		fmt.Println()
	}
	
	if err := iter.Err(); err != nil {
		fmt.Printf("❌ Error listing transfers: %v\n", err)
		return
	}
	
	if count == 0 {
		fmt.Println("📭 No transfers found to this account")
		fmt.Println("💡 Transfers only happen when payments are successfully completed")
	}
}

func provideTroubleshootingGuidance() {
	fmt.Println("\n🔧 Troubleshooting Guide:")
	fmt.Println()
	
	fmt.Println("1. 📋 Payment Intent Status:")
	fmt.Println("   - 'requires_payment_method': Payment created but not attempted")
	fmt.Println("   - 'requires_confirmation': Payment method attached but not confirmed")
	fmt.Println("   - 'succeeded': Payment completed - transfers should appear")
	fmt.Println("   - 'canceled' or 'failed': Payment didn't complete")
	fmt.Println()
	
	fmt.Println("2. 🏢 Connect Account Issues:")
	fmt.Println("   - Account must have 'charges_enabled: true'")
	fmt.Println("   - Complete all required verification steps")
	fmt.Println("   - Check country/currency compatibility")
	fmt.Println()
	
	fmt.Println("3. 💸 Transfer Timing:")
	fmt.Println("   - Transfers happen automatically when payment succeeds")
	fmt.Println("   - Check both Platform and Connect dashboards")
	fmt.Println("   - Transfers appear in Connect account's 'Payments' section")
	fmt.Println()
	
	fmt.Println("4. 🧪 Testing Tips:")
	fmt.Println("   - Use test card 4242424242424242 for successful payments")
	fmt.Println("   - Complete the payment flow (don't just create intent)")
	fmt.Println("   - Check webhook events for payment.succeeded")
	fmt.Println()
	
	fmt.Println("5. 🔍 Where to Look:")
	fmt.Println("   - Platform Dashboard: All payment intents and transfers")
	fmt.Println("   - Connect Dashboard: Received payments and payouts")
	fmt.Println("   - Both show the same payments from different perspectives")
}

// Helper function to debug a specific payment intent
func debugSpecificPaymentIntent(paymentIntentID string) {
	fmt.Printf("\n🔍 Debugging Payment Intent: %s\n", paymentIntentID)
	
	pi, err := paymentintent.Get(paymentIntentID, nil)
	if err != nil {
		fmt.Printf("❌ Error retrieving payment intent: %v\n", err)
		return
	}
	
	fmt.Printf("Status: %s\n", pi.Status)
	fmt.Printf("Amount: €%.2f\n", float64(pi.Amount)/100)
	fmt.Printf("Currency: %s\n", pi.Currency)
	fmt.Printf("Created: %d\n", pi.Created)
	
	if pi.TransferData != nil {
		fmt.Printf("Transfer destination: %s\n", pi.TransferData.Destination.ID)
		if pi.TransferData.Amount > 0 {
			fmt.Printf("Transfer amount: €%.2f\n", float64(pi.TransferData.Amount)/100)
		}
	}
	
	if pi.ApplicationFeeAmount > 0 {
		fmt.Printf("Application fee: €%.2f\n", float64(pi.ApplicationFeeAmount)/100)
	}
	
	// Check if there are any charges
	if len(pi.Charges.Data) > 0 {
		fmt.Println("Charges:")
		for _, charge := range pi.Charges.Data {
			fmt.Printf("  - %s: %s (€%.2f)\n", charge.ID, charge.Status, float64(charge.Amount)/100)
		}
	}
}