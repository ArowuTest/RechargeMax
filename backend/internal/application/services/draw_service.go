package services

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	crand "crypto/rand"
	"encoding/binary"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"rechargemax/internal/domain/entities"
	"rechargemax/internal/domain/repositories"
)

// DrawService handles draw management and CSV export/import
type DrawService struct {
	db               *gorm.DB
	drawRepo         repositories.DrawRepository
	rechargeRepo     repositories.RechargeRepository
	subscriptionRepo repositories.SubscriptionRepository
	wheelSpinRepo    repositories.SpinResultRepository
}

// DrawEntryExport represents a draw entry for CSV export
type DrawEntryExport struct {
	MSISDN string
	Points int64
}

// WinnerImport represents a winner from CSV import
type WinnerImport struct {
	MSISDN   string
	Position int
	Prize    string
	Amount   int64
}

// NewDrawService creates a new draw service
func NewDrawService(
	db *gorm.DB,
	drawRepo repositories.DrawRepository,
	rechargeRepo repositories.RechargeRepository,
	subscriptionRepo repositories.SubscriptionRepository,
	wheelSpinRepo repositories.SpinResultRepository,
) *DrawService {
	return &DrawService{
		db:               db,
		drawRepo:         drawRepo,
		rechargeRepo:     rechargeRepo,
		subscriptionRepo: subscriptionRepo,
		wheelSpinRepo:    wheelSpinRepo,
	}
}

// generateDrawCode generates a unique draw code in the format DRAW-YYYYMMDD-XXXX
func generateDrawCode() string {
	var b [8]byte
	crand.Read(b[:]) //nolint:errcheck — Read never fails on Linux
	n := int(binary.BigEndian.Uint64(b[:]) % 9000)
	return fmt.Sprintf("DRAW-%s-%04d", time.Now().Format("20060102"), n+1000)
}

// CreateDraw creates a new draw record
func (s *DrawService) CreateDraw(ctx context.Context, name, description string, drawDate time.Time, prizePool int64) (*entities.Draw, error) {
	descPtr := &description
	drawTimePtr := &drawDate
	draw := &entities.Draw{
		ID:          uuid.New(),
		DrawCode:    generateDrawCode(),
		Name:        name,
		Type:        "DAILY",
		Description: descPtr,
		StartTime:   drawDate.Add(-24 * time.Hour), // Start 24h before draw
		EndTime:     drawDate,
		DrawTime:    drawTimePtr,
		Status:      "UPCOMING",
		PrizePool:   float64(prizePool),
	}

	err := s.drawRepo.Create(ctx, draw)
	if err != nil {
		return nil, fmt.Errorf("failed to create draw: %w", err)
	}

	return draw, nil
}

// CreateDrawWithTemplate creates a new draw with a prize template
func (s *DrawService) CreateDrawWithTemplate(
	ctx context.Context,
	name, description string,
	drawDate time.Time,
	drawTypeID, prizeTemplateID uuid.UUID,
) (*entities.Draw, error) {
	// Get prize template to calculate total prize pool
	var totalPrizePool float64
	var categories []entities.PrizeCategory
	
	err := s.db.Where("template_id = ?", prizeTemplateID).Find(&categories).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get prize categories: %w", err)
	}
	
	for _, cat := range categories {
		totalPrizePool += cat.PrizeAmount * float64(cat.WinnerCount)
	}
	
	descPtr := &description
	drawTimePtr := &drawDate
	draw := &entities.Draw{
		ID:              uuid.New(),
		DrawCode:        generateDrawCode(),
		Name:            name,
		Type:            "DAILY", // Will be updated based on draw type
		Description:     descPtr,
		StartTime:       drawDate.Add(-24 * time.Hour),
		EndTime:         drawDate,
		DrawTime:        drawTimePtr,
		Status:          "UPCOMING",
		PrizePool:       totalPrizePool,
		DrawTypeID:      &drawTypeID,
		PrizeTemplateID: &prizeTemplateID,
	}
	
	err = s.drawRepo.Create(ctx, draw)
	if err != nil {
		return nil, fmt.Errorf("failed to create draw: %w", err)
	}
	
	return draw, nil
}

