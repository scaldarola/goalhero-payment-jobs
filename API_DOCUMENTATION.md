# GoalHero Payment Jobs API Documentation

## Overview

This document provides detailed API specifications for the GoalHero Payment Jobs Service, including request/response formats, error handling, and integration examples.

## Base URL

- **Local Development**: `http://localhost:8081`
- **Production**: `https://your-vercel-deployment.vercel.app`

## Authentication

### Admin Endpoints
- **Method**: Firebase Authentication
- **Header**: `Authorization: Bearer <firebase_token>`
- **Scope**: Admin users only

### Public Endpoints
- No authentication required
- Used for payment operations and health checks

### Internal Endpoints
- No authentication required
- Designed for service-to-service communication

## Payment Endpoints

### Create Game Payment
Creates a payment intent for a game application.

**Endpoint**: `POST /api/payments/games`

**Request Body**:
```json
{
  "userId": "player_uid_123",
  "gameId": "game_id_456", 
  "applicationId": "application_id_789",
  "organizerId": "acct_stripe_connect_id",
  "amount": 25.0
}
```

**Validation Rules**:
- `userId`: Required, non-empty string
- `gameId`: Required, non-empty string  
- `applicationId`: Required, non-empty string
- `organizerId`: Required, valid Stripe Connect account ID
- `amount`: Required, float between 5.0 and 50.0

**Success Response** (200):
```json
{
  "success": true,
  "payment": {
    "id": "payment_123",
    "userId": "player_uid_123",
    "gameId": "game_id_456",
    "applicationId": "application_id_789", 
    "amount": 25.0,
    "platformFee": 1.0,
    "paymentFee": 0.66,
    "netAmount": 24.0,
    "currency": "EUR",
    "status": "pending",
    "paymentMethod": "stripe",
    "createdAt": "2025-01-15T10:30:00Z"
  },
  "client_secret": "pi_xxx_secret_xxx",
  "payment_intent": "pi_xxx"
}
```

**Error Responses**:
- `400`: Invalid request format or validation errors
- `500`: Payment creation failed

---

### Confirm Game Payment
Confirms a completed payment and creates escrow transaction.

**Endpoint**: `POST /api/payments/confirm`

**Request Body**:
```json
{
  "paymentId": "payment_123"
}
```

**Success Response** (200):
```json
{
  "success": true,
  "payment": {
    "id": "payment_123",
    "status": "confirmed",
    "confirmedAt": "2025-01-15T10:35:00Z"
  },
  "escrow": {
    "id": "escrow_456",
    "gameId": "game_id_456",
    "organizerId": "acct_stripe_connect_id",
    "paymentId": "payment_123",
    "amount": 24.0,
    "status": "held",
    "heldAt": "2025-01-15T10:35:00Z",
    "releaseEligibleAt": "2025-01-16T10:35:00Z",
    "ratingReceived": false,
    "minRatingRequired": 3.0
  }
}
```

---

### Release Escrow Funds
Manually releases escrow funds to organizer.

**Endpoint**: `POST /api/payments/escrow/release`

**Request Body**:
```json
{
  "escrowId": "escrow_456",
  "releaseReason": "manual_approval"
}
```

**Success Response** (200):
```json
{
  "success": true,
  "message": "Escrow released successfully",
  "escrowId": "escrow_456"
}
```

---

### Process Refund
Creates a refund for a payment.

**Endpoint**: `POST /api/payments/refund`

**Request Body**:
```json
{
  "paymentId": "payment_123",
  "amount": 25.0,
  "reason": "game_cancelled"
}
```

**Success Response** (200):
```json
{
  "success": true,
  "message": "Refund processed successfully",
  "paymentId": "payment_123",
  "amount": 25.0
}
```

---

### Get Eligible Escrow Releases
Returns escrow transactions eligible for automatic release.

**Endpoint**: `GET /api/payments/escrow/eligible`

