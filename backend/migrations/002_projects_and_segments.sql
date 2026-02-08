-- Таблица проектов
CREATE TABLE IF NOT EXISTS projects (
  id CHAR(36) PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  description TEXT NULL,
  status VARCHAR(50) NOT NULL DEFAULT 'active',
  user_session_id VARCHAR(64) NOT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NULL,
  INDEX idx_projects_session (user_session_id),
  CONSTRAINT fk_projects_session
    FOREIGN KEY (user_session_id) REFERENCES user_sessions(session_id)
    ON DELETE CASCADE
);

-- Добавить project_id в files (nullable для обратной совместимости)
ALTER TABLE files ADD COLUMN project_id CHAR(36) NULL AFTER user_session_id;
ALTER TABLE files ADD INDEX idx_files_project (project_id);
ALTER TABLE files ADD CONSTRAINT fk_files_project
  FOREIGN KEY (project_id) REFERENCES projects(id)
  ON DELETE SET NULL;

-- Расширить transcription_tasks
ALTER TABLE transcription_tasks ADD COLUMN processing_time INT NULL AFTER error_message;
ALTER TABLE transcription_tasks ADD COLUMN speaker_data JSON NULL AFTER processing_time;

-- Таблица сегментов транскрипции
CREATE TABLE IF NOT EXISTS transcription_segments (
  id CHAR(36) PRIMARY KEY,
  task_id CHAR(36) NOT NULL,
  speaker_id VARCHAR(50) NULL,
  speaker_name VARCHAR(255) NULL,
  start_time INT NOT NULL COMMENT 'начало в миллисекундах',
  end_time INT NOT NULL COMMENT 'конец в миллисекундах',
  text TEXT NOT NULL,
  has_fillers TINYINT(1) NOT NULL DEFAULT 0,
  is_corrected TINYINT(1) NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL,
  INDEX idx_segments_task (task_id),
  INDEX idx_segments_speaker (task_id, speaker_id),
  CONSTRAINT fk_segments_task
    FOREIGN KEY (task_id) REFERENCES transcription_tasks(id)
    ON DELETE CASCADE
);
