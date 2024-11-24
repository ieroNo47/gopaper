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
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dghubble/oauth1"
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
	signingKey := os.Getenv("IP_OAUTH_CONSUMER_SECRET") + "&"
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
		"oauth_nonce":            nonce,
		"oauth_signature_method": "HMAC-SHA1",
		"oauth_timestamp":        strconv.Itoa(int(time.Now().Unix())),
		"oauth_version":          "1.0",
		"x_auth_mode":            "client_auth",
		"x_auth_password":        os.Getenv("IP_PASSWORD"),
		"x_auth_username":        os.Getenv("IP_USER"),
	}
	encParameters := url.QueryEscape(urlEncodeParameters(parameters))
	signatureBaseString := fmt.Sprintf("%s&%s&%s", method, encAccessTokenURL, encParameters)
	parameters["oauth_signature"] = generateSignature(signingKey, signatureBaseString)
	authorizationHeader := getAuthorizationHeader(parameters)
	// fmt.Printf("%s\n%s\n%s\n", signatureBaseString, parameters["oauth_signature"], authorizationHeader)
	values := url.Values{}
	values.Set("x_auth_mode", "client_auth")
	values.Set("x_auth_password", os.Getenv("IP_PASSWORD"))
	values.Set("x_auth_username", os.Getenv("IP_USER"))
	req, err := http.NewRequest("POST", accessTokenURL, strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", authorizationHeader)
	// requestDump, err := httputil.DumpRequest(req, true)
	// if err != nil {
	// 	fmt.Println("Error dumping request:", err)
	// 	return err
	// }
	// fmt.Println(string(requestDump))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println("Status code: ", resp.StatusCode)
	fmt.Println("Response: ", string(body))

	fmt.Println(">> Make authorized API request <<")
	tokenValues, err := url.ParseQuery(string(body))
	if err != nil {
		return err
	}

	tokenSecret := tokenValues.Get("oauth_token_secret")
	token := tokenValues.Get("oauth_token")
	config := oauth1.NewConfig(os.Getenv("IP_OAUTH_CONSUMER_ID"), os.Getenv("IP_OAUTH_CONSUMER_SECRET"))
	otoken := oauth1.NewToken(token, tokenSecret)
	httpClient := config.Client(oauth1.NoContext, otoken)
	titleLimit := 5
	bookmarksURL := fmt.Sprintf("%s/%s/%s",
		os.Getenv("IP_API"),
		os.Getenv("IP_API_VERSION"),
		"bookmarks/list")
	values = url.Values{}
	values.Add("limit", strconv.Itoa(titleLimit))
	// req, err = http.NewRequest("POST", bookmarksURL, strings.NewReader(values.Encode()))
	// if err != nil {
	// 	return err
	// }
	// resp, err = httpClient.Do(req)
	// if err != nil {
	// 	return err
	// }
	resp, err = httpClient.Post(bookmarksURL,
		"application/x-www-form-urlencoded",
		strings.NewReader(values.Encode()),
	)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Println("Status code: ", resp.StatusCode)
	// body, err = io.ReadAll(resp.Body)
	// if err != nil {
	// 	return err
	// }
	// fmt.Println("Response body: ", string(body))
	if resp.StatusCode == 200 {
		// parse json response
		var response Response
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			return err
		}
		titles := response.GetBookmarkTitles()
		fmt.Printf("%d titles for user '%s':\n", titleLimit, response.User.Username)
		for _, title := range titles {
			fmt.Printf("\t- %s\n", title)
		}
	} else {
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		fmt.Println("Response body: ", string(body))
		return fmt.Errorf("failed to make authorized request: %s", string(body))
	}
	return nil
}

func getAuthorizationHeader(parameters map[string]string) string {
	headerValue := "OAuth "
	headerParams := []string{}
	headerParams = append(headerParams,
		"oauth_consumer_key"+"="+parameters["oauth_consumer_key"],
		"oauth_nonce"+"="+parameters["oauth_nonce"],
		"oauth_signature_method"+"="+parameters["oauth_signature_method"],
		"oauth_timestamp"+"="+parameters["oauth_timestamp"],
		"oauth_signature"+"="+parameters["oauth_signature"],
		"oauth_version"+"="+parameters["oauth_version"],
	)
	headerValue += strings.Join(headerParams, ",")
	return headerValue
}

func generateSignature(signingKey string, data string) string {
	h := hmac.New(sha1.New, []byte(signingKey))
	h.Write([]byte(data))
	hash := h.Sum(nil)
	return base64.StdEncoding.EncodeToString(hash)
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
