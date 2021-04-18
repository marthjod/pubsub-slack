#!/bin/bash

set -o nounset
set -o errexit
set -o errtrace
set -o pipefail

go build -o cmd

slack_token="${SLACK_TOKEN:-}"
gcp_project="${GCP_PROJECT:-}"

if [ -z "${slack_token}" ]; then echo "need SLACK_TOKEN"; exit 1; fi
if [ -z "${gcp_project}" ]; then echo "need GCP_PROJECT"; exit 1; fi
if [ ! -f "service-account.json" ]; then echo "need service-account.json"; exit 1; fi

export GCP_PROJECT="" # add GCP project here
export SLACK_TOKEN="" # add Slack token here
export PUBSUB_SUBSCRIPTION="slack-chatops-notifier"
export LOGLEVEL="debug"

export GOOGLE_APPLICATION_CREDENTIALS="./service-account.json"
export SLACK_CHANNEL="chatops-dev"

./cmd | jq -R 'fromjson? | .'

# gcloud pubsub topics publish projects/my-project/topics/slack-chatops-notifications \
#   --message "<@slackhandle> something has happened" \
#   --attribute publish_time="$(date +%s)"
