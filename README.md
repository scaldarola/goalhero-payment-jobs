# GoalHero Payment Jobs Service

Background jobs service for managing payments, escrow, ratings, and automated processes for the GoalHero platform. Designed to run on Vercel as a serverless application.

## Overview

This service handles critical background operations including:
- **Rating Reminders**: Automated reminders for players to rate completed matches
- **Auto Release**: Automatic release of escrowed payments when conditions are met
- **Dispute Escalation**: Escalation of payment disputes requiring admin intervention

## Architecture

- **Serverless First**: Built for Vercel deployment with proper Handler function
- **Firebase Integration**: Uses Firebase Auth and Firestore for data persistence
- **Job Management**: Background jobs with status tracking and health monitoring
- **RESTful API**: Endpoints for job control and monitoring

## Environment Variables

Required environment variables for production deployment:

```bash
# Firebase/Google Cloud (Required for Vercel)
GOOGLE_APPLICATION_CREDENTIALS=<firebase-service-account-json-as-string>
GOOGLE_CLOUD_PROJECT=your-firebase-project-id

# Application Configuration
GO_ENV=production
MAIN_API_URL=https://your-main-api-url.com
JWT_SECRET=your-jwt-secret

# Job Intervals (optional, uses defaults if not set)
RATING_REMINDER_INTERVAL=24h
AUTO_RELEASE_INTERVAL=1h
DISPUTE_ESCALATION_INTERVAL=24h

# Job Configuration (optional)
RATING_DEADLINE_DAYS=7
MIN_RATING_FOR_AUTO_RELEASE=3.0
DISPUTE_ESCALATION_HOURS=72
```

### Setting up Firebase Credentials for Vercel

1. Download your Firebase service account JSON key
2. Convert it to a single-line string (remove newlines)
3. Set `GOOGLE_APPLICATION_CREDENTIALS` to this JSON string in Vercel dashboard
4. Set `GOOGLE_CLOUD_PROJECT` to your Firebase project ID

## API Endpoints

### Health & Status
- `GET /` - Service health check
- `GET /ping` - Simple ping endpoint
- `GET /api/jobs/status` - Detailed job statuses
- `GET /api/jobs/health` - Job health information

### Job Control (Admin Only - Requires Firebase Auth)
- `POST /api/jobs/trigger/:jobName` - Manually trigger a job
- `GET /api/jobs/config` - Get current job configuration
- `POST /api/jobs/config` - Update job configuration
- `POST /api/jobs/restart` - Restart job system

### Internal Service Communication
- `POST /api/jobs/internal/trigger-rating-reminder` - Trigger rating reminders
- `POST /api/jobs/internal/trigger-auto-release` - Trigger auto release
- `POST /api/jobs/internal/trigger-dispute-escalation` - Trigger dispute escalation

## Local Development

1. **Install dependencies**:
   ```bash
   go mod download
   ```

2. **Set up Firebase credentials**:
   ```bash
   # Place your Firebase service account key in:
   # auth/firebase_credentials.json
   ```

3. **Run locally**:
   ```bash
   GO_ENV=development go run main.go
   ```

   The service will start on `http://localhost:8081`

## Deployment to Vercel

1. **Connect repository to Vercel**
2. **Set environment variables** in Vercel dashboard
3. **Deploy** - Vercel will automatically use the `vercel.json` configuration

### Important Vercel Setup Notes

- Firebase credentials should be provided as base64-encoded JSON in `GOOGLE_APPLICATION_CREDENTIALS`
- The service uses the `Handler` function for serverless execution
- Background jobs are handled differently in production (triggered via external cron or API calls)

## Security

- Firebase credentials are excluded from version control (see `.gitignore`)
- Admin endpoints require Firebase authentication
- Internal endpoints are unprotected (design for inter-service communication)
- JWT tokens used for internal service authentication

## Monitoring

The service provides comprehensive job monitoring:
- Individual job status tracking
- Error counting and health metrics
- Runtime performance tracking
- Last run and next scheduled time tracking

Access monitoring via `/api/jobs/status` and `/api/jobs/health` endpoints.

## Contributing

1. Ensure all tests pass: `go test ./...`
2. Run linting: `go fmt ./...`
3. Update documentation as needed
4. Follow existing code patterns and conventions