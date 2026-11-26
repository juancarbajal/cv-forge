CREATE TABLE IF NOT EXISTS cvs (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    primary_language TEXT NOT NULL DEFAULT 'es',
    favorite_version_id TEXT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS cv_versions (
    id TEXT PRIMARY KEY,
    cv_id TEXT NOT NULL,
    version_number INTEGER NOT NULL,
    file_path TEXT NOT NULL,
    source_version_id TEXT,
    change_reason TEXT,
    ai_provider TEXT,
    ai_model TEXT,
    is_favorite BOOLEAN NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL,
    FOREIGN KEY(cv_id) REFERENCES cvs(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS job_applications (
    id TEXT PRIMARY KEY,
    company TEXT NOT NULL,
    role TEXT NOT NULL,
    url TEXT,
    job_posting_raw TEXT NOT NULL,
    detected_language TEXT NOT NULL DEFAULT 'es',
    target_language TEXT NOT NULL DEFAULT 'es',
    cv_id TEXT NOT NULL,
    cv_version_id TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'Borrador',
    total_ats_score REAL NOT NULL DEFAULT 0.0,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY(cv_id) REFERENCES cvs(id) ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS requirements (
    id TEXT PRIMARY KEY,
    job_application_id TEXT NOT NULL,
    name TEXT NOT NULL,
    category TEXT NOT NULL,
    priority TEXT NOT NULL,
    FOREIGN KEY(job_application_id) REFERENCES job_applications(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS match_results (
    id TEXT PRIMARY KEY,
    job_application_id TEXT NOT NULL,
    requirement_id TEXT NOT NULL,
    requirement_name TEXT NOT NULL,
    status TEXT NOT NULL,
    confidence REAL NOT NULL,
    evidence TEXT,
    explanation TEXT,
    FOREIGN KEY(job_application_id) REFERENCES job_applications(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS ats_evaluations (
    id TEXT PRIMARY KEY,
    job_application_id TEXT NOT NULL,
    hard_skills_score REAL NOT NULL,
    experience_score REAL NOT NULL,
    keywords_score REAL NOT NULL,
    education_score REAL NOT NULL,
    language_score REAL NOT NULL,
    total_score REAL NOT NULL,
    recommendations_json TEXT,
    evaluated_at DATETIME NOT NULL,
    FOREIGN KEY(job_application_id) REFERENCES job_applications(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS optimization_changes (
    id TEXT PRIMARY KEY,
    job_application_id TEXT NOT NULL,
    cv_version_generated_id TEXT,
    change_summary TEXT NOT NULL,
    reason TEXT NOT NULL,
    related_requirement_name TEXT,
    expected_ats_impact REAL NOT NULL DEFAULT 0.0,
    evidence_used TEXT,
    FOREIGN KEY(job_application_id) REFERENCES job_applications(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS prompts (
    id TEXT PRIMARY KEY,
    prompt_type TEXT NOT NULL,
    name TEXT NOT NULL,
    version INTEGER NOT NULL,
    content TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS provider_configs (
    provider_name TEXT PRIMARY KEY,
    encrypted_api_key BLOB,
    default_model TEXT NOT NULL,
    base_url TEXT,
    is_active BOOLEAN NOT NULL DEFAULT 0,
    updated_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS audit_logs (
    id TEXT PRIMARY KEY,
    timestamp DATETIME NOT NULL,
    action TEXT NOT NULL,
    provider TEXT,
    model TEXT,
    result_status TEXT NOT NULL,
    details_json TEXT
);

CREATE INDEX IF NOT EXISTS idx_cv_versions_cv_id ON cv_versions(cv_id);
CREATE INDEX IF NOT EXISTS idx_requirements_job_id ON requirements(job_application_id);
CREATE INDEX IF NOT EXISTS idx_match_results_job_id ON match_results(job_application_id);
CREATE INDEX IF NOT EXISTS idx_prompts_type ON prompts(prompt_type);
CREATE INDEX IF NOT EXISTS idx_audit_timestamp ON audit_logs(timestamp);
