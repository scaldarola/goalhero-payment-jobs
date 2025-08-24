package services

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/sebastiancaldarola/goalhero-payment-jobs/config"
	"github.com/sebastiancaldarola/goalhero-payment-jobs/models"
	"google.golang.org/api/iterator"
)

// JobConfig holds configuration for background jobs
type JobConfig struct {
	RatingReminderInterval   time.Duration `json:"ratingReminderInterval"`
	AutoReleaseInterval      time.Duration `json:"autoReleaseInterval"`
	DisputeEscalationInterval time.Duration `json:"disputeEscalationInterval"`
	RatingDeadlineDays       int           `json:"ratingDeadlineDays"`
	MinRatingForAutoRelease  float64       `json:"minRatingForAutoRelease"`
	DisputeEscalationHours   int           `json:"disputeEscalationHours"`
}

// BackgroundJobManager manages all background jobs
type BackgroundJobManager struct {
	config   *JobConfig
	shutdown chan struct{}
	wg       sync.WaitGroup
	running  bool
	mu       sync.Mutex
}

// JobStatus represents the status of a background job
type JobStatus struct {
	JobName        string    `json:"jobName"`
	LastRun        time.Time `json:"lastRun"`
	NextScheduled  time.Time `json:"nextScheduled"`
	LastResult     string    `json:"lastResult"`
	RunCount       int       `json:"runCount"`
	ErrorCount     int       `json:"errorCount"`
	AverageRuntime string    `json:"averageRuntime"`
	IsRunning      bool      `json:"isRunning"`
	Enabled        bool      `json:"enabled"`
}

// JobHealth represents overall health of the job system
type JobHealth struct {
	Healthy          bool                  `json:"healthy"`
	TotalJobs        int                   `json:"totalJobs"`
	RunningJobs      int                   `json:"runningJobs"`
	FailedJobs       int                   `json:"failedJobs"`
	LastHealthCheck  time.Time             `json:"lastHealthCheck"`
	JobStatuses      map[string]*JobStatus `json:"jobStatuses"`
}

var (
	jobManager  *BackgroundJobManager
	jobStatuses = make(map[string]*JobStatus)
	statusMutex sync.RWMutex
)

// StartBackgroundJobs initializes and starts all background jobs
func StartBackgroundJobs() *BackgroundJobManager {
	jobsConf := config.GetJobsConfig()
	
	config := &JobConfig{
		RatingReminderInterval:    jobsConf.RatingReminderInterval,
		AutoReleaseInterval:       jobsConf.AutoReleaseInterval,
		DisputeEscalationInterval: jobsConf.DisputeEscalationInterval,
		RatingDeadlineDays:        jobsConf.RatingDeadlineDays,
		MinRatingForAutoRelease:   jobsConf.MinRatingForAutoRelease,
		DisputeEscalationHours:    jobsConf.DisputeEscalationHours,
	}

	jobManager = &BackgroundJobManager{
		config:   config,
		shutdown: make(chan struct{}),
		running:  true,
	}

	// Initialize job statuses
	initializeJobStatuses(config)

	log.Printf("[BackgroundJobs] Starting job system with intervals: Rating=%v, Release=%v, Dispute=%v", 
		config.RatingReminderInterval, config.AutoReleaseInterval, config.DisputeEscalationInterval)

	// Start each job in its own goroutine
	jobManager.wg.Add(3)
	go jobManager.runRatingReminderJob()
	go jobManager.runAutoReleaseJob()
	go jobManager.runDisputeEscalationJob()

	log.Printf("[BackgroundJobs] All jobs started successfully")
	return jobManager
}

// StopBackgroundJobs gracefully shuts down all background jobs
func (jm *BackgroundJobManager) StopBackgroundJobs() {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	if !jm.running {
		return
	}

	log.Printf("[BackgroundJobs] Shutting down all jobs...")
	jm.running = false
	close(jm.shutdown)
	jm.wg.Wait()
	log.Printf("[BackgroundJobs] All jobs stopped")
}

// GetJobStatuses returns current status of all jobs
func GetJobStatuses() map[string]*JobStatus {
	statusMutex.RLock()
	defer statusMutex.RUnlock()

	statuses := make(map[string]*JobStatus)
	for k, v := range jobStatuses {
		statusCopy := *v
		statuses[k] = &statusCopy
	}
	return statuses
}

// GetJobHealth returns overall health information
func GetJobHealth() *JobHealth {
	statuses := GetJobStatuses()
	
	health := &JobHealth{
		TotalJobs:       len(statuses),
		RunningJobs:     0,
		FailedJobs:      0,
		LastHealthCheck: time.Now(),
		JobStatuses:     statuses,
	}

	for _, status := range statuses {
		if status.IsRunning {
			health.RunningJobs++
		}
		if status.ErrorCount > status.RunCount/2 { // More than 50% error rate
			health.FailedJobs++
		}
	}

	health.Healthy = health.FailedJobs == 0
	return health
}

