package main

import (
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	// Serve static file
	router.Use(static.Serve("/download", static.LocalFile("../../storage", false)))
	router.Run(":8002")
}
