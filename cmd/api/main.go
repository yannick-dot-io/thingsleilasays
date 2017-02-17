package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/joeshaw/envdecode"
)

type config struct {
	Port int `env:"PORT,default=5000"`

	AWS struct {
		Region          string `env:"AWS_REGION,required"`
		AccessKeyID     string `env:"AWS_ACCESS_KEY_ID,required"`
		SecretAccessKey string `env:"AWS_SECRET_ACCESS_KEY,required"`
		Bucket          string `env:"S3_BUCKET,default=thingsleilasays"`
		ObjectName      string `env:"S3_OBJECT_NAME,default=tweets.json"`
	}
}

type page struct {
	Title  string
	Tweets []twitter.Tweet
}

type pageHandler struct {
	s3     *s3.S3
	bucket string
	name   string
}

func (h *pageHandler) getTemplatePath(urlPath string) (string, error) {
	fp := filepath.Join("templates", filepath.Clean(urlPath))
	if fp == "templates" {
		fp = "templates/index.html"
	}

	info, err := os.Stat(fp)
	if err != nil {
		if os.IsNotExist(err) {
			return "", err
		}
	}

	if info.IsDir() {
		return "", errors.New("template path is directory")
	}

	return fp, nil
}

func (h *pageHandler) getTweets() ([]twitter.Tweet, error) {
	result, err := h.s3.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(h.bucket),
		Key:    aws.String(h.name),
	})
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()
	data, err := ioutil.ReadAll(result.Body)
	if err != nil {
		return nil, err
	}
	tweets := make([]twitter.Tweet, 0)
	json.Unmarshal(data, &tweets)
	return tweets, nil
}

func (h *pageHandler) getPage() (*page, error) {
	tweets, err := h.getTweets()
	if err != nil {
		return nil, err
	}
	p := &page{
		Title:  "Things Leila says…",
		Tweets: tweets,
	}
	return p, nil
}

func (h *pageHandler) getTemplate(path string) (*template.Template, error) {
	funcMap := template.FuncMap{
		"formatDate": func(date string) (string, error) {
			t, err := time.Parse(time.RubyDate, date)
			if err != nil {
				return "", err
			}
			if t.Year() != time.Now().Year() {
				return t.Format("Mon Jan 2, 2006"), nil
			}
			return t.Format("Mon Jan 2"), nil
		},
	}
	return template.New("index.html").Funcs(funcMap).ParseFiles(path)
}

func (h *pageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p, err := h.getPage()
	if err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(503), 503)
	}

	fp, err := h.getTemplatePath(r.URL.Path)
	if err != nil {
		log.Println(err.Error())
		http.NotFound(w, r)
		return
	}

	tmpl, err := h.getTemplate(fp)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(500), 500)
		return
	}

	if err := tmpl.Execute(w, p); err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(500), 500)
	}
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

func main() {
	var cfg config
	if err := envdecode.Decode(&cfg); err != nil {
		log.Fatal(err)
	}

	s3, err := newS3(cfg.AWS.Region, cfg.AWS.AccessKeyID, cfg.AWS.SecretAccessKey)
	if err != nil {
		log.Fatal(err)
	}

	handler := &pageHandler{
		s3:     s3,
		bucket: cfg.AWS.Bucket,
		name:   cfg.AWS.ObjectName,
	}
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.Handle("/", handler)
	log.Printf("binding to port %d", cfg.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), nil)
}
