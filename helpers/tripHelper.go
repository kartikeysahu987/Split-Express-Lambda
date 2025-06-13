package helpers

import (
	"connection/database"
	"connection/models"
	"context"
	"fmt"

	// "net/http"
	"strconv"
	"time"

	// "github.com/gin-gonic/gin"
	"math"
	"sort"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var linkedMemberCollection *mongo.Collection = database.OpenCollection(database.Client, "LinkedMembers")

func GetAllFreeMembers(Trip_Id string, Members []string) (FreeMembers []string, NotFree []string) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	// Find all linked members for this trip
	matchStage := bson.D{
		{"$match", bson.D{
			{"trip_id", Trip_Id},
		}},
	}

	cursor, err := linkedMemberCollection.Aggregate(ctx, mongo.Pipeline{matchStage})
	if err != nil {
		fmt.Printf("Error finding linked members: %v\n", err)
		return Members, []string{} // If error, consider all members as free
	}

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		fmt.Printf("Error decoding linked members: %v\n", err)
		return Members, []string{} // If error, consider all members as free
	}

	// Create a map of linked member names
	linked := make(map[string]bool)
	for _, doc := range results {
		if name, ok := doc["name"].(string); ok {
			linked[name] = true
		}
	}

	// Separate members into free and not free
	var freeMembers []string
	var notFree []string
	for _, m := range Members {
		if !linked[m] {
			freeMembers = append(freeMembers, m)
		} else {
			notFree = append(notFree, m)
		}
	}

	return freeMembers, notFree
}

type Settlement struct {
	From   string  `json:"from"`
	To     string  `json:"to"`
	Amount float64 `json:"amount"`
}

func CalculateSettlements(transactions []models.Transaction) []Settlement {
	// Create a map to store net balances for each person
	balances := make(map[string]float64)

	// Calculate net balance for each person
	for _, t := range transactions {
		// Skip if any required field is nil
		if t.PayerName == nil || t.ReciverName == nil || t.Amount == nil || t.Type == nil {
			continue
		}

		amount, err := strconv.ParseFloat(*t.Amount, 64)
		if err != nil {
			continue // Skip if amount can't be parsed
		}

		if *t.Type == "Paid" {
			balances[*t.PayerName] -= amount
			balances[*t.ReciverName] += amount
		}
	}

	// Separate debtors and creditors
	var debtors []struct {
		name   string
		amount float64
	}
	var creditors []struct {
		name   string
		amount float64
	}

	for name, amount := range balances {
		if amount < -0.01 { // Only consider significant debts
			debtors = append(debtors, struct {
				name   string
				amount float64
			}{name, -amount})
		} else if amount > 0.01 { // Only consider significant credits
			creditors = append(creditors, struct {
				name   string
				amount float64
			}{name, amount})
		}
	}

	// Sort debtors and creditors by amount
	sort.Slice(debtors, func(i, j int) bool {
		return debtors[i].amount > debtors[j].amount
	})
	sort.Slice(creditors, func(i, j int) bool {
		return creditors[i].amount > creditors[j].amount
	})

	var settlements []Settlement
	debtorIdx := 0
	creditorIdx := 0

	// Match debtors with creditors
	for debtorIdx < len(debtors) && creditorIdx < len(creditors) {
		debtor := debtors[debtorIdx]
		creditor := creditors[creditorIdx]

		amount := math.Min(debtor.amount, creditor.amount)
		if amount > 0.01 { // Only create settlement if amount is significant
			settlements = append(settlements, Settlement{
				From:   debtor.name,
				To:     creditor.name,
				Amount: amount,
			})
		}

		debtor.amount -= amount
		creditor.amount -= amount

		if debtor.amount < 0.01 {
			debtorIdx++
		}
		if creditor.amount < 0.01 {
			creditorIdx++
		}
	}

	return settlements
}
