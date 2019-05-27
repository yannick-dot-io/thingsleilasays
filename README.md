# Things Tom says…

We post things my young daughter says to Twitter, but several of our family
and friends don't use Twitter, so I created this little app to pull tweets
from Twitter and present them in a simple form. It's fairly straightforward:

- A `fetch` program runs every 10 minutes to pull all Leila's tweets from
  Twitter and stores them as JSON in an S3 bucket.
- An `api` program pulls the tweets from the S3 bucket and renders and serves
  them as HTML.

Both these programs run on Heroku using free dynos.

## Setup

The app can be adapted to work with a different Twitter account. Follow
the [setup instructions](docs/install.md) and rename `thingsleilasays` to a
name that suits your needs.
