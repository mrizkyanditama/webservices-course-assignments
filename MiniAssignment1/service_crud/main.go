package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

var db *gorm.DB
var err error

type Client struct {
	ID       uint   `gorm:"primary_key;auto_increment" json:"id"`
	ClientID string `gorm:"not null"`
	SecretID string `gorm:"not null"`
}

type Token struct {
	ID       uint    `gorm:"primary_key;auto_increment" json:"id"`
	Token    string  `gorm:"not null"`
	Refresh  string  `gorm:"not null"`
	UserID   uint    `gorm:"not null"`
	ClientID *string `gorm:"not null"`
	Expired  time.Time
}

type User struct {
	ID       uint   `gorm:"primary_key;auto_increment" json:"id"`
	Username string `gorm:"not null" json:"username"`
	Password string `gorm:"not null" json:"password"`
	FullName string `gorm:"not null" json:"full_name"`
	NPM      string `gorm:"not null" json:"NPM"`
}

type RequestToken struct {
	Username  *string `json:"username"`
	Password  *string `json:"password"`
	GrantType *string `json:"grant_type"`
	ClientID  *string `json:"client_id"`
	SecretID  *string `json:"client_secret"`
}

func main() {
	db, err = gorm.Open(sqlite.Open("miniassignment1.db"), &gorm.Config{})
	db.AutoMigrate(
		Client{},
		Token{},
		User{},
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
	api := router.Group("/api")
	{
		api.GET("/user", GetAll)
		api.POST("/user", CreateUser)
		api.GET("/user/:id", GetUser)
	}

	// Main api
	oauth := router.Group("/oauth")
	{
		oauth.POST("/token", GetToken)
		oauth.POST("/resource", TokenAuthMiddleware(), GetUserFromToken)
	}

	// Start and run the server
	router.Run(":8002")
}

func TokenValid(r *http.Request) error {
	// Extract token
	tokenString := ExtractToken(r)

	fmt.Println(tokenString)

	token := Token{}
	// Check if token exist
	err = db.Debug().Model(&Token{}).Where("token = ?", tokenString).Take(&token).Error
	if err != nil {
		fmt.Println(err)
		return errors.New("Token is invalid or missing")
	}

	// Check if token not expired yet
	isValid := time.Now().Local().Before(token.Expired)
	fmt.Println(isValid)

	if !isValid {
		fmt.Println(err)
		return errors.New("Token is expired")
	}

	return nil
}

func ExtractToken(r *http.Request) string {
	// Extract token from header request
	keys := r.URL.Query()
	token := keys.Get("token")
	if token != "" {
		return token
	}
	bearerToken := r.Header.Get("Authorization")
	if len(strings.Split(bearerToken, " ")) == 2 {
		return strings.Split(bearerToken, " ")[1]
	}
	return ""
}

func TokenAuthMiddleware() gin.HandlerFunc {
	// If use this middleware, each request must pass the check from this middleware, or else return error
	return func(c *gin.Context) {
		err := TokenValid(c.Request)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":             "invalid_token",
				"error_description": "Token Salah masbro",
			})
			c.Abort()
			return
		}
		c.Next()
	}
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

func CheckRequest(r *RequestToken) error {
	// Helper to check request
	if r.ClientID != nil && r.GrantType != nil && r.Password != nil && r.SecretID != nil && r.Username != nil {
		return nil
	}
	return errors.New("Request is not valid")
}

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

func GetToken(c *gin.Context) {
	// Parse request
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid_request", "Error_description": "ada kesalahan masbro!"})
		return
	}

	requestToken := RequestToken{}
	err = json.Unmarshal(body, &requestToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid_request", "Error_description": "ada kesalahan masbro!"})
		return
	}

	// Validate request
	err = CheckRequest(&requestToken)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid_request", "Error_description": "ada kesalahan masbro!"})
		return
	}

	client := Client{}

	// Check if clientID and clientSecret combination exist
	err = db.Debug().Model(&Client{}).Where("client_id = ? AND secret_id = ?", requestToken.ClientID, requestToken.SecretID).Take(&client).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"error":  "Sorry, no client ID and secret ID combination found",
		})
		return
	}

	// Check if username and password combination exist
	user := User{}
	err = db.Debug().Model(&User{}).Where("username = ? AND password = ?", requestToken.Username, requestToken.Password).Take(&user).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"error":  "Sorry, username or/and password is incorrect",
		})
		return
	}

	// Generate new token
	tokenStr := RandomString(40)
	expired := time.Now().Local().Add(time.Minute * 5)
	refreshToken := RandomString(40)

	newToken := Token{UserID: user.ID, Token: tokenStr, Expired: expired, Refresh: refreshToken, ClientID: requestToken.ClientID}

	err = db.Debug().Model(&Token{}).Create(&newToken).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"error":  "Sorry there has been an error when generating token",
		})
		return
	}

	// Return appropriate response
	c.JSON(http.StatusOK, gin.H{
		"access_token":  newToken.Token,
		"expires_in":    300,
		"token_type":    "Bearer",
		"scope":         nil,
		"refresh_token": refreshToken,
	})

}

func GetUserFromToken(c *gin.Context) {
	tokenString := ExtractToken(c.Request)

	token := Token{}
	// Check if token exist
	err = db.Debug().Model(&Token{}).Where("token = ?", tokenString).Take(&token).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"error":  "Sorry there has been an internal server error",
		})
		return
	}

	// Search for user from token
	user := User{}
	err = db.Debug().Model(&User{}).Where("id = ?", token.UserID).Take(&user).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"error":  "Sorry there has been an internal server error",
		})
		return
	}

	// Return appropriate response
	c.JSON(http.StatusOK, gin.H{
		"access_token":  token.Token,
		"client_id":     token.ClientID,
		"user_id":       user.ID,
		"full_name":     user.FullName,
		"npm":           user.NPM,
		"expires":       nil,
		"refresh_token": token.Refresh,
	})
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
	userID := c.Param("id")
	tid, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"error":  "Sorry there has been an internal server error",
		})
		return
	}

	user := User{}
	err = db.Debug().Model(&User{}).Where("id = ?", tid).Take(&user).Error
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

	newUser := User{}
	err = json.Unmarshal(body, &newUser)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": http.StatusUnprocessableEntity,
			"error":  "Unable to parse requeest",
		})
		return
	}

	err = db.Debug().Model(&User{}).Create(&newUser).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"error":  "Sorry there has been an internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   http.StatusOK,
		"response": newUser,
	})
}
