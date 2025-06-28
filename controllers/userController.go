package controllers

import (
	"connection/database"
	"connection/helpers"
	"connection/models"
	"context"
	"math/rand"
	"strconv"

	// "fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// Here we get a instanve in mongoDb for use
var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")
var otpCollection *mongo.Collection = database.OpenCollection(database.Client, "otp")
var validate = validator.New()

func HashPassword(password string) string {

	hashedpassword, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
		// return
	}
	return string(hashedpassword)
}
func VerifyPassword(userPassword string, foundUserPassword string) (passwordIsvalid bool, msg string) {

	// passwordIsvalid = false
	// msg = ""
	// err := bcrypt.CompareHashAndPassword([]byte(foundUserPassword), []byte(userPassword))
	// if err != nil {
	// 	msg = "email and password is incorrect "
	// 	return passwordIsvalid, msg
	// }
	// passwordIsvalid = true
	// return passwordIsvalid, "password and email is correct "

	if err := bcrypt.CompareHashAndPassword([]byte(foundUserPassword), []byte(userPassword)); err != nil {
		return false, "email and password is incorrect "
	}
	return true, "password and email is correct "
}

// here we bind whole user but we use only email and password so only this two are required
func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User
		var foundUser models.User
		// here the context c store all the user struct data which is present in body (simply) jo json hm bhejte hai vo store krta
		//hai bind json and idhr user structure jaise store krega
		// here from c we bind all the json in user
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
			return
		}
		// Here we find the user with the email as user email
		//here we find user and bind it with founduser
		// bson.M map the key and value it just find the req item in database
		err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
			}
			return
		}

		if foundUser.Email == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}

		// here in user we have a verify password
		passwordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
		if !passwordIsValid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": msg})
			return
		}
		//here we generate refreshtoken and token
		token, refreshToken := helpers.GenerateAllTokens(*foundUser.Email, *foundUser.First_Name, *foundUser.Last_Name, *foundUser.User_type, *foundUser.User_id)

		// Update tokens in database
		helpers.UpdateAllTokens(token, refreshToken, *foundUser.User_id)

		// Fetch updated user data
		err = userCollection.FindOne(ctx, bson.M{"user_id": foundUser.User_id}).Decode(&foundUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching updated user data: " + err.Error()})
			return
		}
		// the response you will receive after succefull login
		c.JSON(http.StatusOK, gin.H{
			"message":       "Login successful",
			"user":          foundUser,
			"token":         token,
			"refresh_token": refreshToken,
		})
	}
}
func Signup() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User
		//bind everything to user
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
			return
		}
		//validate that everything with validation is present or not
		validationErr := validate.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Validation error: " + validationErr.Error()})
			return
		}

		// count the no of time user email is present in database
		count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error while checking email: " + err.Error()})
			return
		}

		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "This email is already registered"})
			return
		}

		// password was sent as it is so we hashes the password for safety
		password := HashPassword(*user.Password)
		user.Password = &password

		// Check if phone exists
		count, err = userCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error while checking phone: " + err.Error()})
			return
		}

		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "This phone number is already registered"})
			return
		}

		// Set user metadata
		user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		//get a new object as id
		user.ID = primitive.NewObjectID()
		//set that userid in hex for uid
		uid := user.ID.Hex()
		user.User_id = &uid

		// Generate tokens
		token, refreshToken := helpers.GenerateAllTokens(*user.Email, *user.First_Name, *user.Last_Name, *user.User_type, *user.User_id)
		user.Token = &token
		user.Refresh_token = &refreshToken

		// Insert user it give back a insertion no just plane syntax
		resultInsertionNumber, inserterr := userCollection.InsertOne(ctx, user)
		if inserterr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user: " + inserterr.Error()})
			return
		}
		// mess we sent back to front end
		c.JSON(http.StatusCreated, gin.H{
			"message": "User created successfully",
			"user_id": resultInsertionNumber.InsertedID,
		})
	}
}

