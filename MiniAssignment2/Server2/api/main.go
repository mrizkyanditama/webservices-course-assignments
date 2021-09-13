package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/machinebox/progress"
	"github.com/streadway/amqp"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Message struct {
	Content string `json:"content"`
	Status  string `json:"status"`
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	router := gin.Default()

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"progress",  // name
		"direct", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	// Serve static file
	router.Use(static.Serve("/compressed", static.LocalFile("./", false)))

	// Handler
	router.POST("/compress", func(c *gin.Context) {
		// Retrieve file from request
		file, err := c.FormFile("file")
		if err != nil {
			fmt.Println(c.Request.Body)
			fmt.Println(err)
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"status": http.StatusUnprocessableEntity,
				"error":  "Unable to parse requeest",
			})
			return
		}

		// Get routing key
		routingKey := c.Request.Header.Get("X-ROUTING-KEY")

		if routingKey == "" {
			fmt.Println("Routing key is empty")
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"status": http.StatusUnprocessableEntity,
				"error":  "Request header X-ROUTING-KEY is empty",
			})
			return
		}

		// Create new goroutines so it does its task on background
		go func() {
			// Open file to be read
			fileContent, err := file.Open()
			if err != nil {
				fmt.Println(err)
				c.JSON(http.StatusUnprocessableEntity, gin.H{
					"status": http.StatusUnprocessableEntity,
					"error":  "Unable to open file",
				})
				return
			}

			// Read bytes of file
			fileBytes, err := ioutil.ReadAll(fileContent)
			if err != nil {
				fmt.Println(err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"message": "Unable to convert to bytes",
				})
				return
			}

			// Declare buffer for result of compressed file
			var b bytes.Buffer
			size := len(fileBytes)
			w := gzip.NewWriter(&b)

			// Use progress library to track progress change
			r := progress.NewWriter(w)

			// Get context background
			ctx := context.Background()

			// Start a goroutine again, asynchronously update everytime progress of compress changes
			go func() {
				// Declare new ticker
				progressChan := progress.NewTicker(ctx, r, int64(size), 30*time.Millisecond)

				// Wait for ticker update and publish to exchange with routing key from previous
				for _ = range progressChan {
					progresCurrent := (b.Len() * 100 / len(fileBytes) )
					payload := Message{Content: string(strconv.Itoa(progresCurrent)), Status: "In Progress"}
					payloadStr, err := json.Marshal(payload)
					err = ch.Publish(
						"progress", // exchange
						routingKey,      // routing key
						false,   // mandatory
						false,   // immediate
						amqp.Publishing{
							ContentType: "text/plain",
							Body:        []byte(payloadStr),
						})
					failOnError(err, "Failed to publish a message")
				}
				payload := Message{Content: "Saving file", Status: "Saving file"}
				payloadStr, err := json.Marshal(payload)
				err = ch.Publish(
					"progress", // exchange
					routingKey,      // routing key
					false,   // mandatory
					false,   // immediate
					amqp.Publishing{
						ContentType: "text/plain",
						Body:        []byte(payloadStr),
					})
				failOnError(err, "Failed to publish a message")
				fmt.Println("\rcompress is completed")
			}()

			// Write to buffer
			r.Write(fileBytes)
			w.Close()

			// Save to disk
			err = ioutil.WriteFile(file.Filename+".gz", b.Bytes(), 0666)
			if err != nil {
				fmt.Println(err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"message": "Unable to compress the file",
				})
				return
			}
			payload := Message{Content: "/compressed" + file.Filename + ".gz", Status: "File saved"}
			payloadStr, err := json.Marshal(payload)
			err = ch.Publish(
				"progress", // exchange
				routingKey,      // routing key
				false,   // mandatory
				false,   // immediate
				amqp.Publishing{
					ContentType: "text/plain",
					Body:        []byte(payloadStr),
				})
			failOnError(err, "Failed to publish a message")
		}()

		c.JSON(http.StatusOK, gin.H{
			"status": "success upload",
		})

	})
	router.Run(":8081")
}
