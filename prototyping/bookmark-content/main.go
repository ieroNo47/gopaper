// Get the html contents of a bookmark and display them in the terminal
package main

import (
	"fmt"
	"log"

	"github.com/ieroNo47/gopaper/prototyping/instapaper"
	"github.com/joho/godotenv"
)

const bookmarkLimit = 1

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	client, err := instapaper.NewClient()
	if err != nil {
		log.Fatalf("Failed to init Instapaper client: %v\n", err)
	}
	bookmarks, err := client.GetBookmarks(bookmarkLimit)
	if err != nil {
		log.Fatalf("Failed to get bookmarks: %v\n", err)
	}
	item := bookmarks[0]
	fmt.Printf("%v\n", item)
}
