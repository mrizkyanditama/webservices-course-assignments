package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
)

type Message struct {
	Content string `json:"message"`
	Client  string `json:"client"`
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"chats",  // name
		"fanout", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(
		q.Name,  // queue name
		"",      // routing key
		"chats", // exchange
		false,
		nil,
	)
	failOnError(err, "Failed to bind a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	// Declare ticker for every n interval
	ticker := time.NewTicker(60 * time.Second)

	go func() {
		// Accept message on channels if received
		for {
			select {
			// Accept subscribed RabbitMQ exchange message if there is any
			case d := <-msgs:
				// For debugging and logging
				log.Printf(" [x] %s", d.Body)
			// Accept message every n interval
			case t := <-ticker.C:
				fmt.Println("Tick at", t)
				// Create proper payload
				contentStr := "Current time is " + time.Now().Format(time.RFC850)
				clientStr := "BOT"
				payload := Message{Content: contentStr, Client: clientStr}
				payloadStr, err := json.Marshal(payload)
				// Publish to exchange
				err = ch.Publish(
					"chats", // exchange
					"",      // routing key
					false,   // mandatory
					false,   // immediate
					amqp.Publishing{
						ContentType: "text/plain",
						Body:        []byte(payloadStr),
					})
				failOnError(err, "Failed to publish a message")
			}
		}
	}()

	log.Printf(" [*] Waiting for logs. To exit press CTRL+C")
	<-forever
}
