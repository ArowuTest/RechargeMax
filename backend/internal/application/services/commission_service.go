package services

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"gorm.io/gorm"
)

// ─────────────────────────────────────────────────────────────────────────────
// Request / response DTOs
// ─────────────────────────────────────────────────────────────────────────────

// CommissionFilter constrains which transactions are included in a report.
type CommissionFilter struct {
	StartDate string // YYYY-MM-DD
	EndDate   string // YYYY-MM-DD
	Network   string // optional: filter by network name
	Provider  string // optional: filter by provider name
}

// CommissionReport is the full reconciliation payload.
type CommissionReport struct {
	Summary            CommissionSummary            `json:"summary"`
	ByNetwork          []CommissionByNetwork         `json:"by_network"`
	ByProvider         []CommissionByProvider        `json:"by_provider"`
	ByDate             []CommissionByDate             `json:"by_date"`
	RecentTransactions []CommissionTransaction        `json:"recent_transactions"`
}

// CommissionSummary aggregates across the entire filtered period.
type CommissionSummary struct {
	TotalTransactions   int64   `json:"total_transactions"`
	TotalRechargeAmount int64   `json:"total_recharge_amount"`
	TotalCommission     int64   `json:"total_commission"`
	AverageCommission   float64 `json:"average_commission"`
	CommissionRate      float64 `json:"commission_rate"`
}

// CommissionByNetwork breaks down commission per network operator.
type CommissionByNetwork struct {
	Network           string  `json:"network"`
	TransactionCount  int64   `json:"transaction_count"`
	TotalAmount       int64   `json:"total_amount"`
	TotalCommission   int64   `json:"total_commission"`
	AverageCommission float64 `json:"average_commission"`
	CommissionRate    float64 `json:"commission_rate"`
}

// CommissionByProvider breaks down commission per recharge provider.
type CommissionByProvider struct {
	Provider          string  `json:"provider"`
	TransactionCount  int64   `json:"transaction_count"`
	TotalAmount       int64   `json:"total_amount"`
	TotalCommission   int64   `json:"total_commission"`
	AverageCommission float64 `json:"average_commission"`
	CommissionRate    float64 `json:"commission_rate"`
}

// CommissionByDate aggregates per calendar day.
type CommissionByDate struct {
	Date             string `json:"date"`
	TransactionCount int64  `json:"transaction_count"`
	TotalAmount      int64  `json:"total_amount"`
	TotalCommission  int64  `json:"total_commission"`
}

