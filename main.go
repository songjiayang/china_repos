package main

import (
	"flag"
	"github.com/songjiayang/china_repos/github"
	"github.com/songjiayang/china_repos/worker"
)

var (
	cookieFile, language string
	minStars             int
)

func main() {
	flag.StringVar(&cookieFile, "cookie", "./cookie", "github cookie file path.")
	flag.StringVar(&language, "l", "Go", "the language you want to search.")
	flag.IntVar(&minStars, "stars", 100, "minimum stars of the repos")
	flag.Parse()

	client := github.NewClient()
	client.LoadCookie(cookieFile)

	worker.New(
		language,
		minStars,
		client,
	).Run()
}
