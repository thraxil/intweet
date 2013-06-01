package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/garyburd/go-oauth/oauth"
	"github.com/gorilla/feeds"
	"github.com/xiam/twitter"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

var (
	MAX_TWEETS    = 50
	POLL_INTERVAL = 60
)

// a "class" for tweets
type tweet struct {
	Handle   string
	Text     string
	Created  string
	Id       string
	FullName string
}

func (t tweet) String() string {
	return fmt.Sprintf("%s\t@%s\t%v",
		t.Created,
		t.Handle,
		t.Text)
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
	Newest string
}

func newTweetCollection() *tweetCollection {
	t := &tweetCollection{
		tweets: make([]tweet, 0),
		ids:    make(map[string]tweet),
		chF:    make(chan func()),
		Newest: "",
	}
	go t.backend()
	return t
}

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
			if tw.Id > t.Newest {
				t.Newest = tw.Id
			}
		}
	}
}

func (t tweetCollection) All() []tweet {
	rch := make(chan []tweet)
	go func() {
		t.chF <- func() {
			rch <- t.tweets
		}
	}()
	return <-rch
}

type ConfigData struct {
	ConsumerKey    string
	ConsumerSecret string
	OauthToken     string
	OauthSecret    string
	MaxTweets      int
	PollInterval   int
	Port           string
}

func poll(client *twitter.Client, tweets *tweetCollection) {
	for {
		p := url.Values{}
		if tweets.Newest != "" {
			p.Set("since_id", tweets.Newest)
		}
		results, _ := client.HomeTimeline(p)
		for _, v := range *results {
			var u map[string]interface{}
			u = v["user"].(map[string]interface{})
			t := tweet{
				u["screen_name"].(string),
				v["text"].(string),
				v["created_at"].(string),
				v["id_str"].(string),
				u["name"].(string),
			}
			tweets.Add(t)
		}
		time.Sleep(time.Duration(POLL_INTERVAL) * time.Second)
	}
}

func main() {
	var configfile string
	flag.StringVar(&configfile, "config", "./config.json", "JSON config file")
	flag.Parse()

	file, err := ioutil.ReadFile(configfile)
	if err != nil {
		panic(err.Error())
	}

	f := ConfigData{}
	err = json.Unmarshal(file, &f)
	if err != nil {
		panic(err.Error())
	}

	MAX_TWEETS = f.MaxTweets
	POLL_INTERVAL = f.PollInterval

	client := twitter.New(&oauth.Credentials{
		f.ConsumerKey,
		f.ConsumerSecret,
	})

	client.SetAuth(&oauth.Credentials{
		f.OauthToken,
		f.OauthSecret,
	})

	_, err = client.VerifyCredentials(nil)
	if err != nil {
		panic("error: " + err.Error())
	}

	tweets := newTweetCollection()
	go poll(client, tweets)
	http.HandleFunc("/atom.xml", makeHandler(atomHandler, tweets))
	http.HandleFunc("/", makeHandler(indexHandler, tweets))
	log.Fatal(http.ListenAndServe(f.Port, nil))
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
<h2>{{.FullName}} @{{.Handle}}</h2>
<p>{{.Text}}</p>
<small><a href="{{.URL}}">{{.Created}}</a></small>
{{end}}
</body>
</html>
`

func atomHandler(w http.ResponseWriter, r *http.Request,
	tweets *tweetCollection) {

	now := time.Now()
	feed := &feeds.Feed{
		Title:       "@thraxil twitter feed",
		Link:        &feeds.Link{Href: "http://tweets.thraxil.org/"},
		Description: "My Twitter Feed",
		Author:      &feeds.Author{"Anders Pearson", "anders@columbia.edu"},
		Created:     now,
	}

	feed.Items = []*feeds.Item{}
	for _, t := range tweets.All() {
		created, err := time.Parse("Mon Jan 2 15:04:05 -0700 2006", t.Created)
		if err != nil {
			created = now
		}
		feed.Items = append(feed.Items,
			&feeds.Item{
				Title:       t.FullName + " (@" + t.Handle + ") " + t.Created,
				Link:        &feeds.Link{Href: t.URL()},
				Description: t.Text,
				Author:      &feeds.Author{t.FullName, "@" + t.Handle},
				Created:     created,
			})
	}
	atom, _ := feed.ToAtom()
	w.Header().Set("Content-Type", "application/atom+xml")
	fmt.Fprintf(w, atom)
}
