package db

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/google/uuid"
	"rev_core/internal/models"
)

type DB struct {
	Conn          *sql.DB
	encryptionKey []byte
}

func New(databaseURL string) (*DB, error) {
	conn, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(); err != nil {
		return nil, err
	}
	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(10)
	conn.SetConnMaxLifetime(5 * time.Minute)

	key := os.Getenv("ENCRYPTION_KEY")
	if key == "" {
		key = "changeme_32_byte_encryption_key!!"
	}
	derivedKey := sha256.Sum256([]byte(key))
	db := &DB{Conn: conn, encryptionKey: derivedKey[:]}
	return db, nil
}

func (d *DB) encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(d.encryptionKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (d *DB) decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(d.encryptionKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	nonce, cipherdata := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, cipherdata, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func (d *DB) GetActiveProviders() ([]models.Provider, error) {
	rows, err := d.Conn.Query(`SELECT id, name, base_url, api_token, http_proxy, socks5_proxy, enable_proxy, max_concurrent_requests, is_active, created_at, updated_at FROM providers WHERE is_active = true`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var providers []models.Provider
	for rows.Next() {
		var p models.Provider
		if err := rows.Scan(&p.ID, &p.Name, &p.BaseURL, &p.APIToken, &p.HTTPProxy, &p.Socks5Proxy, &p.EnableProxy, &p.MaxConcurrentRequests, &p.IsActive, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		decrypted, err := d.decrypt(p.APIToken)
		if err == nil {
			p.APIToken = decrypted
		}
		providers = append(providers, p)
	}
	return providers, rows.Err()
}

func (d *DB) GetAllModels() ([]models.Model, error) {
	rows, err := d.Conn.Query(`SELECT id, provider_id, model_id, display_model_id, max_concurrent_requests, queue_size, is_active, created_at FROM models`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var modelsList []models.Model
	for rows.Next() {
		var m models.Model
		if err := rows.Scan(&m.ID, &m.ProviderID, &m.ModelID, &m.DisplayModelID, &m.MaxConcurrentRequests, &m.QueueSize, &m.IsActive, &m.CreatedAt); err != nil {
			return nil, err
		}
		modelsList = append(modelsList, m)
	}
	return modelsList, rows.Err()
}

func (d *DB) GetActiveModels() ([]models.Model, error) {
	rows, err := d.Conn.Query(`SELECT id, provider_id, model_id, display_model_id, max_concurrent_requests, queue_size, is_active, created_at FROM models WHERE is_active = true`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var modelsList []models.Model
	for rows.Next() {
		var m models.Model
		if err := rows.Scan(&m.ID, &m.ProviderID, &m.ModelID, &m.DisplayModelID, &m.MaxConcurrentRequests, &m.QueueSize, &m.IsActive, &m.CreatedAt); err != nil {
			return nil, err
		}
		modelsList = append(modelsList, m)
	}
	return modelsList, rows.Err()
}

func (d *DB) GetTokenByHash(keyHash string) (*models.Token, error) {
	var t models.Token
	row := d.Conn.QueryRow(`SELECT id, name, key_hash, max_input_tokens, max_output_tokens, requests_per_minute, is_active, created_at FROM tokens WHERE key_hash = $1 AND is_active = true`, keyHash)
	err := row.Scan(&t.ID, &t.Name, &t.KeyHash, &t.MaxInputTokens, &t.MaxOutputTokens, &t.RequestsPerMinute, &t.IsActive, &t.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (d *DB) GetTokenPermissions(tokenID uuid.UUID) ([]models.TokenModelPermission, error) {
	rows, err := d.Conn.Query(`SELECT id, token_id, model_id, max_input_tokens, max_output_tokens, created_at FROM token_model_permissions WHERE token_id = $1`, tokenID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []models.TokenModelPermission
	for rows.Next() {
		var p models.TokenModelPermission
		if err := rows.Scan(&p.ID, &p.TokenID, &p.ModelID, &p.MaxInputTokens, &p.MaxOutputTokens, &p.CreatedAt); err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}
	return perms, rows.Err()
}

func (d *DB) GetConfig(key string) (string, error) {
	var value string
	err := d.Conn.QueryRow(`SELECT value FROM config WHERE key = $1`, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return value, nil
}

func (d *DB) InsertUsageLogBatch(logs []models.UsageLog) error {
	if len(logs) == 0 {
		return nil
	}
	tx, err := d.Conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO usage_logs (id, token_id, provider_id, model_id, model_name, provider_name, request_path, input_tokens, output_tokens, total_tokens, latency_ms, status_code, is_successful, error_message, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, logEntry := range logs {
		_, err := stmt.Exec(
			logEntry.ID,
			logEntry.TokenID,
			logEntry.ProviderID,
			logEntry.ModelID,
			logEntry.ModelName,
			logEntry.ProviderName,
			logEntry.RequestPath,
			logEntry.InputTokens,
			logEntry.OutputTokens,
			logEntry.TotalTokens,
			logEntry.LatencyMs,
			logEntry.StatusCode,
			logEntry.IsSuccessful,
			logEntry.ErrorMessage,
			logEntry.CreatedAt,
		)
		if err != nil {
			log.Printf("Failed to insert usage log: %v", err)
		}
	}
	return tx.Commit()
}
