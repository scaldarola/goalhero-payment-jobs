package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestJobConfig(t *testing.T) {
	t.Run("should create job config with correct values", func(t *testing.T) {
		config := &JobConfig{
			RatingReminderInterval:    24 * time.Hour,
			AutoReleaseInterval:       1 * time.Hour,
			DisputeEscalationInterval: 48 * time.Hour,
			RatingDeadlineDays:        7,
			MinRatingForAutoRelease:   3.0,
			DisputeEscalationHours:    72,
		}

		assert.Equal(t, 24*time.Hour, config.RatingReminderInterval)
		assert.Equal(t, 1*time.Hour, config.AutoReleaseInterval)
		assert.Equal(t, 48*time.Hour, config.DisputeEscalationInterval)
		assert.Equal(t, 7, config.RatingDeadlineDays)
		assert.Equal(t, 3.0, config.MinRatingForAutoRelease)
		assert.Equal(t, 72, config.DisputeEscalationHours)
	})
}

func TestJobStatus(t *testing.T) {
	t.Run("should create job status with correct fields", func(t *testing.T) {
		now := time.Now()
		status := &JobStatus{
			JobName:        "Test Job",
			LastRun:        now,
			NextScheduled:  now.Add(1 * time.Hour),
			LastResult:     "Success",
			RunCount:       5,
			ErrorCount:     1,
			AverageRuntime: "100ms",
			IsRunning:      false,
			Enabled:        true,
		}

		assert.Equal(t, "Test Job", status.JobName)
		assert.Equal(t, now, status.LastRun)
		assert.Equal(t, now.Add(1*time.Hour), status.NextScheduled)
		assert.Equal(t, "Success", status.LastResult)
		assert.Equal(t, 5, status.RunCount)
		assert.Equal(t, 1, status.ErrorCount)
		assert.Equal(t, "100ms", status.AverageRuntime)
		assert.False(t, status.IsRunning)
		assert.True(t, status.Enabled)
	})
}

func TestJobHealth(t *testing.T) {
	t.Run("should create job health with correct fields", func(t *testing.T) {
		now := time.Now()
		statuses := make(map[string]*JobStatus)
		statuses["test_job"] = &JobStatus{
			JobName:    "Test Job",
			RunCount:   10,
			ErrorCount: 2,
		}

		health := &JobHealth{
			Healthy:          true,
			TotalJobs:        1,
			RunningJobs:      0,
			FailedJobs:       0,
			LastHealthCheck:  now,
			JobStatuses:      statuses,
		}

		assert.True(t, health.Healthy)
		assert.Equal(t, 1, health.TotalJobs)
		assert.Equal(t, 0, health.RunningJobs)
		assert.Equal(t, 0, health.FailedJobs)
		assert.Equal(t, now, health.LastHealthCheck)
		assert.Len(t, health.JobStatuses, 1)
	})
}

func TestBackgroundJobManager(t *testing.T) {
	t.Run("should create job manager with config", func(t *testing.T) {
		config := &JobConfig{
			RatingReminderInterval:    1 * time.Hour,
			AutoReleaseInterval:       30 * time.Minute,
			DisputeEscalationInterval: 2 * time.Hour,
		}

		manager := &BackgroundJobManager{
			config:   config,
			shutdown: make(chan struct{}),
			running:  true,
		}

		assert.Equal(t, config, manager.config)
		assert.True(t, manager.running)
		assert.NotNil(t, manager.shutdown)
	})
}

func TestJobStatuses(t *testing.T) {
	t.Run("should initialize job statuses", func(t *testing.T) {
		jobConfig := &JobConfig{
			RatingReminderInterval:    1 * time.Hour,
			AutoReleaseInterval:       30 * time.Minute,
			DisputeEscalationInterval: 2 * time.Hour,
		}

		// Clear existing job statuses
		statusMutex.Lock()
		jobStatuses = make(map[string]*JobStatus)
		statusMutex.Unlock()

		initializeJobStatuses(jobConfig)

		statuses := GetJobStatuses()
		assert.Len(t, statuses, 3)

		assert.Contains(t, statuses, "rating_reminder")
		assert.Contains(t, statuses, "auto_release")
		assert.Contains(t, statuses, "dispute_escalation")

		ratingStatus := statuses["rating_reminder"]
		assert.Equal(t, "Rating Reminder", ratingStatus.JobName)
		assert.Equal(t, "Not run yet", ratingStatus.LastResult)
		assert.True(t, ratingStatus.Enabled)
		assert.False(t, ratingStatus.IsRunning)
	})

	t.Run("should get job statuses", func(t *testing.T) {
		// Initialize test data
		statusMutex.Lock()
		jobStatuses = map[string]*JobStatus{
			"test_job": {
				JobName:       "Test Job",
				LastResult:    "Success",
				RunCount:      5,
				ErrorCount:    0,
				IsRunning:     false,
				Enabled:       true,
			},
		}
		statusMutex.Unlock()

		statuses := GetJobStatuses()
		assert.Len(t, statuses, 1)
		
		testJob := statuses["test_job"]
		assert.Equal(t, "Test Job", testJob.JobName)
		assert.Equal(t, "Success", testJob.LastResult)
		assert.Equal(t, 5, testJob.RunCount)
	})

	t.Run("should calculate job health correctly", func(t *testing.T) {
		// Initialize test data with healthy and unhealthy jobs
		statusMutex.Lock()
		jobStatuses = map[string]*JobStatus{
			"healthy_job": {
				JobName:    "Healthy Job",
				RunCount:   10,
				ErrorCount: 1, // 10% error rate
				IsRunning:  false,
			},
			"unhealthy_job": {
				JobName:    "Unhealthy Job",
				RunCount:   10,
				ErrorCount: 8, // 80% error rate
				IsRunning:  false,
			},
			"running_job": {
				JobName:    "Running Job",
				RunCount:   5,
				ErrorCount: 0,
				IsRunning:  true,
			},
		}
		statusMutex.Unlock()

		health := GetJobHealth()
		
		assert.Equal(t, 3, health.TotalJobs)
		assert.Equal(t, 1, health.RunningJobs)
		assert.Equal(t, 1, health.FailedJobs) // Only unhealthy_job should be marked as failed
		assert.False(t, health.Healthy) // Should be unhealthy due to failed jobs
		assert.Len(t, health.JobStatuses, 3)
	})
}

