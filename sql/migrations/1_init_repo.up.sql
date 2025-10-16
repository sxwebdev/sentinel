PRAGMA foreign_keys = ON;

-- =========================
-- Users
-- =========================
CREATE TABLE IF NOT EXISTS users (
  id          TEXT PRIMARY KEY,
  email       TEXT NOT NULL UNIQUE,
  password    TEXT NOT NULL,
  full_name   TEXT NOT NULL,
  role        TEXT NOT NULL CHECK (role != ''), -- root/admin/user
  avatar      TEXT NOT NULL,
  created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- =========================
-- Projects
-- =========================
CREATE TABLE IF NOT EXISTS projects (
  id          TEXT PRIMARY KEY,
  name        TEXT NOT NULL,
  description TEXT NOT NULL,
  settings    JSONB NOT NULL DEFAULT (jsonb('{}')),
  created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- =========================
-- Agents
-- =========================
CREATE TABLE IF NOT EXISTS agents (
  id            TEXT PRIMARY KEY,
  name          TEXT NOT NULL,
  description   TEXT NOT NULL,
  secret_hash   TEXT NOT NULL,
  token_hint    TEXT NOT NULL,
  fingerprint   TEXT,
  kind          TEXT NOT NULL DEFAULT 'agent' CHECK (kind IN ('hub', 'external')),
  status        TEXT NOT NULL DEFAULT 'unknown',
  is_enabled    BOOLEAN NOT NULL DEFAULT 1,
  "location"    TEXT,
  tags          JSONB NOT NULL DEFAULT (jsonb('[]')),
  config        JSONB NOT NULL DEFAULT (jsonb('{}')),
  system_info   JSONB NOT NULL DEFAULT (jsonb('{}')),
  project_id    TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  last_seen_at  DATETIME,
  created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_agents_kind        ON agents(kind);
CREATE INDEX IF NOT EXISTS idx_agents_is_enabled  ON agents(is_enabled);
CREATE INDEX IF NOT EXISTS idx_agents_location    ON agents("location");
CREATE INDEX IF NOT EXISTS idx_agents_project     ON agents(project_id);

-- =========================
-- Resources per project
-- =========================
CREATE TABLE IF NOT EXISTS resources (
  id          TEXT PRIMARY KEY,
  project_id  TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  name        TEXT NOT NULL,
  description TEXT NOT NULL,
  tags        JSONB NOT NULL DEFAULT (jsonb('[]')),
  payload     JSONB NOT NULL DEFAULT (jsonb('{}')),
  created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_resources_project ON resources(project_id);
CREATE INDEX IF NOT EXISTS idx_resources_name    ON resources(name);

-- =========================
-- Resource <-> Agents links
-- =========================
CREATE TABLE IF NOT EXISTS resource_agents (
  id          TEXT PRIMARY KEY,
  resource_id TEXT NOT NULL REFERENCES resources(id) ON DELETE CASCADE,
  agent_id    TEXT NOT NULL REFERENCES agents(id)    ON DELETE CASCADE,
  project_id  TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_resource_agent ON resource_agents(resource_id, agent_id);
CREATE INDEX IF NOT EXISTS idx_resource_agents_resource ON resource_agents(resource_id);
CREATE INDEX IF NOT EXISTS idx_resource_agents_agent    ON resource_agents(agent_id);
CREATE INDEX IF NOT EXISTS idx_resource_agents_project  ON resource_agents(project_id);

-- =========================
-- Monitors <-> Agents links
-- =========================
CREATE TABLE monitor_agents (
  id          TEXT PRIMARY KEY,
  project_id  TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  monitor_id  TEXT NOT NULL REFERENCES monitors(id) ON DELETE CASCADE,
  agent_id    TEXT NOT NULL REFERENCES agents(id)   ON DELETE CASCADE,
  created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(monitor_id, agent_id)
);
CREATE INDEX IF NOT EXISTS idx_monitor_agents_monitor ON monitor_agents(monitor_id);
CREATE INDEX IF NOT EXISTS idx_monitor_agents_agent   ON monitor_agents(agent_id);
CREATE INDEX IF NOT EXISTS idx_monitor_agents_project ON monitor_agents(project_id);

-- =========================
-- Monitors & their revisions
-- =========================
CREATE TABLE IF NOT EXISTS monitors (
  id          TEXT PRIMARY KEY,
  project_id  TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  resource_id TEXT NOT NULL REFERENCES resources(id) ON DELETE CASCADE,
  name        TEXT NOT NULL,
  kind        TEXT NOT NULL CHECK (kind != ''), -- http/grpc/tcp/x509/kafka/pg/...
  is_enabled     BOOLEAN NOT NULL DEFAULT 1,
  tags        JSONB NOT NULL DEFAULT (jsonb('[]')),
  created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_monitors_project     ON monitors(project_id);
CREATE INDEX IF NOT EXISTS idx_monitors_resource    ON monitors(resource_id);
CREATE INDEX IF NOT EXISTS idx_monitors_is_enabled  ON monitors(is_enabled);
CREATE INDEX IF NOT EXISTS idx_monitors_kind        ON monitors(kind);


-- =========================
-- Monitor Revisions (config history)
-- =========================
CREATE TABLE IF NOT EXISTS monitor_revisions (
  id                   INTEGER PRIMARY KEY AUTOINCREMENT,
  monitor_id           TEXT NOT NULL REFERENCES monitors(id) ON DELETE CASCADE,
  config               JSONB NOT NULL DEFAULT (jsonb('{}')),  -- schedules/endpoints/rules/и т.д.
  content_hash_uint64  INTEGER NOT NULL,                      -- xxhash64 normalized JSON
  is_active            BOOLEAN NOT NULL DEFAULT 0,
  created_at           DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_monrev_monitor ON monitor_revisions(monitor_id);
CREATE UNIQUE INDEX IF NOT EXISTS uq_monrev_active ON monitor_revisions(monitor_id, is_active) WHERE is_active = 1;

-- =========================
-- Results: Monitor Checks
-- =========================
CREATE TABLE IF NOT EXISTS monitor_checks (
  id                TEXT PRIMARY KEY,
  project_id        TEXT NOT NULL REFERENCES projects(id)  ON DELETE CASCADE,
  monitor_id        TEXT NOT NULL REFERENCES monitors(id)  ON DELETE CASCADE,
  revision_id       INTEGER NOT NULL REFERENCES monitor_revisions(id) ON DELETE RESTRICT,
  resource_id       TEXT NOT NULL REFERENCES resources(id) ON DELETE CASCADE,
  agent_id          TEXT NOT NULL REFERENCES agents(id)    ON DELETE CASCADE,
  started_at        DATETIME NOT NULL,
  completed_at      DATETIME NOT NULL,
  severity          INTEGER NOT NULL CHECK (severity BETWEEN 0 AND 5),
  response_time     INTEGER NOT NULL DEFAULT 0,
  evaluation        JSONB NOT NULL DEFAULT (jsonb('{}')),
  meta              JSONB NOT NULL DEFAULT (jsonb('{}')),
  created_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_checks_monitor_time   ON monitor_checks(monitor_id, completed_at);
CREATE INDEX IF NOT EXISTS idx_checks_resource_time  ON monitor_checks(resource_id, completed_at);
CREATE INDEX IF NOT EXISTS idx_checks_agent_time     ON monitor_checks(agent_id, completed_at);
CREATE INDEX IF NOT EXISTS idx_checks_severity_time  ON monitor_checks(severity, completed_at DESC);

-- =========================
-- Current state (global & per-agent)
-- =========================
CREATE TABLE IF NOT EXISTS monitor_states (
  id                     TEXT PRIMARY KEY,
  project_id             TEXT NOT NULL REFERENCES projects(id)  ON DELETE CASCADE,
  monitor_id             TEXT NOT NULL REFERENCES monitors(id)  ON DELETE CASCADE,
  resource_id            TEXT NOT NULL REFERENCES resources(id) ON DELETE CASCADE,
  agent_id               TEXT REFERENCES agents(id)             ON DELETE SET NULL,
  severity               INTEGER NOT NULL DEFAULT 0 CHECK (severity BETWEEN 0 AND 5),
  status                 TEXT NOT NULL DEFAULT 'unknown',
  last_check             DATETIME,
  last_error             TEXT,
  consecutive_fails      INTEGER NOT NULL DEFAULT 0,
  consecutive_success    INTEGER NOT NULL DEFAULT 0,
  total_checks           INTEGER NOT NULL DEFAULT 0,
  avg_response_time      INTEGER NOT NULL DEFAULT 0,
  updated_at             DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_monstate_key ON monitor_states(monitor_id, resource_id, agent_id);
CREATE INDEX IF NOT EXISTS idx_monstate_project    ON monitor_states(project_id);
CREATE INDEX IF NOT EXISTS idx_monstate_severity   ON monitor_states(severity);
CREATE INDEX IF NOT EXISTS idx_monstate_last_check ON monitor_states(last_check DESC);

-- =========================
-- Incidents & their events
-- =========================
CREATE TABLE IF NOT EXISTS incidents (
  id            TEXT PRIMARY KEY,
  project_id    TEXT NOT NULL REFERENCES projects(id)  ON DELETE CASCADE,
  origin        TEXT NOT NULL CHECK (origin IN ('monitor','manual','external')),
  monitor_id    TEXT REFERENCES monitors(id)  ON DELETE CASCADE,
  resource_id   TEXT REFERENCES resources(id) ON DELETE CASCADE,
  agent_id      TEXT REFERENCES agents(id)    ON DELETE SET NULL,
  kind          TEXT NOT NULL CHECK (kind != ''),
  status        TEXT NOT NULL CHECK (status IN ('open','ack','resolved')),
  severity      INTEGER NOT NULL CHECK (severity BETWEEN 0 AND 5),
  summary       TEXT NOT NULL,
  first_seen_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  last_seen_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  resolved_at   DATETIME,
  created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_incident_open
ON incidents(monitor_id, resource_id, agent_id, kind)
WHERE status IN ('open','ack');

CREATE INDEX IF NOT EXISTS idx_incidents_project      ON incidents(project_id);
CREATE INDEX IF NOT EXISTS idx_incidents_status       ON incidents(status);
CREATE INDEX IF NOT EXISTS idx_incidents_first_seen   ON incidents(first_seen_at DESC);
CREATE INDEX IF NOT EXISTS idx_incidents_resolved_at  ON incidents(resolved_at);

-- =========================
-- Incident events (history of state changes)
-- =========================
CREATE TABLE IF NOT EXISTS incident_events (
  id           INTEGER PRIMARY KEY AUTOINCREMENT,
  incident_id  TEXT NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
  created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  type         TEXT NOT NULL CHECK (type IN ('open','escalate','ack','close','note','touch')),
  payload      JSONB NOT NULL DEFAULT (jsonb('{}'))
);
CREATE INDEX IF NOT EXISTS idx_incident_events_incident ON incident_events(incident_id);
CREATE INDEX IF NOT EXISTS idx_incident_events_time     ON incident_events(created_at DESC);

-- =========================
-- Notification providers
-- =========================
CREATE TABLE IF NOT EXISTS notification_providers (
  id            TEXT PRIMARY KEY,
  provider_type TEXT NOT NULL,
  config        JSONB NOT NULL DEFAULT (jsonb('{}')),
  is_enabled    BOOLEAN NOT NULL DEFAULT 1,
  project_id    TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_notif_providers_enabled ON notification_providers(is_enabled);
CREATE INDEX IF NOT EXISTS idx_notif_providers_project ON notification_providers(project_id);

-- =========================
-- Alert Policies
-- =========================
CREATE TABLE IF NOT EXISTS alert_policies (
  id          TEXT PRIMARY KEY,
  name        TEXT NOT NULL,
  rules       JSONB NOT NULL,
  is_enabled  BOOLEAN NOT NULL DEFAULT 1,
  project_id  TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_alert_policies_project    ON alert_policies(project_id);
CREATE INDEX IF NOT EXISTS idx_alert_policies_enabled    ON alert_policies(is_enabled);

-- ==========================
-- Alert Policies links to resources/monitors
-- ==========================
CREATE TABLE IF NOT EXISTS alert_policy_resources (
  policy_id  TEXT NOT NULL REFERENCES alert_policies(id) ON DELETE CASCADE,
  resource_id TEXT NOT NULL REFERENCES resources(id)     ON DELETE CASCADE,
  PRIMARY KEY (policy_id, resource_id)
);
CREATE TABLE IF NOT EXISTS alert_policy_monitors (
  policy_id  TEXT NOT NULL REFERENCES alert_policies(id) ON DELETE CASCADE,
  monitor_id TEXT NOT NULL REFERENCES monitors(id)       ON DELETE CASCADE,
  PRIMARY KEY (policy_id, monitor_id)
);

-- =========================
-- Alerts (created alerts for incidents)
-- =========================
CREATE TABLE IF NOT EXISTS alerts (
  id           TEXT PRIMARY KEY,
  incident_id  TEXT NOT NULL REFERENCES incidents(id)       ON DELETE CASCADE,
  policy_id    TEXT NOT NULL REFERENCES alert_policies(id) ON DELETE RESTRICT,
  status       TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','sent','silenced','failed')),
  created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_alerts_incident ON alerts(incident_id);
CREATE INDEX IF NOT EXISTS idx_alerts_status   ON alerts(status);

-- =========================
-- Notifications history
-- =========================
CREATE TABLE IF NOT EXISTS notification_history (
  id             TEXT PRIMARY KEY,
  alert_id       TEXT REFERENCES alerts(id) ON DELETE CASCADE,
  provider_id    TEXT NOT NULL REFERENCES notification_providers(id) ON DELETE CASCADE,
  message        TEXT NOT NULL CHECK (message != ''),
  status         TEXT NOT NULL DEFAULT 'pending',
  response       TEXT,
  attempts       INTEGER NOT NULL DEFAULT 0,
  error_message  TEXT,
  last_attempt_at DATETIME,
  sent_at        DATETIME,
  created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_notification_history_provider ON notification_history(provider_id);
CREATE INDEX IF NOT EXISTS idx_notification_history_status   ON notification_history(status);

-- =========================
-- Maintenance windows (one-off) + explicit links
-- =========================
CREATE TABLE IF NOT EXISTS maintenance_windows (
  id          TEXT PRIMARY KEY,
  name        TEXT NOT NULL,
  start_at    DATETIME NOT NULL,   -- UTC
  end_at      DATETIME NOT NULL,   -- UTC
  reason      TEXT,
  created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CHECK (end_at > start_at)
);

-- =========================
-- Maintenance windows links to projects/resources/monitors
-- =========================
CREATE TABLE IF NOT EXISTS mw_projects (
  mw_id     TEXT NOT NULL REFERENCES maintenance_windows(id) ON DELETE CASCADE,
  project_id TEXT NOT NULL REFERENCES projects(id)           ON DELETE CASCADE,
  PRIMARY KEY (mw_id, project_id)
);
CREATE TABLE IF NOT EXISTS mw_resources (
  mw_id     TEXT NOT NULL REFERENCES maintenance_windows(id) ON DELETE CASCADE,
  resource_id TEXT NOT NULL REFERENCES resources(id)         ON DELETE CASCADE,
  PRIMARY KEY (mw_id, resource_id)
);
CREATE TABLE IF NOT EXISTS mw_monitors (
  mw_id     TEXT NOT NULL REFERENCES maintenance_windows(id) ON DELETE CASCADE,
  monitor_id TEXT NOT NULL REFERENCES monitors(id)           ON DELETE CASCADE,
  PRIMARY KEY (mw_id, monitor_id)
);
