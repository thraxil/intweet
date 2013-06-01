package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/garyburd/go-oauth/oauth"
	"github.com/xiam/twitter"
	"io/ioutil"
)

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

type ConfigData struct {
	ConsumerKey    string
	ConsumerSecret string
	OauthToken     string
	OauthSecret    string
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
	results, err := client.HomeTimeline(nil)
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
		fmt.Println(t.String())
		fmt.Println(t.URL())
	}
}
