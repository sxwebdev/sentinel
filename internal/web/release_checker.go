package web

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// checkNewVersion checks if a new version is available from github releases
func (s *Server) checkNewVersionWrapper(ctx context.Context) {
	ticker := time.NewTicker(time.Minute * 30)
	defer ticker.Stop()

	go func() {
		if err := s.checkNewVersion(); err != nil {
			s.logger.Errorf("failed to check for new version: %v", err)
		}
	}()

	for {
		select {
		case <-ticker.C:
			// Check for new version
			if err := s.checkNewVersion(); err != nil {
				s.logger.Errorf("failed to check for new version: %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

// checkNewVersion checks if a new version is available from GitHub releases
func (s *Server) checkNewVersion() error {
	s.logger.Info("Checking for new version...")

	// get latest release from GitHub
	httpClient := http.DefaultClient
	httpClient.Timeout = time.Second * 10

	resp, err := httpClient.Get("https://api.github.com/repos/sxwebdev/sentinel/releases/latest")
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var releaseInfo struct {
		TagName string `json:"tag_name"`
		Body    string `json:"body"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&releaseInfo); err != nil {
		return nil
	}

	// Compare with current version
	if releaseInfo.TagName != s.serverInfo.Version {
		s.logger.Infof("New version available: %s", releaseInfo.TagName)

		s.serverInfo.AvailableUpdate = &AvailableUpdate{
			IsAvailableManual: s.config.Upgrader.IsEnabled,
			TagName:           releaseInfo.TagName,
			URL:               "https://github.com/sxwebdev/sentinel/releases/tag/" + releaseInfo.TagName,
			Description:       releaseInfo.Body,
		}
	}

	return nil
}
