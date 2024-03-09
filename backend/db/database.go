package db

import (
	"fmt"

	"backend/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	DB *gorm.DB
)

func InitDB() {
	var err error
	dsn := "user=postgres password=empatpuluh6 dbname=postgres sslmode=disable" // Enter your setting parameter of database here
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("Failed to connect to database")
		panic(err)
	}

	// Auto-miigrate the Order model
	DB.AutoMigrate(&models.Order{})
}