func TestJobTriggers(t *testing.T) {
	t.Run("should return error when job manager is nil", func(t *testing.T) {
		jobManager = nil

		err := TriggerRatingReminder()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "job manager not initialized")

		err = TriggerAutoRelease()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "job manager not initialized")

		err = TriggerDisputeEscalation()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "job manager not initialized")
	})

	t.Run("should trigger jobs when job manager exists", func(t *testing.T) {
		// Mock a job manager
		mockConfig := &JobConfig{
			RatingReminderInterval:    1 * time.Hour,
			AutoReleaseInterval:       30 * time.Minute,
			DisputeEscalationInterval: 2 * time.Hour,
		}
		jobManager = &BackgroundJobManager{
			config:   mockConfig,
			shutdown: make(chan struct{}),
			running:  true,
		}

		err := TriggerRatingReminder()
		assert.NoError(t, err)

		err = TriggerAutoRelease()
		assert.NoError(t, err)

		err = TriggerDisputeEscalation()
		assert.NoError(t, err)
	})
}

func TestUpdateJobConfig(t *testing.T) {
	t.Run("should return error when job manager is nil", func(t *testing.T) {
		jobManager = nil

		newConfig := &JobConfig{
			RatingReminderInterval: 2 * time.Hour,
		}

		err := UpdateJobConfig(newConfig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "job manager not initialized")
	})

	t.Run("should update job configuration successfully", func(t *testing.T) {
		// Mock a job manager
		mockConfig := &JobConfig{
			RatingReminderInterval: 1 * time.Hour,
		}
		jobManager = &BackgroundJobManager{
			config:   mockConfig,
			shutdown: make(chan struct{}),
			running:  true,
		}

		// Initialize some job statuses
		statusMutex.Lock()
		jobStatuses = map[string]*JobStatus{
			"rating_reminder": {
				JobName:       "Rating Reminder",
				NextScheduled: time.Now().Add(1 * time.Hour),
			},
		}
		statusMutex.Unlock()

		newConfig := &JobConfig{
			RatingReminderInterval:    2 * time.Hour,
			AutoReleaseInterval:       1 * time.Hour,
			DisputeEscalationInterval: 3 * time.Hour,
		}

		err := UpdateJobConfig(newConfig)
		assert.NoError(t, err)
		
		// Verify the config was updated
		assert.Equal(t, newConfig, jobManager.config)
	})
}

func TestGetJobConfig(t *testing.T) {
	t.Run("should return job configuration when manager exists", func(t *testing.T) {
		mockConfig := &JobConfig{
			RatingReminderInterval: 1 * time.Hour,
			AutoReleaseInterval:    30 * time.Minute,
		}
		jobManager = &BackgroundJobManager{
			config: mockConfig,
		}

		config := GetJobConfig()
		assert.Equal(t, mockConfig, config)
	})

	t.Run("should return nil when job manager is nil", func(t *testing.T) {
		jobManager = nil

		config := GetJobConfig()
		assert.Nil(t, config)
	})
}

func TestUpdateJobStatus(t *testing.T) {
	t.Run("should update job status correctly", func(t *testing.T) {
		// Initialize test job status
		statusMutex.Lock()
		jobStatuses = map[string]*JobStatus{
			"test_job": {
				JobName:    "Test Job",
				RunCount:   5,
				ErrorCount: 1,
				IsRunning:  true,
			},
		}
		statusMutex.Unlock()

		// Mock job manager for next scheduled calculation
		mockConfig := &JobConfig{
			RatingReminderInterval: 1 * time.Hour,
		}
		jobManager = &BackgroundJobManager{
			config: mockConfig,
		}

		runtime := 500 * time.Millisecond
		updateJobStatus("test_job", "Success", runtime, false)

		statuses := GetJobStatuses()
		updatedStatus := statuses["test_job"]

		assert.Equal(t, "Success", updatedStatus.LastResult)
		assert.Equal(t, 6, updatedStatus.RunCount) // Should increment
		assert.Equal(t, 1, updatedStatus.ErrorCount) // Should not increment for success
		assert.Equal(t, runtime.String(), updatedStatus.AverageRuntime)
		assert.False(t, updatedStatus.IsRunning)
		assert.True(t, updatedStatus.LastRun.After(time.Now().Add(-1*time.Second)))
	})

	t.Run("should increment error count on error", func(t *testing.T) {
		// Initialize test job status
		statusMutex.Lock()
		jobStatuses = map[string]*JobStatus{
			"error_job": {
				JobName:    "Error Job",
				RunCount:   3,
				ErrorCount: 1,
				IsRunning:  true,
			},
		}
		statusMutex.Unlock()

		runtime := 200 * time.Millisecond
		updateJobStatus("error_job", "Failed", runtime, true)

		statuses := GetJobStatuses()
		updatedStatus := statuses["error_job"]

		assert.Equal(t, "Failed", updatedStatus.LastResult)
		assert.Equal(t, 4, updatedStatus.RunCount)
		assert.Equal(t, 2, updatedStatus.ErrorCount) // Should increment for error
		assert.False(t, updatedStatus.IsRunning)
	})
}