// UpdateJobConfig updates the job configuration
func UpdateJobConfig(newConfig *JobConfig) error {
	if jobManager == nil {
		return fmt.Errorf("job manager not initialized")
	}

	jobManager.mu.Lock()
	defer jobManager.mu.Unlock()

	log.Printf("[Config] Updating job configuration...")
	jobManager.config = newConfig

	// Update next scheduled times based on new intervals
	statusMutex.Lock()
	if status, exists := jobStatuses["rating_reminder"]; exists {
		status.NextScheduled = time.Now().Add(newConfig.RatingReminderInterval)
	}
	if status, exists := jobStatuses["auto_release"]; exists {
		status.NextScheduled = time.Now().Add(newConfig.AutoReleaseInterval)
	}
	if status, exists := jobStatuses["dispute_escalation"]; exists {
		status.NextScheduled = time.Now().Add(newConfig.DisputeEscalationInterval)
	}
	statusMutex.Unlock()

	log.Printf("[Config] Job configuration updated successfully")
	return nil
}

// GetJobConfig returns the current job configuration
func GetJobConfig() *JobConfig {
	if jobManager == nil {
		return nil
	}
	return jobManager.config
}

// Trigger methods for manual job execution
func TriggerRatingReminder() error {
	if jobManager == nil {
		return fmt.Errorf("job manager not initialized")
	}
	go jobManager.runRatingReminder()
	return nil
}

func TriggerAutoRelease() error {
	if jobManager == nil {
		return fmt.Errorf("job manager not initialized")
	}
	go jobManager.runAutoRelease()
	return nil
}

func TriggerDisputeEscalation() error {
	if jobManager == nil {
		return fmt.Errorf("job manager not initialized")
	}
	go jobManager.runDisputeEscalation()
	return nil
}

// Internal job execution methods
func initializeJobStatuses(config *JobConfig) {
	statusMutex.Lock()
	defer statusMutex.Unlock()

	jobStatuses["rating_reminder"] = &JobStatus{
		JobName:       "Rating Reminder",
		LastResult:    "Not run yet",
		NextScheduled: time.Now().Add(config.RatingReminderInterval),
		Enabled:       true,
	}

	jobStatuses["auto_release"] = &JobStatus{
		JobName:       "Auto Release",
		LastResult:    "Not run yet",
		NextScheduled: time.Now().Add(config.AutoReleaseInterval),
		Enabled:       true,
	}

	jobStatuses["dispute_escalation"] = &JobStatus{
		JobName:       "Dispute Escalation",
		LastResult:    "Not run yet",
		NextScheduled: time.Now().Add(config.DisputeEscalationInterval),
		Enabled:       true,
	}
}

func updateJobStatus(jobName string, result string, runTime time.Duration, hasError bool) {
	statusMutex.Lock()
	defer statusMutex.Unlock()

	if status, exists := jobStatuses[jobName]; exists {
		status.LastRun = time.Now()
		status.LastResult = result
		status.RunCount++
		status.AverageRuntime = runTime.String()
		status.IsRunning = false

		if hasError {
			status.ErrorCount++
		}

		// Calculate next scheduled run
		if jobManager != nil {
			switch jobName {
			case "rating_reminder":
				status.NextScheduled = time.Now().Add(jobManager.config.RatingReminderInterval)
			case "auto_release":
				status.NextScheduled = time.Now().Add(jobManager.config.AutoReleaseInterval)
			case "dispute_escalation":
				status.NextScheduled = time.Now().Add(jobManager.config.DisputeEscalationInterval)
			}
		}
	}
}

// Job runner methods (implement the actual job logic from original background_jobs.go)
func (jm *BackgroundJobManager) runRatingReminderJob() {
	defer jm.wg.Done()
	
	ticker := time.NewTicker(jm.config.RatingReminderInterval)
	defer ticker.Stop()

	log.Printf("[RatingReminderJob] Started (interval: %v)", jm.config.RatingReminderInterval)

	for {
		select {
		case <-jm.shutdown:
			log.Printf("[RatingReminderJob] Shutting down")
			return
		case <-ticker.C:
			jm.runRatingReminder()
		}
	}
}

func (jm *BackgroundJobManager) runAutoReleaseJob() {
	defer jm.wg.Done()
	
	ticker := time.NewTicker(jm.config.AutoReleaseInterval)
	defer ticker.Stop()

	log.Printf("[AutoReleaseJob] Started (interval: %v)", jm.config.AutoReleaseInterval)

	for {
		select {
		case <-jm.shutdown:
			log.Printf("[AutoReleaseJob] Shutting down")
			return
		case <-ticker.C:
			jm.runAutoRelease()
		}
	}
}

func (jm *BackgroundJobManager) runDisputeEscalationJob() {
	defer jm.wg.Done()
	
	ticker := time.NewTicker(jm.config.DisputeEscalationInterval)
	defer ticker.Stop()

	log.Printf("[DisputeEscalationJob] Started (interval: %v)", jm.config.DisputeEscalationInterval)

	for {
		select {
		case <-jm.shutdown:
			log.Printf("[DisputeEscalationJob] Shutting down")
			return
		case <-ticker.C:
			jm.runDisputeEscalation()
		}
	}
}

