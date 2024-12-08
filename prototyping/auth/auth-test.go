// Instapaper API tests
package main

// ENV VARS:
// IP_USER
// IP_PASSWORD
// IP_API
// IP_API_VERSION
// IP_ACCESS_TOKEN_ENDPOINT
// IP_OAUTH_CONSUMER_ID
// IP_OAUTH_CONSUMER_SECRET

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/ieroNo47/gopaper/prototyping/instapaper"
	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("Hey!")

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	if err := testBasicAuth(); err != nil {
		log.Fatalf("BASIC AUTH FAILED: %v\n", err)
	}
	if err := testXAuth(); err != nil {
		log.Fatalf("XAUTH FAILED: %v\n", err)
	}

}

func testXAuth() error {
	fmt.Println("== Testing xAuth ==")
	// https://web.archive.org/web/20130308050830/https://dev.twitter.com/docs/oauth/xauth
	fmt.Println(">> Get oAuth token <<")
	client, err := instapaper.NewClient()
	if err != nil {
		return err
	}
	titleLimit := 5
	titles, err := client.GetBookmarkTitles(titleLimit)
	if err != nil {
		return err
	}
	fmt.Printf("%d titles:\n", titleLimit)
	for _, title := range titles {
		fmt.Printf("\t- %s\n", title)
	}
	return nil
}

func testBasicAuth() error {
	fmt.Println("== Testing Basic Auth ==")
	client := &http.Client{}
	req, _ := http.NewRequest("GET", os.Getenv("IP_API")+"/authenticate", nil)

	req.SetBasicAuth(os.Getenv("IP_USER"), os.Getenv("IP_PASSWORD"))

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error: Failed to verify authentication: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Status: %d\nBody: %s\n", resp.StatusCode, string(body))
	if resp.StatusCode == 200 {
		fmt.Println("Result: Successfully Authenticated")
	} else {
		return fmt.Errorf("result: Failed to authenticate. Check user and password")
	}
	return nil
}
