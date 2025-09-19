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
