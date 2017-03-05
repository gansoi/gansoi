package slack

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gansoi/gansoi/plugins"
)

type (
	// Slack will notify Slack using an incoming webhook.
	// Please see: https://api.slack.com/incoming-webhooks
	Slack struct {
		Username string `json:"username"`
		URL      string `json:"url"`
		Channel  string `json:"channel"`
	}

	attachment struct {
		Fallback  string `json:"fallback,omitempty"`
		Title     string `json:"title,omitempty"`
		Text      string `json:"text,omitempty"`
		Footer    string `json:"footer,omitempty"`
		Timestamp int64  `json:"ts,omitempty"`
	}

	message struct {
		Username    string       `json:"username,omitempty"`
		Channel     string       `json:"channel,omitempty"`
		IconEmoji   string       `json:"icon_emoji,omitempty"`
		Attachments []attachment `json:"attachments,omitempty"`
	}
)

func init() {
	plugins.RegisterNotifier("slack", Slack{})
}

// Notify implement plugins.Notifier
func (s *Slack) Notify(text string) error {
	msg := &message{
		Username:  s.Username,
		Channel:   s.Channel,
		IconEmoji: ":alien:",
		Attachments: []attachment{
			{
				Fallback:  text,
				Title:     text,
				Footer:    "Gansoi",
				Timestamp: time.Now().Unix(),
			},
		},
	}

	// This should never fail.
	b, _ := json.Marshal(msg)

	resp, err := http.Post(s.URL, "application/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
