package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/omareloui/skrewscore/internal/mongodb"
	"github.com/omareloui/skrewscore/internal/router"
)

func init() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("error loading .env file")
	}
}

func main() {
	mongodb.Init()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		mongodb.Disconnect(ctx)
	}()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", router.Router)
	log.Printf("Skrew scorer running on :%s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
