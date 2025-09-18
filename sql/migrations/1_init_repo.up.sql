-- Create services table
CREATE TABLE IF NOT EXISTS services (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  protocol TEXT NOT NULL,
  interval TEXT NOT NULL,
  timeout TEXT NOT NULL,
  retries INTEGER NOT NULL DEFAULT 3,
  tags jsonb NOT NULL DEFAULT '[]',
  config jsonb NOT NULL DEFAULT '{}',
  is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create incidents table
CREATE TABLE IF NOT EXISTS incidents (
  id TEXT PRIMARY KEY,
  service_id TEXT NOT NULL REFERENCES services(id),
  start_time DATETIME NOT NULL,
  end_time DATETIME,
  error TEXT NOT NULL,
  duration_ns INTEGER,
  resolved BOOLEAN NOT NULL DEFAULT FALSE,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create service_states table for current service states
CREATE TABLE IF NOT EXISTS service_states (
  id TEXT PRIMARY KEY,
  service_id TEXT NOT NULL REFERENCES services(id),
  status TEXT NOT NULL DEFAULT 'unknown',
  last_check DATETIME,
  next_check DATETIME,
  last_error TEXT,
  consecutive_fails INTEGER NOT NULL DEFAULT 0,
  consecutive_success INTEGER NOT NULL DEFAULT 0,
  total_checks INTEGER NOT NULL DEFAULT 0,
  response_time_ns INTEGER,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(service_id)
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_services_name ON services(name);
CREATE INDEX IF NOT EXISTS idx_services_enabled ON services(is_enabled);
CREATE INDEX IF NOT EXISTS idx_services_created_at ON services(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_incidents_service_id ON incidents(service_id);
CREATE INDEX IF NOT EXISTS idx_incidents_start_time_desc ON incidents(start_time DESC);
CREATE INDEX IF NOT EXISTS idx_incidents_resolved ON incidents(resolved);

CREATE INDEX IF NOT EXISTS idx_service_states_service_id ON service_states(service_id);
CREATE INDEX IF NOT EXISTS idx_service_states_status ON service_states(status);
CREATE INDEX IF NOT EXISTS idx_service_states_last_check ON service_states(last_check DESC);
CREATE INDEX IF NOT EXISTS idx_service_states_next_check ON service_states(next_check);