// To learn go mongo db aggragrate function
func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {

		//only admin used api so admin permission to out
		if err := helpers.CheckUserType(c, "ADMIN"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// Parse pagination parameters
		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}

		page, err := strconv.Atoi(c.Query("page"))
		if err != nil || page < 1 {
			page = 1
		}

		startIndex := (page - 1) * recordPerPage

		// Create aggregation pipeline
		//its like a filter to make where
		// here match means where
		matchStage := bson.D{{Key: "$match", Value: bson.D{{}}}}

		groupStage := bson.D{
			{
				Key: "$group",
				Value: bson.D{
					{Key: "_id", Value: "null"},
					{Key: "total_count", Value: bson.D{{Key: "$sum", Value: 1}}},
					{Key: "data", Value: bson.D{{Key: "$push", Value: "$$ROOT"}}},
				},
			},
		}
		projectStage := bson.D{
			{
				Key: "$project",
				Value: bson.D{
					{Key: "_id", Value: 0},
					{Key: "total_count", Value: 1},
					{Key: "user_items", Value: bson.D{
						{Key: "$slice", Value: bson.A{"$data", startIndex, recordPerPage}},
					}},
				},
			},
		}

		// Execute aggregation
		result, err := userCollection.Aggregate(ctx, mongo.Pipeline{
			matchStage, groupStage, projectStage,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching users: " + err.Error()})
			return
		}

		// Decode results
		var allUsers []bson.M
		if err = result.All(ctx, &allUsers); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding results: " + err.Error()})
			return
		}

		if len(allUsers) == 0 {
			c.JSON(http.StatusOK, gin.H{
				"total_count": 0,
				"user_items":  []bson.M{},
			})
			return
		}

		c.JSON(http.StatusOK, allUsers[0])
	}
}
func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {

		//If your route is defined as /users/:user_id, then c.Param("user_id") grabs whatever value was in that slot of the URL.
		userId := c.Param("user_id")

		if err := helpers.MatchUserTypeToUid(c, userId); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var user models.User
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		err := userCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&user)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error})
		}
		c.JSON(http.StatusOK, user)

	}

}

func GetOTP() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var otpRequest models.OTPRequest

		// Bind JSON request
		if err := c.BindJSON(&otpRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
			return
		}

		// Validate email
		validationErr := validate.Struct(otpRequest)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Validation error: " + validationErr.Error()})
			return
		}

		// Check if user exists
		var foundUser models.User
		err := userCollection.FindOne(ctx, bson.M{"email": otpRequest.Email}).Decode(&foundUser)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found with this email"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
			}
			return
		}

		// Generate 6-digit OTP
		rand.Seed(time.Now().UnixNano())
		otp := strconv.Itoa(rand.Intn(900000) + 100000) // Generates number between 100000-999999

		// Create OTP record
		otpRecord := models.OTP{
			ID:        primitive.NewObjectID(),
			Email:     otpRequest.Email,
			OTP:       otp,
			ExpiresAt: time.Now().Add(10 * time.Minute), // OTP expires in 10 minutes
			CreatedAt: time.Now(),
			Used:      false,
		}

		// Delete any existing OTP for this email
		_, err = otpCollection.DeleteMany(ctx, bson.M{"email": otpRequest.Email})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error clearing existing OTP: " + err.Error()})
			return
		}

		// Insert new OTP
		_, err = otpCollection.InsertOne(ctx, otpRecord)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving OTP: " + err.Error()})
			return
		}

		// Send OTP via email
		err = helpers.SendOTPEmail(otpRequest.Email, otp)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error sending OTP email: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "OTP sent successfully to your email",
			"email":   otpRequest.Email,
		})
	}
}

func VerifyOTP() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var otpVerification models.OTPVerification

		// Bind JSON request
		if err := c.BindJSON(&otpVerification); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
			return
		}

		// Validate request
		validationErr := validate.Struct(otpVerification)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Validation error: " + validationErr.Error()})
			return
		}

		// Find OTP record
		var otpRecord models.OTP
		err := otpCollection.FindOne(ctx, bson.M{
			"email": otpVerification.Email,
			"otp":   otpVerification.OTP,
			"used":  false,
		}).Decode(&otpRecord)

		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid OTP or OTP already used"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
			}
			return
		}

		// Check if OTP is expired
		if time.Now().After(otpRecord.ExpiresAt) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "OTP has expired"})
			return
		}

		// Mark OTP as used
		_, err = otpCollection.UpdateOne(
			ctx,
			bson.M{"_id": otpRecord.ID},
			bson.M{"$set": bson.M{"used": true}},
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating OTP status: " + err.Error()})
			return
		}

		// Get user data
		var foundUser models.User
		err = userCollection.FindOne(ctx, bson.M{"email": otpVerification.Email}).Decode(&foundUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching user data: " + err.Error()})
			return
		}

		// Generate tokens
		token, refreshToken := helpers.GenerateAllTokens(*foundUser.Email, *foundUser.First_Name, *foundUser.Last_Name, *foundUser.User_type, *foundUser.User_id)

		// Update tokens in database
		helpers.UpdateAllTokens(token, refreshToken, *foundUser.User_id)

		// Fetch updated user data
		err = userCollection.FindOne(ctx, bson.M{"user_id": foundUser.User_id}).Decode(&foundUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching updated user data: " + err.Error()})
			return
		}

		// Return success response with tokens (same as login)
		c.JSON(http.StatusOK, gin.H{
			"message":       "OTP verified successfully",
			"user":          foundUser,
			"token":         token,
			"refresh_token": refreshToken,
		})
	}
}
