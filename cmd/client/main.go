package main

import (
	"fmt"
	"log"

	"github.com/bmamha/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bmamha/learn-pub-sub-starter/internal/pubsub"
	"github.com/bmamha/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril client...")
	rabbitConnectionServer := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(rabbitConnectionServer)
	if err != nil {
		fmt.Printf("Failed to connect to RabbitMQ: %v\n", err)
	}

	defer conn.Close()

	user, err := gamelogic.ClientWelcome()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	pauseQueue := routing.PauseKey + "." + user
	gameState := gamelogic.NewGameState(user)
	err = pubsub.Subscribe(conn,
		routing.ExchangePerilDirect,
		pauseQueue,
		routing.PauseKey,
		pubsub.TransientQueue,
		handlerPause(gameState),
		pubsub.Unmarshal)
	if err != nil {
		log.Fatalf("Failed to subscribe to pause messages: %v", err)
	}

	pub := func(exchange, routingKey string, body any) error {
		ch, err := conn.Channel()
		if err != nil {
			return err
		}
		return pubsub.PublishJSON(ch, exchange, routingKey, body)
	}

	// Client detects army moves
	err = pubsub.Subscribe(conn,
		routing.ExchangePerilTopic,
		"army_moves."+user,
		"army_moves.*",
		pubsub.TransientQueue,
		handlerMove(gameState, pub, user),
		pubsub.Unmarshal)
	if err != nil {
		log.Fatalf("Failed to declare and bind army moves queue: %v", err)
	}

	pubLog := func(exchange, routingKey string, logBody routing.GameLog) error {
		ch, err := conn.Channel()
		if err != nil {
			return err
		}
		return pubsub.PublishGob(ch, exchange, routingKey, logBody)
	}

	// Client recognizes wars
	err = pubsub.Subscribe(conn,
		routing.ExchangePerilTopic,
		routing.WarRecognitionsPrefix,
		routing.WarRecognitionsPrefix+".*",
		pubsub.DurableQueue,
		warHandler(gameState, pubLog, user),
		pubsub.Unmarshal)
	if err != nil {
		log.Fatalf("Failed to declare and bind war recognitions queue: %v", err)
	}

	for {

		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		switch words[0] {
		case "spawn":
			err = gameState.CommandSpawn(words)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
		case "move":
			am, err := gameState.CommandMove(words)
			if err != nil {
				fmt.Println(err)
				continue
			}
			ch, err := conn.Channel()
			if err != nil {
				fmt.Printf("Failed to open a channel: %v\n", err)
				continue
			}
			pubsub.PublishJSON(ch, routing.ExchangePerilTopic, "army_moves."+user, am)

		case "status":
			gameState.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed")
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Println("Unknown command. Possible commands are: spawn, move, status, help, spam, quit")
		}
	}
}
