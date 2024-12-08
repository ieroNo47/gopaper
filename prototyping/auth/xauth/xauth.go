// xAuth functions
package xauth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

func GetToken() (url.Values, error) {
	signingKey := os.Getenv("IP_OAUTH_CONSUMER_SECRET") + "&"
	method := "POST"
	nonce, err := generateNonce(32)
	if err != nil {
		return nil, err
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
		return nil, err
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
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// fmt.Println("Status code: ", resp.StatusCode)
	// fmt.Println("Response: ", string(body))

	tokenValues, err := url.ParseQuery(string(body))
	if err != nil {
		return nil, err
	}
	return tokenValues, nil
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
