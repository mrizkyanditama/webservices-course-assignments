package main

import (
	"fmt"
	"log"
	"net/http"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

var db *gorm.DB
var err error

type User struct {
	ID   uint   `gorm:"primary_key;auto_increment" json:"id"`
	Name string `gorm:"not null" json:"name"`
	NPM  string `gorm:"not null" json:"NPM"`
}

type Transaction struct {
	ID     uint   `gorm:"primary_key;auto_increment" json:"id"`
	UserID uint   `gorm:"not null"`
	Name   string `gorm:"not null" json:"name"`
}

func main() {
	db, err = gorm.Open(sqlite.Open("../assignment5.db"), &gorm.Config{})
	db.AutoMigrate(
		User{},
		Transaction{},
	)
	if err != nil {
		fmt.Printf("Cannot connect to mini assignment 1 database")
		log.Fatal("This is the error:", err)
	} else {
		fmt.Println("We are connected to the mini assignment 1 database")
	}

	router := gin.Default()

	router.Use(CORSMiddleware())

	// Used for checking and debugging
	api := router.Group("/")
	{
		api.GET("/read", GetAll)
		api.GET("/read/:npm", GetUser)
		api.GET("/read/:npm/:idTrx", GetUserTrx)
	}

	// Start and run the server
	router.Run(":8002")
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func GetAll(c *gin.Context) {
	// Help to check user
	var err error
	users := []User{}
	err = db.Debug().Model(&User{}).Find(&users).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"error":  "Sorry there has been an internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   http.StatusOK,
		"response": users,
	})
}

func GetUser(c *gin.Context) {
	// Help to check user
	userNPM := c.Param("npm")

	user := User{}
	err = db.Debug().Model(&User{}).Where("npm = ?", userNPM).Take(&user).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"error":  "Sorry there has been an internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   http.StatusOK,
		"response": user,
	})
}

func GetUserTrx(c *gin.Context) {
	// Help to check user
	userNPM := c.Param("npm")
	trxId := c.Param("idTrx")

	user := User{}
	err = db.Debug().Model(&User{}).Where("npm = ?", userNPM).Take(&user).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"error":  "Sorry there has been an internal server error",
		})
		return
	}

	transaction := Transaction{}
	err = db.Debug().Model(&Transaction{}).Where("user_id = ? AND id = ?", user.ID, trxId).Take(&transaction).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"error":  "Sorry there has been an internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   http.StatusOK,
		"response": transaction,
	})
}
