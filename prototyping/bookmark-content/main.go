// Get the html contents of a bookmark and display them in the terminal
package main

import (
	"fmt"
	"log"

	"github.com/charmbracelet/glamour"
	"github.com/ieroNo47/gopaper/prototyping/instapaper"
	"github.com/joho/godotenv"
	"jaytaylor.com/html2text"
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
	text, err := client.GetBookmarkText(item.BookmarkID)
	if err != nil {
		log.Fatalf("Failed to get bookrmark text: %v\n", err)
	}
	// fmt.Printf("Text: %s", text)
	t, err := html2text.FromString(text)
	if err != nil {
		log.Fatalf("Failed to convert html to text: %v\n", err)
	}
	// fmt.Println(t)
	out, err := glamour.Render(t, "dark")
	fmt.Print(out)
}
