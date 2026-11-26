package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jcarbajal/cv-checker/internal/domain"
)

type SQLiteConfigRepository struct {
	db *sql.DB
}

func NewSQLiteConfigRepository(db *sql.DB) *SQLiteConfigRepository {
	return &SQLiteConfigRepository{db: db}
}

func (r *SQLiteConfigRepository) SaveConfig(ctx context.Context, config *domain.ProviderConfig) error {
	config.UpdatedAt = time.Now().UTC()
	query := `INSERT INTO provider_configs (provider_name, encrypted_api_key, default_model, base_url, is_active, updated_at)
			  VALUES (?, ?, ?, ?, ?, ?)
			  ON CONFLICT(provider_name) DO UPDATE SET
			    encrypted_api_key = excluded.encrypted_api_key,
			    default_model = excluded.default_model,
			    base_url = excluded.base_url,
			    is_active = excluded.is_active,
			    updated_at = excluded.updated_at`

	_, err := r.db.ExecContext(ctx, query,
		config.ProviderName, config.EncryptedAPIKey, config.DefaultModel,
		config.BaseURL, config.IsActive, config.UpdatedAt,
	)
	return err
}

func (r *SQLiteConfigRepository) GetConfig(ctx context.Context, providerName string) (*domain.ProviderConfig, error) {
	query := `SELECT provider_name, encrypted_api_key, default_model, base_url, is_active, updated_at
			  FROM provider_configs WHERE provider_name = ?`
	var c domain.ProviderConfig
	var url sql.NullString
	err := r.db.QueryRowContext(ctx, query, providerName).Scan(
		&c.ProviderName, &c.EncryptedAPIKey, &c.DefaultModel, &url, &c.IsActive, &c.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if url.Valid {
		c.BaseURL = url.String
	}
	return &c, nil
}

func (r *SQLiteConfigRepository) ListConfigs(ctx context.Context) ([]domain.ProviderConfig, error) {
	query := `SELECT provider_name, encrypted_api_key, default_model, base_url, is_active, updated_at
			  FROM provider_configs ORDER BY provider_name ASC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.ProviderConfig
	for rows.Next() {
		var c domain.ProviderConfig
		var url sql.NullString
		if err := rows.Scan(&c.ProviderName, &c.EncryptedAPIKey, &c.DefaultModel, &url, &c.IsActive, &c.UpdatedAt); err != nil {
			return nil, err
		}
		if url.Valid {
			c.BaseURL = url.String
		}
		list = append(list, c)
	}
	return list, nil
}

func (r *SQLiteConfigRepository) SetActiveProvider(ctx context.Context, providerName string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `UPDATE provider_configs SET is_active = 0`); err != nil {
		return err
	}

	res, err := tx.ExecContext(ctx, `UPDATE provider_configs SET is_active = 1, updated_at = ? WHERE provider_name = ?`, time.Now().UTC(), providerName)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrNotFound
	}

	return tx.Commit()
}

func (r *SQLiteConfigRepository) GetActiveConfig(ctx context.Context) (*domain.ProviderConfig, error) {
	query := `SELECT provider_name, encrypted_api_key, default_model, base_url, is_active, updated_at
			  FROM provider_configs WHERE is_active = 1 LIMIT 1`
	var c domain.ProviderConfig
	var url sql.NullString
	err := r.db.QueryRowContext(ctx, query).Scan(
		&c.ProviderName, &c.EncryptedAPIKey, &c.DefaultModel, &url, &c.IsActive, &c.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrProviderNotActive
	}
	if err != nil {
		return nil, err
	}
	if url.Valid {
		c.BaseURL = url.String
	}
	return &c, nil
}
