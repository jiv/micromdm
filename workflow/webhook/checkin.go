package webhook

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/micromdm/micromdm/mdm/checkin"
	"github.com/micromdm/micromdm/platform/pubsub"
	"github.com/pkg/errors"
)

type CheckinWebhook struct {
	Topic       string
	CallbackURL string
	HTTPClient  *http.Client
}

func NewCheckinWebhook(httpClient *http.Client, topic, callbackURL string) (*CheckinWebhook, error) {
	if topic == "" {
		return nil, errors.New("webhook: topic should not be empty")
	}

	if callbackURL == "" {
		return nil, errors.New("webhook: callbackURL should not be empty")
	}

	return &CheckinWebhook{HTTPClient: httpClient, Topic: topic, CallbackURL: callbackURL}, nil
}

func (cw CheckinWebhook) StartListener(sub pubsub.Subscriber) error {
	checkinEvents, err := sub.Subscribe(context.TODO(), "checkinWebhook", cw.Topic)

	if err != nil {
		return errors.Wrapf(err,
			"subscribing checkinWebhook to %s topic", cw.Topic)
	}

	go func() {
		for {
			select {
			case event := <-checkinEvents:
				var ev checkin.Event
				if err := checkin.UnmarshalEvent(event.Message, &ev); err != nil {
					fmt.Println(err)
					continue
				}

				_, err := cw.HTTPClient.Post(cw.CallbackURL, "application/x-apple-aspen-mdm", bytes.NewBuffer([]byte(ev.Command.UDID)))
				if err != nil {
					fmt.Printf("error sending command response: %s\n", err)
				}
			}
		}
	}()

	return nil
}
