package main

import (
	"fmt"
	"time"

	"github.com/bmamha/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bmamha/learn-pub-sub-starter/internal/pubsub"
	"github.com/bmamha/learn-pub-sub-starter/internal/routing"
)

type logPublisher func(exchange, routingKey string, body routing.GameLog) error

func warHandler(gs *gamelogic.GameState, pub logPublisher, username string) func(gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(rw gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")
		outcome, winner, loser := gs.HandleWar(rw)

		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeYouWon:
			err := pub(routing.ExchangePerilTopic,
				routing.GameLogSlug+"."+username,
				routing.GameLog{CurrentTime: time.Now(), Message: fmt.Sprintf("%s won a war against %s.", winner, loser), Username: username})
			if err != nil {
				fmt.Println("Error publishing war log:", err)
				return pubsub.NackRequeue
			}

			return pubsub.Ack
		case gamelogic.WarOutcomeOpponentWon:
			err := pub(routing.ExchangePerilTopic,
				routing.GameLogSlug+"."+username,
				routing.GameLog{CurrentTime: time.Now(), Message: fmt.Sprintf("%s won a war against %s.", winner, loser), Username: username})
			if err != nil {
				fmt.Println("Error publishing war log:", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeDraw:
			err := pub(routing.ExchangePerilTopic,
				routing.GameLogSlug+"."+username,
				routing.GameLog{CurrentTime: time.Now(), Message: fmt.Sprintf("A war between %s and %s resulted in a draw", winner, loser), Username: username})
			if err != nil {
				fmt.Println("Error publishing war log:", err)
				return pubsub.NackRequeue
			}

			return pubsub.Ack
		default:
			fmt.Println("Unknown error")
			return pubsub.Ack
		}
	}
}
