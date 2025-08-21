# Stripe Integration Testing Guide

This guide explains how to test your Stripe integration using test API keys and test cards.

## Setup

### 1. Environment Configuration

1. Copy the example environment file:
   ```bash
   cp .env.example .env
   ```

2. Fill in your Stripe test credentials in `.env`:
   ```bash
   STRIPE_SECRET_KEY=sk_test_your_actual_test_key_here
   STRIPE_CONNECT_ACCOUNT=acct_test_your_connect_account_here
   STRIPE_TEST_MODE=true
   ```

### 2. Get Your Stripe Test Keys

1. Visit [Stripe Dashboard](https://dashboard.stripe.com/)
2. Toggle to "Test mode" (top right switch)
3. Go to Developers → API keys
4. Copy your "Secret key" (starts with `sk_test_`)
5. For Connect testing, you'll also need a Connect account ID

## Running Tests

### Unit Tests
Run the existing unit tests:
```bash
go test ./services/... -v
```

### Integration Tests
Run tests that make actual API calls to Stripe:
```bash
STRIPE_SECRET_KEY=sk_test_... go test ./services/... -v -run Integration
```

### Manual Integration Test
Run the integration test script:
```bash
go run test_stripe_integration.go
```

This script will:
- ✅ Validate your environment setup
- 💰 Test fee calculations
- 🔐 Test account validation
- 💳 Create a real payment intent (if Connect account is configured)
- 📋 Show available test cards

## Test Cards

Use these test card numbers in your frontend:

### Successful Payments
- **4242 4242 4242 4242** - Visa (most common for testing)
- **5555 5555 5555 4444** - Mastercard
- **3782 8224 6310 005** - American Express

### Failed Payments
- **4000 0000 0000 0002** - Generic decline
- **4000 0000 0000 9995** - Insufficient funds
- **4000 0000 0000 0069** - Expired card
- **4000 0000 0000 0127** - Incorrect CVC
- **4000 0000 0000 0119** - Processing error

### Test Details
- **Expiration**: Use any valid future date (e.g., 12/34)
- **CVC**: Any 3 digits for Visa/MC, 4 digits for Amex
- **ZIP**: Any valid postal code

## Payment Flow Testing

### 1. Create Payment Intent
```bash
curl -X POST http://localhost:8080/api/payments \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 15.0,
    "game_id": "test_game_123",
    "organizer_id": "acct_test_..."
  }'
```

### 2. Complete Payment (Frontend)
Use the returned `client_secret` with Stripe Elements or mobile SDKs to complete the payment.

### 3. Verify Payment
```bash
curl -X GET http://localhost:8080/api/payments/{payment_id}
```

## Connect Account Testing

For full testing, you'll need a Stripe Connect test account:

1. Create a Connect platform in test mode
2. Create an Express or Standard account
3. Use the account ID (starts with `acct_test_` or `acct_`) in your environment

## Webhook Testing

### Local Development
1. Install Stripe CLI: `stripe login`
2. Forward webhooks: `stripe listen --forward-to localhost:8080/webhooks/stripe`
3. Use the webhook secret in your `.env` file

### Test Webhooks
```bash
stripe trigger payment_intent.succeeded
stripe trigger payment_intent.payment_failed
```

## Common Test Scenarios

### ✅ Successful Payment Flow
1. Create payment intent with valid amount (€5-50)
2. Use `4242424242424242` test card
3. Verify payment succeeds and funds are held in escrow
4. Test releasing funds to organizer

### ❌ Failed Payment Handling
1. Create payment intent
2. Use `4000000000000002` (decline) test card
3. Verify payment fails gracefully
4. Test retry mechanisms

### 💸 Refund Testing
1. Complete a successful payment
2. Create refund for full or partial amount
3. Verify refund processes correctly

### 🔄 Escrow Release Testing
1. Complete payment (funds held in escrow)
2. Simulate game completion
3. Release funds to organizer
4. Verify transfer completes

## Monitoring and Debugging

### Stripe Dashboard
- Monitor test payments in real-time
- View detailed logs and error messages
- Test webhook deliveries

### Application Logs
- Check server logs for payment processing
- Monitor fee calculations
- Track escrow operations

## Troubleshooting

### Common Issues

**"Invalid API Key"**
- Ensure you're using a test key starting with `sk_test_`
- Verify the key is correctly set in environment

**"No such customer/account"**
- Ensure Connect account ID is valid
- Check you're using the correct test account

**"Amount too small"**
- Stripe requires minimum €0.50 for most currencies
- Check your amount calculations

**Webhook 401/403 errors**
- Verify webhook secret matches Stripe CLI output
- Check endpoint URL is accessible

## Production Considerations

⚠️ **Never use test keys in production!**

Before going live:
- [ ] Switch to live API keys
- [ ] Test with real bank accounts
- [ ] Set up production webhook endpoints
- [ ] Configure proper error handling
- [ ] Set up monitoring and alerting