// ExportDrawEntries aggregates draw entries from the draw_entries table for
// the given date range and writes a CSV to outputPath.
// Each row = one MSISDN + its total entries count across the period.
func (s *DrawService) ExportDrawEntries(ctx context.Context, startDate, endDate time.Time, outputPath string) (string, error) {
	// Aggregate entries count from draw_entries joined to draws (by date)
	type row struct {
		MSISDN       string
		TotalEntries int64
	}
	var rows []row

	err := s.db.WithContext(ctx).
		Table("draw_entries de").
		Select("de.msisdn, COALESCE(SUM(de.entries_count), 0) AS total_entries").
		Joins("JOIN draws d ON d.id = de.draw_id").
		Where("d.created_at BETWEEN ? AND ?", startDate, endDate).
		Group("de.msisdn").
		Order("total_entries DESC").
		Scan(&rows).Error
	if err != nil {
		return "", fmt.Errorf("failed to aggregate draw entries: %w", err)
	}

	// Create CSV file
	file, err := os.Create(outputPath)
	if err != nil {
		return "", fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write([]string{"MSISDN", "TotalEntries"}); err != nil {
		return "", fmt.Errorf("failed to write CSV header: %w", err)
	}
	for _, r := range rows {
		if err := writer.Write([]string{r.MSISDN, strconv.FormatInt(r.TotalEntries, 10)}); err != nil {
			return "", fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return outputPath, nil
}

func (s *DrawService) ImportWinners(ctx context.Context, drawID uuid.UUID, csvPath string) ([]*WinnerImport, error) {
	file, err := os.Open(csvPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Read header
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Validate header
	expectedHeader := []string{"MSISDN", "Position", "Prize", "Amount"}
	if len(header) != len(expectedHeader) {
		return nil, fmt.Errorf("invalid CSV header format. Expected: %v", expectedHeader)
	}

	var winners []*WinnerImport

	// Read winners
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV row: %w", err)
		}

		if len(record) != 4 {
			continue // Skip invalid rows
		}

		position, err := strconv.Atoi(record[1])
		if err != nil {
			continue // Skip invalid position
		}

		amount, err := strconv.ParseInt(record[3], 10, 64)
		if err != nil {
			continue // Skip invalid amount
		}

		winners = append(winners, &WinnerImport{
			MSISDN:   record[0],
			Position: position,
			Prize:    record[2],
			Amount:   amount,
		})
	}

	return winners, nil
}

// GetDrawByID retrieves a draw by ID
func (s *DrawService) GetDrawByID(ctx context.Context, drawID uuid.UUID) (*entities.Draw, error) {
	return s.drawRepo.FindByID(ctx, drawID)
}

// GetDraws retrieves all draws with pagination
func (s *DrawService) GetDraws(ctx context.Context, page, limit int) ([]*entities.Draw, int64, error) {
	// Calculate offset from page number
	offset := (page - 1) * limit
	
	draws, err := s.drawRepo.FindAll(ctx, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get draws: %w", err)
	}

	total, err := s.drawRepo.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count draws: %w", err)
	}

	return draws, total, nil
}

// UpdateDrawStatus updates the status of a draw
func (s *DrawService) UpdateDrawStatus(ctx context.Context, drawID uuid.UUID, status string) error {
	draw, err := s.drawRepo.FindByID(ctx, drawID)
	if err != nil {
		return fmt.Errorf("draw not found: %w", err)
	}

	draw.Status = status
	if status == "completed" {
		draw.CompletedAt = timePtr(time.Now())
	}

	return s.drawRepo.Update(ctx, draw)
}

// GetActiveDraw gets the currently active draw
func (s *DrawService) GetActiveDraw(ctx context.Context) (*entities.Draw, error) {
	draws, err := s.drawRepo.FindByStatus(ctx, "ACTIVE", 1, 0)
	if err != nil {
		return nil, err
	}
	if len(draws) == 0 {
		return nil, fmt.Errorf("no active draw found")
	}
	return draws[0], nil
}

// GetUpcomingDraws gets upcoming draws
func (s *DrawService) GetUpcomingDraws(ctx context.Context, limit int) ([]*entities.Draw, error) {
	return s.drawRepo.FindUpcoming(ctx, limit)
}

// GetCompletedDraws gets completed draws
func (s *DrawService) GetCompletedDraws(ctx context.Context, page, limit int) ([]*entities.Draw, int64, error) {
	draws, err := s.drawRepo.FindByStatus(ctx, "COMPLETED", 100, 0)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get completed draws: %w", err)
	}

	// For pagination, we'd need a proper method in the repository
	// This is a simplified version
	total := int64(len(draws))

	return draws, total, nil
}

