package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jcarbajal/cv-checker/internal/domain"
)

type SQLiteCVRepository struct {
	db *sql.DB
}

func NewSQLiteCVRepository(db *sql.DB) *SQLiteCVRepository {
	return &SQLiteCVRepository{db: db}
}

func (r *SQLiteCVRepository) Create(ctx context.Context, cv *domain.CV) error {
	query := `INSERT INTO cvs (id, title, description, primary_language, favorite_version_id, created_at, updated_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?)`
	now := time.Now().UTC()
	if cv.CreatedAt.IsZero() {
		cv.CreatedAt = now
	}
	if cv.UpdatedAt.IsZero() {
		cv.UpdatedAt = now
	}

	_, err := r.db.ExecContext(ctx, query, cv.ID, cv.Title, cv.Description, string(cv.PrimaryLanguage), cv.FavoriteVersionID, cv.CreatedAt, cv.UpdatedAt)
	return err
}

func (r *SQLiteCVRepository) Update(ctx context.Context, cv *domain.CV) error {
	query := `UPDATE cvs SET title = ?, description = ?, primary_language = ?, favorite_version_id = ?, updated_at = ?
			  WHERE id = ?`
	cv.UpdatedAt = time.Now().UTC()
	res, err := r.db.ExecContext(ctx, query, cv.Title, cv.Description, string(cv.PrimaryLanguage), cv.FavoriteVersionID, cv.UpdatedAt, cv.ID)
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
	return nil
}

func (r *SQLiteCVRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM cvs WHERE id = ?`
	res, err := r.db.ExecContext(ctx, query, id)
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
	return nil
}

func (r *SQLiteCVRepository) GetByID(ctx context.Context, id string) (*domain.CV, error) {
	query := `SELECT id, title, description, primary_language, favorite_version_id, created_at, updated_at
			  FROM cvs WHERE id = ?`
	var cv domain.CV
	var favVer sql.NullString
	var lang string
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&cv.ID, &cv.Title, &cv.Description, &lang, &favVer, &cv.CreatedAt, &cv.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	cv.PrimaryLanguage = domain.Language(lang)
	if favVer.Valid {
		cv.FavoriteVersionID = favVer.String
	}

	versions, err := r.ListVersions(ctx, cv.ID)
	if err == nil {
		cv.Versions = versions
	}

	return &cv, nil
}

func (r *SQLiteCVRepository) List(ctx context.Context, search string) ([]domain.CV, error) {
	var query string
	var args []interface{}
	if search != "" {
		query = `SELECT id, title, description, primary_language, favorite_version_id, created_at, updated_at
				 FROM cvs WHERE title LIKE ? OR description LIKE ? ORDER BY updated_at DESC`
		pattern := "%" + search + "%"
		args = []interface{}{pattern, pattern}
	} else {
		query = `SELECT id, title, description, primary_language, favorite_version_id, created_at, updated_at
				 FROM cvs ORDER BY updated_at DESC`
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.CV
	for rows.Next() {
		var cv domain.CV
		var favVer sql.NullString
		var lang string
		if err := rows.Scan(&cv.ID, &cv.Title, &cv.Description, &lang, &favVer, &cv.CreatedAt, &cv.UpdatedAt); err != nil {
			return nil, err
		}
		cv.PrimaryLanguage = domain.Language(lang)
		if favVer.Valid {
			cv.FavoriteVersionID = favVer.String
		}
		result = append(result, cv)
	}
	return result, nil
}

func (r *SQLiteCVRepository) AddVersion(ctx context.Context, version *domain.CVVersion) error {
	query := `INSERT INTO cv_versions (id, cv_id, version_number, file_path, source_version_id, change_reason, ai_provider, ai_model, is_favorite, created_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	if version.CreatedAt.IsZero() {
		version.CreatedAt = time.Now().UTC()
	}

	_, err := r.db.ExecContext(ctx, query,
		version.ID, version.CVID, version.VersionNumber, version.FilePath,
		version.SourceVersionID, version.ChangeReason, version.AIProvider,
		version.AIModel, version.IsFavorite, version.CreatedAt,
	)
	return err
}

func (r *SQLiteCVRepository) GetVersion(ctx context.Context, versionID string) (*domain.CVVersion, error) {
	query := `SELECT id, cv_id, version_number, file_path, source_version_id, change_reason, ai_provider, ai_model, is_favorite, created_at
			  FROM cv_versions WHERE id = ?`
	var v domain.CVVersion
	var srcVer, aiProv, aiMod sql.NullString
	err := r.db.QueryRowContext(ctx, query, versionID).Scan(
		&v.ID, &v.CVID, &v.VersionNumber, &v.FilePath, &srcVer, &v.ChangeReason, &aiProv, &aiMod, &v.IsFavorite, &v.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if srcVer.Valid {
		v.SourceVersionID = srcVer.String
	}
	if aiProv.Valid {
		v.AIProvider = aiProv.String
	}
	if aiMod.Valid {
		v.AIModel = aiMod.String
	}
	return &v, nil
}

func (r *SQLiteCVRepository) ListVersions(ctx context.Context, cvID string) ([]domain.CVVersion, error) {
	query := `SELECT id, cv_id, version_number, file_path, source_version_id, change_reason, ai_provider, ai_model, is_favorite, created_at
			  FROM cv_versions WHERE cv_id = ? ORDER BY version_number ASC`
	rows, err := r.db.QueryContext(ctx, query, cvID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []domain.CVVersion
	for rows.Next() {
		var v domain.CVVersion
		var srcVer, aiProv, aiMod sql.NullString
		if err := rows.Scan(&v.ID, &v.CVID, &v.VersionNumber, &v.FilePath, &srcVer, &v.ChangeReason, &aiProv, &aiMod, &v.IsFavorite, &v.CreatedAt); err != nil {
			return nil, err
		}
		if srcVer.Valid {
			v.SourceVersionID = srcVer.String
		}
		if aiProv.Valid {
			v.AIProvider = aiProv.String
		}
		if aiMod.Valid {
			v.AIModel = aiMod.String
		}
		versions = append(versions, v)
	}
	return versions, nil
}

func (r *SQLiteCVRepository) SetFavoriteVersion(ctx context.Context, cvID string, versionID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Desmarcar todas las versiones de este CV
	if _, err := tx.ExecContext(ctx, `UPDATE cv_versions SET is_favorite = 0 WHERE cv_id = ?`, cvID); err != nil {
		return err
	}

	// Marcar únicamente la seleccionada (RF-004)
	res, err := tx.ExecContext(ctx, `UPDATE cv_versions SET is_favorite = 1 WHERE cv_id = ? AND id = ?`, cvID, versionID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("versión no encontrada para este CV")
	}

	// Actualizar en tabla cvs
	if _, err := tx.ExecContext(ctx, `UPDATE cvs SET favorite_version_id = ?, updated_at = ? WHERE id = ?`, versionID, time.Now().UTC(), cvID); err != nil {
		return err
	}

	return tx.Commit()
}
