package controllers

import (
	"connection/database"
	"connection/helpers"
	"connection/models"
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	// "golang.org/x/net/idna"
)

var tripCollection *mongo.Collection = database.OpenCollection(database.Client, "trips")
var linkedMemberCollection *mongo.Collection = database.OpenCollection(database.Client, "LinkedMembers")
var transactionCollection *mongo.Collection = database.OpenCollection(database.Client, "transaction")

// var userCollection *mongo.Collection =database.OpenCollection(database.Client,"user")
// Always add creator in members
// Members can be added with invite code
func CreateTrip() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("Starting CreateTrip handler")
		// 1. Create a context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		fmt.Println("Binding JSON request")
		// 2. Bind incoming JSON into your Trip struct
		var trip models.Trip
		if err := c.BindJSON(&trip); err != nil {
			fmt.Println("Error binding JSON:", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
			return
		}

		// Validate required fields
		if trip.Name == nil {
			fmt.Println("Trip name is required")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Trip name is required"})
			return
		}

		fmt.Println("Getting user ID from context")
		// 3. Extract the authenticated user's UID from the Gin context
		creatorID := c.GetString("uid")
		if creatorID == "" {
			fmt.Println("No user ID found in context")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Could not find user ID in context"})
			return
		}

		// Get user's first and last name from context
		firstName := c.GetString("first_name")
		lastName := c.GetString("last_name")
		if firstName == "" || lastName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User's name information is missing"})
			return
		}

		// Create member name in format firstname_lastname
		memberName := firstName + "_" + lastName

		fmt.Println("Creating new trip document")
		trip.ID = primitive.NewObjectID()
		trip_id := trip.ID.Hex()
		trip.Trip_ID = &trip_id
		//name given in json
		//description given in json

		// Initialize members array if it's nil
		if trip.Members == nil {
			trip.Members = &[]string{}
		}

		// Add creator as first member
		*trip.Members = append(*trip.Members, memberName)

		trip.Creator_ID = &creatorID

		// Create invite code safely
		if trip.Name != nil && trip.Trip_ID != nil {
			invite_code := *trip.Name + *trip.Trip_ID
			trip.Invite_Code = &invite_code
		} else {
			fmt.Println("Error: Cannot create invite code - name or trip_id is nil")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create trip: Invalid trip data"})
			return
		}

		trip.Created_At = time.Now()

		fmt.Println("Inserting trip into database")
		// 7. Insert the fully populated `trip` into MongoDB:
		insertResult, err := tripCollection.InsertOne(ctx, trip)
		if err != nil {
			fmt.Println("Error inserting trip:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create trip: " + err.Error()})
			return
		}

		// Create link for the creator
		linkMember := models.Member{
			ID:      primitive.NewObjectID(),
			Trip_ID: trip.Trip_ID,
			Name:    &memberName,
			Uid:     &creatorID,
		}
		_, err = linkedMemberCollection.InsertOne(ctx, linkMember)
		if err != nil {
			// If linking fails, we should probably delete the trip
			_, _ = tripCollection.DeleteOne(ctx, bson.M{"_id": trip.ID})
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to link creator as member: " + err.Error()})
			return
		}

		fmt.Println("Trip created successfully")
		// 8. Return success with the new trip's ID
		c.JSON(http.StatusCreated, gin.H{
			"message": "Trip created successfully",
			"tripID":  insertResult.InsertedID,
		})
	}
}

