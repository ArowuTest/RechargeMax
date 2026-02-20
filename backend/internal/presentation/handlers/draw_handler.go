package handlers

import (
	"strings"
	"time"
	
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"rechargemax/internal/application/services"
	"rechargemax/internal/domain/entities"
	"rechargemax/internal/errors"
	"rechargemax/internal/middleware"
	"rechargemax/internal/validation"
)

type DrawHandler struct {
	drawService *services.DrawService
}

func NewDrawHandler(drawService *services.DrawService) *DrawHandler {
	return &DrawHandler{drawService: drawService}
}

// CreateDraw godoc
// @Summary Create a new draw
// @Description Create a new lottery draw (Admin only)
// @Tags admin, draws
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Draw creation request"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /admin/draws [post]
func (h *DrawHandler) CreateDraw(c *gin.Context) {
	var req struct {
		Name            string    `json:"name" binding:"required"`
		Description     string    `json:"description"`
		DrawDate        time.Time `json:"draw_date"`
		DrawTypeID      uint      `json:"draw_type_id"`
		PrizeTemplateID *uint     `json:"prize_template_id"`
		PrizePool       *int64    `json:"prize_pool"`
		DurationHours   int       `json:"duration_hours"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, errors.BadRequest("Invalid request format"))
		return
	}

	// Set default draw date if not provided
	if req.DrawDate.IsZero() {
		if req.DurationHours > 0 {
			req.DrawDate = time.Now().Add(time.Duration(req.DurationHours) * time.Hour)
		} else {
			req.DrawDate = time.Now().Add(24 * time.Hour)
		}
	}

	// Validate either prize_template_id or prize_pool is provided
	if req.PrizeTemplateID == nil && req.PrizePool == nil {
		middleware.RespondWithError(c, errors.BadRequest("Either prize_template_id or prize_pool must be provided"))
		return
	}

	// If prize_pool is provided, validate it
	if req.PrizePool != nil && *req.PrizePool <= 0 {
		middleware.RespondWithError(c, errors.BadRequest("Prize pool must be greater than 0"))
		return
	}

	// Create draw with appropriate method
	var draw *entities.Draw
	var err error
	
	if req.PrizeTemplateID != nil {
		// Create draw with prize template
		draw, err = h.drawService.CreateDrawWithTemplate(
			c.Request.Context(),
			req.Name,
			req.Description,
			req.DrawDate,
			req.DrawTypeID,
			*req.PrizeTemplateID,
		)
	} else {
		// Create draw with manual prize pool (legacy)
		draw, err = h.drawService.CreateDraw(
			c.Request.Context(),
			req.Name,
			req.Description,
			req.DrawDate,
			*req.PrizePool,
		)
	}
	if err != nil {
		middleware.RespondWithError(c, err)
		return
	}

	// Log draw creation
	errors.Info("Draw created", map[string]interface{}{
		"draw_id":    draw.ID,
		"name":       req.Name,
		"draw_date":  req.DrawDate,
		"prize_pool": req.PrizePool,
	})

	middleware.RespondWithSuccess(c, draw)
}

// ExportEntries godoc
// @Summary Export draw entries to CSV
// @Description Export all draw entries for a date range to CSV file (Admin only)
// @Tags admin, draws
// @Produce text/csv
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Success 200 {file} file "CSV file"
// @Failure 400 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /admin/draws/export [get]
func (h *DrawHandler) ExportEntries(c *gin.Context) {
	// Parse date range from query parameters
	startDateStr := c.DefaultQuery("start_date", time.Now().AddDate(0, 0, -30).Format("2006-01-02"))
	endDateStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		middleware.RespondWithError(c, errors.BadRequest("Invalid start_date format. Use YYYY-MM-DD"))
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		middleware.RespondWithError(c, errors.BadRequest("Invalid end_date format. Use YYYY-MM-DD"))
		return
	}

	// Validate date range
	if err := validation.ValidateDateRange(startDateStr, endDateStr); err != nil {
		middleware.RespondWithError(c, errors.BadRequest(err.Error()))
		return
	}

	csvData, err := h.drawService.ExportDrawEntries(c.Request.Context(), startDate, endDate, "/tmp/draw_entries.csv")
	if err != nil {
		middleware.RespondWithError(c, err)
		return
	}

	// Log export
	errors.Info("Draw entries exported", map[string]interface{}{
		"start_date": startDate,
		"end_date":   endDate,
	})

	c.Header("Content-Disposition", "attachment; filename=draw_entries_"+startDateStr+"_to_"+endDateStr+".csv")
	c.Data(200, "text/csv", []byte(csvData))
}

// ImportWinners godoc
// @Summary Import winners from CSV
// @Description Import draw winners from CSV file (Admin only)
// @Tags admin, draws
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Draw ID"
// @Param file formData file true "CSV file with winners"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /admin/draws/{id}/import-winners [post]
func (h *DrawHandler) ImportWinners(c *gin.Context) {
	drawIDStr := c.Param("id")
	drawID, err := uuid.Parse(drawIDStr)
	if err != nil {
		middleware.RespondWithError(c, errors.BadRequest("Invalid draw ID"))
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		middleware.RespondWithError(c, errors.BadRequest("No file uploaded"))
		return
	}

	// Validate file type
	if file.Header.Get("Content-Type") != "text/csv" && !strings.HasSuffix(strings.ToLower(file.Filename), ".csv") {
		middleware.RespondWithError(c, errors.BadRequest("File must be a CSV file"))
		return
	}
	
	// Save file temporarily
	filepath := "/tmp/" + file.Filename
	if err := c.SaveUploadedFile(file, filepath); err != nil {
		middleware.RespondWithError(c, errors.Internal("Failed to save uploaded file").WithError(err))
		return
	}

	count, err := h.drawService.ImportWinners(c.Request.Context(), drawID, filepath)
	if err != nil {
		middleware.RespondWithError(c, err)
		return
	}

	// Log import
	errors.Info("Winners imported", map[string]interface{}{
		"draw_id": drawID,
		"count":   count,
	})

	middleware.RespondWithSuccess(c, map[string]interface{}{
		"message": "Winners imported successfully",
		"count":   count,
	})
}

// GetActiveDraws godoc
// @Summary Get active draws
// @Description Get list of currently active draws
// @Tags draws
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} errors.ErrorResponse
// @Router /draws/active [get]
func (h *DrawHandler) GetActiveDraws(c *gin.Context) {
	draws, err := h.drawService.GetActiveDraws(c.Request.Context())
	if err != nil {
		middleware.RespondWithError(c, err)
		return
	}

	middleware.RespondWithSuccess(c, draws)
}

// GetDrawByID godoc
// @Summary Get draw by ID
// @Description Get details of a specific draw
// @Tags draws
// @Produce json
// @Param id path string true "Draw ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Router /draws/{id} [get]
func (h *DrawHandler) GetDrawByID(c *gin.Context) {
	drawIDStr := c.Param("id")
	drawID, err := uuid.Parse(drawIDStr)
	if err != nil {
		middleware.RespondWithError(c, errors.BadRequest("Invalid draw ID"))
		return
	}

	draw, err := h.drawService.GetDrawByID(c.Request.Context(), drawID)
	if err != nil {
		middleware.RespondWithError(c, errors.NotFound("Draw not found"))
		return
	}

	middleware.RespondWithSuccess(c, draw)
}

// GetMyEntries godoc
// @Summary Get user's draw entries
// @Description Get all draw entries for the authenticated user
// @Tags draws
// @Produce json
// @Param draw_id query string false "Draw ID (defaults to active draw)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Router /draws/my-entries [get]
func (h *DrawHandler) GetMyEntries(c *gin.Context) {
	msisdn := c.GetString("msisdn")
	if msisdn == "" {
		middleware.RespondWithError(c, errors.Unauthorized("User not authenticated"))
		return
	}

	// Get draw ID from query parameter, or use active draw
	drawIDStr := c.Query("draw_id")
	var drawID uuid.UUID
	var err error

	if drawIDStr != "" {
		drawID, err = uuid.Parse(drawIDStr)
		if err != nil {
			middleware.RespondWithError(c, errors.BadRequest("Invalid draw ID"))
			return
		}
	} else {
		// Get active draw
		activeDraw, err := h.drawService.GetActiveDraw(c.Request.Context())
		if err != nil {
			middleware.RespondWithError(c, errors.NotFound("No active draw found"))
			return
		}
		drawID = activeDraw.ID
	}

	entries, err := h.drawService.GetUserEntries(c.Request.Context(), drawID, msisdn)
	if err != nil {
		middleware.RespondWithError(c, err)
		return
	}

	middleware.RespondWithSuccess(c, entries)
}

// GetDrawWinners godoc
// @Summary Get draw winners
// @Description Get list of winners for a specific draw
// @Tags draws
// @Produce json
// @Param id path string true "Draw ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /draws/{id}/winners [get]
func (h *DrawHandler) GetDrawWinners(c *gin.Context) {
	drawIDStr := c.Param("id")
	drawID, err := uuid.Parse(drawIDStr)
	if err != nil {
		middleware.RespondWithError(c, errors.BadRequest("Invalid draw ID"))
		return
	}

	winners, err := h.drawService.GetDrawWinners(c.Request.Context(), drawID)
	if err != nil {
		middleware.RespondWithError(c, err)
		return
	}

	middleware.RespondWithSuccess(c, winners)
}

// GetDraw godoc
// @Summary Get draw (alias for GetDrawByID)
// @Description Get details of a specific draw
// @Tags draws
// @Produce json
// @Param id path string true "Draw ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Router /draw/{id} [get]
func (h *DrawHandler) GetDraw(c *gin.Context) {
	// Alias for GetDrawByID
	h.GetDrawByID(c)
}

// GetDraws godoc
// @Summary Get all draws
// @Description Get list of all draws with pagination
// @Tags draws
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /draws [get]
func (h *DrawHandler) GetDraws(c *gin.Context) {
	// Parse pagination parameters
	var pagination validation.PaginationRequest
	if err := c.ShouldBindQuery(&pagination); err != nil {
		middleware.RespondWithError(c, errors.BadRequest("Invalid pagination parameters"))
		return
	}

	// Validate pagination
	if err := pagination.Validate(); err != nil {
		middleware.RespondWithValidationError(c, err)
		return
	}

	// Get all draws (this would need to be implemented in DrawService)
	draws, err := h.drawService.GetActiveDraws(c.Request.Context())
	if err != nil {
		middleware.RespondWithError(c, err)
		return
	}

	middleware.RespondWithSuccess(c, map[string]interface{}{
		"draws": draws,
		"page":  pagination.Page,
		"limit": pagination.Limit,
	})
}

// UploadDrawEntries godoc
// @Summary Upload draw entries via CSV
// @Description Upload MSISDN and Points data for a draw via CSV file (Admin only)
// @Tags admin, draws
// @Accept multipart/form-data
// @Produce json
// @Param draw_id path string true "Draw ID"
// @Param file formData file true "CSV file with MSISDN,Points format"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /admin/draws/{draw_id}/upload-entries [post]
func (h *DrawHandler) UploadDrawEntries(c *gin.Context) {
	drawIDStr := c.Param("draw_id")
	drawID, err := uuid.Parse(drawIDStr)
	if err != nil {
		middleware.RespondWithError(c, errors.BadRequest("Invalid draw ID"))
		return
	}

	// Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		middleware.RespondWithError(c, errors.BadRequest("No file uploaded"))
		return
	}

	// Validate file extension
	if !strings.HasSuffix(strings.ToLower(file.Filename), ".csv") {
		middleware.RespondWithError(c, errors.BadRequest("File must be a CSV"))
		return
	}

	// Open the file
	src, err := file.Open()
	if err != nil {
		middleware.RespondWithError(c, errors.Internal("Failed to open file"))
		return
	}
	defer src.Close()

	// Process CSV entries
	entriesCreated, err := h.drawService.ProcessCSVEntries(c.Request.Context(), drawID, src)
	if err != nil {
		middleware.RespondWithError(c, errors.Internal(err.Error()))
		return
	}

	middleware.RespondWithSuccess(c, gin.H{
		"message": "Entries uploaded successfully",
		"draw_id": drawID,
		"entries_created": entriesCreated,
	})
}
