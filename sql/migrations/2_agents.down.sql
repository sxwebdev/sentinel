-- Drop notification history table
DROP TABLE IF EXISTS notification_history;

-- Drop notification providers table
DROP TABLE IF EXISTS notification_providers;

-- Drop incident states table
DROP TABLE IF EXISTS incident_states;

-- Revert service_states response_time back to response_time_ns (from milliseconds to nanoseconds)
ALTER TABLE service_states RENAME COLUMN response_time TO response_time_ns;
UPDATE service_states SET response_time_ns = response_time_ns * 1000000 WHERE response_time_ns IS NOT NULL AND response_time_ns > 0;

-- Revert incidents duration back to duration_ns (from milliseconds to nanoseconds)
ALTER TABLE incidents RENAME COLUMN duration TO duration_ns;
UPDATE incidents SET duration_ns = duration_ns * 1000000 WHERE duration_ns IS NOT NULL AND duration_ns > 0;

-- Drop agents table
DROP TABLE IF EXISTS agents;
