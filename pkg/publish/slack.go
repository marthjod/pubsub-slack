package publish

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/nlopes/slack"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"gocloud.dev/pubsub"
)

const publishTimeMetadataKey = "publish_time"

// Slack is a slack publisher subscribed to a Google Pub/Sub subscription.
type Slack struct {
	sub                     *pubsub.Subscription
	client                  *slack.Client
	channel                 string
	messageOpts             []slack.MsgOption
	logger                  zerolog.Logger
	errChan                 chan error
	ignoreMessagesOlderThan time.Duration
}

// NewSlack returns a new slack publisher.
func NewSlack(
	sub *pubsub.Subscription,
	client *slack.Client,
	channel string,
	ignoreMessagesOlderThan time.Duration,
	logger zerolog.Logger,
) *Slack {
	return &Slack{
		sub:                     sub,
		client:                  client,
		channel:                 channel,
		logger:                  logger,
		errChan:                 make(chan error),
		ignoreMessagesOlderThan: ignoreMessagesOlderThan,
	}
}

// Publish receives pubsub messages and posts them to a slack channel.
// Should be called as a goroutine.
func (s *Slack) Publish(ctx context.Context, errChan chan error) {
	for {

		msg, err := s.receiveMessage(ctx)
		if err != nil {
			errChan <- err
			continue
		}

		go func() {
			defer msg.Ack()

			shouldPost, err := s.isRecent(msg)
			if err != nil {
				errChan <- errors.Wrap(err, "determining publish time")
				return
			}

			if shouldPost {
				if err := s.postMessage(msg); err != nil {
					errChan <- errors.Wrap(err, "posting message to slack")
				}
			}
		}()
	}
}

func (s *Slack) receiveMessage(ctx context.Context) (*pubsub.Message, error) {
	msg, err := s.sub.Receive(ctx)
	if err != nil && err != context.Canceled {
		return nil, errors.Wrap(err, "receiving message from subscription")
	}
	s.logger.Debug().Str("pubsubMessage", fmt.Sprintf("%s", msg.Body)).Str("metadata", fmt.Sprintf("%v", msg.Metadata)).Msg("received message from Pub/Sub")
	return msg, nil
}

func getPublishTime(m *pubsub.Message) (time.Time, error) {
	var _t time.Time
	if val, ok := m.Metadata[publishTimeMetadataKey]; ok {
		t, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return _t, fmt.Errorf("unable to convert '%s' metadata value", publishTimeMetadataKey)
		}
		return time.Unix(t, 0), nil
	}
	return _t, fmt.Errorf("key '%s' not found in message metadata", publishTimeMetadataKey)
}

func (s *Slack) isRecent(m *pubsub.Message) (bool, error) {
	publishTime, err := getPublishTime(m)
	if err != nil {
		return false, err
	}
	return publishTime.After(time.Now().Add(-s.ignoreMessagesOlderThan)), nil
}

func (s *Slack) postMessage(m *pubsub.Message) error {
	body := string(m.Body)

	publishTime, err := getPublishTime(m)
	if err != nil {
		s.logger.Warn().Msg("unable to extract publish time")
	} else {
		body += fmt.Sprintf(" (%s)", publishTime)
	}
	opts := append(
		s.messageOpts,
		slack.MsgOptionText(body, false),
	)
	_, _, err = s.client.PostMessage(
		s.channel,
		opts...,
	)
	return err
}
