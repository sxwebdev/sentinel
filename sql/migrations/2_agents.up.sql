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
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(token_ct),
  UNIQUE(fingerprint)
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_agents_name ON agents(name); 
CREATE INDEX IF NOT EXISTS idx_agents_enabled ON agents(is_enabled);

-- Update incidents duration. Converting from nanoseconds to milliseconds and renaming the column to duration
UPDATE incidents SET duration_ns = duration_ns / 1000000 WHERE duration_ns IS NOT NULL AND duration_ns > 0;
ALTER TABLE incidents RENAME COLUMN duration_ns TO duration;

-- Update service_states response_time_ns to response_time (from nanoseconds to milliseconds)
UPDATE service_states SET response_time_ns = response_time_ns / 1000000 WHERE response_time_ns IS NOT NULL AND response_time_ns > 0;
ALTER TABLE service_states RENAME COLUMN response_time_ns TO response_time;
