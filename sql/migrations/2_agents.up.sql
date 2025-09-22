-- Create agents table
CREATE TABLE IF NOT EXISTS agents (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  description TEXT,
  host TEXT NOT NULL,
  port INT NOT NULL,
  token_ct BLOB,
  token_nonce BLOB,
  token_hint TEXT NOT NULL,
  fingerprint TEXT,
  "status" TEXT NOT NULL DEFAULT 'unknown',
  is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
  tags jsonb NOT NULL DEFAULT '[]',
  config jsonb NOT NULL DEFAULT '{}',
  system_info jsonb NOT NULL DEFAULT '{}',
  last_seen_at DATETIME,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(token_ct),
  UNIQUE(fingerprint)
);
CREATE INDEX IF NOT EXISTS idx_agents_name ON agents(name); 
CREATE INDEX IF NOT EXISTS idx_agents_enabled ON agents(is_enabled);

-- Update incidents duration. Converting from nanoseconds to milliseconds and renaming the column to duration
UPDATE incidents SET duration_ns = duration_ns / 1000000 WHERE duration_ns IS NOT NULL AND duration_ns > 0;
ALTER TABLE incidents RENAME COLUMN duration_ns TO duration;

-- Update service_states response_time_ns to response_time (from nanoseconds to milliseconds)
UPDATE service_states SET response_time_ns = response_time_ns / 1000000 WHERE response_time_ns IS NOT NULL AND response_time_ns > 0;
ALTER TABLE service_states RENAME COLUMN response_time_ns TO response_time;

-- Create incidents states table
CREATE TABLE IF NOT EXISTS incident_states (
  id TEXT PRIMARY KEY,
  incident_id TEXT NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
  "status" TEXT NOT NULL CHECK (status != ''),
  level INTEGER NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_incident_states_incident_id ON incident_states(incident_id);
CREATE INDEX IF NOT EXISTS idx_incident_states_status ON incident_states(status);

-- Create notifications providers table
CREATE TABLE IF NOT EXISTS notification_providers (
  id TEXT PRIMARY KEY,
  provider_type TEXT NOT NULL,
  config jsonb NOT NULL DEFAULT '{}',
  is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_notification_providers_enabled ON notification_providers(is_enabled);

-- Create notifications history table
CREATE TABLE IF NOT EXISTS notification_history (
  id TEXT PRIMARY KEY,
  provider_id TEXT NOT NULL REFERENCES notification_providers(id) ON DELETE CASCADE,
  incident_id TEXT REFERENCES incidents(id) ON DELETE SET NULL,
  message TEXT NOT NULL CHECK (message != ''),
  "status" TEXT NOT NULL DEFAULT 'pending' CHECK (status != ''),
  response TEXT,
  attempts INTEGER NOT NULL DEFAULT 0,
  error_message TEXT,
  last_attempt_at DATETIME,
  sent_at DATETIME,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_notification_history_provider ON notification_history(provider_id);
CREATE INDEX IF NOT EXISTS idx_notification_history_status ON notification_history(status);
