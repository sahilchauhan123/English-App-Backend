package notifications

import (
	"context"
	"fmt"
	"log"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

func InitializeAppWithServiceAccount() *firebase.App {
	// [START initialize_app_service_account_golang]
	opt := option.WithCredentialsFile("strango.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}
	// [END initialize_app_service_account_golang]

	return app
}

func SendToToken(app *firebase.App) {
	ctx := context.Background()
	client, err := app.Messaging(ctx)
	if err != nil {
		log.Fatalf("error getting Messaging client: %v\n", err)
	}

	registrationToken := "eXi_K_sGTByU-R59syK86I:APA91bEpBaCA2JBtRDoiP1NalTDjHxb-N9WHgHhYTPsofgW7ERU3wJhuBkVmkP7GTYZxg3SaFausqpGpTyQYblKqqgiADxChNdI-ibGR1Lmi1Tqu55LO3LI"

	message := &messaging.Message{
		Token: registrationToken,

		// ✅ 1. Add both notification + data
		Notification: &messaging.Notification{
			Title: "Upcoming Lecture 🔔",
			Body:  "Your next class starts at 2:45 PM in Room 101.",
		},
		Data: map[string]string{
			"score": "850",
			"time":  "2:45",
			"title": "Upcoming Lecture 🔔",
			"body":  "Your next class starts at 2:45 PM in Room 101.",
		},

		// ✅ 2. Android config for heads-up behavior
		Android: &messaging.AndroidConfig{
			Priority: "high", // 🔥 critical for heads-up
			Notification: &messaging.AndroidNotification{
				ChannelID: "default", // must match your Notifee channel id
				Sound:     "default",
				Color:     "#2196F3", // optional accent color
			},
		},
	}

	response, err := client.Send(ctx, message)
	if err != nil {
		log.Fatalf("Error sending message: %v\n", err)
	}
	fmt.Println("✅ Successfully sent message:", response)
}
