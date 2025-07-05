package helpers

import (
	"connection/models"
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetContactInfoHelper(contacts []models.Contact) ([]models.ContactInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var result []models.ContactInfo

	for _, contact := range contacts {
		if contact.ContactNo == nil {
			log.Printf("Skipping contact with nil ContactNo")
			continue
		}

		// Validate phone number format (basic validation)
		if *contact.ContactNo == "" {
			log.Printf("Skipping contact with empty phone number")
			continue
		}

		// Find user by phone (matches contact.ContactNo)
		filter := bson.M{"phone": *contact.ContactNo}
		log.Printf("Searching for user with phone: %s", *contact.ContactNo)

		var user struct {
			User_id    *string `bson:"user_id"`
			First_Name *string `bson:"First_Name"`
			Last_Name  *string `bson:"Last_Name"`
		}

		err := userCollection.FindOne(ctx, filter).Decode(&user)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				log.Printf("No user found for phone: %s", *contact.ContactNo)
				continue
			}
			log.Printf("Error querying user for phone %s: %v", *contact.ContactNo, err)
			return nil, fmt.Errorf("database error while querying user: %w", err)
		}

		var fullName string
		if user.First_Name != nil {
			fullName += *user.First_Name
		}
		if user.Last_Name != nil {
			if fullName != "" {
				fullName += " "
			}
			fullName += *user.Last_Name
		}

		contactInfo := models.ContactInfo{
			Name:      contact.Name,
			ContactNo: contact.ContactNo,
			Uid:       user.User_id,
			UserName:  &fullName,
		}

		result = append(result, contactInfo)
		log.Printf("Successfully processed contact: %s -> User: %s", *contact.ContactNo, fullName)
	}

	log.Printf("Processed %d contacts, found %d users", len(contacts), len(result))
	return result, nil
}
