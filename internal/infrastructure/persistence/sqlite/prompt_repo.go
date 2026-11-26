package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jcarbajal/cv-checker/internal/domain"
)

type SQLitePromptRepository struct {
	db *sql.DB
}

func NewSQLitePromptRepository(db *sql.DB) *SQLitePromptRepository {
	return &SQLitePromptRepository{db: db}
}

func (r *SQLitePromptRepository) Create(ctx context.Context, prompt *domain.Prompt) error {
	if prompt.ID == "" {
		prompt.ID = uuid.NewString()
	}
	if prompt.CreatedAt.IsZero() {
		prompt.CreatedAt = time.Now().UTC()
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Si este nuevo prompt se marca activo, desmarcar los anteriores del mismo tipo
	if prompt.IsActive {
		if _, err := tx.ExecContext(ctx, `UPDATE prompts SET is_active = 0 WHERE prompt_type = ?`, string(prompt.Type)); err != nil {
			return err
		}
	}

	query := `INSERT INTO prompts (id, prompt_type, name, version, content, is_active, created_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?)`
	if _, err := tx.ExecContext(ctx, query, prompt.ID, string(prompt.Type), prompt.Name, prompt.Version, prompt.Content, prompt.IsActive, prompt.CreatedAt); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *SQLitePromptRepository) GetActive(ctx context.Context, pType domain.PromptType) (*domain.Prompt, error) {
	query := `SELECT id, prompt_type, name, version, content, is_active, created_at
			  FROM prompts WHERE prompt_type = ? AND is_active = 1 LIMIT 1`
	var p domain.Prompt
	var tStr string
	err := r.db.QueryRowContext(ctx, query, string(pType)).Scan(
		&p.ID, &tStr, &p.Name, &p.Version, &p.Content, &p.IsActive, &p.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNoActivePrompt
	}
	if err != nil {
		return nil, err
	}
	p.Type = domain.PromptType(tStr)
	return &p, nil
}

func (r *SQLitePromptRepository) GetByVersion(ctx context.Context, pType domain.PromptType, version int) (*domain.Prompt, error) {
	query := `SELECT id, prompt_type, name, version, content, is_active, created_at
			  FROM prompts WHERE prompt_type = ? AND version = ? LIMIT 1`
	var p domain.Prompt
	var tStr string
	err := r.db.QueryRowContext(ctx, query, string(pType), version).Scan(
		&p.ID, &tStr, &p.Name, &p.Version, &p.Content, &p.IsActive, &p.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	p.Type = domain.PromptType(tStr)
	return &p, nil
}

func (r *SQLitePromptRepository) ListByType(ctx context.Context, pType domain.PromptType) ([]domain.Prompt, error) {
	query := `SELECT id, prompt_type, name, version, content, is_active, created_at
			  FROM prompts WHERE prompt_type = ? ORDER BY version DESC`
	rows, err := r.db.QueryContext(ctx, query, string(pType))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.Prompt
	for rows.Next() {
		var p domain.Prompt
		var tStr string
		if err := rows.Scan(&p.ID, &tStr, &p.Name, &p.Version, &p.Content, &p.IsActive, &p.CreatedAt); err != nil {
			return nil, err
		}
		p.Type = domain.PromptType(tStr)
		result = append(result, p)
	}
	return result, nil
}

func (r *SQLitePromptRepository) SetActive(ctx context.Context, pType domain.PromptType, version int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `UPDATE prompts SET is_active = 0 WHERE prompt_type = ?`, string(pType)); err != nil {
		return err
	}

	res, err := tx.ExecContext(ctx, `UPDATE prompts SET is_active = 1 WHERE prompt_type = ? AND version = ?`, string(pType), version)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("versión de prompt no encontrada")
	}

	return tx.Commit()
}

func (r *SQLitePromptRepository) SeedDefaults(ctx context.Context, defaultPrompts []domain.Prompt) error {
	for _, p := range defaultPrompts {
		var count int
		err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM prompts WHERE prompt_type = ?`, string(p.Type)).Scan(&count)
		if err != nil {
			return err
		}
		if count == 0 {
			p.IsActive = true
			if err := r.Create(ctx, &p); err != nil {
				return err
			}
		}
	}
	return nil
}
