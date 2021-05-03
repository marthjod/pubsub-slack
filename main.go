package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/go-chi/chi"
	"github.com/marthjod/pubsub-slack/config"
	"github.com/marthjod/pubsub-slack/pkg/publish"
	"github.com/nlopes/slack"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"gocloud.dev/pubsub"
	_ "gocloud.dev/pubsub/gcppubsub"
)

const subscriptionScheme = "gcppubsub"

func main() {
	ctx := context.Background()

	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()

	cfg, err := config.New()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to get config")
	}
	lvl, err := zerolog.ParseLevel(cfg.Loglevel)
	if err != nil {
		logger.Fatal().Err(err).Msg("unable to parse level")
	}
	logger = logger.Level(lvl)

	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, os.Interrupt)

	errChan := make(chan error)

	// implicit credentials via GOOGLE_APPLICATION_CREDENTIALS
	subscription := fmt.Sprintf("%s://projects/%s/subscriptions/%s", subscriptionScheme, cfg.GCPProject, cfg.PubsubSubscription)
	sub, err := pubsub.OpenSubscription(ctx, subscription)
	if err != nil {
		logger.Fatal().Err(err).Str("subscription", subscription).Msg("opening subscription")
	}
	defer func() {
		if err := sub.Shutdown(ctx); err != nil {
			logger.Error().Err(err).Msg("shutting down subscription")
		}
	}()
	logger.Debug().Str("subscription", subscription).Msg("connected to Pub/Sub subscription")

	slackClient := slack.New(cfg.SlackToken)
	slackPublisher := publish.NewSlack(sub, slackClient, cfg.SlackChannel, cfg.IgnoreMessagesOlderThan, logger)
	go slackPublisher.Publish(ctx, errChan)

	router := chi.NewRouter()
	router.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("OK")); err != nil {
			logger.Error().Err(err)
		}
	})
	router.Mount("/metrics", promhttp.Handler())

	go func() {
		logger.Info().Str("address", cfg.ListenAddr).Msg("listening")
		if err := http.ListenAndServe(cfg.ListenAddr, router); err != http.ErrServerClosed {
			logger.Error().Err(err)
		}
	}()

	for {
		select {
		case err := <-errChan:
			logger.Error().Err(err).Msg("received error")
		case <-termChan:
			logger.Info().Msg("shutting down")
			os.Exit(0)
		}
	}
}
