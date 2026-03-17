package services

import (
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"rechargemax/internal/domain/entities"
	"rechargemax/internal/domain/repositories"
)

// PointsService handles user points management and tracking
type PointsService struct {
	userRepo         repositories.UserRepository
	rechargeRepo     repositories.RechargeRepository
	ussdRepo         repositories.USSDRechargeRepository
	subscriptionRepo repositories.SubscriptionRepository
	spinRepo         repositories.SpinRepository
	adjustmentRepo   repositories.PointsAdjustmentRepository
	notificationService *NotificationService
}

// NewPointsService creates a new points service
func NewPointsService(
	userRepo repositories.UserRepository,
	rechargeRepo repositories.RechargeRepository,
	ussdRepo repositories.USSDRechargeRepository,
	subscriptionRepo repositories.SubscriptionRepository,
	spinRepo repositories.SpinRepository,
	adjustmentRepo repositories.PointsAdjustmentRepository,
	notificationService *NotificationService,
) *PointsService {
	return &PointsService{
		userRepo:         userRepo,
		rechargeRepo:     rechargeRepo,
		ussdRepo:         ussdRepo,
		subscriptionRepo: subscriptionRepo,
		spinRepo:         spinRepo,
		adjustmentRepo:   adjustmentRepo,
		notificationService: notificationService,
	}
}

// UserPointsSummary represents a user's points summary
type UserPointsSummary struct {
	UserID           uuid.UUID `json:"user_id"`
	MSISDN           string    `json:"msisdn"`
	Email            string    `json:"email"`
	FullName         string    `json:"full_name"`
	TotalPoints      int       `json:"total_points"`
	AvailablePoints  int       `json:"available_points"`
	LockedPoints     int       `json:"locked_points"`
	LifetimePoints   int       `json:"lifetime_points"`
	LastEarnedAt     *time.Time `json:"last_earned_at"`
	PointsBySource   map[string]int `json:"points_by_source"`
}

// PointsHistoryEntry represents a single points transaction
type PointsHistoryEntry struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"user_id"`
	MSISDN      string     `json:"msisdn"`
	Points      int        `json:"points"`
	Source      string     `json:"source"`
	Description string     `json:"description"`
	ReferenceID *uuid.UUID `json:"reference_id"`
	Status      string     `json:"status"`
	CreatedBy   *uuid.UUID `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
}

// PointsStatistics represents overall points statistics
type PointsStatistics struct {
	TotalUsers           int            `json:"total_users"`
	TotalPointsIssued    int            `json:"total_points_issued"`
	TotalPointsAvailable int            `json:"total_points_available"`
	TotalPointsLocked    int            `json:"total_points_locked"`
	PointsBySource       map[string]int `json:"points_by_source"`
	TopUsers             []UserPointsSummary `json:"top_users"`
}

// GetUsersWithPoints retrieves all users with their points summary
func (s *PointsService) GetUsersWithPoints(ctx context.Context, searchQuery string, dateFrom, dateTo *time.Time) ([]*UserPointsSummary, error) {
	users, err := s.userRepo.FindAll(ctx, 10000, 0) // Get all users
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve users: %w", err)
	}

	var summaries []*UserPointsSummary
	for _, user := range users {
		// Apply search filter
		if searchQuery != "" {
			lowerQuery := strings.ToLower(searchQuery)
			if !strings.Contains(strings.ToLower(user.MSISDN), lowerQuery) &&
				!strings.Contains(strings.ToLower(user.Email), lowerQuery) &&
				!strings.Contains(strings.ToLower(user.FullName), lowerQuery) {
				continue
			}
		}

		summary, err := s.getUserPointsSummary(ctx, user.ID, dateFrom, dateTo)
		if err != nil {
			continue // Skip users with errors
		}

		summaries = append(summaries, summary)
	}

	return summaries, nil
}

