# GoalHero Payment Workflow Documentation

## Overview
This document describes the complete payment workflow for the GoalHero sports platform, from initial game application to final payout to organizers.

## Business Logic

### Core Concept
GoalHero is a sports platform where:
- **Organizers** create games/matches
- **Players** apply to join games with a bid price
- **Payments** are processed through Stripe Connect with escrow functionality
- **Funds** are held in escrow until the game is completed and rated

### Fee Structure
- **Platform Fee**: 4% of payment amount
- **Stripe Processing Fee**: 1.65% + €0.25 per transaction
- **Total User Pays**: Game amount + Stripe processing fee
- **Organizer Receives**: Game amount - Platform fee (after escrow release)

## Payment Workflow

### 1. Game Creation & Application
```
Organizer creates game → Players apply with price bids → Organizer selects player
```

**Key Points:**
- Players set their own price when applying (€5-50 range)
- Organizer reviews applications and selects preferred player
- No payment happens until player is accepted

### 2. Payment Intent Creation
```
POST /api/payments/games
{
  "userId": "player_123",
  "gameId": "game_456", 
  "applicationId": "app_789",
  "organizerId": "acct_stripe_connect_id",
  "amount": 25.0
}
```

**What Happens:**
- Creates Payment record in database (status: `pending`)
- Calculates all fees (platform, Stripe processing)
- Creates Stripe PaymentIntent with:
  - Total amount (game price + processing fee)
  - Transfer destination (organizer's Connect account)
  - Application fee (platform fee)
- Returns `client_secret` for frontend completion

**Database Records:**
```go
Payment {
  Amount: 25.0,           // Game price
  PlatformFee: 1.0,       // 4% of 25.0
  PaymentFee: 0.66,       // 1.65% + 0.25
  NetAmount: 24.0,        // Amount - PlatformFee
  Status: "pending"
}
```

### 3. Payment Completion (Frontend)
```javascript
// Frontend uses Stripe Elements/Mobile SDK
const result = await stripe.confirmPayment({
  elements,
  clientSecret: 'pi_xxx_secret_xxx',
  confirmParams: {
    return_url: 'https://app.goalhero.com/payment-success'
  }
});
```

**What Happens:**
- User completes payment with test/real card
- Stripe processes payment and transfers funds to organizer account
- Webhook notification sent to platform (if configured)

### 4. Payment Confirmation
```
POST /api/payments/confirm
{
  "paymentId": "payment_123"
}
```

**What Happens:**
- Updates Payment status to `confirmed`
- Creates EscrowTransaction record:
  ```go
  EscrowTransaction {
    Status: "held",
    Amount: 24.0,  // Net amount (after platform fee)
    ReleaseEligibleAt: game_end_time + 24_hours,
    RatingReceived: false,
    MinRatingRequired: 3.0
  }
  ```
- Funds are now held in escrow on the organizer's Connect account

### 5. Game Completion & Rating
```
Game is played → Organizer marks game complete → Players rate performance
```

**Rating Process:**
- Players have 7 days to rate the game/organizer
- Ratings affect escrow release eligibility
- Background job sends reminders for pending ratings

### 6. Automatic Escrow Release
**Conditions for Release:**
- Game completed + 24 hours passed
- Average rating ≥ 3.0/5.0 (configurable)
- No active disputes

**Release Process:**
```go
// Background job runs periodically
func (jm *BackgroundJobManager) runAutoRelease() {
  // Find eligible escrows
  eligibleEscrows := GetEscrowsEligibleForRelease()
  
  for _, escrow := range eligibleEscrows {
    ProcessEscrowRelease(escrow.ID, "automatic_release")
  }
}
```

### 7. Manual Processes

**Manual Escrow Release:**
```
POST /api/payments/escrow/release
{
  "escrowId": "escrow_123",
  "releaseReason": "manual_approval"
}
```

**Refund Processing:**
```
POST /api/payments/refund
{
  "paymentId": "payment_123",
  "amount": 25.0,
  "reason": "game_cancelled"
}
```

## State Transitions

### Payment States
```
pending → confirmed → [refunded]
```

### Escrow States  
```
held → pending_rating → approved → released
  ↓
disputed → resolved → [refunded]
```

### Application States
```
pending → accepted → payment_completed
  ↓
rejected/withdrawn
```

## Background Jobs

### 1. Rating Reminder Job
- **Frequency**: Every 6 hours
- **Purpose**: Reminds players to rate completed games
- **Target**: Games completed 1-7 days ago without ratings

### 2. Auto Release Job
- **Frequency**: Every hour
- **Purpose**: Automatically releases eligible escrow funds
- **Conditions**: 24h+ after game, good ratings, no disputes

### 3. Dispute Escalation Job
- **Frequency**: Every 4 hours  
- **Purpose**: Escalates unresolved disputes to admin review
- **Target**: Disputes open for 48+ hours

## Error Handling & Edge Cases

### Payment Failures
- Card declined → Payment stays `pending`, user can retry
- Insufficient funds → Same as above
- Processing error → Log error, allow retry

### Refund Scenarios
- Game cancelled by organizer → Full refund
- Player no-show → Partial/no refund (organizer decision)
- Dispute resolution → Admin-determined refund amount

### Escrow Release Blocks
- Poor ratings (< 3.0) → Hold for manual review
- Active disputes → Hold until resolved
- Missing ratings → Extended hold period

## Security & Compliance

### Stripe Connect Integration
- Uses destination charges with application fees
- Funds flow: Customer → Platform → Organizer (with fees)
- Automatic transfers when payments succeed

### Data Protection
- PCI compliance through Stripe
- No card data stored in application
- Payment tokens used for recurring operations

### Financial Controls
- All transactions logged and auditable  
- Automated reconciliation possible
- Clear separation of platform fees and organizer payouts

## Monitoring & Observability

### Key Metrics
- Payment success rate
- Average escrow hold time
- Rating completion rate
- Dispute resolution time
- Fee collection efficiency

### Alerts
- Failed payment processing
- Stuck escrow transactions
- High dispute rates
- Background job failures

## Testing Strategy

### Unit Tests
- Fee calculations
- State transitions
- Validation logic
- Background job logic

### Integration Tests
- Stripe API interactions
- Database operations
- End-to-end payment flows

### Test Environment
- Stripe test mode with test cards
- Firebase test project
- Isolated test data