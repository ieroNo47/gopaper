// Instapaper API tests
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("Hey!")

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	testBasicAuth()
	testXAuth()

}

func testXAuth() error {
	fmt.Println("== Testing XAuth ==")
	return nil
}

func testBasicAuth() error {
	fmt.Println("== Testing Basic Auth ==")
	client := &http.Client{}
	req, _ := http.NewRequest("GET", os.Getenv("IP_API")+"/authenticate", nil)

	req.SetBasicAuth(os.Getenv("IP_USER"), os.Getenv("IP_PASSWORD"))

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Error: Failed to verify authentication: %v\n", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Status: %d\nBody: %s\n", resp.StatusCode, string(body))
	if resp.StatusCode == 200 {
		fmt.Println("Result: Successfully Authenticated")
	} else {
		return fmt.Errorf("Result: Failed to authenticate. Check user and password.")
	}
	return nil
}