// getUserPointsSummary calculates points summary for a single user
func (s *PointsService) getUserPointsSummary(ctx context.Context, userID uuid.UUID, dateFrom, dateTo *time.Time) (*UserPointsSummary, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	summary := &UserPointsSummary{
		UserID:         user.ID,
		MSISDN:         user.MSISDN,
		Email:          user.Email,
		FullName:       user.FullName,
		TotalPoints:    user.TotalPoints,
		LifetimePoints: user.TotalPoints,
		PointsBySource: make(map[string]int),
	}

	// Calculate points by source
	pointsBySource, err := s.calculatePointsBySource(ctx, userID, dateFrom, dateTo)
	if err == nil {
		summary.PointsBySource = pointsBySource
	}

	// Get last earned date
	lastEarned, err := s.getLastPointsEarnedDate(ctx, userID)
	if err == nil && lastEarned != nil {
		summary.LastEarnedAt = lastEarned
	}

	// For now, assume all points are available (locked points would be in active draws)
	summary.AvailablePoints = user.TotalPoints
	summary.LockedPoints = 0

	return summary, nil
}

// calculatePointsBySource calculates points distribution by source
func (s *PointsService) calculatePointsBySource(ctx context.Context, userID uuid.UUID, dateFrom, dateTo *time.Time) (map[string]int, error) {
	pointsBySource := make(map[string]int)

	// Platform recharges
	recharges, err := s.rechargeRepo.FindByUserID(ctx, userID, 10000, 0)
	if err == nil {
		for _, r := range recharges {
			if dateFrom != nil && r.CreatedAt.Before(*dateFrom) {
				continue
			}
			if dateTo != nil && r.CreatedAt.After(*dateTo) {
				continue
			}
			if r.Status == "SUCCESS" {
				pointsBySource["platform_recharge"] += int(r.Amount / 20000) // ₦200 = 1 point
			}
		}
	}

	// USSD recharges
	user, err := s.userRepo.FindByID(ctx, userID)
	if err == nil {
		var startDate, endDate time.Time
		if dateFrom != nil {
			startDate = *dateFrom
		}
		if dateTo != nil {
			endDate = *dateTo
		}
		ussdRecharges, err := s.ussdRepo.FindByMSISDN(ctx, user.MSISDN, startDate, endDate)
		if err == nil {
			for _, ur := range ussdRecharges {
			if dateFrom != nil && ur.ReceivedAt.Before(*dateFrom) {
				continue
			}
			if dateTo != nil && ur.ReceivedAt.After(*dateTo) {
				continue
			}
			pointsBySource["ussd_recharge"] += ur.PointsEarned
			}
		}
	}

	// Wheel spins
	spins, err := s.spinRepo.FindByUserID(ctx, userID, 10000, 0)
	if err == nil {
		for _, spin := range spins {
			if dateFrom != nil && spin.CreatedAt.Before(*dateFrom) {
				continue
			}
			if dateTo != nil && spin.CreatedAt.After(*dateTo) {
				continue
			}
			if spin.PrizeType == "points" {
				pointsBySource["wheel_spin"] += int(spin.PrizeValue)
			}
		}
	}

	// Admin adjustments
	adjustments, err := s.adjustmentRepo.FindByUserID(ctx, userID)
	if err == nil {
		for _, adj := range adjustments {
			if dateFrom != nil && adj.CreatedAt.Before(*dateFrom) {
				continue
			}
			if dateTo != nil && adj.CreatedAt.After(*dateTo) {
				continue
			}
			if adj.Points > 0 {
				pointsBySource["admin_added"] += adj.Points
			} else {
				pointsBySource["admin_deducted"] += -adj.Points
			}
		}
	}

	return pointsBySource, nil
}

// getLastPointsEarnedDate gets the last date points were earned
func (s *PointsService) getLastPointsEarnedDate(ctx context.Context, userID uuid.UUID) (*time.Time, error) {
	var lastDate *time.Time

	// Check latest recharge
	recharges, err := s.rechargeRepo.FindByUserID(ctx, userID, 10000, 0)
	if err == nil {
		for _, r := range recharges {
			if r.Status == "SUCCESS" {
				if lastDate == nil || r.CreatedAt.After(*lastDate) {
					lastDate = &r.CreatedAt
				}
			}
		}
	}

	// Check latest USSD recharge
	user2, err := s.userRepo.FindByID(ctx, userID)
	if err == nil {
		ussdRecharges, err := s.ussdRepo.FindByMSISDN(ctx, user2.MSISDN, time.Time{}, time.Time{})
		if err == nil {
		for _, ur := range ussdRecharges {
			if lastDate == nil || ur.ReceivedAt.After(*lastDate) {
				lastDate = &ur.ReceivedAt
			}
		}
		}
	}

	// Check latest spin
	spins, err := s.spinRepo.FindByUserID(ctx, userID, 10000, 0)
	if err == nil {
		for _, spin := range spins {
			if spin.PrizeType == "points" {
				if lastDate == nil || spin.CreatedAt.After(*lastDate) {
					lastDate = &spin.CreatedAt
				}
			}
		}
	}

	return lastDate, nil
}

