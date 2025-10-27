package main

import (
	"fmt"

	"github.com/bmamha/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bmamha/learn-pub-sub-starter/internal/pubsub"
	"github.com/bmamha/learn-pub-sub-starter/internal/routing"
)

func logHandler() func(routing.GameLog) pubsub.AckType {
	return func(gl routing.GameLog) pubsub.AckType {
		defer fmt.Print("> ")
		err := gamelogic.WriteLog(gl)
		if err != nil {
			fmt.Println("Error writing log:", err)
			return pubsub.NackRequeue
		}
		fmt.Println("Game log written to disk.")
		return pubsub.Ack
	}
}
