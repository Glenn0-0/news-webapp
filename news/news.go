package news

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"net/http"
	"time"
)

type Article struct {
	Source struct {
		ID   interface{} `json:"id"`
		Name string      `json:"name"`
	} `json:"source"`
	Author      string    `json:"author"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	URLToImage  string    `json:"urlToImage"`
	PublishedAt time.Time `json:"publishedAt"`
	Content     string    `json:"content"`
}

// formats published date from standard time.Time to "January 10, 2009" format.
func (a *Article) FormatPublishedDate() string {
	year, month, day := a.PublishedAt.Date()
	return fmt.Sprintf("%v %d, %d", month, day, year)
}

type Results struct {
	Status       string	   `json:"status"`	
	TotalResults int	   `json:"totalResults"`
	Articles     []Article `json:"articles"`
}

// client for working with the News API.
type Client struct {
	http     *http.Client
	key      string
	PageSize int
}

// returns results (according to given query and current page) and an error.
func (c *Client) FetchEverything(query, page string) (*Results, error) {
	// creating the URL for given query and page number.
	endpoint := fmt.Sprintf("https://newsapi.org/v2/everything?q=%s&pageSize=%d&page=%s&apiKey=%s&sortBy=publishedAt", url.QueryEscape(query), c.PageSize, page, c.key)
	resp, err := c.http.Get(endpoint)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	// converting the response body to []byte. 
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK { //200 OK
		return nil, fmt.Errorf(string(body))
	}

	// decode body into the Result.
	res := &Results{}
	return res, json.Unmarshal(body, res)
}

// returns a new Client instance for making requests to rhe NewsAPi.
func NewClient(httpClient *http.Client, key string, pageSize int) *Client {
	if pageSize > 100 {
		pageSize = 100
	}

	return &Client{httpClient, key, pageSize}
}