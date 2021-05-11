package config

import (
	"time"

	"github.com/alexflint/go-arg"
	"github.com/pkg/errors"
)

// Config represents the command configuration.
type Config struct {
	// GOOGLE_APPLICATION_CREDENTIALS used implicitly by gocloud.dev!
	GoogleApplicationCredentials string        `arg:"env:GOOGLE_APPLICATION_CREDENTIALS"`
	Loglevel                     string        `arg:"env:LOGLEVEL"`
	GCPProject                   string        `arg:"env:GCP_PROJECT"`
	ListenAddr                   string        `arg:"env:LISTEN_ADDR"`
	SlackToken                   string        `arg:"env:SLACK_TOKEN"`
	SlackChannel                 string        `arg:"env:SLACK_CHANNEL"`
	PubsubSubscription           string        `arg:"env:PUBSUB_SUBSCRIPTION"`
	IgnoreMessagesOlderThan      time.Duration `arg:"env:IGNORE_MESSAGES_OLDER_THAN"`
	MetricsNamespace             string        `arg:"env:METRICS_NAMESPACE"`
	MetadataKeys                 []string      `arg:"env:METADATA_KEYS"`
}

// New returns a pre-filled Config.
func New() (Config, error) {
	c := Config{
		ListenAddr:              ":8080",
		SlackChannel:            "chatops-dev",
		IgnoreMessagesOlderThan: 10 * time.Minute,
		MetricsNamespace:        "pubsubslack",
	}
	if err := arg.Parse(&c); err != nil {
		return c, errors.Wrap(err, "failed to parse config")
	}
	if c.GCPProject == "" {
		return c, errors.New("need GCP_PROJECT")
	}
	if c.GoogleApplicationCredentials == "" {
		return c, errors.New("need GOOGLE_APPLICATION_CREDENTIALS")
	}
	if c.PubsubSubscription == "" {
		return c, errors.New("no PUBSUB_SUBSCRIPTION")
	}
	if c.PubsubSubscription != "" {
		if c.SlackToken == "" {
			return c, errors.New("no Slack token provided")
		}
		if c.SlackChannel == "" {
			return c, errors.New("no Slack channel provided")
		}
	}
	return c, nil
}
