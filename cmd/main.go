package main

import (
	"fmt"
	"log"

	"github.com/siwiec987/notes-api/internal/api"
	"github.com/siwiec987/notes-api/internal/database"
)

func main() {
	db, err := database.GetDB()
	if err != nil {
		log.Fatal("Unable to connect to database: ", err)
	}
	err = database.DBInit(db)
	if err != nil {
		log.Fatal("Unable to initialize database: ", err)
	}
	fmt.Println("Connected to database")

	api := api.NewAPIServer(":8080", db)
	api.Run()
}