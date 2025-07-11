//	@title						Notes API
//	@version					1.0
//	@description				A simple API for managing notes and categories.
//	@host						localhost:8080
//	@BasePath					/
//	@schemes					http
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
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

	s := api.NewAPIServer(":8080", db)
	s.Run()
}