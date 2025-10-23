package main

import (
	"fmt"

	"github.com/bmamha/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bmamha/learn-pub-sub-starter/internal/pubsub"
	"github.com/bmamha/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")
	rabbitConnectionServer := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(rabbitConnectionServer)
	if err != nil {
		fmt.Printf("Failed to connect to RabbitMQ: %v\n", err)
	}
	defer conn.Close()
	fmt.Println("Connected to RabbitMQ successfully.")

	ch, err := conn.Channel()
	if err != nil {
		fmt.Printf("Failed to open a channel: %v\n", err)
	}

	defer ch.Close()
	_, queue, err := pubsub.DeclareAndBind(conn, routing.ExchangePerilTopic, routing.GameLogSlug, "game_logs.*", pubsub.DurableQueue)
	if err != nil {
		fmt.Printf("Failed to open a channel: %v\n", err)
	}
	fmt.Printf("Queue Name: %s\n", queue.Name)
	gamelogic.PrintServerHelp()
	for {
		words := gamelogic.GetInput()
		if words == nil {
			continue
		}
		switch firstWord := words[0]; firstWord {
		case "pause":
			fmt.Println("Sending a Pause Message...")
			pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})
		case "resume":
			fmt.Println("Sending a Resume Message...")
			pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: false})
		case "quit":
			fmt.Println("Quitting...")
			return
		default:
			fmt.Println("Unknown command. Possible commands are: pause, resume, quit, help")
		}

	}
}