// GetActiveDraws returns all active draws
func (s *DrawService) GetActiveDraws(ctx context.Context) ([]*entities.Draw, error) {
	draws, err := s.drawRepo.FindByStatus(ctx, "ACTIVE", 100, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get active draws: %w", err)
	}
	return draws, nil
}

// GetUserEntries returns user's entries for a specific draw
func (s *DrawService) GetUserEntries(ctx context.Context, drawID uuid.UUID, msisdn string) ([]DrawEntryResponse, error) {
	// Get draw to verify it exists and get date range
	draw, err := s.GetDrawByID(ctx, drawID)
	if err != nil {
		return nil, err
	}

	// Get user's recharges during draw period
	// This is a simplified implementation
	var entries []DrawEntryResponse

	// Implement actual entry retrieval logic
	// In production, this would:
	// 1. Query recharges for this user during draw period
	// 2. Calculate points from each recharge (₦200 = 1 point)
	// 3. Query subscriptions for this user during draw period
	// 4. Calculate subscription points (₦20/day = 1 point)
	// 5. Query wheel spins for bonus points
	// 6. Create entry records for each point
	//
	// Example implementation:
	// user, err := s.userRepo.FindByMSISDN(ctx, msisdn)
	// if err != nil {
	//     return nil, fmt.Errorf("user not found: %w", err)
	// }
	// 
	// // Get recharges during draw period
	// recharges, _ := s.rechargeRepo.FindByUserIDAndDateRange(ctx, user.ID, draw.StartDate, draw.EndDate)
	// for _, r := range recharges {
	//     if r.Status == "completed" {
	//         points := r.Amount / 20000 // ₦200 = 1 point
	//         for i := int64(0); i < points; i++ {
	//             entries = append(entries, DrawEntryResponse{
	//                 DrawID:      draw.ID,
	//                 MSISDN:      msisdn,
	//                 EntryNumber: fmt.Sprintf("%s-%d", r.ID.String(), i),
	//                 Source:      "recharge",
	//                 Amount:      r.Amount,
	//                 CreatedAt:   r.CreatedAt,
	//             })
	//         }
	//     }
	// }
	
	// Query draw_entries table for this user/draw combination
	type entryRow struct {
		ID          string    `gorm:"column:id"`
		DrawID      string    `gorm:"column:draw_id"`
		MSISDN      string    `gorm:"column:msisdn"`
		EntrySource string    `gorm:"column:entry_source"`
		Amount      int64     `gorm:"column:amount"`
		CreatedAt   time.Time `gorm:"column:created_at"`
	}
	var rows []entryRow
	s.db.WithContext(ctx).Raw(`
		SELECT id, draw_id, msisdn, entry_source, amount, created_at
		FROM draw_entries
		WHERE draw_id = ? AND msisdn = ?
		ORDER BY created_at ASC
	`, draw.ID, msisdn).Scan(&rows)

	userID := uuid.Nil
	if user, err2 := s.rechargeRepo.FindByReference(ctx, ""); err2 == nil && user != nil {
		// user lookup not needed here; MSISDN is enough
		_ = user
	}

	for _, r := range rows {
		id, _ := uuid.Parse(r.ID)
		drawUUID, _ := uuid.Parse(r.DrawID)
		entries = append(entries, DrawEntryResponse{
			ID:        id,
			DrawID:    drawUUID,
			UserID:    userID,
			MSISDN:    r.MSISDN,
			Amount:    r.Amount,
			EntryDate: r.CreatedAt,
		})
	}

	return entries, nil
}

