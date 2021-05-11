package metadata

import (
	"fmt"
	"strings"

	"gocloud.dev/pubsub"
)

type Extractor interface {
	Extract(msg *pubsub.Message) map[string]string
	ExtractString(msg *pubsub.Message) string
}

type extractor struct {
	metadataKeys []string
}

func NewExtractor(metadataKeys []string) *extractor {
	return &extractor{
		metadataKeys: metadataKeys,
	}
}

func (e *extractor) Extract(msg *pubsub.Message) map[string]string {
	var m = map[string]string{}
	for _, key := range e.metadataKeys {
		if val, ok := msg.Metadata[key]; ok {
			m[key] = val
		}
	}
	return m
}

func (e *extractor) ExtractString(msg *pubsub.Message) string {
	var sb []string
	for k, v := range e.Extract(msg) {
		sb = append(sb, fmt.Sprintf("%s: %s", k, v))
	}
	return strings.Join(sb, ", ")
}