// GetPointsHistory retrieves points transaction history
func (s *PointsService) GetPointsHistory(ctx context.Context, userID *uuid.UUID, source string, dateFrom, dateTo *time.Time) ([]*PointsHistoryEntry, error) {
	var history []*PointsHistoryEntry

	// If userID is specified, get history for that user
	if userID != nil {
		userHistory, err := s.getUserPointsHistory(ctx, *userID, source, dateFrom, dateTo)
		if err != nil {
			return nil, err
		}
		history = append(history, userHistory...)
	} else {
		// Get history for all users
		users, err := s.userRepo.FindAll(ctx, 10000, 0) // Get all users
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve users: %w", err)
		}

		for _, user := range users {
			userHistory, err := s.getUserPointsHistory(ctx, user.ID, source, dateFrom, dateTo)
			if err != nil {
				continue
			}
			history = append(history, userHistory...)
		}
	}

	return history, nil
}

// getUserPointsHistory gets points history for a specific user
func (s *PointsService) getUserPointsHistory(ctx context.Context, userID uuid.UUID, source string, dateFrom, dateTo *time.Time) ([]*PointsHistoryEntry, error) {
	var history []*PointsHistoryEntry

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Platform recharges
	if source == "" || source == "platform_recharge" {
		recharges, err := s.rechargeRepo.FindByUserID(ctx, userID, 10000, 0)
		if err == nil {
			for _, r := range recharges {
				if dateFrom != nil && r.CreatedAt.Before(*dateFrom) {
					continue
				}
				if dateTo != nil && r.CreatedAt.After(*dateTo) {
					continue
				}
				if r.Status == "SUCCESS" {
					points := int(r.Amount / 20000)
					history = append(history, &PointsHistoryEntry{
						ID:          r.ID,
						UserID:      userID,
						MSISDN:      user.MSISDN,
						Points:      points,
						Source:      "platform_recharge",
						Description: fmt.Sprintf("Recharge of ₦%.2f", float64(r.Amount)/100),
						ReferenceID: &r.ID,
						Status:      "completed",
						CreatedAt:   r.CreatedAt,
					})
				}
			}
		}
	}

	// USSD recharges
	if source == "" || source == "ussd_recharge" {
		user3, err := s.userRepo.FindByID(ctx, userID)
		if err == nil {
			ussdRecharges, err := s.ussdRepo.FindByMSISDN(ctx, user3.MSISDN, time.Time{}, time.Time{})
			if err == nil {
				for _, ur := range ussdRecharges {
				if dateFrom != nil && ur.ReceivedAt.Before(*dateFrom) {
					continue
				}
				if dateTo != nil && ur.ReceivedAt.After(*dateTo) {
					continue
				}
				history = append(history, &PointsHistoryEntry{
					ID:          ur.ID,
					UserID:      userID,
					MSISDN:      user.MSISDN,
					Points:      ur.PointsEarned,
					Source:      "ussd_recharge",
					Description: fmt.Sprintf("USSD recharge of ₦%.2f on %s", float64(ur.Amount)/100, ur.Network),
					ReferenceID: &ur.ID,
					Status:      "completed",
					CreatedAt:   ur.ReceivedAt,
				})
				}
			}
		}
	}

	// Wheel spins
	if source == "" || source == "wheel_spin" {
		spins, err := s.spinRepo.FindByUserID(ctx, userID, 10000, 0)
		if err == nil {
			for _, spin := range spins {
				if dateFrom != nil && spin.CreatedAt.Before(*dateFrom) {
					continue
				}
				if dateTo != nil && spin.CreatedAt.After(*dateTo) {
					continue
				}
				if spin.PrizeType == "points" {
					history = append(history, &PointsHistoryEntry{
						ID:          spin.ID,
						UserID:      userID,
						MSISDN:      user.MSISDN,
						Points:      int(spin.PrizeValue),
						Source:      "wheel_spin",
						Description: fmt.Sprintf("Won %d points from wheel spin", spin.PrizeValue),
						ReferenceID: &spin.ID,
						Status:      "completed",
						CreatedAt:   spin.CreatedAt,
					})
				}
			}
		}
	}

	// Admin adjustments
	if source == "" || source == "admin_adjustment" {
		adjustments, err := s.adjustmentRepo.FindByUserID(ctx, userID)
		if err == nil {
			for _, adj := range adjustments {
				if dateFrom != nil && adj.CreatedAt.Before(*dateFrom) {
					continue
				}
				if dateTo != nil && adj.CreatedAt.After(*dateTo) {
					continue
				}
				history = append(history, &PointsHistoryEntry{
					ID:          adj.ID,
					UserID:      userID,
					MSISDN:      user.MSISDN,
					Points:      adj.Points,
					Source:      "admin_adjustment",
					Description: fmt.Sprintf("%s - %s", adj.Reason, adj.Description),
					ReferenceID: &adj.ID,
					Status:      "completed",
					CreatedBy:   &adj.AdminID,
					CreatedAt:   adj.CreatedAt,
				})
			}
		}
	}

	return history, nil
}