// DrawEntryResponse represents a draw entry
type DrawEntryResponse struct {
	ID        uuid.UUID `json:"id"`
	DrawID    uuid.UUID `json:"draw_id"`
	UserID    uuid.UUID `json:"user_id"`
	MSISDN    string    `json:"msisdn"`
	Amount    int64     `json:"amount"` // Amount in kobo
	EntryDate time.Time `json:"entry_date"`
}

// GetDrawWinners returns winners for a specific draw
func (s *DrawService) GetDrawWinners(ctx context.Context, drawID uuid.UUID) ([]DrawWinnerResponse, error) {
	// Get draw to verify it exists
	_, err := s.GetDrawByID(ctx, drawID)
	if err != nil {
		return nil, err
	}

	// Get winners from winner repository
	// In production, this would:
	// 1. Query winners table for this draw_id
	// 2. Join with users table to get user details
	// 3. Return winner information with prize details
	//
	// Example implementation:
	// winners, err := s.winnerRepo.FindByDrawID(ctx, drawID)
	// if err != nil {
	//     return nil, fmt.Errorf("failed to get winners: %w", err)
	// }
	// 
	// var response []DrawWinnerResponse
	// for _, winner := range winners {
	//     user, _ := s.userRepo.FindByMSISDN(ctx, winner.MSISDN)
	//     response = append(response, DrawWinnerResponse{
	//         ID:          winner.ID,
	//         DrawID:      winner.DrawID,
	//         MSISDN:      winner.MSISDN,
	//         UserName:    user.FullName,
	//         PrizeName:   winner.PrizeName,
	//         PrizeValue:  winner.PrizeValue,
	//         ClaimStatus: winner.ClaimStatus,
	//         WonAt:       winner.CreatedAt,
	//     })
	// }
	// return response, nil
	
	// Query winners table for this draw
	type winnerRow struct {
		ID               string     `gorm:"column:id"`
		DrawID           string     `gorm:"column:draw_id"`
		MSISDN           string     `gorm:"column:msisdn"`
		PrizeType        string     `gorm:"column:prize_type"`
		PrizeDescription string     `gorm:"column:prize_description"`
		PrizeAmount      *int64     `gorm:"column:prize_amount"`
		ClaimStatus      string     `gorm:"column:claim_status"`
		CreatedAt        time.Time  `gorm:"column:created_at"`
	}
	var rows []winnerRow
	s.db.WithContext(ctx).Raw(`
		SELECT id, draw_id, msisdn, prize_type, prize_description, prize_amount, claim_status, created_at
		FROM winners
		WHERE draw_id = ?
		ORDER BY position ASC
	`, drawID).Scan(&rows)

	var response []DrawWinnerResponse
	for _, r := range rows {
		id, _ := uuid.Parse(r.ID)
		dID, _ := uuid.Parse(r.DrawID)
		prizeVal := int64(0)
		if r.PrizeAmount != nil {
			prizeVal = *r.PrizeAmount
		}
		response = append(response, DrawWinnerResponse{
			ID:         id,
			DrawID:     dID,
			MSISDN:     r.MSISDN,
			PrizeType:  r.PrizeType,
			PrizeValue: float64(prizeVal),
			Status:     r.ClaimStatus,
			WonAt:      r.CreatedAt,
		})
	}
	return response, nil
}

// DrawWinnerResponse represents a draw winner
type DrawWinnerResponse struct {
	ID         uuid.UUID `json:"id"`
	DrawID     uuid.UUID `json:"draw_id"`
	UserID     uuid.UUID `json:"user_id"`
	MSISDN     string    `json:"msisdn"`
	FullName   string    `json:"full_name"`
	PrizeType  string    `json:"prize_type"`
	PrizeValue float64   `json:"prize_value"`
	Status     string    `json:"status"`
	WonAt      time.Time `json:"won_at"`
}



