package config

// TelecomConfig holds all telecom network API configurations

// MTNConfig holds MTN API configuration
type MTNConfig struct {
	APIKey      string `json:"api_key"`
	SecretKey   string `json:"secret_key"`
	BaseURL     string `json:"base_url"`
	AirtimeURL  string `json:"airtime_url"`
	DataURL     string `json:"data_url"`
	CallbackURL string `json:"callback_url"`
	Environment string `json:"environment"` // sandbox, production
}

// AirtelConfig holds Airtel API configuration
type AirtelConfig struct {
	APIKey      string `json:"api_key"`
	SecretKey   string `json:"secret_key"`
	BaseURL     string `json:"base_url"`
	AirtimeURL  string `json:"airtime_url"`
	DataURL     string `json:"data_url"`
	CallbackURL string `json:"callback_url"`
	Environment string `json:"environment"`
}

// GloConfig holds Glo API configuration
type GloConfig struct {
	APIKey      string `json:"api_key"`
	SecretKey   string `json:"secret_key"`
	BaseURL     string `json:"base_url"`
	AirtimeURL  string `json:"airtime_url"`
	DataURL     string `json:"data_url"`
	CallbackURL string `json:"callback_url"`
	Environment string `json:"environment"`
}

// NineMobileConfig holds 9mobile API configuration
type NineMobileConfig struct {
	APIKey      string `json:"api_key"`
	SecretKey   string `json:"secret_key"`
	BaseURL     string `json:"base_url"`
	AirtimeURL  string `json:"airtime_url"`
	DataURL     string `json:"data_url"`
	CallbackURL string `json:"callback_url"`
	Environment string `json:"environment"`
}

// LoadTelecomConfig loads telecom configuration from environment variables
func LoadTelecomConfig() TelecomConfig {
	return TelecomConfig{
		MTN: MTNConfig{
			APIKey:      getEnv("MTN_API_KEY", ""),
			SecretKey:   getEnv("MTN_SECRET_KEY", ""),
			BaseURL:     getEnv("MTN_BASE_URL", "https://api.mtn.ng"),
			AirtimeURL:  getEnv("MTN_AIRTIME_URL", "https://api.mtn.ng/v1/airtime"),
			DataURL:     getEnv("MTN_DATA_URL", "https://api.mtn.ng/v1/data"),
			CallbackURL: getEnv("MTN_CALLBACK_URL", "https://rechargemax.com/api/v1/recharge/webhook/mtn"),
			Environment: getEnv("MTN_ENVIRONMENT", "sandbox"),
		},
		Airtel: AirtelConfig{
			APIKey:      getEnv("AIRTEL_API_KEY", ""),
			SecretKey:   getEnv("AIRTEL_SECRET_KEY", ""),
			BaseURL:     getEnv("AIRTEL_BASE_URL", "https://api.airtel.ng"),
			AirtimeURL:  getEnv("AIRTEL_AIRTIME_URL", "https://api.airtel.ng/merchant/v1/airtime"),
			DataURL:     getEnv("AIRTEL_DATA_URL", "https://api.airtel.ng/merchant/v1/data"),
			CallbackURL: getEnv("AIRTEL_CALLBACK_URL", "https://rechargemax.com/api/v1/recharge/webhook/airtel"),
			Environment: getEnv("AIRTEL_ENVIRONMENT", "sandbox"),
		},
		Glo: GloConfig{
			APIKey:      getEnv("GLO_API_KEY", ""),
			SecretKey:   getEnv("GLO_SECRET_KEY", ""),
			BaseURL:     getEnv("GLO_BASE_URL", "https://api.gloworld.com"),
			AirtimeURL:  getEnv("GLO_AIRTIME_URL", "https://api.gloworld.com/airtime"),
			DataURL:     getEnv("GLO_DATA_URL", "https://api.gloworld.com/data"),
			CallbackURL: getEnv("GLO_CALLBACK_URL", "https://rechargemax.com/api/v1/recharge/webhook/glo"),
			Environment: getEnv("GLO_ENVIRONMENT", "sandbox"),
		},
		NineMobile: NineMobileConfig{
			APIKey:      getEnv("9MOBILE_API_KEY", ""),
			SecretKey:   getEnv("9MOBILE_SECRET_KEY", ""),
			BaseURL:     getEnv("9MOBILE_BASE_URL", "https://api.9mobile.com.ng"),
			AirtimeURL:  getEnv("9MOBILE_AIRTIME_URL", "https://api.9mobile.com.ng/airtime"),
			DataURL:     getEnv("9MOBILE_DATA_URL", "https://api.9mobile.com.ng/data"),
			CallbackURL: getEnv("9MOBILE_CALLBACK_URL", "https://rechargemax.com/api/v1/recharge/webhook/9mobile"),
			Environment: getEnv("9MOBILE_ENVIRONMENT", "sandbox"),
		},
	}
}
