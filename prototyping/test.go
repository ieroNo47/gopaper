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
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("Hey!")

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// if err := testBasicAuth(); err != nil {
	// 	log.Fatal(err)
	// }
	if err := testXAuth(); err != nil {
		log.Fatal(err)
	}

}

func testXAuth() error {
	fmt.Println("== Testing XAuth ==")
	// https://web.archive.org/web/20130308050830/https://dev.twitter.com/docs/oauth/xauth
	// enc := url.QueryEscape(os.Getenv("IP_API"))
	// fmt.Printf("%s: %s\n", os.Getenv("IP_API"), enc)
	// signingKey := os.Getenv("IP_OAUTH_CONSUMER_SECRET") + "&"
	// h := hmac.New(sha1.New, []byte(signingKey))
	// h.Write([]byte(enc))
	// hash := h.Sum(nil)
	// bhash := base64.StdEncoding.EncodeToString(hash)
	// fmt.Print(bhash)
	method := "POST"
	nonce, err := generateNonce(32)
	if err != nil {
		return err
	}
	accessTokenURL := fmt.Sprintf("%s/%s/%s",
		os.Getenv("IP_API"),
		os.Getenv("IP_API_VERSION"),
		os.Getenv("IP_ACCESS_TOKEN_ENDPOINT"))
	encAccessTokenURL := url.QueryEscape(accessTokenURL)
	parameters := map[string]string{
		"oauth_consumer_key":     os.Getenv("IP_OAUTH_CONSUMER_ID"),
		"oauth_consumer_secret":  os.Getenv("IP_OAUTH_CONSUMER_SECRET"),
		"oauth_nonce":            nonce,
		"oauth_signature_method": "HMAC-SHA1",
		"oauth_timestamp":        string(time.Now().Unix()),
		"oauth_version":          "1.0",
		"x_auth_mode":            "client_auth",
		"x_auth_password":        os.Getenv("IP_USER"),
		"x_auth_username":        os.Getenv("IP_PASSWORD"),
	}
	encParameters := url.QueryEscape(urlEncodeParameters(parameters))
	signatureBaseString := fmt.Sprintf("%s&%s&%s", method, encAccessTokenURL, encParameters)
	fmt.Printf("%s\n", signatureBaseString)
	// TODO: create body
	// sign
	// create oauth header string
	// make request
	return nil
}

func urlEncodeParameters(parameters map[string]string) string {
	encParams := []string{}
	keys := make([]string, 0, len(parameters))
	for k := range parameters {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		encKey := url.QueryEscape(k)
		encValue := url.QueryEscape(parameters[k])
		encParams = append(encParams, fmt.Sprintf("%s=%s", encKey, encValue))
	}
	return strings.Join(encParams, "&")
}

func generateNonce(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
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
