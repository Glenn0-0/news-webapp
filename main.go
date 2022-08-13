package main

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
	"time"

	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	
	"github.com/freshman-tech/news-demo-starter-files/news"
	"github.com/joho/godotenv"
)

type Search struct {
	Query      string
	NextPage   int
	TotalPages int
	Results    *news.Results
}

// determines if the last page was reached.
func (s *Search) IsLastPage() bool {
	return s.NextPage >= s.TotalPages
}

// returns previous page number.
func (s *Search) PreviousPage() int {
	return s.CurrentPage() - 1
}

// returns current page number.
func (s *Search) CurrentPage() int {
	if s.NextPage == 1 {
		return s.NextPage
	}

	return s.NextPage - 1
}

// temlate to parse and validate .html file.
var tpl = template.Must(template.ParseFiles("index.html"))

func main() {
	// handling errors opening .env file.
	if err := godotenv.Load(); err != nil {
		log.Println("Error loading .env file")
	}

	// getting PORT from .env.
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	// getting API key from .env.
	apiKey := os.Getenv("NEWS_API_KEY")
	if apiKey == "" {
		log.Fatal("Env: API key must be set")
	}

	myClient := &http.Client{Timeout: 10 * time.Second}
	newsapi := news.NewClient(myClient, apiKey, 20)

	// initiallize a file server object by passing the directory with all static files.
	fs := http.FileServer(http.Dir("assets"))

	mux := http.NewServeMux()
	mux.Handle("/assets/", http.StripPrefix("/assets", fs)) // connecting CSS files from assets.
	mux.HandleFunc("/search", searchHandler(newsapi))
	mux.HandleFunc("/", indexHandler)

	fmt.Printf("Starting server at port %s\n", port)
	http.ListenAndServe(":" + port, mux)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	// executing template: checking for errors and being written to ResponseWriter.
	buf := &bytes.Buffer{}
	err := tpl.Execute(buf, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	buf.WriteTo(w)
}

func searchHandler(newsapi *news.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// parsing raw URL into u (URL structure).
		u, err := url.Parse(r.URL.String())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// getting parameters from URL structure
		parameters := u.Query()
		searchQuery := parameters.Get("q")

		// getting page number from Query field. Setting page to 1 if not specified.
		page := parameters.Get("page")
		if page == "" {
			page = "1"
		}

		// fetching results; handling errors.
		results, err := newsapi.FetchEverything(searchQuery, page)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// converting and assigning page (string) to nextPage (int).
		nextPage, err := strconv.Atoi(page)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// declaring the Search struct.
		// computing total pages by dividing number of results by the page size, rounding to int.
		search := &Search{
			Query: searchQuery,
			NextPage: nextPage,
			TotalPages: int(math.Ceil(float64(results.TotalResults) / float64(newsapi.PageSize))),
			Results: results,

		}

		// increments the page number every time a new page of results is received.
		if ok := !search.IsLastPage(); ok {
			search.NextPage++
		}

		// executing the template, passing the search struct as the data object.
		buf := &bytes.Buffer{}
		err = tpl.Execute(buf, search)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		buf.WriteTo(w)
	}
	
}