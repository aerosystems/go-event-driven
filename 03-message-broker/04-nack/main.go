package main

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
)

type AlarmClient interface {
	StartAlarm() error
	StopAlarm() error
}

func ConsumeMessages(sub message.Subscriber, alarmClient AlarmClient) {
	messages, err := sub.Subscribe(context.Background(), "smoke_sensor")
	if err != nil {
		panic(err)
	}

	for msg := range messages {
		switch string(msg.Payload) {
		case "1":
			if err := alarmClient.StartAlarm(); err != nil {
				msg.Nack()
			}
		case "0":
			if err := alarmClient.StopAlarm(); err != nil {
				msg.Nack()
			}
		default:
			msg.Nack()
		}
		msg.Ack()
	}
}
