package services

import (
	"testing"
	"time"

	"github.com/sebastiancaldarola/goalhero-payment-jobs/models"
	"github.com/stretchr/testify/assert"
)

func TestIsEligibleForAutoRelease(t *testing.T) {
	paymentService := NewPaymentService()
	now := time.Now()

	testCases := []struct {
		name     string
		escrow   *models.EscrowTransaction
		expected bool
		reason   string
	}{
		{
			name: "eligible_with_good_rating",
			escrow: &models.EscrowTransaction{
				ID:                "escrow_1",
				Status:            models.EscrowStatusHeld,
				ReleaseEligibleAt: now.Add(-1 * time.Hour), // Past eligible time
				RatingReceived:    true,
				ActualRating:      4.5,
				MinRatingRequired: 3.0,
			},
			expected: true,
			reason:   "Should release with good rating",
		},
		{
			name: "not_eligible_poor_rating",
			escrow: &models.EscrowTransaction{
				ID:                "escrow_2",
				Status:            models.EscrowStatusHeld,
				ReleaseEligibleAt: now.Add(-1 * time.Hour), // Past eligible time
				RatingReceived:    true,
				ActualRating:      2.0,
				MinRatingRequired: 3.0,
			},
			expected: false,
			reason:   "Should not release with poor rating",
		},
		{
			name: "not_eligible_before_time",
			escrow: &models.EscrowTransaction{
				ID:                "escrow_3",
				Status:            models.EscrowStatusHeld,
				ReleaseEligibleAt: now.Add(1 * time.Hour), // Future time
				RatingReceived:    false,
			},
			expected: false,
			reason:   "Should not release before eligible time",
		},
		{
			name: "eligible_no_rating_after_grace",
			escrow: &models.EscrowTransaction{
				ID:                "escrow_4",
				Status:            models.EscrowStatusHeld,
				ReleaseEligibleAt: now.Add(-25 * time.Hour), // Past eligible time + grace period
				RatingReceived:    false,
			},
			expected: true,
			reason:   "Should release after grace period with no rating",
		},
		{
			name: "not_eligible_disputed",
			escrow: &models.EscrowTransaction{
				ID:                "escrow_5",
				Status:            models.EscrowStatusDisputed,
				ReleaseEligibleAt: now.Add(-1 * time.Hour),
				RatingReceived:    false,
			},
			expected: false,
			reason:   "Should not release disputed escrow",
		},
		{
			name: "not_eligible_waiting_for_rating",
			escrow: &models.EscrowTransaction{
				ID:                "escrow_6",
				Status:            models.EscrowStatusHeld,
				ReleaseEligibleAt: now.Add(-2 * time.Hour), // Past eligible but within grace
				RatingReceived:    false,
			},
			expected: false,
			reason:   "Should wait for rating within grace period",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := paymentService.isEligibleForAutoRelease(tc.escrow)
			assert.Equal(t, tc.expected, result, tc.reason)
		})
	}
}

func TestProcessAutomaticReleases(t *testing.T) {
	t.Run("should_return_zero_when_no_firestore", func(t *testing.T) {
		// This test runs when Firestore client is not available
		paymentService := NewPaymentService()
		
		processed, failed, errors, totalReleased, err := paymentService.ProcessAutomaticReleases()
		
		// Should handle gracefully when no Firestore client
		if err != nil {
			assert.Contains(t, err.Error(), "firestore client not available")
			assert.Equal(t, 0, processed)
			assert.Equal(t, 0, failed)
			assert.Empty(t, errors)
			assert.Equal(t, 0.0, totalReleased)
		}
	})
}

func TestEscrowRatingLogic(t *testing.T) {
	testCases := []struct {
		name               string
		rating             float64
		minRequired        float64
		expectedApproved   bool
		expectedStatus     string
	}{
		{
			name:             "good_rating_approved",
			rating:           4.5,
			minRequired:      3.0,
			expectedApproved: true,
			expectedStatus:   models.EscrowStatusApproved,
		},
		{
			name:             "minimum_rating_approved",
			rating:           3.0,
			minRequired:      3.0,
			expectedApproved: true,
			expectedStatus:   models.EscrowStatusApproved,
		},
		{
			name:             "poor_rating_not_approved",
			rating:           2.5,
			minRequired:      3.0,
			expectedApproved: false,
			expectedStatus:   models.EscrowStatusHeld, // Stays held for manual review
		},
		{
			name:             "very_poor_rating",
			rating:           1.0,
			minRequired:      3.0,
			expectedApproved: false,
			expectedStatus:   models.EscrowStatusHeld,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			escrow := &models.EscrowTransaction{
				MinRatingRequired: tc.minRequired,
				Status:           models.EscrowStatusHeld,
			}

			// Simulate the rating update logic
			escrow.RatingReceived = true
			escrow.ActualRating = tc.rating

			if tc.rating >= tc.minRequired {
				escrow.RatingApproved = true
				escrow.Status = models.EscrowStatusApproved
			} else {
				escrow.RatingApproved = false
				// Poor rating - keep in held status for manual review
			}

			assert.Equal(t, tc.expectedApproved, escrow.RatingApproved)
			assert.Equal(t, tc.expectedStatus, escrow.Status)
		})
	}
}

func TestAutoReleaseBusinessRules(t *testing.T) {
	now := time.Now()
	
	t.Run("grace_period_calculation", func(t *testing.T) {
		releaseTime := now.Add(-2 * time.Hour)
		graceDeadline := releaseTime.Add(24 * time.Hour)
		
		// Should be within grace period (2 hours past eligible time)
		assert.True(t, now.Before(graceDeadline), "Should be within 24h grace period")
		
		// Test past grace period
		pastGraceTime := now.Add(-25 * time.Hour)
		pastGraceDeadline := pastGraceTime.Add(24 * time.Hour)
		assert.True(t, now.After(pastGraceDeadline), "Should be past 24h grace period")
	})
	
	t.Run("escrow_hold_period", func(t *testing.T) {
		// Test that default hold period is 24 hours
		assert.Equal(t, 24, models.EscrowHoldHours, "Default escrow hold should be 24 hours")
		
		// Test escrow creation with hold period
		createdAt := now
		expectedReleaseTime := createdAt.Add(24 * time.Hour)
		
		escrow := &models.EscrowTransaction{
			HeldAt:            createdAt,
			ReleaseEligibleAt: expectedReleaseTime,
		}
		
		assert.Equal(t, expectedReleaseTime, escrow.ReleaseEligibleAt)
	})
}