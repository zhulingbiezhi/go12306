package slack

import (
	"context"
	"fmt"
	"github.com/nlopes/slack"
)

func Notification(ctx context.Context) error {
	api := slack.New("YOUR_TOKEN_HERE")
	attachment := slack.Attachment{
		Pretext: "some pretext",
		Text:    "some text",
		// Uncomment the following part to send a field too
		/*
			Fields: []slack.AttachmentField{
				slack.AttachmentField{
					Title: "a",
					Value: "no",
				},
			},
		*/
	}
	channelID, timestamp, resp, err := api.SendMessage("", slack.MsgOptionText("Some text", false), slack.MsgOptionAttachments(attachment))
	if err != nil {
		fmt.Printf("%s\n", err)
		return nil
	}
	fmt.Printf("Message successfully sent to channel %s at %s, resp: %s", channelID, timestamp, resp)
	return nil
}
