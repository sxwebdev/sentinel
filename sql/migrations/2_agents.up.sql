-- Create agents table
CREATE TABLE IF NOT EXISTS agents (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  description TEXT,
  secret_hash TEXT NOT NULL,
  token_hint TEXT NOT NULL,
  fingerprint TEXT,
  last_assignment_rev TEXT,
  "status" TEXT NOT NULL DEFAULT 'unknown',
  is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
  tags jsonb NOT NULL DEFAULT '[]',
  config jsonb NOT NULL DEFAULT '{}',
  system_info jsonb NOT NULL DEFAULT '{}',
  last_online_at DATETIME,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_agents_name ON agents(name); 
CREATE INDEX IF NOT EXISTS idx_agents_enabled ON agents(is_enabled);

-- Create services_agents table (many-to-many relationship between services and agents)
CREATE TABLE IF NOT EXISTS services_agents (
  id TEXT PRIMARY KEY,
  service_id TEXT NOT NULL REFERENCES services(id) ON DELETE CASCADE,
  agent_id TEXT NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
  revision TEXT NOT NULL CHECK (revision != '')
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_services_agents_unique ON services_agents(service_id, agent_id);

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
  service_id TEXT REFERENCES services(id) ON DELETE SET NULL,
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
