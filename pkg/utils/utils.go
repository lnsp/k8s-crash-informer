package utils

import (
	"fmt"
	"io/ioutil"

	"github.com/kelseyhightower/envconfig"

	"github.com/mattermost/mattermost-server/model"
)

const namespaceFilePath = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"

// Namespace returns the namespace this pod is running in.
func Namespace() (string, error) {
	nsfile, err := ioutil.ReadFile(namespaceFilePath)
	if err != nil {
		return "", fmt.Errorf("could not read namespace: %v", err)
	}
	return string(nsfile), nil
}

type MattermostConfig struct {
	User          string
	Password      string
	URL           string
	Team, Channel string
}

type MattermostClient struct {
	mattermost *model.Client4
	user       *model.User
	channel    *model.Channel
}

func (client *MattermostClient) SendAttachements(attachements ...*model.SlackAttachment) {
	post := &model.Post{ChannelId: client.channel.Id}
	model.ParseSlackAttachment(post, attachements)
	client.mattermost.CreatePost(post)
}

func (client *MattermostClient) Send(msg string) {
	post := &model.Post{
		ChannelId: client.channel.Id,
		Message:   msg,
	}
	client.mattermost.CreatePost(post)
}

func NewMattermostClient() (*MattermostClient, error) {
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
	return &MattermostClient{client, user, channel}, nil
}
