package main

import (
	"context"
	"log"

	"github.com/dev3mike/go-api-swagger-boilerplate/cmd/server/setup"
	"github.com/dev3mike/go-api-swagger-boilerplate/internal/database"
	"github.com/dev3mike/go-api-swagger-boilerplate/internal/entities"
)

func main() {
	setup.SetupServerPrerequisites()

	seedProfilesTable()

	// Log the result of the insertion
	log.Printf("Seed data inserted successfully")

	setup.DisconnectDatabase()
}

func seedProfilesTable() {
	// Seed data here
	profile := entities.ProfileEntity{
		PrivateEmail: stringPointer("private@gmail.com"),
		ProfileImage: stringPointer("https://example.com/profile.jpg"),
		Fullname:     stringPointer("John Doe"),
	}
	ctx := context.Background()
	// Insert seed data into the database
	_, err := database.DB.NewInsert().Model(&profile).Exec(ctx)
	if err != nil {
		log.Fatalf("Failed to insert seed data: %v", err)
	}
}

func stringPointer(s string) *string {
	return &s
}