// AdjustUserPoints manually adjusts user points (admin function)
func (s *PointsService) AdjustUserPoints(ctx context.Context, userID uuid.UUID, points int, reason, description string, adminID uuid.UUID) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Update user points
	newTotal := user.TotalPoints + points
	if newTotal < 0 {
		return fmt.Errorf("adjustment would result in negative points balance")
	}

	user.TotalPoints = newTotal
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user points: %w", err)
	}

	// Create adjustment record
	adjustment := &entities.PointsAdjustment{
		ID:          uuid.New(),
		UserID:      userID,
		Points:      points,
		Reason:      reason,
		Description: description,
		AdminID:     adminID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.adjustmentRepo.Create(ctx, adjustment); err != nil {
		return fmt.Errorf("failed to create adjustment record: %w", err)
	}

	// Notify user of points adjustment
	if s.notificationService != nil {
		adjustmentType := "added"
		absPoints := points
		if points < 0 {
			adjustmentType = "deducted"
			absPoints = -points
		}
		msg := fmt.Sprintf("%d loyalty points have been %s to your RechargeMax account. Reason: %s", absPoints, adjustmentType, reason)
		go s.notificationService.SendSMS(ctx, user.MSISDN, msg)
	}

	return nil
}

// GetPointsStatistics retrieves overall points statistics
func (s *PointsService) GetPointsStatistics(ctx context.Context, dateFrom, dateTo *time.Time) (*PointsStatistics, error) {
	stats := &PointsStatistics{
		PointsBySource: make(map[string]int),
	}

	users, err := s.userRepo.FindAll(ctx, 10000, 0) // Get all users
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve users: %w", err)
	}

	stats.TotalUsers = len(users)

	// Calculate statistics
	var userSummaries []UserPointsSummary
	for _, user := range users {
		stats.TotalPointsIssued += user.TotalPoints
		stats.TotalPointsAvailable += user.TotalPoints // Simplified

		// Get points by source for this user
		pointsBySource, err := s.calculatePointsBySource(ctx, user.ID, dateFrom, dateTo)
		if err == nil {
			for source, points := range pointsBySource {
				stats.PointsBySource[source] += points
			}
		}

		// Collect user summary for top users calculation
		summary, err := s.getUserPointsSummary(ctx, user.ID, dateFrom, dateTo)
		if err == nil {
			userSummaries = append(userSummaries, *summary)
		}
	}

	// Get top 10 users by points using efficient sorting
	stats.TopUsers = s.getTopUsersByPoints(userSummaries, 10)

	return stats, nil
}