**Success Response** (200):
```json
{
  "success": true,
  "escrows": [
    {
      "id": "escrow_456",
      "gameId": "game_id_456",
      "organizerId": "acct_stripe_connect_id",
      "amount": 24.0,
      "status": "approved",
      "releaseEligibleAt": "2025-01-15T10:35:00Z",
      "ratingReceived": true,
      "actualRating": 4.5,
      "ratingApproved": true
    }
  ],
  "count": 1
}
```

---

### Process Eligible Releases
Processes all eligible escrow releases automatically.

**Endpoint**: `POST /api/payments/escrow/process-eligible`

**Success Response** (200):
```json
{
  "success": true,
  "total_eligible": 5,
  "processed": 4,
  "failed": 1,
  "errors": [
    "Escrow escrow_789: insufficient balance"
  ]
}
```

---

### Get Test Cards
Returns test card numbers for Stripe testing (test mode only).

**Endpoint**: `GET /api/payments/test-cards`

**Success Response** (200):
```json
{
  "success": true,
  "test_cards": {
    "visa_success": "4242424242424242",
    "visa_decline": "4000000000000002",
    "mastercard_success": "5555555555554444",
    "amex_success": "378282246310005",
    "insufficient_funds": "4000000000009995",
    "expired_card": "4000000000000069",
    "incorrect_cvc": "4000000000000127",
    "processing_error": "4000000000000119"
  },
  "note": "These are test card numbers for Stripe testing"
}
```

## Job Management Endpoints

### Get Job Statuses
Returns detailed status information for all background jobs.

**Endpoint**: `GET /api/jobs/status`

**Success Response** (200):
```json
{
  "rating_reminder": {
    "jobName": "Rating Reminder",
    "lastRun": "2025-01-15T10:00:00Z",
    "nextScheduled": "2025-01-15T16:00:00Z",
    "lastResult": "Successfully sent 12 rating reminders",
    "runCount": 145,
    "errorCount": 2,
    "averageRuntime": "1.234s",
    "isRunning": false,
    "enabled": true
  },
  "auto_release": {
    "jobName": "Auto Release",
    "lastRun": "2025-01-15T10:30:00Z",
    "nextScheduled": "2025-01-15T11:30:00Z", 
    "lastResult": "Released 3 escrow transactions",
    "runCount": 876,
    "errorCount": 5,
    "averageRuntime": "2.456s",
    "isRunning": false,
    "enabled": true
  },
  "dispute_escalation": {
    "jobName": "Dispute Escalation",
    "lastRun": "2025-01-15T08:00:00Z",
    "nextScheduled": "2025-01-15T12:00:00Z",
    "lastResult": "No disputes requiring escalation",
    "runCount": 234,
    "errorCount": 0,
    "averageRuntime": "456ms",
    "isRunning": false,
    "enabled": true
  }
}
```

---

### Get Job Health
Returns overall health information for the job system.

**Endpoint**: `GET /api/jobs/health`

**Success Response** (200):
```json
{
  "healthy": true,
  "totalJobs": 3,
  "runningJobs": 0,
  "failedJobs": 0,
  "lastHealthCheck": "2025-01-15T10:45:00Z",
  "jobStatuses": {
    // Same format as /api/jobs/status
  }
}
```

---

### Trigger Job (Admin)
Manually triggers a specific background job.

**Endpoint**: `POST /api/jobs/trigger/:jobName`
**Authentication**: Required (Firebase Auth)

**Path Parameters**:
- `jobName`: One of `rating_reminder`, `auto_release`, `dispute_escalation`

**Success Response** (200):
```json
{
  "success": true,
  "message": "Job rating_reminder triggered successfully",
  "jobName": "rating_reminder"
}
```

---

### Get Job Configuration (Admin)
Returns current job configuration settings.

**Endpoint**: `GET /api/jobs/config`
**Authentication**: Required (Firebase Auth)

**Success Response** (200):
```json
{
  "ratingReminderInterval": "6h0m0s",
  "autoReleaseInterval": "1h0m0s", 
  "disputeEscalationInterval": "4h0m0s",
  "ratingDeadlineDays": 7,
  "minRatingForAutoRelease": 3.0,
  "disputeEscalationHours": 48
}
```

---

### Update Job Configuration (Admin)
Updates job configuration settings.

