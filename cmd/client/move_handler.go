package main

import (
	"fmt"

	gamelogic "github.com/bmamha/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bmamha/learn-pub-sub-starter/internal/pubsub"
	"github.com/bmamha/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerMove(gs *gamelogic.GameState, publishCh *amqp.Channel) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(am gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		moveOutcome := gs.HandleMove(am)
		switch moveOutcome {
		case gamelogic.MoveOutcomeNotInvolved:
			return pubsub.NackRequeue
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			rk := routing.WarRecognitionsPrefix + "." + gs.GetUsername()
			if err := pubsub.PublishJSON(publishCh, routing.ExchangePerilTopic,
				rk,
				gamelogic.RecognitionOfWar{Attacker: am.Player, Defender: gs.GetPlayerSnap()}); err != nil {
				return pubsub.NackRequeue
			}

			return pubsub.Ack
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.NackDiscard
		default:
			return pubsub.NackDiscard
		}
	}
}