// getTopUsersByPoints returns top N users sorted by total points
func (s *PointsService) getTopUsersByPoints(users []UserPointsSummary, topN int) []UserPointsSummary {
	// Quick sort implementation for better performance
	if len(users) <= 1 {
		return users
	}

	// Sort users by total points (descending)
	s.quickSortUsers(users, 0, len(users)-1)

	// Return top N
	if len(users) > topN {
		return users[:topN]
	}
	return users
}

// quickSortUsers sorts users by total points in descending order
func (s *PointsService) quickSortUsers(users []UserPointsSummary, low, high int) {
	if low < high {
		pivot := s.partitionUsers(users, low, high)
		s.quickSortUsers(users, low, pivot-1)
		s.quickSortUsers(users, pivot+1, high)
	}
}

// partitionUsers partitions the users array for quicksort
func (s *PointsService) partitionUsers(users []UserPointsSummary, low, high int) int {
	pivot := users[high].TotalPoints
	i := low - 1

	for j := low; j < high; j++ {
		if users[j].TotalPoints > pivot { // Descending order
			i++
			users[i], users[j] = users[j], users[i]
		}
	}

	users[i+1], users[high] = users[high], users[i+1]
	return i + 1
}

// ExportUsersWithPointsToCSV exports users with points to CSV format
func (s *PointsService) ExportUsersWithPointsToCSV(ctx context.Context, searchQuery string, dateFrom, dateTo *time.Time) (string, error) {
	users, err := s.GetUsersWithPoints(ctx, searchQuery, dateFrom, dateTo)
	if err != nil {
		return "", err
	}

	var csvBuilder strings.Builder
	writer := csv.NewWriter(&csvBuilder)

	// Write header
	header := []string{"MSISDN", "Email", "Full Name", "Total Points", "Available Points", "Locked Points", "Lifetime Points", "Last Earned", "Platform Recharge", "USSD Recharge", "Wheel Spin", "Admin Added", "Admin Deducted"}
	writer.Write(header)

	// Write data
	for _, user := range users {
		lastEarned := ""
		if user.LastEarnedAt != nil {
			lastEarned = user.LastEarnedAt.Format("2006-01-02 15:04:05")
		}

		row := []string{
			user.MSISDN,
			user.Email,
			user.FullName,
			strconv.Itoa(user.TotalPoints),
			strconv.Itoa(user.AvailablePoints),
			strconv.Itoa(user.LockedPoints),
			strconv.Itoa(user.LifetimePoints),
			lastEarned,
			strconv.Itoa(user.PointsBySource["platform_recharge"]),
			strconv.Itoa(user.PointsBySource["ussd_recharge"]),
			strconv.Itoa(user.PointsBySource["wheel_spin"]),
			strconv.Itoa(user.PointsBySource["admin_added"]),
			strconv.Itoa(user.PointsBySource["admin_deducted"]),
		}
		writer.Write(row)
	}

	writer.Flush()
	return csvBuilder.String(), nil
}

// ExportPointsHistoryToCSV exports points history to CSV format
func (s *PointsService) ExportPointsHistoryToCSV(ctx context.Context, userID *uuid.UUID, source string, dateFrom, dateTo *time.Time) (string, error) {
	history, err := s.GetPointsHistory(ctx, userID, source, dateFrom, dateTo)
	if err != nil {
		return "", err
	}

	var csvBuilder strings.Builder
	writer := csv.NewWriter(&csvBuilder)

	// Write header
	header := []string{"Date", "MSISDN", "Points", "Source", "Description", "Status", "Created By"}
	writer.Write(header)

	// Write data
	for _, entry := range history {
		createdBy := ""
		if entry.CreatedBy != nil {
			createdBy = entry.CreatedBy.String()
		}

		row := []string{
			entry.CreatedAt.Format("2006-01-02 15:04:05"),
			entry.MSISDN,
			strconv.Itoa(entry.Points),
			entry.Source,
			entry.Description,
			entry.Status,
			createdBy,
		}
		writer.Write(row)
	}

	writer.Flush()
	return csvBuilder.String(), nil
}
