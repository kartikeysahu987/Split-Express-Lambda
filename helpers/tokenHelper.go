package helpers

import (
	// "bytes"
	"connection/database"
	"context"
	"errors"
	"log"
	"os"
	"time"

	// "github.com/dgrijalva/jwt-go"
	jwt "github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SignedDetails struct {
	Email      string
	First_name string
	Last_Name  string
	Uid        string
	User_type  string
	jwt.StandardClaims
}

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

var SECRET_KEY string

func init() {
	SECRET_KEY = os.Getenv("SECRET_KEY")
	if SECRET_KEY == "" {
		log.Fatal("SECRET_KEY environment variable is not set")
	}
}

// to understand
func ValidateToken(clientToken string) (claims *SignedDetails, msg string) {

	// Step 1: If no token was given, tell the user it's required
	if clientToken == "" {
		return nil, "token is required"
	}

	// Step 2: Try to read (parse) the token and pull the data (claims) from it
	token, err := jwt.ParseWithClaims(
		clientToken,       // the token string sent by client (like in headers)
		&SignedDetails{},  // where we want to store the data from the token
		func(token *jwt.Token) (interface{}, error) {
			// this function gives back the secret key used to verify the token
			return []byte(SECRET_KEY), nil
		},
	)

	// Step 3: If there's any error while reading token (maybe tampered, expired etc.)
	if err != nil {
		return nil, "invalid token: " + err.Error()
	}

	// Step 4: Try to convert the token's data into our SignedDetails structure
	claims, ok := token.Claims.(*SignedDetails)

	// Step 5: If conversion fails OR the token is invalid
	if !ok || !token.Valid {
		return nil, "invalid token claims"
	}

	// Step 6: Everything went fine, return the claims and empty message (no error)
	return claims, ""
}



func GenerateAllTokens(email, first_name, last_name, user_type, user_id string) (signedToken string, signedRefreshToken string) {
	claims := &SignedDetails{
		Email:      email,
		First_name: first_name,
		Last_Name:  last_name,
		Uid:        user_id,
		User_type:  user_type,

		//standard syntax 
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add( 720 * time.Hour).Unix(), // token expires in 1 day
			// IssuedAt:  time.Now().Unix(),
		},
	}

	refreshClaims := &SignedDetails{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(168 * time.Hour).Unix(), // refresh token expires in 7 days
			// IssuedAt:  time.Now().Unix(),
		},
	}
	// syntax we ues HS256 method and get it signed with our secretkey to generat security token 
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		log.Printf("Error generating token: %v", err)
		return "", ""
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		log.Printf("Error generating refresh token: %v", err)
		return "", ""
	}

	return token, refreshToken
}

func UpdateAllTokens(signedToken string, signedRefreshToken string, userId string) error {
	if signedToken == "" || signedRefreshToken == "" || userId == "" {
		return errors.New("invalid token data")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var updateObj primitive.D
	// bson.E puut the value as key value pair in json like structure 
	updateObj = append(updateObj, bson.E{"token", signedToken})
	updateObj = append(updateObj, bson.E{"refresh_token", signedRefreshToken})
	//letting upsert := true means "if no existing document matches filter, create a new one with these values."
	upsert := true
	//Here, you're saying: "Find the document where the user_id field equals the value of userId."
	filter := bson.M{"user_id": userId}

	opt := options.UpdateOptions{
		Upsert: &upsert,
	}
	// here we are updating the mongodb with item where user id is same
	// bson d uis used to update the value in key
	_, err := userCollection.UpdateOne(
		ctx, filter, bson.D{
			{"$set", updateObj},
		},
		&opt,
	)
	if err != nil {
		log.Printf("Error updating tokens: %v", err)
		return err
	}

	return nil
}
