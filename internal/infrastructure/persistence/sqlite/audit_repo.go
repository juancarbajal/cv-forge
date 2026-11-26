package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jcarbajal/cv-checker/internal/domain"
)

type SQLiteAuditRepository struct {
	db *sql.DB
}

func NewSQLiteAuditRepository(db *sql.DB) *SQLiteAuditRepository {
	return &SQLiteAuditRepository{db: db}
}

func (r *SQLiteAuditRepository) Log(ctx context.Context, entry *domain.AuditLog) error {
	if entry.ID == "" {
		entry.ID = uuid.NewString()
	}
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now().UTC()
	}

	query := `INSERT INTO audit_logs (id, timestamp, action, provider, model, result_status, details_json)
			  VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query,
		entry.ID, entry.Timestamp, string(entry.Action),
		entry.Provider, entry.Model, entry.ResultStatus, entry.DetailsJSON,
	)
	return err
}

func (r *SQLiteAuditRepository) List(ctx context.Context, limit int) ([]domain.AuditLog, error) {
	if limit <= 0 {
		limit = 100
	}
	query := `SELECT id, timestamp, action, provider, model, result_status, details_json
			  FROM audit_logs ORDER BY timestamp DESC LIMIT ?`
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []domain.AuditLog
	for rows.Next() {
		var l domain.AuditLog
		var act string
		var prov, mod, det sql.NullString
		if err := rows.Scan(&l.ID, &l.Timestamp, &act, &prov, &mod, &l.ResultStatus, &det); err != nil {
			return nil, err
		}
		l.Action = domain.AuditAction(act)
		if prov.Valid {
			l.Provider = prov.String
		}
		if mod.Valid {
			l.Model = mod.String
		}
		if det.Valid {
			l.DetailsJSON = det.String
		}
		logs = append(logs, l)
	}
	return logs, nil
}
