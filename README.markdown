intweet
=======

Simple little Twitter -> Atom Feed tool.

I think Twitter is really fun, but I don't like how most twitter
clients work.

Reverse chronological order drives me crazy. I'm not glued to an
iPhone all day reading every tweet as they come in. I prefer to live
my life and focus on the world around me and catch up on things like
Twitter when I have some free time. With most Twitter clients, that
means starting with the newest tweets, being confused by the previous
tweets that they refer to, then going down the list until I hit one
that I'm pretty sure I saw the last time I looked at my twitter
client. Then starting the process over again the next time.

I already have an interface for consuming information that works very
well for me and that I'm happy with. It's a web-based feed reader that
gives me all the entries in chronological order and keeps track (for
me) of which ones I've already seen and which ones are new. I use it
for keeping up with hundreds of news feeds every day.

So I figured the most sensible thing was just to take the twitter feed
and spit it out as an Atom feed that my reader could handle.

That's all intweet is. You configure it to connect to your Twitter
account, set it running somewhere that your feed reader can access,
and then subscribe to it. It fetches new tweets every minute or so and
makes them available as an Atom feed.

Installation
-------------

To use it, you'll need a Go development environment set up.

Pull down the code and build:

    $ git clone https://github.com/thraxil/intweet.git
    $ cd intweet
    $ make install_deps
    $ make

Then you need to get a set of API keys and tokens from
Twitter. Register at: https://dev.twitter.com/apps and create an
access token. You'll end up with four things: Consumer Key, Consumer
Secret, Access Token, and Access token secret.

Make a file named something like `config.toml` and put in it something
like:

    oauth_token = "this value you get from twitter"
    oauth_secret = "this value you get from twitter"
    consumer_key = "this value you get from twitter"
    consumer_secret = "this value you get from twitter"
    max_tweets = 100
    poll_interval = 300
    port = ":8000"
    feed_title = "my awesome twitter feed"
    feed_link = "http://tweets.example.com/"
    feed_description = "My Twitter Feed"
    feed_author_name = "Your Name"
    feed_author_email = "you@example.com"


`max_tweets` is the number of tweets to keep around. If you follow a
lot of people and they tweet a lot and your reader doesn't fetch feeds
very often, you may want this higher. `poll_interval` is how often it
asks Twitter for new feeds. Don't set this too low or you'll get shut
down by Twitter for hitting their API too often. `port` is the port
that the server will run on. Yes, the ":" should be in there. The
other fields are just for the feed metadata.

Then you can run it with:

    $ ./intweet -config=config.toml

If the Twitter authentication fails, you'll get an ugly error and
you'll need to figure out what you did wrong.

Otherwise, you should be able to point your browser at
http://localhost:8000/ (or wherever you are running it) and see a bare
bones HTML view of the most recent tweets in your timeline. That's
mainly to verify that things are working right. The actual Atom feed
is then at http://localhost:8000/atom.xml so that's what you would
subscribe to.

I've included a sample upstart config as well. If you're running
Ubuntu, you can copy `upstart.conf` to `/etc/init/intweet.conf`, edit
it to point to whereever you have the intweet binary and config files,
change the user that it runs as, then run `sudo start intweet` and
upstart will keep the process running for you.

There's also a Docker container that's easy to run (if you use
Docker). Make your config.toml as described above and run like so:

   docker run \
       -p 8890:8890 \
       -v /path/to/config.toml:/intweet/config.toml \
       thraxil/intweet