// UpdateDraw updates draw details (admin operation)
func (s *DrawService) UpdateDraw(ctx context.Context, drawID string, updates map[string]interface{}) (*entities.Draw, error) {
	// Parse UUID
	did, err := uuid.Parse(drawID)
	if err != nil {
		return nil, fmt.Errorf("invalid draw ID format: %w", err)
	}
	
	// Get existing draw
	draw, err := s.drawRepo.FindByID(ctx, did)
	if err != nil {
		return nil, fmt.Errorf("draw not found: %w", err)
	}
	
	// Apply updates
	if name, ok := updates["name"].(string); ok {
		draw.Name = name
	}
	
	if description, ok := updates["description"].(string); ok {
		draw.Description = &description
	}
	
	if drawDate, ok := updates["draw_date"].(string); ok {
		parsedDate, err := time.Parse(time.RFC3339, drawDate)
		if err == nil {
			draw.DrawTime = &parsedDate
		}
	}
	
	if status, ok := updates["status"].(string); ok {
		validStatuses := []string{"PENDING", "ACTIVE", "COMPLETED", "CANCELLED", "UPCOMING"}
		isValid := false
		for _, s := range validStatuses {
			if status == s {
				isValid = true
				break
			}
		}
		if isValid {
			draw.Status = status
		}
	}
	
	if prizePool, ok := updates["prize_pool"].(float64); ok {
		draw.PrizePool = prizePool
	}
	
	// Update winners count and runner-ups count
	if winnersCount, ok := updates["winners_count"].(float64); ok {
		draw.WinnersCount = int(winnersCount)
	}
	
	if runnerUpsCount, ok := updates["runner_ups_count"].(float64); ok {
		draw.RunnerUpsCount = int(runnerUpsCount)
	}

	if prizeTemplateIDStr, ok := updates["prize_template_id"].(string); ok && prizeTemplateIDStr != "" {
		ptid, err := uuid.Parse(prizeTemplateIDStr)
		if err == nil {
			draw.PrizeTemplateID = &ptid
		}
	}

	// Save updated draw - use UpdateStatus for status-only updates to avoid draw_code unique constraint issues
	if len(updates) == 1 {
		if status, ok := updates["status"].(string); ok {
			if err := s.drawRepo.UpdateStatus(ctx, did, status); err != nil {
				return nil, fmt.Errorf("failed to update draw status: %w", err)
			}
			draw.Status = status
			return draw, nil
		}
	}
	// For other updates, use targeted column updates to avoid overwriting draw_code
	updateMap := map[string]interface{}{}
	if _, ok := updates["name"]; ok { updateMap["name"] = draw.Name }
	if _, ok := updates["description"]; ok { updateMap["description"] = draw.Description }
	if _, ok := updates["draw_date"]; ok { updateMap["draw_time"] = draw.DrawTime }
	if _, ok := updates["status"]; ok { updateMap["status"] = draw.Status }
	if _, ok := updates["prize_pool"]; ok { updateMap["prize_pool"] = draw.PrizePool }
	if _, ok := updates["winners_count"]; ok { updateMap["winners_count"] = draw.WinnersCount }
	if _, ok := updates["runner_ups_count"]; ok { updateMap["runner_ups_count"] = draw.RunnerUpsCount }
	if _, ok := updates["prize_template_id"]; ok { updateMap["prize_template_id"] = draw.PrizeTemplateID }
	if len(updateMap) > 0 {
		if err := s.db.Model(draw).Updates(updateMap).Error; err != nil {
			return nil, fmt.Errorf("failed to update draw: %w", err)
		}
	}
	
	return draw, nil
}

