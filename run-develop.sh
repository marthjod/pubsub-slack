#!/bin/bash

set -o nounset
set -o errexit
set -o errtrace
set -o pipefail


slack_token="${SLACK_TOKEN:-}"
slack_channel="${SLACK_CHANNEL:-#chatops-dev}"
gcp_project="${GCP_PROJECT:-}"
pubsub_subscription="${PUBSUB_SUBSCRIPTION:-cluster-upgrades-dev}"

if [ -z "${slack_token}" ]; then echo "need SLACK_TOKEN"; exit 1; fi
if [ -z "${gcp_project}" ]; then echo "need GCP_PROJECT"; exit 1; fi
# needs Pub/Sub Subscriber role
if [ ! -f "service-account.json" ]; then echo "need service-account.json"; exit 1; fi

go build -o cmd

export GCP_PROJECT="${gcp_project}"
export SLACK_TOKEN="${slack_token}"
export PUBSUB_SUBSCRIPTION="${pubsub_subscription}"
export SLACK_CHANNEL="${slack_channel}"
export LOGLEVEL="debug"
export GOOGLE_APPLICATION_CREDENTIALS="./service-account.json"

./cmd | jq -R 'fromjson? | .'

# gcloud pubsub topics publish projects/${gcp_project}/topics/slack-chatops-notifications \
#   --message "<@slackhandle> something has happened" \
#   --attribute publish_time="$(date +%s)"
