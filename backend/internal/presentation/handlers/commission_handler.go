package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CommissionHandler struct {
	db *gorm.DB
}

func NewCommissionHandler(db *gorm.DB) *CommissionHandler {
	return &CommissionHandler{db: db}
}

// CommissionReconciliationRequest represents the request for commission reconciliation
type CommissionReconciliationRequest struct {
	StartDate string `json:"start_date" binding:"required"`
	EndDate   string `json:"end_date" binding:"required"`
	Network   string `json:"network"`   // Optional: filter by network
	Provider  string `json:"provider"`  // Optional: filter by provider
}

// CommissionReconciliationResponse represents the commission reconciliation data
type CommissionReconciliationResponse struct {
	Summary struct {
		TotalTransactions   int64   `json:"total_transactions"`
		TotalRechargeAmount int64   `json:"total_recharge_amount"`
		TotalCommission     int64   `json:"total_commission"`
		AverageCommission   float64 `json:"average_commission"`
		CommissionRate      float64 `json:"commission_rate"`
	} `json:"summary"`

	ByNetwork []struct {
		Network           string  `json:"network"`
		TransactionCount  int64   `json:"transaction_count"`
		TotalAmount       int64   `json:"total_amount"`
		TotalCommission   int64   `json:"total_commission"`
		AverageCommission float64 `json:"average_commission"`
		CommissionRate    float64 `json:"commission_rate"`
	} `json:"by_network"`

	ByProvider []struct {
		Provider          string  `json:"provider"`
		TransactionCount  int64   `json:"transaction_count"`
		TotalAmount       int64   `json:"total_amount"`
		TotalCommission   int64   `json:"total_commission"`
		AverageCommission float64 `json:"average_commission"`
		CommissionRate    float64 `json:"commission_rate"`
	} `json:"by_provider"`

	ByDate []struct {
		Date             string `json:"date"`
		TransactionCount int64  `json:"transaction_count"`
		TotalAmount      int64  `json:"total_amount"`
		TotalCommission  int64  `json:"total_commission"`
	} `json:"by_date"`

	RecentTransactions []struct {
		ID             string    `json:"id"`
		MSISDN         string    `json:"msisdn"`
		Network        string    `json:"network"`
		Provider       string    `json:"provider"`
		Amount         int64     `json:"amount"`
		Commission     int64     `json:"commission"`
		CommissionRate float64   `json:"commission_rate"`
		Status         string    `json:"status"`
		CreatedAt      time.Time `json:"created_at"`
	} `json:"recent_transactions"`
}