**Endpoint**: `POST /api/jobs/config`
**Authentication**: Required (Firebase Auth)

**Request Body**:
```json
{
  "ratingReminderInterval": "6h0m0s",
  "autoReleaseInterval": "1h0m0s",
  "disputeEscalationInterval": "4h0m0s", 
  "ratingDeadlineDays": 7,
  "minRatingForAutoRelease": 3.0,
  "disputeEscalationHours": 48
}
```

**Success Response** (200):
```json
{
  "success": true,
  "message": "Job configuration updated successfully"
}
```

## Internal Endpoints

These endpoints are designed for service-to-service communication and do not require authentication.

### Trigger Rating Reminder
**Endpoint**: `POST /api/jobs/internal/trigger-rating-reminder`

**Success Response** (200):
```json
{
  "success": true,
  "message": "Rating reminder job triggered"
}
```

### Trigger Auto Release  
**Endpoint**: `POST /api/jobs/internal/trigger-auto-release`

**Success Response** (200):
```json
{
  "success": true,
  "message": "Auto release job triggered"
}
```

### Trigger Dispute Escalation
**Endpoint**: `POST /api/jobs/internal/trigger-dispute-escalation`

**Success Response** (200):
```json
{
  "success": true,
  "message": "Dispute escalation job triggered"
}
```

## Health Check Endpoints

### Service Health
**Endpoint**: `GET /` or `GET /ping`

**Success Response** (200):
```json
{
  "service": "goalhero-payment-jobs",
  "status": "healthy"
}
```

## Error Responses

### Standard Error Format
All error responses follow this format:

```json
{
  "success": false,
  "error": "Brief error description",
  "details": "Detailed error message or validation errors"
}
```

### Common HTTP Status Codes
- `200`: Success
- `400`: Bad Request (validation errors, invalid input)
- `401`: Unauthorized (missing or invalid authentication)
- `403`: Forbidden (insufficient permissions)
- `404`: Not Found (resource doesn't exist)
- `500`: Internal Server Error (server-side errors)
- `501`: Not Implemented (feature not yet implemented)

## Integration Examples

### Frontend Payment Flow
```javascript
// 1. Create payment intent
const response = await fetch('/api/payments/games', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    userId: 'user_123',
    gameId: 'game_456',
    applicationId: 'app_789',
    organizerId: 'acct_connect_id',
    amount: 25.0
  })
});

const { client_secret } = await response.json();

// 2. Complete payment with Stripe
const result = await stripe.confirmPayment({
  elements,
  clientSecret: client_secret,
  confirmParams: {
    return_url: 'https://app.example.com/payment-success'
  }
});

// 3. Confirm payment on backend
if (result.paymentIntent?.status === 'succeeded') {
  await fetch('/api/payments/confirm', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      paymentId: paymentId
    })
  });
}
```

### Admin Job Management
```javascript
// Get Firebase auth token
const token = await firebase.auth().currentUser.getIdToken();

// Trigger job manually
const response = await fetch('/api/jobs/trigger/rating_reminder', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${token}`
  }
});

// Update job configuration  
await fetch('/api/jobs/config', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    ratingReminderInterval: "8h0m0s",
    minRatingForAutoRelease: 3.5
  })
});
```

### Service-to-Service Communication
```javascript
// Trigger jobs from external services
await fetch('https://payment-jobs.vercel.app/api/jobs/internal/trigger-auto-release', {
  method: 'POST'
});
```

## Rate Limits

- **Payment Endpoints**: No explicit rate limits (handled by Stripe)
- **Job Triggers**: Max 1 request per minute per job type
- **Health Checks**: No limits
- **Admin Operations**: Max 60 requests per minute per user

## Webhooks

While not directly provided by this service, you can set up Stripe webhooks to notify external services:

### Recommended Webhook Events
- `payment_intent.succeeded` - Payment completed successfully
- `payment_intent.payment_failed` - Payment failed
- `transfer.created` - Funds transferred to Connect account
- `application_fee.created` - Platform fee collected

### Webhook URL Format
Point Stripe webhooks to your main API service, which can then trigger this service's internal endpoints as needed.