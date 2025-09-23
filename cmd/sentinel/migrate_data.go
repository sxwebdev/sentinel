package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sxwebdev/sentinel/pkg/sqlite"
	"github.com/urfave/cli/v3"
	_ "modernc.org/sqlite"
)

// -------- Old schema structs (export format) --------

type ServiceOld struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Protocol  string          `json:"protocol"`
	Interval  string          `json:"interval"` // time.Duration string (e.g., "1s", "250ms")
	Timeout   string          `json:"timeout"`
	Retries   int64           `json:"retries"`
	Tags      json.RawMessage `json:"tags"`
	Config    json.RawMessage `json:"config"`
	IsEnabled bool            `json:"is_enabled"`
	CreatedAt *time.Time      `json:"created_at,omitempty"`
	UpdatedAt *time.Time      `json:"updated_at,omitempty"`
}

type StateOld struct {
	ID               string     `json:"id"`
	ServiceID        string     `json:"service_id"`
	Status           string     `json:"status"`
	LastCheck        *time.Time `json:"last_check"`
	NextCheck        *time.Time `json:"next_check"`
	LastError        *string    `json:"last_error"`
	ConsecutiveFails int64      `json:"consecutive_fails"`
	ConsecutiveSucc  int64      `json:"consecutive_success"`
	TotalChecks      int64      `json:"total_checks"`
	ResponseTimeNS   *int64     `json:"response_time_ns"`
	CreatedAt        *time.Time `json:"created_at"`
	UpdatedAt        *time.Time `json:"updated_at"`
}

type IncidentOld struct {
	ID         string     `json:"id"`
	ServiceID  string     `json:"service_id"`
	StartTime  time.Time  `json:"start_time"`
	EndTime    *time.Time `json:"end_time"`
	Error      string     `json:"error"`
	DurationNS *int64     `json:"duration_ns"`
	Resolved   bool       `json:"resolved"`
	CreatedAt  *time.Time `json:"created_at"`
	UpdatedAt  *time.Time `json:"updated_at"`
}

func parseDurationMS(s string) (int64, error) {
	if strings.TrimSpace(s) == "" {
		return 0, nil
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, err
	}
	return d.Milliseconds(), nil
}

func readAll[T any](ctx context.Context, db *sql.DB, query string, scan func(*sql.Rows) (T, error)) ([]T, error) {
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []T
	for rows.Next() {
		v, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func exportCmd() *cli.Command {
	return &cli.Command{
		Name:  "export",
		Usage: "Export old schema tables to JSON files (services/service_states/incidents)",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "db", Required: true, Usage: "Path to SQLite DB file"},
			&cli.StringFlag{Name: "out", Required: true, Usage: "Output directory for JSON files"},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			dbpath := c.String("db")
			outDir := c.String("out")
			if err := os.MkdirAll(outDir, 0o755); err != nil {
				return err
			}
			db, err := sql.Open("sqlite", fmt.Sprintf("file:%s?_busy_timeout=5000&_pragma=foreign_keys(ON)", dbpath))
			if err != nil {
				return err
			}
			defer db.Close()

			ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			// services
			services, err := readAll(ctx, db, `SELECT id,name,protocol,interval,timeout,retries,tags,config,is_enabled,created_at,updated_at FROM services`, func(r *sql.Rows) (ServiceOld, error) {
				var s ServiceOld
				var tags, cfg []byte
				if err := r.Scan(&s.ID, &s.Name, &s.Protocol, &s.Interval, &s.Timeout, &s.Retries, &tags, &cfg, &s.IsEnabled, &s.CreatedAt, &s.UpdatedAt); err != nil {
					return s, err
				}
				s.Tags = append([]byte(nil), tags...)
				s.Config = append([]byte(nil), cfg...)
				return s, nil
			})
			if err != nil {
				return err
			}

			// service_states
			states, err := readAll(ctx, db, `SELECT id,service_id,status,last_check,next_check,last_error,consecutive_fails,consecutive_success,total_checks,response_time_ns,created_at,updated_at FROM service_states`, func(r *sql.Rows) (StateOld, error) {
				var s StateOld
				if err := r.Scan(&s.ID, &s.ServiceID, &s.Status, &s.LastCheck, &s.NextCheck, &s.LastError, &s.ConsecutiveFails, &s.ConsecutiveSucc, &s.TotalChecks, &s.ResponseTimeNS, &s.CreatedAt, &s.UpdatedAt); err != nil {
					return s, err
				}
				return s, nil
			})
			if err != nil {
				return err
			}

			// incidents
			incidents, err := readAll(ctx, db, `SELECT id,service_id,start_time,end_time,error,duration_ns,resolved,created_at,updated_at FROM incidents`, func(r *sql.Rows) (IncidentOld, error) {
				var s IncidentOld
				if err := r.Scan(&s.ID, &s.ServiceID, &s.StartTime, &s.EndTime, &s.Error, &s.DurationNS, &s.Resolved, &s.CreatedAt, &s.UpdatedAt); err != nil {
					return s, err
				}
				return s, nil
			})
			if err != nil {
				return err
			}

			// write files
			write := func(name string, v any) error {
				b, err := json.MarshalIndent(v, "", "  ")
				if err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(outDir, name), b, 0o644)
			}

			if err := write("services.old.json", services); err != nil {
				return err
			}
			if err := write("service_states.old.json", states); err != nil {
				return err
			}
			if err := write("incidents.old.json", incidents); err != nil {
				return err
			}

			fmt.Printf("Exported %d services, %d states, %d incidents to %s\n", len(services), len(states), len(incidents), outDir)
			return nil
		},
	}
}