func GetAllTrip() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helpers.CheckUserType(c, "ADMIN"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
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
		result, err := tripCollection.Aggregate(ctx, mongo.Pipeline{
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

func GetAllMyTrip() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		uid := c.GetString("uid")

		matchStage := bson.D{
			{"$match", bson.D{
				{"creator_id", uid},
			}},
		}

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
		// matchStage := bson.D{{Key: "$match", Value: bson.D{{}}}}

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
		result, err := tripCollection.Aggregate(ctx, mongo.Pipeline{
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

func GetAllNotFreeMemberOnInviteCode() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// Get invite code from request body
		var requestBody struct {
			InviteCode string `json:"invite_code" binding:"required"`
		}
		if err := c.BindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
			return
		}

		// Create match stage for aggregation
		matchStage := bson.D{
			{"$match", bson.D{
				{"invite_code", requestBody.InviteCode},
			}},
		}

		// Execute aggregation
		cursor, err := tripCollection.Aggregate(ctx, mongo.Pipeline{matchStage})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No trip with this invite code " + err.Error()})
			return
		}

		// Decode results
		var results []bson.M
		if err := cursor.All(ctx, &results); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if len(results) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Trip not found with the given invite code"})
			return
		}

		tripDoc := results[0]
		rawMembers, ok := tripDoc["members"].(bson.A)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid members format"})
			return
		}

		var members []string
		for _, m := range rawMembers {
			if name, ok := m.(string); ok {
				members = append(members, name)
			}
		}

		// Get trip ID
		tripID := tripDoc["trip_id"].(string)

		// Get free and non-free members
		free, notFree := helpers.GetAllFreeMembers(tripID, members)

		c.JSON(http.StatusOK, gin.H{
			"trip_id":          tripID,
			"trip_name":        tripDoc["trip_name"],
			"free_members":     free,
			"not_free_members": notFree,
			"total_members":    len(members),
			"total_free":       len(free),
			"total_not_free":   len(notFree),
		})
	}
}

func LinkMember() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// Step 1: Bind request JSON
		var requestBody struct {
			InviteCode string `json:"invite_code" binding:"required"`
			MemberName string `json:"name" binding:"required"`
		}
		if err := c.BindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
			return
		}

		// Step 2: Get user ID from context
		uid := c.GetString("uid")
		if uid == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// Step 3: Find trip by invite code
		var trip models.Trip
		err := tripCollection.FindOne(ctx, bson.M{"invite_code": requestBody.InviteCode}).Decode(&trip)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{"error": "No trip found with this invite code"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error finding trip: " + err.Error()})
			}
			return
		}

		// Step 4: Check if member exists in trip members
		memberExists := false
		if trip.Members != nil {
			for _, member := range *trip.Members {
				if member == requestBody.MemberName {
					memberExists = true
					break
				}
			}
		}
		if !memberExists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Member not found in trip members"})
			return
		}

		// Step 5: Check if member is already linked
		var existingLink models.Member
		err = linkedMemberCollection.FindOne(ctx, bson.M{
			"trip_id": trip.Trip_ID,
			"name":    requestBody.MemberName,
		}).Decode(&existingLink)
		if err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Member is already linked"})
			return
		} else if err != mongo.ErrNoDocuments {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking existing link: " + err.Error()})
			return
		}

		// Step 5.1: Check if member is linked with any username in this trip
		var existingMemberLink models.Member
		err = linkedMemberCollection.FindOne(ctx, bson.M{
			"trip_id": trip.Trip_ID,
			"uid":     uid,
		}).Decode(&existingMemberLink)
		if err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "You have already linked with another member in this trip"})
			return
		} else if err != mongo.ErrNoDocuments {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking existing member link: " + err.Error()})
			return
		}

		// Step 6: Create and insert new link
		linkMember := models.Member{
			ID:      primitive.NewObjectID(),
			Trip_ID: trip.Trip_ID,
			Name:    &requestBody.MemberName,
			Uid:     &uid,
		}
		_, err = linkedMemberCollection.InsertOne(ctx, linkMember)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert linked member: " + err.Error()})
			return
		}

		// Step 7: Return success with updated member status
		free, notFree := helpers.GetAllFreeMembers(*trip.Trip_ID, *trip.Members)
		c.JSON(http.StatusOK, gin.H{
			"message":          "Member linked successfully",
			"trip_id":          trip.Trip_ID,
			"trip_name":        trip.Name,
			"free_members":     free,
			"not_free_members": notFree,
			"total_members":    len(*trip.Members),
			"total_free":       len(free),
			"total_not_free":   len(notFree),
		})
	}
}

