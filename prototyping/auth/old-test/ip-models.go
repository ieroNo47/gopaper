// Instapaper models
package main

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
