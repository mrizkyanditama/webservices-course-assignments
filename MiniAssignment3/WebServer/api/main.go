package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func RandomInteger(min int, max int) int {
	// Helper func
	return min + rand.Intn(max-min)
}

func RandomString(n int) string {
	//Helper func
	bytes := make([]byte, n)

	for i := 0; i < n; i++ {
		bytes[i] = byte(RandomInteger(97, 122))
	}

	return string(bytes)
}

func main() {
	router := gin.Default()
	router.LoadHTMLGlob("templates/*")
	router.GET("/upload", func(c *gin.Context) {
		c.HTML(http.StatusOK, "upload.html", gin.H{})
	})

	router.POST("/upload", func(c *gin.Context) {
		// single file
		fileInput, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"status": http.StatusUnprocessableEntity,
				"error":  "Unable to parse requeest",
			})
			return
		}

		// Buffer for temporary saving the ifle
		buf := new(bytes.Buffer)
		writer := multipart.NewWriter(buf)

		// Create multipart form file from file input
		part, err := writer.CreateFormFile("file", fileInput.Filename)
		if err != nil {
			fmt.Println(err)
		}

		// Open file IO
		file, err := fileInput.Open()
		if err != nil {
			fmt.Println(err)
		}

		// Copy content of file to multi part
		if _, err = io.Copy(part, file); err != nil {
			fmt.Println(err)
		}

		writer.Close()

		// Initiate http client and add body from previous buf
		client := &http.Client{}
		req, err := http.NewRequest("POST", "http://localhost:8081/compress", buf)

		// Add request header
		randomString := RandomString(10)
		req.Header.Set("X-ROUTING-KEY", randomString)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		if err != nil {
			fmt.Println(err)
		}

		// Save response
		res, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
		}

		// Read response body
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println(string(body))

		// Check if server 2 succesfully accept the file
		if strings.Contains(string(body), "success upload") {
			c.HTML(http.StatusOK, "progress.html", gin.H{"routingKey": randomString})
		} else {
			c.String(http.StatusOK, fmt.Sprintf("'%s' failed to be upload", fileInput.Filename))
		}

	})
	router.Run(":8001")
}
