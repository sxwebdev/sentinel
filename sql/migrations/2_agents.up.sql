-- Create agents table
CREATE TABLE IF NOT EXISTS agents (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  description TEXT,
  host TEXT,
  port INT,
  token_ct BLOB,
  token_nonce BLOB,
  token_hint TEXT,
  fingerprint TEXT,
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  system_info jsonb NOT NULL DEFAULT '{}',
  last_seen_at DATETIME,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(token_ct),
  UNIQUE(fingerprint)
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_agents_name ON agents(name); 
CREATE INDEX IF NOT EXISTS idx_agents_active ON agents(is_active);