// GetCommissionReconciliation returns real commission data from DB
func (h *CommissionHandler) GetCommissionReconciliation(c *gin.Context) {
	var req CommissionReconciliationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid start_date format. Use YYYY-MM-DD"})
		return
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid end_date format. Use YYYY-MM-DD"})
		return
	}
	// Extend end to end-of-day
	endDate = endDate.Add(24*time.Hour - time.Second)

	response := CommissionReconciliationResponse{}

	// ── Summary ────────────────────────────────────────────────────────────────
	type summaryRow struct {
		TotalTransactions   int64
		TotalRechargeAmount int64
		TotalCommission     int64
	}
	var sumRow summaryRow

	q := h.db.Table("transactions").
		Select("COUNT(*) AS total_transactions, COALESCE(SUM(amount),0) AS total_recharge_amount, COALESCE(SUM(commission_amount),0) AS total_commission").
		Where("created_at BETWEEN ? AND ? AND status = 'SUCCESS'", startDate, endDate)

	if req.Network != "" {
		q = q.Where("network = ?", strings.ToUpper(req.Network))
	}
	if req.Provider != "" {
		q = q.Where("provider = ?", req.Provider)
	}
	q.Scan(&sumRow)

	response.Summary.TotalTransactions = sumRow.TotalTransactions
	response.Summary.TotalRechargeAmount = sumRow.TotalRechargeAmount
	response.Summary.TotalCommission = sumRow.TotalCommission
	if sumRow.TotalTransactions > 0 {
		response.Summary.AverageCommission = float64(sumRow.TotalCommission) / float64(sumRow.TotalTransactions)
	}
	if sumRow.TotalRechargeAmount > 0 {
		response.Summary.CommissionRate = float64(sumRow.TotalCommission) / float64(sumRow.TotalRechargeAmount) * 100
	}

	// ── By Network ─────────────────────────────────────────────────────────────
	type netRow struct {
		Network          string
		TransactionCount int64
		TotalAmount      int64
		TotalCommission  int64
	}
	var netRows []netRow
	h.db.Table("transactions").
		Select("network, COUNT(*) AS transaction_count, COALESCE(SUM(amount),0) AS total_amount, COALESCE(SUM(commission_amount),0) AS total_commission").
		Where("created_at BETWEEN ? AND ? AND status = 'SUCCESS'", startDate, endDate).
		Group("network").
		Scan(&netRows)

	for _, r := range netRows {
		avg := float64(0)
		rate := float64(0)
		if r.TransactionCount > 0 {
			avg = float64(r.TotalCommission) / float64(r.TransactionCount)
		}
		if r.TotalAmount > 0 {
			rate = float64(r.TotalCommission) / float64(r.TotalAmount) * 100
		}
		response.ByNetwork = append(response.ByNetwork, struct {
			Network           string  `json:"network"`
			TransactionCount  int64   `json:"transaction_count"`
			TotalAmount       int64   `json:"total_amount"`
			TotalCommission   int64   `json:"total_commission"`
			AverageCommission float64 `json:"average_commission"`
			CommissionRate    float64 `json:"commission_rate"`
		}{
			Network: r.Network, TransactionCount: r.TransactionCount,
			TotalAmount: r.TotalAmount, TotalCommission: r.TotalCommission,
			AverageCommission: avg, CommissionRate: rate,
		})
	}

	// ── By Provider ────────────────────────────────────────────────────────────
	type provRow struct {
		Provider         string
		TransactionCount int64
		TotalAmount      int64
		TotalCommission  int64
	}
	var provRows []provRow
	h.db.Table("transactions").
		Select("provider, COUNT(*) AS transaction_count, COALESCE(SUM(amount),0) AS total_amount, COALESCE(SUM(commission_amount),0) AS total_commission").
		Where("created_at BETWEEN ? AND ? AND status = 'SUCCESS'", startDate, endDate).
		Group("provider").
		Scan(&provRows)

	for _, r := range provRows {
		avg := float64(0)
		rate := float64(0)
		if r.TransactionCount > 0 {
			avg = float64(r.TotalCommission) / float64(r.TransactionCount)
		}
		if r.TotalAmount > 0 {
			rate = float64(r.TotalCommission) / float64(r.TotalAmount) * 100
		}
		response.ByProvider = append(response.ByProvider, struct {
			Provider          string  `json:"provider"`
			TransactionCount  int64   `json:"transaction_count"`
			TotalAmount       int64   `json:"total_amount"`
			TotalCommission   int64   `json:"total_commission"`
			AverageCommission float64 `json:"average_commission"`
			CommissionRate    float64 `json:"commission_rate"`
		}{
			Provider: r.Provider, TransactionCount: r.TransactionCount,
			TotalAmount: r.TotalAmount, TotalCommission: r.TotalCommission,
			AverageCommission: avg, CommissionRate: rate,
		})
	}

	// ── By Date ────────────────────────────────────────────────────────────────
	type dateRow struct {
		Day              string
		TransactionCount int64
		TotalAmount      int64
		TotalCommission  int64
	}
	var dateRows []dateRow
	h.db.Table("transactions").
		Select("DATE(created_at) AS day, COUNT(*) AS transaction_count, COALESCE(SUM(amount),0) AS total_amount, COALESCE(SUM(commission_amount),0) AS total_commission").
		Where("created_at BETWEEN ? AND ? AND status = 'SUCCESS'", startDate, endDate).
		Group("DATE(created_at)").
		Order("day ASC").
		Scan(&dateRows)

	for _, r := range dateRows {
		response.ByDate = append(response.ByDate, struct {
			Date             string `json:"date"`
			TransactionCount int64  `json:"transaction_count"`
			TotalAmount      int64  `json:"total_amount"`
			TotalCommission  int64  `json:"total_commission"`
		}{Date: r.Day, TransactionCount: r.TransactionCount, TotalAmount: r.TotalAmount, TotalCommission: r.TotalCommission})
	}

	// ── Recent Transactions (last 20) ─────────────────────────────────────────
	type txnRow struct {
		ID             string
		Msisdn         string
		Network        string
		Provider       string
		Amount         int64
		CommissionAmt  int64
		Status         string
		CreatedAt      time.Time
	}
	var txns []txnRow
	h.db.Table("transactions").
		Select("id, msisdn, network, provider, amount, commission_amount AS commission_amt, status, created_at").
		Where("created_at BETWEEN ? AND ? AND status = 'SUCCESS'", startDate, endDate).
		Order("created_at DESC").
		Limit(20).
		Scan(&txns)

	for _, t := range txns {
		rate := float64(0)
		if t.Amount > 0 {
			rate = float64(t.CommissionAmt) / float64(t.Amount) * 100
		}
		msisdn := t.Msisdn
		if len(msisdn) > 7 {
			msisdn = msisdn[:4] + "****" + msisdn[len(msisdn)-3:]
		}
		response.RecentTransactions = append(response.RecentTransactions, struct {
			ID             string    `json:"id"`
			MSISDN         string    `json:"msisdn"`
			Network        string    `json:"network"`
			Provider       string    `json:"provider"`
			Amount         int64     `json:"amount"`
			Commission     int64     `json:"commission"`
			CommissionRate float64   `json:"commission_rate"`
			Status         string    `json:"status"`
			CreatedAt      time.Time `json:"created_at"`
		}{
			ID: t.ID, MSISDN: msisdn, Network: t.Network, Provider: t.Provider,
			Amount: t.Amount, Commission: t.CommissionAmt, CommissionRate: rate,
			Status: t.Status, CreatedAt: t.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": response})
}

// ExportCommissionReport exports commission data as CSV
func (h *CommissionHandler) ExportCommissionReport(c *gin.Context) {
	var req CommissionReconciliationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid start_date"})
		return
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid end_date"})
		return
	}
	endDate = endDate.Add(24*time.Hour - time.Second)

	type txnRow struct {
		CreatedAt     time.Time
		ID            string
		Msisdn        string
		Network       string
		Provider      string
		Amount        int64
		CommissionAmt int64
		Status        string
	}

	var txns []txnRow
	h.db.Table("transactions").
		Select("created_at, id, msisdn, network, provider, amount, commission_amount AS commission_amt, status").
		Where("created_at BETWEEN ? AND ? AND status = 'SUCCESS'", startDate, endDate).
		Order("created_at ASC").
		Scan(&txns)

	var sb strings.Builder
	sb.WriteString("Date,Transaction ID,Phone Number,Network,Provider,Amount (₦),Commission (₦),Commission Rate (%),Status\n")
	for _, t := range txns {
		commRate := float64(0)
		if t.Amount > 0 {
			commRate = float64(t.CommissionAmt) / float64(t.Amount) * 100
		}
		sb.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s,%.2f,%.2f,%.2f,%s\n",
			t.CreatedAt.Format("2006-01-02"),
			t.ID,
			t.Msisdn,
			t.Network,
			t.Provider,
			float64(t.Amount)/100,
			float64(t.CommissionAmt)/100,
			commRate,
			t.Status,
		))
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=commission_report_"+req.StartDate+"_to_"+req.EndDate+".csv")
	c.String(http.StatusOK, sb.String())
}
