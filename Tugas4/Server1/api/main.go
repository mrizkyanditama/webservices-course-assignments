package main

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.LoadHTMLGlob("templates/*")
	router.GET("/upload", func(c *gin.Context) {
		c.HTML(http.StatusOK, "upload.html", gin.H{})
	})

	router.POST("/upload", func(c *gin.Context) {
		// single file
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"status": http.StatusUnprocessableEntity,
				"error":  "Unable to parse requeest",
			})
			return
		}

		// Open file
		src, err := file.Open()
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"status": http.StatusUnprocessableEntity,
				"error":  "Unable to open file",
			})
			return

		}
		defer src.Close()

		// Get parent directory
		wd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		parent := filepath.Dir(filepath.Dir(wd))

		// Storage path is in parent directory with name 'storage'
		storagePath := filepath.Join(parent, "storage")

		// If no directory with that name, create one then
		if _, err := os.Stat(storagePath); os.IsNotExist(err) {
			os.Mkdir(storagePath, 0755)
		}

		fileName := file.Filename

		dst, err := os.Create(filepath.Join(storagePath, filepath.Base(fileName))) // dir is directory where you want to save file.
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"status": http.StatusUnprocessableEntity,
				"error":  "Unable to save file",
			})
			return
		}

		defer dst.Close()
		if _, err = io.Copy(dst, src); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": http.StatusInternalServerError,
				"error":  "There was a problem when saving file",
			})
			return
		}

		fileUrl := "http://localhost/download/" + fileName
		c.HTML(http.StatusOK, "download.html", gin.H{"fileName": fileName, "fileUrl": fileUrl})

	})
	router.Run(":8001")
}
