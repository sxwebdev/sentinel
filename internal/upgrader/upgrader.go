package upgrader

import (
	"errors"
	"os/exec"
	"strings"

	"github.com/sxwebdev/sentinel/internal/config"
	"github.com/tkcrm/mx/logger"
)

type Upgrader struct {
	logger logger.Logger
	config config.Upgrader
}

func New(l logger.Logger, cfg config.Upgrader) *Upgrader {
	return &Upgrader{
		logger: l,
		config: cfg,
	}
}

// Do performs the upgrade operation based on the configured command
func (u *Upgrader) Do() error {
	if u.config.IsEnabled == false {
		u.logger.Info("Upgrade is disabled, skipping.")
		return nil
	}

	if u.config.Command == "" {
		return errors.New("upgrade command is not configured")
	}

	u.logger.Infof("Starting upgrade process with command: %s", u.config.Command)

	// init reader and parse every line
	lines := strings.SplitSeq(u.config.Command, "\n")
	for line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue // Skip empty lines
		}

		// Execute the command
		if err := u.executeCommand(line); err != nil {
			u.logger.Errorf("Failed to execute command '%s': %v", line, err)
			return err
		}
	}

	return nil // Return an error if the command execution fails
}

func (u *Upgrader) executeCommand(command string) error {
	// Parse command and arguments
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return errors.New("empty command")
	}

	cmd := exec.Command(parts[0], parts[1:]...)

	// Set up logging for command output
	output, err := cmd.CombinedOutput()
	if err != nil {
		u.logger.Errorf("Command failed: %s, output: %s", command, strings.TrimSpace(string(output)))
		return err
	}

	u.logger.Infof("Command executed successfully: %s", command)
	if len(output) > 0 {
		u.logger.Infof("Command output: %s", strings.TrimSpace(string(output)))
	}

	return nil
}
