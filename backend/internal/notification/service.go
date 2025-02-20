package notification

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/slack-go/slack"
	"go.uber.org/zap"
)

type Service struct {
	config         NotificationConfig
	logger         *zap.Logger
	slackClient    *slack.Client
	discordSession *discordgo.Session
	wg             sync.WaitGroup
}

func NewService(config NotificationConfig, logger *zap.Logger) (*Service, error) {
	s := &Service{
		config: config,
		logger: logger,
	}

	// Initialize Slack client if configured
	if config.SlackToken != "" {
		s.slackClient = slack.New(config.SlackToken)
	}

	// Initialize Discord session if configured
	if config.DiscordToken != "" {
		session, err := discordgo.New("Bot " + config.DiscordToken)
		if err != nil {
			return nil, fmt.Errorf("failed to create Discord session: %w", err)
		}
		s.discordSession = session
	}

	return s, nil
}

func (s *Service) SendNotification(event NotificationEvent) {
	channels := event.Channels
	if len(channels) == 0 {
		channels = s.config.DefaultChannels
	}

	for _, channel := range channels {
		s.wg.Add(1)
		go func(ch NotificationChannel) {
			defer s.wg.Done()

			var err error
			switch ch {
			case ChannelSlack:
				err = s.sendSlackNotification(event)
			case ChannelDiscord:
				err = s.sendDiscordNotification(event)
			}

			if err != nil {
				s.logger.Error("Failed to send notification",
					zap.String("channel", string(ch)),
					zap.Error(err),
				)
			}
		}(channel)
	}
}

func (s *Service) sendSlackNotification(event NotificationEvent) error {
	if s.slackClient == nil {
		return fmt.Errorf("slack client not configured")
	}

	attachment := slack.Attachment{
		Color: s.getColorForEvent(event),
		Fields: []slack.AttachmentField{
			{
				Title: "Task",
				Value: event.Task.Title,
				Short: true,
			},
			{
				Title: "Status",
				Value: string(event.Task.Status),
				Short: true,
			},
			{
				Title: "Priority",
				Value: string(event.Task.Priority),
				Short: true,
			},
			{
				Title: "Due Date",
				Value: event.Task.DueDate.Format("2006-01-02"),
				Short: true,
			},
		},
		Footer: "Task Management System",
		Ts:     json.Number(fmt.Sprintf("%d", time.Now().Unix())),
	}

	_, _, err := s.slackClient.PostMessage(
		s.config.SlackChannel,
		slack.MsgOptionText(s.getNotificationTitle(event), false),
		slack.MsgOptionAttachments(attachment),
	)
	return err
}

func (s *Service) sendDiscordNotification(event NotificationEvent) error {
	if s.discordSession == nil {
		return fmt.Errorf("discord session not configured")
	}

	embed := &discordgo.MessageEmbed{
		Title: s.getNotificationTitle(event),
		Color: s.getDiscordColorForEvent(event),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Task",
				Value:  event.Task.Title,
				Inline: true,
			},
			{
				Name:   "Status",
				Value:  string(event.Task.Status),
				Inline: true,
			},
			{
				Name:   "Priority",
				Value:  string(event.Task.Priority),
				Inline: true,
			},
			{
				Name:   "Due Date",
				Value:  event.Task.DueDate.Format("2006-01-02"),
				Inline: true,
			},
		},
		Timestamp: time.Now().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Task Management System",
		},
	}

	_, err := s.discordSession.ChannelMessageSendEmbed(s.config.DiscordChannelID, embed)
	return err
}

func (s *Service) getNotificationTitle(event NotificationEvent) string {
	switch event.Type {
	case NotificationTypeTaskCreated:
		return "🆕 New Task Created"
	case NotificationTypeTaskUpdated:
		return "📝 Task Updated"
	case NotificationTypeTaskDeleted:
		return "🗑️ Task Deleted"
	case NotificationTypeTaskDue:
		return "⏰ Task Due Soon"
	default:
		return "Task Notification"
	}
}

func (s *Service) getColorForEvent(event NotificationEvent) string {
	switch event.Type {
	case NotificationTypeTaskCreated:
		return "#36a64f" // green
	case NotificationTypeTaskUpdated:
		return "#2196f3" // blue
	case NotificationTypeTaskDeleted:
		return "#f44336" // red
	case NotificationTypeTaskDue:
		return "#ff9800" // orange
	default:
		return "#9e9e9e" // grey
	}
}

func (s *Service) getDiscordColorForEvent(event NotificationEvent) int {
	switch event.Type {
	case NotificationTypeTaskCreated:
		return 0x36a64f // green
	case NotificationTypeTaskUpdated:
		return 0x2196f3 // blue
	case NotificationTypeTaskDeleted:
		return 0xf44336 // red
	case NotificationTypeTaskDue:
		return 0xff9800 // orange
	default:
		return 0x9e9e9e // grey
	}
}

func (s *Service) Close() {
	s.wg.Wait()
	if s.discordSession != nil {
		s.discordSession.Close()
	}
}
