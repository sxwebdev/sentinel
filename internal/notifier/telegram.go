package notifier

import (
	"fmt"
	"log"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sxwebdev/sentinel/internal/config"
)

// Notifier defines the interface for sending notifications
type Notifier interface {
	SendAlert(serviceName string, incident *config.Incident) error
	SendRecovery(serviceName string, incident *config.Incident) error
}

// TelegramNotifier sends notifications via Telegram
type TelegramNotifier struct {
	bot    *tgbotapi.BotAPI
	chatID int64
}

// NewTelegramNotifier creates a new Telegram notifier
func NewTelegramNotifier(botToken, chatIDStr string) (*TelegramNotifier, error) {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot: %w", err)
	}

	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid chat ID: %w", err)
	}

	log.Printf("Telegram bot authorized: %s", bot.Self.UserName)

	return &TelegramNotifier{
		bot:    bot,
		chatID: chatID,
	}, nil
}

// SendAlert sends an alert notification when a service goes down
func (t *TelegramNotifier) SendAlert(serviceName string, incident *config.Incident) error {
	message := t.formatAlertMessage(serviceName, incident)
	return t.sendMessage(message)
}

// SendRecovery sends a recovery notification when a service comes back up
func (t *TelegramNotifier) SendRecovery(serviceName string, incident *config.Incident) error {
	message := t.formatRecoveryMessage(serviceName, incident)
	return t.sendMessage(message)
}

// formatAlertMessage formats an alert message
func (t *TelegramNotifier) formatAlertMessage(serviceName string, incident *config.Incident) string {
	return fmt.Sprintf(
		"🔴 *[ALERT]* %s is DOWN\n\n"+
			"• Service: `%s`\n"+
			"• Error: %s\n"+
			"• Started: %s\n"+
			"• Incident ID: `%s`",
		serviceName,
		serviceName,
		incident.Error,
		incident.StartTime.Format("2006-01-02 15:04:05 UTC"),
		incident.ID,
	)
}

// formatRecoveryMessage formats a recovery message
func (t *TelegramNotifier) formatRecoveryMessage(serviceName string, incident *config.Incident) string {
	var duration string
	if incident.Duration != nil {
		duration = formatDuration(*incident.Duration)
	} else {
		duration = formatDuration(time.Since(incident.StartTime))
	}

	var endTime string
	if incident.EndTime != nil {
		endTime = incident.EndTime.Format("2006-01-02 15:04:05 UTC")
	} else {
		endTime = time.Now().Format("2006-01-02 15:04:05 UTC")
	}

	return fmt.Sprintf(
		"🟢 *[RECOVERY]* %s is UP\n\n"+
			"• Service: `%s`\n"+
			"• Downtime: %s\n"+
			"• Recovered: %s\n"+
			"• Incident ID: `%s`",
		serviceName,
		serviceName,
		duration,
		endTime,
		incident.ID,
	)
}

// sendMessage sends a message to the configured chat
func (t *TelegramNotifier) sendMessage(text string) error {
	msg := tgbotapi.NewMessage(t.chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown

	_, err := t.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send telegram message: %w", err)
	}

	return nil
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh %dm", hours, minutes)
}