func Pay() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// Step 1: Bind request JSON
		var trans models.Transaction
		if err := c.BindJSON(&trans); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
			return
		}

		// Step 2: Validate required fields
		if trans.Trip_ID == nil || trans.PayerName == nil || trans.ReciverName == nil || trans.Amount == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields: trip_id, payer_name,amount and reciever_name are required"})
			return
		}

		// Step 3: Check if trip exists
		var trip models.Trip
		err := tripCollection.FindOne(ctx, bson.M{"trip_id": *trans.Trip_ID}).Decode(&trip)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{"error": "Trip not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error finding trip: " + err.Error()})
			}
			return
		}

		// Step 4: Check if payer and receiver are members of the trip
		if trip.Members != nil {
			payerFound := false
			receiverFound := false
			for _, member := range *trip.Members {
				if member == *trans.PayerName {
					payerFound = true
				}
				if member == *trans.ReciverName {
					receiverFound = true
				}
			}
			if !payerFound || !receiverFound {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Payer or receiver is not a member of this trip"})
				return
			}
		}

		// Step 5: Create transaction record
		trans.ID = primitive.NewObjectID()
		trans.Created_At = time.Now()
		Type := "Paid"
		trans.Type = &Type
		if trans.Description == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Can;t have payment without description"})
			return
		}

		resultNumber, err := transactionCollection.InsertOne(ctx, trans)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record transaction: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":        "Transaction recorded successfully",
			"transaction_id": resultNumber.InsertedID,
			"transaction":    trans,
		})
	}
}
func Settle() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// Step 1: Bind request JSON
		var trans models.Transaction
		if err := c.BindJSON(&trans); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
			return
		}

		// Step 2: Validate required fields
		if trans.Trip_ID == nil || trans.PayerName == nil || trans.ReciverName == nil || trans.Amount == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields: trip_id, payer_name,amount and reciever_name are required"})
			return
		}

		// Step 3: Check if trip exists
		var trip models.Trip
		err := tripCollection.FindOne(ctx, bson.M{"trip_id": *trans.Trip_ID}).Decode(&trip)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{"error": "Trip not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error finding trip: " + err.Error()})
			}
			return
		}

		// Step 4: Check if payer and receiver are members of the trip
		if trip.Members != nil {
			payerFound := false
			receiverFound := false
			for _, member := range *trip.Members {
				if member == *trans.PayerName {
					payerFound = true
				}
				if member == *trans.ReciverName {
					receiverFound = true
				}
			}
			if !payerFound || !receiverFound {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Payer or receiver is not a member of this trip"})
				return
			}
		}

		// Step 5: Create transaction record
		trans.ID = primitive.NewObjectID()
		trans.Created_At = time.Now()
		Type := "Settle"
		trans.Type = &Type
		// if trans.Description==nil{
		// 	c.JSON(http.StatusBadRequest,gin.H{"error":"Can;t have payment without description"})
		// 	return
		// }

		resultNumber, err := transactionCollection.InsertOne(ctx, trans)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record transaction: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":        "Transaction recorded successfully",
			"transaction_id": resultNumber.InsertedID,
			"transaction":    trans,
		})
	}
}

func GetAllTransaction() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// Step 1: Bind request JSON
		var requestBody struct {
			TripId string `json:"trip_id" binding:"required"`
		}
		if err := c.BindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
			return
		}

		// Step 2: Get user ID from context
		uid := c.GetString("uid")
		if uid == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// Create match stage to filter by trip_id
		matchStage := bson.D{
			{"$match", bson.D{
				{"trip_id", requestBody.TripId},
			}},
		}

		// Execute aggregation to get all transactions
		result, err := transactionCollection.Aggregate(ctx, mongo.Pipeline{matchStage})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching transactions: " + err.Error()})
			return
		}

		// Decode results
		var transactions []bson.M
		if err = result.All(ctx, &transactions); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding results: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"total_count":  len(transactions),
			"transactions": transactions,
		})
	}
}

func GetSettlements() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// Step 1: Bind request JSON
		var requestBody struct {
			TripId string `json:"trip_id" binding:"required"`
		}
		if err := c.BindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
			return
		}

		// Step 2: Get all transactions for the trip
		matchStage := bson.D{
			{"$match", bson.D{
				{"trip_id", requestBody.TripId},
			}},
		}

		cursor, err := transactionCollection.Aggregate(ctx, mongo.Pipeline{matchStage})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching transactions: " + err.Error()})
			return
		}

		var transactions []models.Transaction
		if err := cursor.All(ctx, &transactions); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding transactions: " + err.Error()})
			return
		}

		// Step 3: Calculate settlements
		settlements := helpers.CalculateSettlements(transactions)

		c.JSON(http.StatusOK, gin.H{
			"settlements": settlements,
		})
	}
}