// CommissionTransaction is a single masked transaction line item.
type CommissionTransaction struct {
	ID             string    `json:"id"`
	MSISDN         string    `json:"msisdn"` // masked: 0801****234
	Network        string    `json:"network"`
	Provider       string    `json:"provider"`
	Amount         int64     `json:"amount"`
	Commission     int64     `json:"commission"`
	CommissionRate float64   `json:"commission_rate"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

// ─────────────────────────────────────────────────────────────────────────────
// CommissionService
// ─────────────────────────────────────────────────────────────────────────────

// CommissionService runs commission-reconciliation queries.
type CommissionService struct {
	db *gorm.DB
}

// NewCommissionService constructs a CommissionService.
func NewCommissionService(db *gorm.DB) *CommissionService {
	return &CommissionService{db: db}
}

// GetReconciliation builds the full commission report for the given filter.
func (s *CommissionService) GetReconciliation(ctx context.Context, f CommissionFilter) (*CommissionReport, error) {
	start, end, err := parseDateRange(f.StartDate, f.EndDate)
	if err != nil {
		return nil, err
	}
	db := s.db.WithContext(ctx)
	report := &CommissionReport{}

	// ── Summary ───────────────────────────────────────────────────────────────
	type summaryRow struct {
		TotalTransactions   int64
		TotalRechargeAmount int64
		TotalCommission     int64
	}
	var sum summaryRow
	q := db.Table("transactions").
		Select("COUNT(*) AS total_transactions, COALESCE(SUM(amount),0) AS total_recharge_amount, COALESCE(SUM(commission_amount),0) AS total_commission").
		Where("created_at BETWEEN ? AND ? AND status = 'SUCCESS'", start, end)
	if f.Network != "" {
		q = q.Where("network = ?", strings.ToUpper(f.Network))
	}
	if f.Provider != "" {
		q = q.Where("provider = ?", f.Provider)
	}
	q.Scan(&sum)

	report.Summary = CommissionSummary{
		TotalTransactions:   sum.TotalTransactions,
		TotalRechargeAmount: sum.TotalRechargeAmount,
		TotalCommission:     sum.TotalCommission,
	}
	if sum.TotalTransactions > 0 {
		report.Summary.AverageCommission = float64(sum.TotalCommission) / float64(sum.TotalTransactions)
	}
	if sum.TotalRechargeAmount > 0 {
		report.Summary.CommissionRate = float64(sum.TotalCommission) / float64(sum.TotalRechargeAmount) * 100
	}

	// ── By Network ────────────────────────────────────────────────────────────
	type netRow struct {
		Network          string
		TransactionCount int64
		TotalAmount      int64
		TotalCommission  int64
	}
	var netRows []netRow
	db.Table("transactions").
		Select("network, COUNT(*) AS transaction_count, COALESCE(SUM(amount),0) AS total_amount, COALESCE(SUM(commission_amount),0) AS total_commission").
		Where("created_at BETWEEN ? AND ? AND status = 'SUCCESS'", start, end).
		Group("network").
		Scan(&netRows)
	for _, r := range netRows {
		avg, rate := commRates(r.TotalCommission, r.TransactionCount, r.TotalAmount)
		report.ByNetwork = append(report.ByNetwork, CommissionByNetwork{
			Network: r.Network, TransactionCount: r.TransactionCount,
			TotalAmount: r.TotalAmount, TotalCommission: r.TotalCommission,
			AverageCommission: avg, CommissionRate: rate,
		})
	}

	// ── By Provider ───────────────────────────────────────────────────────────
	type provRow struct {
		Provider         string
		TransactionCount int64
		TotalAmount      int64
		TotalCommission  int64
	}
	var provRows []provRow
	db.Table("transactions").
		Select("provider, COUNT(*) AS transaction_count, COALESCE(SUM(amount),0) AS total_amount, COALESCE(SUM(commission_amount),0) AS total_commission").
		Where("created_at BETWEEN ? AND ? AND status = 'SUCCESS'", start, end).
		Group("provider").
		Scan(&provRows)
	for _, r := range provRows {
		avg, rate := commRates(r.TotalCommission, r.TransactionCount, r.TotalAmount)
		report.ByProvider = append(report.ByProvider, CommissionByProvider{
			Provider: r.Provider, TransactionCount: r.TransactionCount,
			TotalAmount: r.TotalAmount, TotalCommission: r.TotalCommission,
			AverageCommission: avg, CommissionRate: rate,
		})
	}

	// ── By Date ───────────────────────────────────────────────────────────────
	type dateRow struct {
		Day              string
		TransactionCount int64
		TotalAmount      int64
		TotalCommission  int64
	}
	var dateRows []dateRow
	db.Table("transactions").
		Select("DATE(created_at) AS day, COUNT(*) AS transaction_count, COALESCE(SUM(amount),0) AS total_amount, COALESCE(SUM(commission_amount),0) AS total_commission").
		Where("created_at BETWEEN ? AND ? AND status = 'SUCCESS'", start, end).
		Group("DATE(created_at)").
		Order("day ASC").
		Scan(&dateRows)
	for _, r := range dateRows {
		report.ByDate = append(report.ByDate, CommissionByDate{
			Date: r.Day, TransactionCount: r.TransactionCount,
			TotalAmount: r.TotalAmount, TotalCommission: r.TotalCommission,
		})
	}

	// ── Recent Transactions ───────────────────────────────────────────────────
	type txnRow struct {
		ID            string
		MSISDN        string
		Network       string
		Provider      string
		Amount        int64
		CommissionAmt int64
		Status        string
		CreatedAt     time.Time
	}
	var txns []txnRow
	db.Table("transactions").
		Select("id, msisdn, network, provider, amount, commission_amount AS commission_amt, status, created_at").
		Where("created_at BETWEEN ? AND ? AND status = 'SUCCESS'", start, end).
		Order("created_at DESC").
		Limit(20).
		Scan(&txns)
	for _, t := range txns {
		_, rate := commRates(t.CommissionAmt, 1, t.Amount)
		msisdn := t.MSISDN
		if len(msisdn) > 7 {
			msisdn = msisdn[:4] + "****" + msisdn[len(msisdn)-3:]
		}
		report.RecentTransactions = append(report.RecentTransactions, CommissionTransaction{
			ID: t.ID, MSISDN: msisdn, Network: t.Network, Provider: t.Provider,
			Amount: t.Amount, Commission: t.CommissionAmt, CommissionRate: rate,
			Status: t.Status, CreatedAt: t.CreatedAt,
		})
	}

	return report, nil
}

// ExportCSV returns a CSV byte slice for the given filter.
func (s *CommissionService) ExportCSV(ctx context.Context, f CommissionFilter) ([]byte, error) {
	start, end, err := parseDateRange(f.StartDate, f.EndDate)
	if err != nil {
		return nil, err
	}

	type txnRow struct {
		CreatedAt     time.Time
		ID            string
		MSISDN        string
		Network       string
		Provider      string
		Amount        int64
		CommissionAmt int64
		Status        string
	}
	var txns []txnRow
	s.db.WithContext(ctx).Table("transactions").
		Select("created_at, id, msisdn, network, provider, amount, commission_amount AS commission_amt, status").
		Where("created_at BETWEEN ? AND ? AND status = 'SUCCESS'", start, end).
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
			t.ID, t.MSISDN, t.Network, t.Provider,
			float64(t.Amount)/100,
			float64(t.CommissionAmt)/100,
			commRate, t.Status,
		))
	}
	return []byte(sb.String()), nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func parseDateRange(startStr, endStr string) (time.Time, time.Time, error) {
	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start_date: use YYYY-MM-DD")
	}
	end, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid end_date: use YYYY-MM-DD")
	}
	end = end.Add(24*time.Hour - time.Second)
	return start, end, nil
}

func commRates(commission, txCount, amount int64) (avg, rate float64) {
	if txCount > 0 {
		avg = math.Round(float64(commission)/float64(txCount)*100) / 100
	}
	if amount > 0 {
		rate = math.Round(float64(commission)/float64(amount)*10000) / 100
	}
	return
}
