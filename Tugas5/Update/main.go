package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
		api.POST("/update", CreateUser)
		api.POST("/update/:npm", CreateTransaction)
	}

	// Start and run the server
	router.Run(":8001")
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

func CreateUser(c *gin.Context) {
	// Help to create user
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": http.StatusUnprocessableEntity,
			"error":  "Unable to parse requeest",
		})
		return
	}

	user := User{}
	err = json.Unmarshal(body, &user)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": http.StatusUnprocessableEntity,
			"error":  "Unable to parse requeest",
		})
		return
	}

	oldUser := User{}
	err = db.Debug().Model(&User{}).Where("npm = ?", user.NPM).Take(&oldUser).Error
	if err != nil {
		// If haven't existed yet, then create one
		err = db.Debug().Model(&User{}).Create(&user).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": http.StatusInternalServerError,
				"error":  "Sorry there has been an internal server error",
			})
			return
		}
	} else {
		// Else update the user
		err = db.Debug().Model(&User{}).Where("npm = ?", oldUser.NPM).Update("name", user.Name).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": http.StatusInternalServerError,
				"error":  "Sorry there has been an internal server error",
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   http.StatusOK,
		"response": user,
	})
}

func CreateTransaction(c *gin.Context) {
	// Help to create user
	userNPM := c.Param("npm")

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": http.StatusUnprocessableEntity,
			"error":  "Unable to parse requeest",
		})
		return
	}

	transaction := Transaction{}
	err = json.Unmarshal(body, &transaction)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": http.StatusUnprocessableEntity,
			"error":  "Unable to parse requeest",
		})
		return
	}

	oldTransaction := Transaction{}
	user := User{}
	err = db.Debug().Model(&User{}).Where("npm = ?", userNPM).Take(&user).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"error":  "Sorry there has been an internal server error",
		})
		return
	}

	err = db.Debug().Model(&Transaction{}).Where("user_id = ? AND id = ?", user.ID, transaction.ID).Take(&oldTransaction).Error
	if err != nil {
		transaction.UserID = user.ID
		err = db.Debug().Model(&Transaction{}).Create(&transaction).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": http.StatusInternalServerError,
				"error":  "Sorry there has been an internal server error",
			})
			return
		}
	} else {
		err = db.Debug().Model(&Transaction{}).Where("user_id = ? AND id = ?", user.ID, transaction.ID).Update("name", transaction.Name).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": http.StatusInternalServerError,
				"error":  "Sorry there has been an internal server error",
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   http.StatusOK,
		"response": transaction,
	})
}
