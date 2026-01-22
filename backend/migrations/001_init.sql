CREATE TABLE IF NOT EXISTS schema_migrations (
  version VARCHAR(64) PRIMARY KEY,
  applied_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS user_sessions (
  session_id VARCHAR(64) PRIMARY KEY,
  created_at DATETIME NOT NULL,
  last_activity DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS files (
  id CHAR(36) PRIMARY KEY,
  original_name VARCHAR(255) NOT NULL,
  storage_path VARCHAR(512) NOT NULL UNIQUE,
  file_size BIGINT NOT NULL,
  mime_type VARCHAR(128) NOT NULL,
  uploaded_at DATETIME NOT NULL,
  user_session_id VARCHAR(64) NOT NULL,
  INDEX idx_files_session (user_session_id),
  CONSTRAINT fk_files_session
    FOREIGN KEY (user_session_id) REFERENCES user_sessions(session_id)
    ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS transcription_tasks (
  id CHAR(36) PRIMARY KEY,
  file_id CHAR(36) NOT NULL,
  status ENUM('ожидает','в процессе','готово','ошибка') NOT NULL DEFAULT 'ожидает',
  provider VARCHAR(64) NOT NULL,
  language VARCHAR(16) NULL,
  transcript_text LONGTEXT NULL,
  raw_response JSON NULL,
  error_message VARCHAR(512) NULL,
  created_at DATETIME NOT NULL,
  started_at DATETIME NULL,
  completed_at DATETIME NULL,
  INDEX idx_tasks_file (file_id),
  INDEX idx_tasks_status (status),
  CONSTRAINT fk_tasks_file
    FOREIGN KEY (file_id) REFERENCES files(id)
    ON DELETE CASCADE
);
