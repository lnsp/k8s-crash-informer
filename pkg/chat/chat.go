package chat

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
	"github.com/mattermost/mattermost-server/model"
	"github.com/nlopes/slack"
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
	_, resp := client.mattermost.CreatePost(post)
	if resp.Error != nil {
		klog.Warningf("Failed to notify Mattermost: %v", resp.Error)
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
	team, resp := client.GetTeamByName(cfg.Team, "")
	if resp.Error != nil {
		return nil, resp.Error
	}
	channel, resp := client.GetChannelByName(cfg.Channel, team.Id, "")
	if resp.Error != nil {
		return nil, resp.Error
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
	attachment := slack.Attachment{
		Color: "#AD2200",
		Text:  note.Message,
		Title: note.Title,
		Fields: []slack.AttachmentField{
			{
				Title: "Logs",
				Value: "```\n" + note.Logs + "```",
			},
		},
	}
	if note.Reason != "" {
		attachment.Fields = append(attachment.Fields, slack.AttachmentField{
			Title: "Reason",
			Value: note.Reason,
		})
	}
	client.sendAttachments(attachment)
}

func (client *SlackClient) sendAttachments(attachments ...slack.Attachment) {
	_, _, _, err := client.Client.SendMessage(client.Channel, slack.MsgOptionAttachments(attachments...))
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
