package chat

import (
	"fmt"
	"strings"
	"unicode/utf16"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kelseyhightower/envconfig"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/slack-go/slack"
	"k8s.io/klog"
)

// CrashNotification contains all data to print out a informative crash note.
type CrashNotification struct {
	Title   string
	Message string
	Reason  string
	Logs    string
}

// Client informs a communication channel about a crash.
type Client interface {
	Send(*CrashNotification)
}

type MattermostConfig struct {
	Token         string
	URL           string
	Team, Channel string
}

type MattermostClient struct {
	mattermost *model.Client4
	channel    *model.Channel
}

// Send constructs a new mattermost message containing information about the crash.
func (client *MattermostClient) Send(note *CrashNotification) {
	attachment := &model.SlackAttachment{
		Color: "#AD2200",
		Text:  note.Message,
		Title: note.Title,
		Fields: []*model.SlackAttachmentField{
			{
				Title: "Logs",
				Value: "```\n" + note.Logs + "```",
			},
		},
	}
	// Check for termination message
	if note.Reason != "" {
		attachment.Fields = append(attachment.Fields, &model.SlackAttachmentField{
			Title: "Reason",
			Value: note.Reason,
		})
	}
	client.sendAttachments(attachment)
}

func (client *MattermostClient) sendAttachments(attachements ...*model.SlackAttachment) {
	post := &model.Post{ChannelId: client.channel.Id}
	model.ParseSlackAttachment(post, attachements)
	_, _, err := client.mattermost.CreatePost(post)
	if err != nil {
		klog.Warningf("Failed to notify Mattermost: %v", err)
	}
}

type ClientConfig struct {
	Type string `default:"mattermost"`
}

// NewClientFromEnv instantiates a new chat client using env configuration.
func NewClientFromEnv() (Client, error) {
	var (
		cfg    ClientConfig
		client Client
		err    error
	)
	if err := envconfig.Process("informer", &cfg); err != nil {
		return nil, err
	}
	switch cfg.Type {
	case "mattermost":
		client, err = NewMattermostClientFromEnv()
	case "slack":
		client, err = NewSlackClientFromEnv()
	case "telegram":
		client, err = NewTelegramClientFromEnv()
	default:
		err = fmt.Errorf("unknown client type: %s", cfg.Type)
	}
	if err != nil {
		return nil, fmt.Errorf("create client from env: %w", err)
	}
	return client, nil
}

// NewMattermostClientFromEnv instantiates and configures a Mattermost client.
func NewMattermostClientFromEnv() (*MattermostClient, error) {
	var cfg MattermostConfig
	if err := envconfig.Process("mattermost", &cfg); err != nil {
		return nil, err
	}
	client := model.NewAPIv4Client(cfg.URL)
	client.SetToken(cfg.Token)

	team, _, err := client.GetTeamByName(cfg.Team, "")
	if err != nil {
		return nil, fmt.Errorf("get team: %w", err)
	}
	channel, _, err := client.GetChannelByName(cfg.Channel, team.Id, "")
	if err != nil {
		return nil, fmt.Errorf("get channel: %w", err)
	}
	return &MattermostClient{client, channel}, nil
}

type SlackConfig struct {
	Token   string
	Channel string
}

type SlackClient struct {
	Client  *slack.Client
	Channel string
}

func (client *SlackClient) Send(note *CrashNotification) {
	blocks := []slack.Block{
		slack.NewHeaderBlock(&slack.TextBlockObject{
			Type: slack.PlainTextType,
			Text: note.Title,
		}),
		slack.NewSectionBlock(&slack.TextBlockObject{
			Type: slack.MarkdownType,
			Text: note.Message,
		}, []*slack.TextBlockObject{}, nil),
		slack.NewDividerBlock(),
		slack.NewSectionBlock(&slack.TextBlockObject{
			Type: slack.MarkdownType,
			Text: "*Logs*\n```\n" + note.Logs + "\n```",
		}, []*slack.TextBlockObject{}, nil),
	}
	if note.Reason != "" {
		blocks = append(blocks, slack.NewSectionBlock(&slack.TextBlockObject{
			Type: slack.MarkdownType,
			Text: "*Reason*\n```\n" + note.Reason + "\n```",
		}, []*slack.TextBlockObject{}, nil))
	}
	_, _, _, err := client.Client.SendMessage(client.Channel, slack.MsgOptionBlocks(blocks...))
	if err != nil {
		klog.Warningf("Failed to notify Slack: %v", err)
	}
}

// NewSlackClientFromEnv instantiates and configures a Slack client.
func NewSlackClientFromEnv() (*SlackClient, error) {
	var cfg SlackConfig
	if err := envconfig.Process("slack", &cfg); err != nil {
		return nil, err
	}
	api := slack.New(cfg.Token)
	return &SlackClient{
		Client:  api,
		Channel: cfg.Channel,
	}, nil
}

type TelegramConfig struct {
	Token  string
	ChatID int64
}

type TelegramClient struct {
	chat tgbotapi.Chat
	bot  *tgbotapi.BotAPI
}

func (client *TelegramClient) Send(note *CrashNotification) {
	// Generate text
	type entityMarker struct {
		Type    string
		Content string
	}
	contents := []entityMarker{
		{"bold", note.Title},
		{"", note.Message},
		{"bold", "Logs"},
		{"pre", note.Logs},
		{"bold", "Reason"},
		{"pre", note.Reason},
	}
	// Generate text
	cleartext := []string{}
	for _, c := range contents {
		cleartext = append(cleartext, c.Content)
	}
	// Get all joined groups
	message := tgbotapi.NewMessage(client.chat.ID, strings.Join(cleartext, "\n"))
	offset := 0
	// Generate list of entities
	for _, c := range contents {
		length := len(utf16.Encode([]rune(c.Content)))
		if c.Type != "" {
			message.Entities = append(message.Entities, tgbotapi.MessageEntity{
				Type:   c.Type,
				Offset: offset,
				Length: length,
			})
		}
		offset += length + 1
	}
	_, err := client.bot.Send(message)
	if err != nil {
		fmt.Println(err)
	}
}

// NewTelegramClientFromEnv instantiates and configures a Telegram client.
func NewTelegramClientFromEnv() (*TelegramClient, error) {
	var cfg TelegramConfig
	if err := envconfig.Process("telegram", &cfg); err != nil {
		return nil, err
	}
	bot, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		return nil, fmt.Errorf("init bot: %w", err)
	}
	// Get chat to verify
	chat, err := bot.GetChat(tgbotapi.ChatInfoConfig{
		ChatConfig: tgbotapi.ChatConfig{
			ChatID: cfg.ChatID,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("get chat: %w", err)
	}
	return &TelegramClient{
		chat: chat,
		bot:  bot,
	}, nil
}