// ExecuteDraw executes a draw (triggers winner selection)
func (s *DrawService) ExecuteDraw(ctx context.Context, drawID string) error {
	// Parse UUID
	did, err := uuid.Parse(drawID)
	if err != nil {
		return fmt.Errorf("invalid draw ID format: %w", err)
	}
	
	// Get draw
	draw, err := s.drawRepo.FindByID(ctx, did)
	if err != nil {
		return fmt.Errorf("draw not found: %w", err)
	}
	
	// Validate draw status
	if draw.Status == "COMPLETED" {
		return fmt.Errorf("draw has already been executed")
	}
	
	if draw.Status == "CANCELLED" {
		return fmt.Errorf("draw has been cancelled")
	}
	
	// Check if draw has entries — query live count from draw_entries table
	var entryCount int64
	if err := s.db.Table("draw_entries").Where("draw_id = ?", did).Count(&entryCount).Error; err != nil {
		return fmt.Errorf("failed to count draw entries: %w", err)
	}
	if entryCount == 0 {
		return fmt.Errorf("no entries found for this draw")
	}
	
	// Update draw status to completed
	draw.Status = "COMPLETED"
	now := time.Now()
	draw.DrawTime = &now
	draw.CompletedAt = &now
	
	if err := s.drawRepo.Update(ctx, draw); err != nil {
		return fmt.Errorf("failed to update draw status: %w", err)
	}
	
	// Prize Tier System: Category-aware winner selection
	// 1. Load prize categories from template
	// 2. Get all unique MSISDNs from draw entries
	// 3. Select winners for each category (no duplicates across categories)
	// 4. Store winners with category information
	
	// Load prize categories
	if draw.PrizeTemplateID == nil {
		return fmt.Errorf("draw does not have a prize template assigned")
	}
	
	var prizeCategories []entities.PrizeCategory
	if err := s.db.Where("template_id = ?", *draw.PrizeTemplateID).
		Order("display_order ASC").
		Find(&prizeCategories).Error; err != nil {
		return fmt.Errorf("failed to load prize categories: %w", err)
	}
	
	if len(prizeCategories) == 0 {
		return fmt.Errorf("no prize categories found for this template")
	}
	
	// Get all unique MSISDNs from draw entries
	var uniqueMSISDNs []string
	if err := s.db.Table("draw_entries").
		Where("draw_id = ?", did).
		Distinct("msisdn").
		Pluck("msisdn", &uniqueMSISDNs).Error; err != nil {
		return fmt.Errorf("failed to get unique MSISDNs: %w", err)
	}
	
	if len(uniqueMSISDNs) == 0 {
		return fmt.Errorf("no unique MSISDNs found in draw entries")
	}
	
	// Track selected MSISDNs to prevent duplicates across categories
	selectedMSISDNs := make(map[string]bool)
	var allWinners []entities.DrawWinners
	position := 1
	
	// Select winners for each prize category
	for _, category := range prizeCategories {
		// Filter out already selected MSISDNs
		availableMSISDNs := make([]string, 0)
		for _, msisdn := range uniqueMSISDNs {
			if !selectedMSISDNs[msisdn] {
				availableMSISDNs = append(availableMSISDNs, msisdn)
			}
		}
		
		// Check if we have enough MSISDNs
		totalNeeded := category.WinnerCount + category.RunnerUpCount
		if len(availableMSISDNs) < totalNeeded {
			return fmt.Errorf("insufficient unique MSISDNs for category %s: need %d, have %d",
				category.CategoryName, totalNeeded, len(availableMSISDNs))
		}
		
		// Crypto-random Fisher-Yates shuffle (SEC-009: crypto/rand, not math/rand)
		shuffledMSISDNs := make([]string, len(availableMSISDNs))
		copy(shuffledMSISDNs, availableMSISDNs)
		cryptoShuffle(shuffledMSISDNs)
		
		// Select winners for this category
		for i := 0; i < category.WinnerCount; i++ {
			msisdn := shuffledMSISDNs[i]
			selectedMSISDNs[msisdn] = true
			
				categoryID := category.ID
				categoryName := category.CategoryName
				winner := entities.DrawWinners{
					ID:              uuid.New(),
					DrawID:          did,
					MSISDN:          msisdn,
					Position:        position,
					PrizeAmount:     int64(category.PrizeAmount),
					IsRunnerUp:      false,
					PrizeCategoryID: &categoryID,
					CategoryName:    &categoryName,
					CreatedAt:       &now,
				}
				allWinners = append(allWinners, winner)
				position++
			}
			
			// Select runner-ups for this category
			for i := category.WinnerCount; i < totalNeeded; i++ {
				msisdn := shuffledMSISDNs[i]
				selectedMSISDNs[msisdn] = true
				
				categoryID := category.ID
				categoryName := category.CategoryName
				runnerUp := entities.DrawWinners{
					ID:              uuid.New(),
					DrawID:          did,
					MSISDN:          msisdn,
					Position:        position,
					PrizeAmount:     int64(category.PrizeAmount),
				IsRunnerUp:      true,
				PrizeCategoryID: &categoryID,
				CategoryName:    &categoryName,
				CreatedAt:       &now,
			}
			allWinners = append(allWinners, runnerUp)
			position++
		}
	}
	
	// Save all winners to database
	if len(allWinners) > 0 {
		if err := s.db.Create(&allWinners).Error; err != nil {
			return fmt.Errorf("failed to save winners: %w", err)
		}
	}
	
	// Update draw statistics
	draw.TotalWinners = len(allWinners)
	if err := s.drawRepo.Update(ctx, draw); err != nil {
		return fmt.Errorf("failed to update draw statistics: %w", err)
	}
	
	return nil
}