func (jm *BackgroundJobManager) runRatingReminder() {
	start := time.Now()
	log.Printf("[RatingReminderJob] Starting execution")

	statusMutex.Lock()
	if status, exists := jobStatuses["rating_reminder"]; exists {
		status.IsRunning = true
	}
	statusMutex.Unlock()

	var result string
	var hasError bool

	defer func() {
		updateJobStatus("rating_reminder", result, time.Since(start), hasError)
	}()

	// Check if Firestore client is available
	firestoreClient := config.FirestoreClient()
	if firestoreClient == nil {
		log.Printf("[RatingReminderJob] Firestore client not available (test environment?)")
		result = "Skipped - no Firestore client available"
		return
	}

	ctx := context.Background()
	
	// Implementation from original background_jobs.go
	sevenDaysAgo := time.Now().AddDate(0, 0, -jm.config.RatingDeadlineDays)
	oneDayAgo := time.Now().Add(-24 * time.Hour)

	query := firestoreClient.Collection("matches").
		Where("status", "==", "completed").
		Where("completedAt", ">=", sevenDaysAgo).
		Where("completedAt", "<=", oneDayAgo)

	iter := query.Documents(ctx)
	remindersSent := 0
	errors := 0

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("[RatingReminderJob] Error iterating matches: %v", err)
			errors++
			continue
		}

		var match models.Match
		if err := doc.DataTo(&match); err != nil {
			log.Printf("[RatingReminderJob] Failed to parse match: %v", err)
			errors++
			continue
		}

		// Send reminders for this match
		for _, playerID := range match.PlayersPresent {
			if playerID != match.CreatedBy {
				sendRatingReminder(playerID, &match)
				remindersSent++
			}
		}
	}

	if errors > 0 {
		hasError = true
		result = fmt.Sprintf("Sent %d reminders with %d errors", remindersSent, errors)
	} else {
		result = fmt.Sprintf("Successfully sent %d rating reminders", remindersSent)
	}

	log.Printf("[RatingReminderJob] Completed: %s (runtime: %v)", result, time.Since(start))
}

func (jm *BackgroundJobManager) runAutoRelease() {
	start := time.Now()
	log.Printf("[AutoReleaseJob] Starting execution")

	statusMutex.Lock()
	if status, exists := jobStatuses["auto_release"]; exists {
		status.IsRunning = true
	}
	statusMutex.Unlock()

	var result string
	var hasError bool

	defer func() {
		updateJobStatus("auto_release", result, time.Since(start), hasError)
	}()

	// Check if Firestore client is available
	firestoreClient := config.FirestoreClient()
	if firestoreClient == nil {
		log.Printf("[AutoReleaseJob] Firestore client not available (test environment?)")
		result = "Skipped - no Firestore client available"
		return
	}

	// Process automatic escrow releases
	paymentService := NewPaymentService()
	processed, failed, errors, totalReleased, err := paymentService.ProcessAutomaticReleases()
	
	if err != nil {
		hasError = true
		result = fmt.Sprintf("Auto release failed: %v", err)
		log.Printf("[AutoReleaseJob] Failed: %s", result)
		return
	}

	// Calculate validated count (total escrows examined)
	validated := processed + failed

	if failed > 0 {
		hasError = true
		result = fmt.Sprintf("Processed %d releases, %d failed (errors: %d)", processed, failed, len(errors))
		// Log the first few errors for debugging
		for i, errMsg := range errors {
			if i >= 3 { // Limit to first 3 errors in logs
				break
			}
			log.Printf("[AutoReleaseJob] Error: %s", errMsg)
		}
	} else {
		result = fmt.Sprintf("Successfully processed %d automatic releases", processed)
	}

	// Send Slack notification with job summary
	paymentService.SendSlackJobSummaryNotification(validated, processed, failed, totalReleased, time.Since(start))

	log.Printf("[AutoReleaseJob] Completed: %s (runtime: %v)", result, time.Since(start))
}

func (jm *BackgroundJobManager) runDisputeEscalation() {
	start := time.Now()
	log.Printf("[DisputeEscalationJob] Starting execution")

	statusMutex.Lock()
	if status, exists := jobStatuses["dispute_escalation"]; exists {
		status.IsRunning = true
	}
	statusMutex.Unlock()

	var result string
	var hasError bool

	defer func() {
		updateJobStatus("dispute_escalation", result, time.Since(start), hasError)
	}()

	// Check if Firestore client is available
	firestoreClient := config.FirestoreClient()
	if firestoreClient == nil {
		log.Printf("[DisputeEscalationJob] Firestore client not available (test environment?)")
		result = "Skipped - no Firestore client available"
		return
	}

	// TODO: Implement dispute escalation logic
	// For now, simulate the result
	result = "Dispute escalation job completed (implementation needed)"
	log.Printf("[DisputeEscalationJob] Completed: %s (runtime: %v)", result, time.Since(start))
}

// Helper functions
func sendRatingReminder(playerID string, match *models.Match) {
	log.Printf("[RatingReminder] Sending reminder to player %s for match %s", playerID, match.ID)
	// TODO: Implement notification sending
}