package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jcarbajal/cv-checker/internal/domain"
)

type SQLiteJobRepository struct {
	db *sql.DB
}

func NewSQLiteJobRepository(db *sql.DB) *SQLiteJobRepository {
	return &SQLiteJobRepository{db: db}
}

func (r *SQLiteJobRepository) Create(ctx context.Context, job *domain.JobApplication) error {
	query := `INSERT INTO job_applications (id, company, role, url, job_posting_raw, detected_language, target_language, cv_id, cv_version_id, status, total_ats_score, created_at, updated_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	now := time.Now().UTC()
	if job.CreatedAt.IsZero() {
		job.CreatedAt = now
	}
	if job.UpdatedAt.IsZero() {
		job.UpdatedAt = now
	}

	_, err := r.db.ExecContext(ctx, query,
		job.ID, job.Company, job.Role, job.URL, job.JobPostingRaw,
		string(job.DetectedLanguage), string(job.TargetLanguage),
		job.CVID, job.CVVersionID, string(job.Status), job.TotalATSScore,
		job.CreatedAt, job.UpdatedAt,
	)
	return err
}

func (r *SQLiteJobRepository) Update(ctx context.Context, job *domain.JobApplication) error {
	query := `UPDATE job_applications 
			  SET company = ?, role = ?, url = ?, job_posting_raw = ?, detected_language = ?, target_language = ?, cv_id = ?, cv_version_id = ?, status = ?, total_ats_score = ?, updated_at = ?
			  WHERE id = ?`
	job.UpdatedAt = time.Now().UTC()
	res, err := r.db.ExecContext(ctx, query,
		job.Company, job.Role, job.URL, job.JobPostingRaw,
		string(job.DetectedLanguage), string(job.TargetLanguage),
		job.CVID, job.CVVersionID, string(job.Status), job.TotalATSScore,
		job.UpdatedAt, job.ID,
	)
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

func (r *SQLiteJobRepository) GetByID(ctx context.Context, id string) (*domain.JobApplication, error) {
	query := `SELECT id, company, role, url, job_posting_raw, detected_language, target_language, cv_id, cv_version_id, status, total_ats_score, created_at, updated_at
			  FROM job_applications WHERE id = ?`
	var j domain.JobApplication
	var url sql.NullString
	var detLang, tgtLang, status string
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&j.ID, &j.Company, &j.Role, &url, &j.JobPostingRaw,
		&detLang, &tgtLang, &j.CVID, &j.CVVersionID,
		&status, &j.TotalATSScore, &j.CreatedAt, &j.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if url.Valid {
		j.URL = url.String
	}
	j.DetectedLanguage = domain.Language(detLang)
	j.TargetLanguage = domain.Language(tgtLang)
	j.Status = domain.JobStatus(status)

	reqs, err := r.GetRequirements(ctx, j.ID)
	if err == nil {
		j.Requirements = reqs
	}

	return &j, nil
}

func (r *SQLiteJobRepository) List(ctx context.Context) ([]domain.JobApplication, error) {
	query := `SELECT id, company, role, url, job_posting_raw, detected_language, target_language, cv_id, cv_version_id, status, total_ats_score, created_at, updated_at
			  FROM job_applications ORDER BY updated_at DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.JobApplication
	for rows.Next() {
		var j domain.JobApplication
		var url sql.NullString
		var detLang, tgtLang, status string
		if err := rows.Scan(
			&j.ID, &j.Company, &j.Role, &url, &j.JobPostingRaw,
			&detLang, &tgtLang, &j.CVID, &j.CVVersionID,
			&status, &j.TotalATSScore, &j.CreatedAt, &j.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if url.Valid {
			j.URL = url.String
		}
		j.DetectedLanguage = domain.Language(detLang)
		j.TargetLanguage = domain.Language(tgtLang)
		j.Status = domain.JobStatus(status)
		result = append(result, j)
	}
	return result, nil
}

func (r *SQLiteJobRepository) SaveRequirements(ctx context.Context, jobAppID string, reqs []domain.Requirement) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Eliminar existentes
	if _, err := tx.ExecContext(ctx, `DELETE FROM requirements WHERE job_application_id = ?`, jobAppID); err != nil {
		return err
	}

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO requirements (id, job_application_id, name, category, priority) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, req := range reqs {
		id := req.ID
		if id == "" {
			id = uuid.NewString()
		}
		if _, err := stmt.ExecContext(ctx, id, jobAppID, req.Name, string(req.Category), string(req.Priority)); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *SQLiteJobRepository) GetRequirements(ctx context.Context, jobAppID string) ([]domain.Requirement, error) {
	query := `SELECT id, job_application_id, name, category, priority FROM requirements WHERE job_application_id = ?`
	rows, err := r.db.QueryContext(ctx, query, jobAppID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reqs []domain.Requirement
	for rows.Next() {
		var req domain.Requirement
		var cat, prio string
		if err := rows.Scan(&req.ID, &req.JobApplicationID, &req.Name, &cat, &prio); err != nil {
			return nil, err
		}
		req.Category = domain.RequirementCategory(cat)
		req.Priority = domain.RequirementPriority(prio)
		reqs = append(reqs, req)
	}
	return reqs, nil
}

func (r *SQLiteJobRepository) SaveMatchResults(ctx context.Context, jobAppID string, matches []domain.MatchResult) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM match_results WHERE job_application_id = ?`, jobAppID); err != nil {
		return err
	}

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO match_results (id, job_application_id, requirement_id, requirement_name, status, confidence, evidence, explanation) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, m := range matches {
		id := m.ID
		if id == "" {
			id = uuid.NewString()
		}
		if _, err := stmt.ExecContext(ctx, id, jobAppID, m.RequirementID, m.RequirementName, string(m.Status), m.Confidence, m.Evidence, m.Explanation); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *SQLiteJobRepository) GetMatchResults(ctx context.Context, jobAppID string) ([]domain.MatchResult, error) {
	query := `SELECT id, job_application_id, requirement_id, requirement_name, status, confidence, evidence, explanation FROM match_results WHERE job_application_id = ?`
	rows, err := r.db.QueryContext(ctx, query, jobAppID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matches []domain.MatchResult
	for rows.Next() {
		var m domain.MatchResult
		var status string
		var ev, exp sql.NullString
		if err := rows.Scan(&m.ID, &m.JobApplicationID, &m.RequirementID, &m.RequirementName, &status, &m.Confidence, &ev, &exp); err != nil {
			return nil, err
		}
		m.Status = domain.MatchStatus(status)
		if ev.Valid {
			m.Evidence = ev.String
		}
		if exp.Valid {
			m.Explanation = exp.String
		}
		matches = append(matches, m)
	}
	return matches, nil
}

func (r *SQLiteJobRepository) SaveATSScore(ctx context.Context, score *domain.ATSScore) error {
	query := `INSERT OR REPLACE INTO ats_evaluations (id, job_application_id, hard_skills_score, experience_score, keywords_score, education_score, language_score, total_score, recommendations_json, evaluated_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	if score.ID == "" {
		score.ID = uuid.NewString()
	}
	if score.EvaluatedAt.IsZero() {
		score.EvaluatedAt = time.Now().UTC()
	}

	_, err := r.db.ExecContext(ctx, query,
		score.ID, score.JobApplicationID, score.HardSkillsScore, score.ExperienceScore,
		score.KeywordsScore, score.EducationScore, score.LanguageScore, score.TotalScore,
		score.RecommendationsJSON, score.EvaluatedAt,
	)
	return err
}

func (r *SQLiteJobRepository) GetATSScore(ctx context.Context, jobAppID string) (*domain.ATSScore, error) {
	query := `SELECT id, job_application_id, hard_skills_score, experience_score, keywords_score, education_score, language_score, total_score, recommendations_json, evaluated_at
			  FROM ats_evaluations WHERE job_application_id = ? ORDER BY evaluated_at DESC LIMIT 1`
	var s domain.ATSScore
	var recs sql.NullString
	err := r.db.QueryRowContext(ctx, query, jobAppID).Scan(
		&s.ID, &s.JobApplicationID, &s.HardSkillsScore, &s.ExperienceScore,
		&s.KeywordsScore, &s.EducationScore, &s.LanguageScore, &s.TotalScore,
		&recs, &s.EvaluatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if recs.Valid {
		s.RecommendationsJSON = recs.String
	}
	return &s, nil
}

func (r *SQLiteJobRepository) SaveOptimizationChanges(ctx context.Context, changes []domain.OptimizationChange) error {
	if len(changes) == 0 {
		return nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO optimization_changes (id, job_application_id, cv_version_generated_id, change_summary, reason, related_requirement_name, expected_ats_impact, evidence_used) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, c := range changes {
		id := c.ID
		if id == "" {
			id = uuid.NewString()
		}
		if _, err := stmt.ExecContext(ctx, id, c.JobApplicationID, c.CVVersionGeneratedID, c.ChangeSummary, c.Reason, c.RelatedRequirementName, c.ExpectedATSImpact, c.EvidenceUsed); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *SQLiteJobRepository) GetOptimizationChanges(ctx context.Context, jobAppID string) ([]domain.OptimizationChange, error) {
	query := `SELECT id, job_application_id, cv_version_generated_id, change_summary, reason, related_requirement_name, expected_ats_impact, evidence_used FROM optimization_changes WHERE job_application_id = ?`
	rows, err := r.db.QueryContext(ctx, query, jobAppID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var changes []domain.OptimizationChange
	for rows.Next() {
		var c domain.OptimizationChange
		var cvVer, reqName, ev sql.NullString
		if err := rows.Scan(&c.ID, &c.JobApplicationID, &cvVer, &c.ChangeSummary, &c.Reason, &reqName, &c.ExpectedATSImpact, &ev); err != nil {
			return nil, err
		}
		if cvVer.Valid {
			c.CVVersionGeneratedID = cvVer.String
		}
		if reqName.Valid {
			c.RelatedRequirementName = reqName.String
		}
		if ev.Valid {
			c.EvidenceUsed = ev.String
		}
		changes = append(changes, c)
	}
	return changes, nil
}