func importCmd() *cli.Command {
	return &cli.Command{
		Name:  "import",
		Usage: "Import JSON into NEW schema (converts durations to ms)",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "db", Required: true, Usage: "Path to NEW SQLite DB file"},
			&cli.StringFlag{Name: "in", Required: true, Usage: "Input directory with JSON files"},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			dbpath := c.String("db")
			inDir := c.String("in")

			dsn := sqlite.GetDSN(dbpath)
			db, err := sql.Open("sqlite", dsn)
			if err != nil {
				return err
			}
			defer db.Close()

			ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
			defer cancel()

			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				return err
			}
			defer func() {
				if err != nil {
					_ = tx.Rollback()
				}
			}()

			// Load JSON
			var svcsOld []ServiceOld
			var stsOld []StateOld
			var incOld []IncidentOld
			if err = readJSON(filepath.Join(inDir, "services.old.json"), &svcsOld); err != nil {
				return err
			}
			if err = readJSON(filepath.Join(inDir, "service_states.old.json"), &stsOld); err != nil {
				return err
			}
			if err = readJSON(filepath.Join(inDir, "incidents.old.json"), &incOld); err != nil {
				return err
			}

			// Convert & insert
			if err = insertServices(ctx, tx, svcsOld); err != nil {
				return fmt.Errorf("services: %w", err)
			}
			if err = insertStates(ctx, tx, stsOld); err != nil {
				return fmt.Errorf("service_states: %w", err)
			}
			if err = insertIncidents(ctx, tx, incOld); err != nil {
				return fmt.Errorf("incidents: %w", err)
			}

			if err = tx.Commit(); err != nil {
				return err
			}
			_, _ = db.Exec(`PRAGMA integrity_check;`)
			_, _ = db.Exec(`PRAGMA optimize;`)
			fmt.Println("Import OK")
			return nil
		},
	}
}

func insertServices(ctx context.Context, tx *sql.Tx, in []ServiceOld) error {
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO services (id,name,protocol,interval,timeout,retries,tags,config,is_enabled,created_at,updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,COALESCE(?,CURRENT_TIMESTAMP),COALESCE(?,CURRENT_TIMESTAMP))`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, s := range in {
		intervalMS, err := parseDurationMS(s.Interval)
		if err != nil {
			return fmt.Errorf("id=%s interval %q: %w", s.ID, s.Interval, err)
		}
		timeoutMS, err := parseDurationMS(s.Timeout)
		if err != nil {
			return fmt.Errorf("id=%s timeout %q: %w", s.ID, s.Timeout, err)
		}
		protocol := s.Protocol
		if protocol == "" {
			protocol = "unknown"
		}
		if s.Tags == nil {
			s.Tags = json.RawMessage("[]")
		}
		if s.Config == nil {
			s.Config = json.RawMessage("{}")
		}
		if _, err := stmt.ExecContext(ctx, s.ID, s.Name, protocol, intervalMS, timeoutMS, s.Retries, s.Tags, s.Config, s.IsEnabled, formatDate(s.CreatedAt), formatDate(s.UpdatedAt)); err != nil {
			return err
		}
	}
	return nil
}

func insertStates(ctx context.Context, tx *sql.Tx, in []StateOld) error {
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO service_states (id,service_id,status,last_check,last_error,consecutive_fails,consecutive_success,total_checks,avg_response_time,created_at,updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,COALESCE(?,CURRENT_TIMESTAMP),COALESCE(?,CURRENT_TIMESTAMP))`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, s := range in {
		var avgMS *int64
		if s.ResponseTimeNS != nil {
			v := *s.ResponseTimeNS / 1_000_000
			avgMS = &v
		}
		if _, err := stmt.ExecContext(ctx, s.ID, s.ServiceID, s.Status, formatDate(s.LastCheck), s.LastError, s.ConsecutiveFails, s.ConsecutiveSucc, s.TotalChecks, avgMS, formatDate(s.CreatedAt), formatDate(s.UpdatedAt)); err != nil {
			return err
		}
	}
	return nil
}

func parseTimeFlexible(s string) (time.Time, error) {
	formats := []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05", "2006-01-02 15:04:05Z07:00"}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, errors.New("unsupported time format: " + s)
}

func insertIncidents(ctx context.Context, tx *sql.Tx, in []IncidentOld) error {
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO incidents (id,service_id,error,duration,started_at,resolved_at,created_at,updated_at)
		VALUES (?,?,?,?,?,?,COALESCE(?,CURRENT_TIMESTAMP),COALESCE(?,CURRENT_TIMESTAMP))`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, s := range in {
		var dur *int64
		if s.DurationNS != nil {
			v := *s.DurationNS / 1_000_000
			dur = &v
		} else if s.EndTime != nil && !s.EndTime.IsZero() && !s.StartTime.IsZero() {
			if st, err1 := parseTimeFlexible(s.StartTime.String()); err1 == nil {
				if et, err2 := parseTimeFlexible(s.EndTime.String()); err2 == nil {
					v := et.Sub(st).Milliseconds()
					dur = &v
				}
			}
		}

		var resolvedAt *string
		if s.Resolved && s.EndTime != nil {
			endTime := formatDate(s.EndTime)
			resolvedAt = endTime
		}

		if _, err := stmt.ExecContext(ctx, s.ID, s.ServiceID, s.Error, dur, formatDate(&s.StartTime), resolvedAt, formatDate(s.CreatedAt), formatDate(s.UpdatedAt)); err != nil {
			return err
		}
	}
	return nil
}

func readJSON(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

func formatDate(t *time.Time) *string {
	if t == nil {
		return nil
	}
	formatted := t.Format("2006-01-02 15:04:05")
	return &formatted
}
