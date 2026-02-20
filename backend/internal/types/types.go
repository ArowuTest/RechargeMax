package types

import "time"

// RechargeResponse represents the response from network APIs
type RechargeResponse struct {
	Success     bool   `json:"success"`
	Reference   string `json:"reference"`
	NetworkRef  string `json:"network_ref"`
	Status      string `json:"status"`
	Message     string `json:"message"`
	Balance     string `json:"balance,omitempty"`
	ValidUntil  string `json:"valid_until,omitempty"`
}

// NetworkProvider represents a telecom network provider
type NetworkProvider struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Logo        string `json:"logo"`
	IsActive    bool   `json:"is_active"`
	AirtimeAPI  string `json:"airtime_api"`
	DataAPI     string `json:"data_api"`
	APIKey      string `json:"api_key"`
	SecretKey   string `json:"secret_key"`
}

// DataPackage represents available data packages
type DataPackage struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Size        string `json:"size"`
	Validity    string `json:"validity"`
	Price       int64  `json:"price"` // Price in kobo
	Network     string `json:"network"`
	IsActive    bool   `json:"is_active"`
}

// RechargeRequest represents a recharge request
type RechargeRequest struct {
	MSISDN       string `json:"msisdn"`
	Amount       int64  `json:"amount"`
	Network      string `json:"network"`
	Type         string `json:"type"` // airtime, data
	DataPackage  string `json:"data_package,omitempty"`
	Reference    string `json:"reference"`
	CallbackURL  string `json:"callback_url"`
}

// WebhookPayload represents incoming webhook data
type WebhookPayload struct {
	Event     string      `json:"event"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
	Signature string      `json:"signature"`
}
