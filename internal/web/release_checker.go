package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
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

// checkNewVersion checks if new versions are available from GitHub releases
func (s *Server) checkNewVersion() error {
	s.logger.Info("Checking for new versions...")

	// Get all releases from GitHub
	httpClient := http.DefaultClient
	httpClient.Timeout = time.Second * 10

	// Get multiple pages of releases to ensure we don't miss any
	resp, err := httpClient.Get("https://api.github.com/repos/sxwebdev/sentinel/releases?per_page=100")
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var releases []struct {
		TagName     string `json:"tag_name"`
		Body        string `json:"body"`
		PublishedAt string `json:"published_at"`
		Draft       bool   `json:"draft"`
		Prerelease  bool   `json:"prerelease"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil
	}

	// Parse current version
	var currentVersion *semver.Version

	// Handle special cases like "local", "dev", etc.
	if s.serverInfo.Version == "local" || s.serverInfo.Version == "dev" || s.serverInfo.Version == "" {
		// For development versions, consider them as version 0.0.0
		s.logger.Debugf("Development version detected (%s), treating as 0.0.0 for comparison", s.serverInfo.Version)
		currentVersion, err = semver.NewVersion("0.0.0")
	} else {
		currentVersion, err = semver.NewVersion(s.serverInfo.Version)
	}

	if err != nil {
		s.logger.Warnf("Failed to parse current version %s: %v", s.serverInfo.Version, err)
		return nil
	}

	// Find newer releases (excluding drafts and prereleases)
	var newerReleases []struct {
		TagName     string
		Body        string
		PublishedAt string
		Version     *semver.Version
	}

	for _, release := range releases {
		// Skip drafts and prereleases
		if release.Draft || release.Prerelease {
			continue
		}

		// Parse release version
		releaseVersion, err := semver.NewVersion(release.TagName)
		if err != nil {
			continue // Skip invalid version tags
		}

		// Check if release is newer than current version
		if releaseVersion.GreaterThan(currentVersion) {
			newerReleases = append(newerReleases, struct {
				TagName     string
				Body        string
				PublishedAt string
				Version     *semver.Version
			}{
				TagName:     release.TagName,
				Body:        release.Body,
				PublishedAt: release.PublishedAt,
				Version:     releaseVersion,
			})
		}
	}

	// If no newer releases found, clear the update info
	if len(newerReleases) == 0 {
		s.serverInfo.AvailableUpdate = nil
		return nil
	}

	// Get the latest version info
	latestRelease := newerReleases[0]

	// Combine descriptions from all newer releases
	var combinedBody strings.Builder
	combinedBody.WriteString(fmt.Sprintf("# Found %d newer version(s):\n\n", len(newerReleases)))

	for idx, release := range newerReleases {
		combinedBody.WriteString(fmt.Sprintf("## Version %s\n", release.TagName))
		if release.Body != "" {
			combinedBody.WriteString(release.Body)
		} else {
			combinedBody.WriteString("No release notes available.")
		}

		if idx < len(newerReleases)-1 {
			combinedBody.WriteString("\n\n---\n\n")
		}
	}

	s.logger.Infof("Found %d newer version(s), latest: %s", len(newerReleases), latestRelease.TagName)

	s.serverInfo.AvailableUpdate = &AvailableUpdate{
		IsAvailableManual: s.config.Upgrader.IsEnabled,
		TagName:           latestRelease.TagName,
		URL:               "https://github.com/sxwebdev/sentinel/releases/tag/" + latestRelease.TagName,
		Description:       combinedBody.String(),
	}

	return nil
}
