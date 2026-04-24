package main

import (
	"log"
	"mini-cloud/cmd"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}
	cmd.Execute()
}
