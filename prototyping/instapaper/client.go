// Instapaper models
package instapaper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dghubble/oauth1"
	"github.com/ieroNo47/gopaper/prototyping/auth/xauth"
)

const (
	defaultTimeout  = 10 * time.Second
	contentType     = "application/x-www-form-urlencoded"
	bookmarksList   = "bookmarks/list"
	bookarmsGetText = "bookmarks/get_text"
)

type Response struct {
	Highlights []Highlight `json:"highlights"` // empty array
	Bookmarks  []Bookmark  `json:"bookmarks"`
	User       User        `json:"user"`
}

type Highlight struct {
	HighlightID int64       `json:"highlight_id"`
	Text        string      `json:"text"`
	Note        interface{} `json:"note"` // using interface{} since it can be null
	BookmarkID  int64       `json:"bookmark_id"`
	Time        int64       `json:"time"`
	Position    int         `json:"position"`
	Type        string      `json:"type"`
}

type Bookmark struct {
	Hash              string  `json:"hash"`
	Description       string  `json:"description"`
	Tags              []Tag   `json:"tags"`
	BookmarkID        int64   `json:"bookmark_id"`
	PrivateSource     string  `json:"private_source"`
	Title             string  `json:"title"`
	URL               string  `json:"url"`
	ProgressTimestamp int64   `json:"progress_timestamp"`
	Time              int64   `json:"time"`
	Progress          float64 `json:"progress"`
	Starred           string  `json:"starred"`
	Type              string  `json:"type"`
}

type Tag struct {
	Count int     `json:"count"`
	Hash  string  `json:"hash"`
	Name  string  `json:"name"`
	ID    int     `json:"id"`
	Time  float64 `json:"time"`
	Slug  string  `json:"slug"`
}

type User struct {
	Username             string `json:"username"`
	UserID               int    `json:"user_id"`
	Type                 string `json:"type"`
	SubscriptionIsActive string `json:"subscription_is_active"`
}

func (r Response) GetBookmarkTitles() []string {
	titles := []string{}
	for _, bookmark := range r.Bookmarks {
		titles = append(titles, bookmark.Title)
	}
	return titles
}

func (r Response) GetBookmarks() []Bookmark {
	return r.Bookmarks
}

// Client
type Client struct {
	httpClient *http.Client
	apiVersion string
	baseURL    string
}

func NewClient() (Client, error) {
	tokenValues, err := xauth.GetToken()
	if err != nil {
		return Client{}, fmt.Errorf("failed to get token: %w", err)
	}

	tokenSecret := tokenValues.Get("oauth_token_secret")
	token := tokenValues.Get("oauth_token")
	config := oauth1.NewConfig(
		os.Getenv("IP_OAUTH_CONSUMER_ID"),
		os.Getenv("IP_OAUTH_CONSUMER_SECRET"))
	otoken := oauth1.NewToken(token, tokenSecret)
	httpClient := config.Client(oauth1.NoContext, otoken)
	httpClient.Timeout = defaultTimeout
	return Client{httpClient: httpClient,
		apiVersion: os.Getenv("IP_API_VERSION"),
		baseURL:    os.Getenv("IP_API")}, nil
}

func (c Client) GetBookmarks(limit int) ([]Bookmark, error) {
	bookmarksURL := fmt.Sprintf("%s/%s/%s",
		c.baseURL,
		c.apiVersion,
		bookmarksList)
	values := url.Values{}
	values.Add("limit", strconv.Itoa(limit))

	resp, err := c.httpClient.Post(bookmarksURL,
		contentType,
		strings.NewReader(values.Encode()),
	)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get bookmarks list: %s", string(body))
	}

	// parse json response
	var response Response
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}
	return response.GetBookmarks(), nil

}

func (c Client) GetBookmarkTitles(limit int) ([]string, error) {
	bookmarks, err := c.GetBookmarks(limit)
	if err != nil {
		return nil, err
	}
	titles := []string{}
	for _, bookmark := range bookmarks {
		titles = append(titles, bookmark.Title)
	}
	return titles, nil
}
