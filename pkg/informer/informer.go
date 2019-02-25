package informer

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/mattermost/mattermost-server/model"
)

// CrashNotification contains all data to print out a informative crash note.
type CrashNotification struct {
	Title   string
	Message string
	Reason  string
	Logs    string
}

// Informer informs a communication channel about a crash.
type Informer interface {
	Inform(*CrashNotification)
}

type MattermostConfig struct {
	User          string
	Password      string
	URL           string
	Team, Channel string
}

type MattermostInformer struct {
	mattermost *model.Client4
	user       *model.User
	channel    *model.Channel
}

// Inform constructs a new mattermost message containing information about the crash.
func (informer *MattermostInformer) Inform(note *CrashNotification) {
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
	informer.sendAttachements(attachment)
}

func (informer *MattermostInformer) sendAttachements(attachements ...*model.SlackAttachment) {
	post := &model.Post{ChannelId: informer.channel.Id}
	model.ParseSlackAttachment(post, attachements)
	informer.mattermost.CreatePost(post)
}

func (informer *MattermostInformer) Send(msg string) {
	post := &model.Post{
		ChannelId: informer.channel.Id,
		Message:   msg,
	}
	informer.mattermost.CreatePost(post)
}

func NewMattermostInformerFromEnv() (*MattermostInformer, error) {
	var cfg MattermostConfig
	if err := envconfig.Process("mattermost", &cfg); err != nil {
		return nil, err
	}
	client := model.NewAPIv4Client(cfg.URL)
	user, resp := client.Login(cfg.User, cfg.Password)
	if resp.Error != nil {
		return nil, resp.Error
	}
	team, resp := client.GetTeamByName(cfg.Team, "")
	if resp.Error != nil {
		return nil, resp.Error
	}
	channel, resp := client.GetChannelByName(cfg.Channel, team.Id, "")
	if resp.Error != nil {
		return nil, resp.Error
	}
	return &MattermostInformer{client, user, channel}, nil
}
