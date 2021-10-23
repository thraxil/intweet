package main // import "github.org/thraxil/intweet"

import (
	_ "expvar"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/ChimeraCoder/anaconda"
	"github.com/gorilla/feeds"
)

var (
	MAX_TWEETS        = 50
	POLL_INTERVAL     = 60
	FEED_TITLE        = ""
	FEED_LINK         = ""
	FEED_DESCRIPTION  = ""
	FEED_AUTHOR_EMAIL = ""
	FEED_AUTHOR_NAME  = ""
)

// a "class" for tweets
type tweet struct {
	Handle  string
	Text    string
	Created string
	Id      string
}

func (t tweet) URL() string {
	return fmt.Sprintf("https://twitter.com/%s/status/%s", t.Handle, t.Id)
}

// all the tweets that we know of.
// maintain a slice of them ordered by date
// and a map indexed by Id

type tweetCollection struct {
	tweets []tweet
	ids    map[string]tweet
	chF    chan func()
	newest string    // id of the newest tweet we've see
	latest time.Time // timestamp for the last change made
}

func newTweetCollection() *tweetCollection {
	t := &tweetCollection{
		tweets: make([]tweet, 0),
		ids:    make(map[string]tweet),
		chF:    make(chan func()),
		newest: "",
		latest: time.Now(),
	}
	go t.backend()
	return t
}

// all access to the collection's internal data structures
// (the list and map) get serialized through this
// channel backend for safe concurrency.
func (t *tweetCollection) backend() {
	for f := range t.chF {
		f()
	}
}

func (t *tweetCollection) Add(tw tweet) {
	t.chF <- func() {
		_, present := t.ids[tw.Id]
		if !present {
			if len(t.tweets) >= MAX_TWEETS {
				// first, throw away old ones
				tt := make([]tweet, MAX_TWEETS)
				copy(tt, t.tweets[len(t.tweets)-MAX_TWEETS:])
				t.tweets = tt
			}
			t.tweets = append(t.tweets, tw)
			t.ids[tw.Id] = tw
			t.latest = time.Now()
			if tw.Id > t.newest {
				t.newest = tw.Id
			}
		}
	}
}

func (t *tweetCollection) All() []tweet {
	rch := make(chan []tweet)
	go func() {
		t.chF <- func() {
			rch <- t.tweets
		}
	}()
	return <-rch
}

func (t *tweetCollection) GetNewest() string {
	rch := make(chan string)
	go func() {
		t.chF <- func() {
			rch <- t.newest
		}
	}()
	return <-rch
}

func (t *tweetCollection) GetLatest() time.Time {
	rch := make(chan time.Time)
	go func() {
		t.chF <- func() {
			rch <- t.latest
		}
	}()
	return <-rch
}

func poll(client *anaconda.TwitterApi, tweets *tweetCollection) {
	for {
		p := url.Values{}
		results, _ := client.GetHomeTimeline(p)
		for _, v := range results {
			t := tweet{
				v.User.ScreenName,
				v.FullText,
				v.CreatedAt,
				v.IdStr,
			}
			tweets.Add(t)
		}
		time.Sleep(time.Duration(POLL_INTERVAL) * time.Second)
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {
	oauth_token := os.Getenv("INTWEET_OAUTH_TOKEN")
	oauth_secret := os.Getenv("INTWEET_OAUTH_SECRET")
	consumer_key := os.Getenv("INTWEET_CONSUMER_KEY")
	consumer_secret := os.Getenv("INTWEET_CONSUMER_SECRET")
	max_tweets, err := strconv.Atoi(getEnv("INTWEET_MAX_TWEETS", "100"))
	if err == nil {
		MAX_TWEETS = max_tweets
	}
	poll_interval, err := strconv.Atoi(getEnv("INTWEET_POLL_INTERVAL", "60"))
	if err == nil {
		POLL_INTERVAL = poll_interval
	}
	port := getEnv("INTWEET_PORT", ":8000")
	FEED_TITLE = getEnv("INTWEET_FEED_TITLE", "tweets")
	FEED_LINK = getEnv("INTWEET_FEED_LINK", "http://localhost:8000/")
	FEED_DESCRIPTION = getEnv("INTWEET_DESCRIPTION", "twitter to atom gateway")
	FEED_AUTHOR_NAME = getEnv("INTWEET_FEED_AUTHOR_NAME", "your name here")
	FEED_AUTHOR_EMAIL = getEnv("INTWEET_FEED_AUTHOR_EMAIL", "you@example.com")

	client := anaconda.NewTwitterApiWithCredentials(
		oauth_token, oauth_secret, consumer_key, consumer_secret)

	tweets := newTweetCollection()
	go poll(client, tweets)
	http.HandleFunc("/atom.xml", makeHandler(atomHandler, tweets))
	http.HandleFunc("/", makeHandler(indexHandler, tweets))
	log.Fatal(http.ListenAndServe(port, nil))
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, *tweetCollection),
	tweets *tweetCollection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, tweets)
	}
}

type PageResponse struct {
	Tweets []tweet
}

func indexHandler(w http.ResponseWriter, r *http.Request,
	tweets *tweetCollection) {
	pr := PageResponse{
		Tweets: tweets.All(),
	}
	t, _ := template.New("index").Parse(index_view_template)
	t.Execute(w, pr)
}

const index_view_template = `
<html>
<head>
<title>Tweets</title>
</head>
<body>

{{range .Tweets}}
>@{{.Handle}}</h2>
<p>{{.Text}}</p>
<small><a href="{{.URL}}">{{.Created}}</a></small>
{{end}}
</body>
</html>
`

func atomHandler(w http.ResponseWriter, r *http.Request,
	tweets *tweetCollection) {
	latest := tweets.GetLatest()

	feed := &feeds.Feed{
		Title:       FEED_TITLE,
		Link:        &feeds.Link{Href: FEED_LINK},
		Description: FEED_DESCRIPTION,
		Author:      &feeds.Author{FEED_AUTHOR_NAME, FEED_AUTHOR_EMAIL},
		Created:     latest,
	}

	feed.Items = []*feeds.Item{}
	all_tweets := tweets.All()
	for _, t := range all_tweets {
		created, err := time.Parse("Mon Jan 2 15:04:05 -0700 2006", t.Created)
		if err != nil {
			created = latest
		}
		feed.Items = append(feed.Items,
			&feeds.Item{
				Title:       "@" + t.Handle + " " + t.Created,
				Link:        &feeds.Link{Href: t.URL()},
				Description: t.Text,
				Author:      &feeds.Author{t.Handle, "@" + t.Handle},
				Created:     created,
			})
	}
	atom, _ := feed.ToAtom()
	w.Header().Set("Content-Type", "application/atom+xml")
	fmt.Fprintf(w, atom)
}
