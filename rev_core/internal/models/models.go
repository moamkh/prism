package models

import (
	"database/sql"

	"github.com/google/uuid"
)

type Provider struct {
	ID                    uuid.UUID `db:"id"`
	Name                  string    `db:"name"`
	BaseURL               string    `db:"base_url"`
	APIToken              string    `db:"api_token"`
	HTTPProxy             *string   `db:"http_proxy"`
	Socks5Proxy           *string   `db:"socks5_proxy"`
	EnableProxy           bool      `db:"enable_proxy"`
	MaxConcurrentRequests int       `db:"max_concurrent_requests"`
	IsActive              bool      `db:"is_active"`
	CreatedAt             int64     `db:"created_at"`
	UpdatedAt             int64     `db:"updated_at"`
}

type Model struct {
	ID             uuid.UUID `db:"id"`
	ProviderID     uuid.UUID `db:"provider_id"`
	ModelID        string    `db:"model_id"`
	DisplayModelID *string   `db:"display_model_id"`
	IsActive       bool      `db:"is_active"`
	CreatedAt      int64     `db:"created_at"`
}

type Token struct {
	ID                uuid.UUID      `db:"id"`
	Name              string         `db:"name"`
	KeyHash           string         `db:"key_hash"`
	MaxInputTokens    sql.NullInt32  `db:"max_input_tokens"`
	MaxOutputTokens   sql.NullInt32  `db:"max_output_tokens"`
	RequestsPerMinute sql.NullInt32  `db:"requests_per_minute"`
	IsActive          bool           `db:"is_active"`
	CreatedAt         int64          `db:"created_at"`
}

type TokenModelPermission struct {
	ID             uuid.UUID     `db:"id"`
	TokenID        uuid.UUID     `db:"token_id"`
	ModelID        uuid.UUID     `db:"model_id"`
	MaxInputTokens sql.NullInt32 `db:"max_input_tokens"`
	MaxOutputTokens sql.NullInt32 `db:"max_output_tokens"`
	CreatedAt      int64         `db:"created_at"`
}

type UsageLog struct {
	ID           uuid.UUID      `db:"id"`
	TokenID      uuid.NullUUID  `db:"token_id"`
	ProviderID   uuid.NullUUID  `db:"provider_id"`
	ModelID      uuid.NullUUID  `db:"model_id"`
	ModelName    sql.NullString `db:"model_name"`
	ProviderName sql.NullString `db:"provider_name"`
	RequestPath  string         `db:"request_path"`
	InputTokens  int            `db:"input_tokens"`
	OutputTokens int            `db:"output_tokens"`
	TotalTokens  int            `db:"total_tokens"`
	LatencyMs    sql.NullInt32  `db:"latency_ms"`
	StatusCode   sql.NullInt32  `db:"status_code"`
	CreatedAt    int64          `db:"created_at"`
}

type ConfigItem struct {
	Key       string `db:"key"`
	Value     string `db:"value"`
	UpdatedAt int64  `db:"updated_at"`
}