// ProcessCSVEntries processes a CSV file containing MSISDN and Points
// Format: MSISDN,Points
// Example: 08012345678,5
func (s *DrawService) ProcessCSVEntries(ctx context.Context, drawID uuid.UUID, csvReader io.Reader) (int, error) {
	reader := csv.NewReader(csvReader)
	reader.FieldsPerRecord = 2
	reader.TrimLeadingSpace = true
	
	entriesCreated := 0
	lineNumber := 0
	
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return entriesCreated, fmt.Errorf("error reading CSV at line %d: %w", lineNumber, err)
		}
		
		lineNumber++
		
		// Skip header row if present
		if lineNumber == 1 && (record[0] == "MSISDN" || record[0] == "msisdn") {
			continue
		}
		
		msisdn := record[0]
		pointsStr := record[1]
		
		// Validate MSISDN format (Nigerian: 080/081/070/090/091 + 8 digits)
		if !isValidNigerianMSISDN(msisdn) {
			return entriesCreated, fmt.Errorf("invalid MSISDN format at line %d: %s", lineNumber, msisdn)
		}
		
		// Parse points
		points, err := strconv.Atoi(pointsStr)
		if err != nil || points <= 0 {
			return entriesCreated, fmt.Errorf("invalid points value at line %d: %s", lineNumber, pointsStr)
		}
		
		// Create N entries for this MSISDN based on points
		for i := 0; i < points; i++ {
			now := time.Now()
			entry := &entities.DrawEntries{
				ID:         uuid.New(),
				DrawID:     drawID,
				MSISDN:     msisdn,
				SourceType: "MANUAL",
				CreatedAt:  &now,
			}

			if err := s.drawRepo.CreateEntry(ctx, entry); err != nil {
				return entriesCreated, fmt.Errorf("failed to create entry for %s: %w", msisdn, err)
			}

			entriesCreated++
		}
	}
	
	return entriesCreated, nil
}

// isValidNigerianMSISDN validates Nigerian phone number format
func isValidNigerianMSISDN(msisdn string) bool {
	// Remove any spaces or dashes
	msisdn = strings.ReplaceAll(msisdn, " ", "")
	msisdn = strings.ReplaceAll(msisdn, "-", "")
	
	// Must be 11 digits starting with 0, or 13 digits starting with 234
	if len(msisdn) == 11 && msisdn[0] == '0' {
		// Check if starts with valid prefix
		validPrefixes := []string{
			// MTN
			"0803", "0806", "0810", "0813", "0814", "0816", "0903", "0906", "0913", "0916",
			"0801", "0804", "0818", "0819", "0909",
			// Airtel
			"0802", "0808", "0812", "0701", "0708", "0902", "0907",
			// Glo
			"0805", "0807", "0811", "0815", "0905", "0915",
			// 9mobile
			"0809", "0817", "0818", "0909", "0908", "0904", "0912",
		}
		for _, vp := range validPrefixes {
			if strings.HasPrefix(msisdn, vp) {
				return true
			}
		}
	} else if len(msisdn) == 13 && strings.HasPrefix(msisdn, "234") {
		// Convert to 0-format and validate
		return isValidNigerianMSISDN("0" + msisdn[3:])
	}
	
	return false
}



// cryptoShuffle performs a Fisher-Yates shuffle using crypto/rand to ensure
// unpredictable draw winner selection (SEC-009).
func cryptoShuffle(s []string) {
	for i := len(s) - 1; i > 0; i-- {
		var b [8]byte
		crand.Read(b[:]) //nolint:errcheck
		j := int(binary.BigEndian.Uint64(b[:]) % uint64(i+1))
		s[i], s[j] = s[j], s[i]
	}
}
