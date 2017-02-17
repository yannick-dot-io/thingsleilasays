package main

import (
	"bytes"
	"encoding/json"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/joeshaw/envdecode"
)

type config struct {
	Username string `env:"TWITTER_USERNAME,required"`

	Twitter struct {
		ConsumerKey    string `env:"TWITTER_CONSUMER_KEY,required"`
		ConsumerSecret string `env:"TWITTER_CONSUMER_SECRET,required"`
		AccessToken    string `env:"TWITTER_ACCESS_TOKEN,required"`
		AccessSecret   string `env:"TWITTER_ACCESS_SECRET,required"`
		TweetLimit     int    `env:"TWITTER_TWEET_LIMIT,default=3200"`
	}

	AWS struct {
		Region          string `env:"AWS_REGION,required"`
		AccessKeyID     string `env:"AWS_ACCESS_KEY_ID,required"`
		SecretAccessKey string `env:"AWS_SECRET_ACCESS_KEY,required"`
		Bucket          string `env:"S3_BUCKET,default=thingsleilasays"`
		ObjectName      string `env:"S3_OBJECT_NAME,default=tweets.json"`
	}
}

func newTwitterClient(consumerKey, consumerSecret, accessToken, accessSecret string) *twitter.Client {
	config := oauth1.NewConfig(consumerKey, consumerSecret)
	token := oauth1.NewToken(accessToken, accessSecret)
	httpClient := config.Client(oauth1.NoContext, token)
	return twitter.NewClient(httpClient)
}

func fetchTweets(client *twitter.Client, username string, limit int) ([]twitter.Tweet, error) {
	params := &twitter.UserTimelineParams{
		ScreenName: username,
		Count:      limit,
	}
	tweets, resp, err := client.Timelines.UserTimeline(params)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()
	return tweets, nil
}

func newS3(region, accessKeyID, secretAccessKey string) (*s3.S3, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
		Credentials: credentials.NewStaticCredentials(
			accessKeyID,
			secretAccessKey,
			"",
		),
	})
	if err != nil {
		return nil, err
	}
	return s3.New(sess), nil
}

func putTweets(s *s3.S3, bucket, name string, tweets []twitter.Tweet) error {
	data, err := json.Marshal(tweets)
	if err != nil {
		return err
	}
	_, err = s.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(name),
		Body:   bytes.NewReader(data),
	})
	return err
}

func main() {
	var cfg config
	if err := envdecode.Decode(&cfg); err != nil {
		log.Fatal(err)
	}

	client := newTwitterClient(cfg.Twitter.ConsumerKey, cfg.Twitter.ConsumerSecret, cfg.Twitter.AccessToken, cfg.Twitter.AccessSecret)
	tweets, err := fetchTweets(client, cfg.Username, cfg.Twitter.TweetLimit)
	if err != nil {
		log.Fatal(err)
	}

	s, err := newS3(cfg.AWS.Region, cfg.AWS.AccessKeyID, cfg.AWS.SecretAccessKey)
	if err != nil {
		log.Fatal(err)
	}
	err = putTweets(s, cfg.AWS.Bucket, cfg.AWS.ObjectName, tweets)
	if err != nil {
		log.Fatal(err)
	}
}
