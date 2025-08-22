# GoalHero Payment Jobs Service

A microservice for handling payment processing, escrow management, and background jobs for the GoalHero sports platform.

## 🎯 Purpose

GoalHero is a sports platform where organizers create games and players apply to join with competitive bids. This service manages the entire payment lifecycle from initial payment creation to final payout, with built-in escrow functionality to ensure fair transactions.

## 🏗️ Architecture

### Core Components

- **Payment Processing**: Stripe Connect integration with escrow functionality
- **Background Jobs**: Automated rating reminders, escrow releases, and dispute handling  
- **REST API**: Endpoints for payment operations and job management
- **Firebase Integration**: Authentication and Firestore database
- **Serverless Ready**: Designed for Vercel deployment with local development support

### Tech Stack

- **Language**: Go 1.21+
- **Web Framework**: Gin HTTP router
- **Database**: Firebase Firestore
- **Payment Provider**: Stripe Connect
- **Authentication**: Firebase Auth
- **Deployment**: Vercel (serverless) + local development

## 💰 Business Model

### Payment Flow
1. **Players apply** to games with bid amounts (€5-50)
2. **Organizers select** preferred players
3. **Payments processed** through Stripe with escrow
4. **Funds held** until game completion + rating
5. **Automatic release** to organizers based on performance ratings

### Fee Structure
- **Platform Fee**: 4% of game amount
- **Stripe Fee**: 1.65% + €0.25 per transaction
- **Total Cost**: Game amount + Stripe processing fee
- **Organizer Nets**: Game amount - platform fee (after escrow release)

## 🚀 Quick Start

### Prerequisites
- Go 1.21 or higher
- Stripe test account with Connect enabled
- Firebase project with Firestore enabled

### Environment Setup
1. Copy environment template:
   ```bash
   cp .env.example .env
   ```

2. Configure your environment variables:
   ```bash
   # Stripe Configuration
   STRIPE_SECRET_KEY=sk_test_your_key_here
   STRIPE_CONNECT_ACCOUNT=acct_test_your_connect_account
   STRIPE_TEST_MODE=true
   
   # Firebase Configuration  
   GOOGLE_APPLICATION_CREDENTIALS=path/to/firebase-service-account.json
   FIREBASE_PROJECT_ID=your-project-id
   
   # Server Configuration
   PORT=8081
   GO_ENV=development
   ```

### Local Development
1. Install dependencies:
   ```bash
   go mod download
   ```

2. Run the service:
   ```bash
   go run main.go
   ```

3. Service available at: `http://localhost:8081`

### Testing
Run the full test suite:
```bash
go test ./... -v
```

Test Stripe integration:
```bash
go run test_stripe_integration.go
```

Debug payment issues:
```bash
go run debug_stripe_payments.go
```

## 🔄 Payment Workflow

### 1. Payment Creation
```bash
POST /api/payments/games
{
  "userId": "player_123",
  "gameId": "game_456",
  "applicationId": "app_789", 
  "organizerId": "acct_stripe_connect_id",
  "amount": 25.0
}
```

**Response**: Returns payment intent with `client_secret` for frontend completion.

### 2. Payment Completion
Frontend completes payment using Stripe Elements/SDK with the client secret.

### 3. Escrow Management
- Funds automatically held in escrow on organizer's Stripe Connect account
- Released after game completion + rating period (24h default)
- Background jobs handle automatic processing

### 4. Background Processing
- **Rating Reminders**: Every 6 hours, reminds players to rate games
- **Auto Release**: Every hour, releases eligible escrow funds
- **Dispute Escalation**: Every 4 hours, escalates unresolved disputes

## 📊 API Endpoints

### Payment Operations
- `POST /api/payments/games` - Create game payment
- `POST /api/payments/confirm` - Confirm payment completion  
- `POST /api/payments/escrow/release` - Manually release escrow
- `POST /api/payments/refund` - Process refund
- `GET /api/payments/escrow/eligible` - Get eligible escrow releases
- `GET /api/payments/test-cards` - Get test card numbers (test mode only)

### Job Management
- `GET /api/jobs/status` - Get job statuses
- `GET /api/jobs/health` - Get job health information
- `POST /api/jobs/trigger/:jobName` - Manually trigger job (admin)
- `GET /api/jobs/config` - Get job configuration
- `POST /api/jobs/config` - Update job configuration (admin)

### Internal Services
- `POST /api/jobs/internal/trigger-rating-reminder` - Trigger rating reminders
- `POST /api/jobs/internal/trigger-auto-release` - Trigger escrow releases
- `POST /api/jobs/internal/trigger-dispute-escalation` - Trigger dispute handling

## 🗄️ Data Models

### Payment
Represents a payment transaction with fees and status tracking.

### EscrowTransaction  
Manages funds held in escrow with release conditions and timing.

### Match
Sports game/match with players, location, and completion status.

### Application
Player application to join a game with bid price and status.

See [PAYMENT_WORKFLOW.md](./PAYMENT_WORKFLOW.md) for detailed business logic documentation.

## 🔧 Configuration

### Background Jobs
Default intervals (configurable via API):
- Rating Reminders: Every 6 hours
- Auto Release: Every 1 hour  
- Dispute Escalation: Every 4 hours

### Business Rules
- Minimum game price: €5
- Maximum game price: €50
- Escrow hold period: 24 hours after game completion
- Minimum rating for auto-release: 3.0/5.0
- Rating deadline: 7 days after game completion

## 🧪 Testing

### Test Cards (Stripe Test Mode)
- **Success**: `4242424242424242`
- **Decline**: `4000000000000002`  
- **Insufficient Funds**: `4000000000009995`
- **Expired**: `4000000000000069`

### Testing Tools
- `test_stripe_integration.go` - Integration testing script
- `debug_stripe_payments.go` - Payment debugging utility
- Comprehensive unit test suite in `*_test.go` files

See [TESTING.md](./TESTING.md) for complete testing guide.

## 🚀 Deployment

### Vercel (Production)
The service is designed for serverless deployment on Vercel:
- Main handler exports `Handler(w, r)` function
- Background jobs disabled in production (use external cron/scheduler)
- Environment variables configured in Vercel dashboard

### Local Development  
- Full background job system runs locally
- Gin router serves HTTP endpoints
- Uses port 8081 by default

## 🔐 Security

### Payment Security
- PCI compliance through Stripe (no card data stored)
- Stripe Connect for secure fund transfers
- Firebase Authentication for API access

### Access Control
- Public endpoints for payment operations
- Admin-only endpoints for job management
- Internal endpoints for service-to-service communication

## 📈 Monitoring

### Health Checks
- `GET /` or `GET /ping` - Service health
- `GET /api/jobs/health` - Background job health
- Job status tracking with error counts and runtime metrics

### Key Metrics
- Payment success rates
- Escrow release timing
- Rating completion rates
- Background job performance

## 🤝 Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Run tests (`go test ./... -v`)
4. Commit changes (`git commit -m 'Add amazing feature'`)
5. Push to branch (`git push origin feature/amazing-feature`)
6. Open Pull Request

## 📝 License

This project is proprietary software for GoalHero platform.

## 📞 Support

- Documentation: See `/docs` folder
- Issues: Create GitHub issue
- Testing: Use provided testing tools and documentation

---

**Built with ❤️ for the GoalHero sports community**