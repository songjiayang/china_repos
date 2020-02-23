package github

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func (client *Client) Repositories(q string, limit int) (items []*Repository) {
	req, err := client.NewRequest(http.MethodGet, searchURL, nil)
	if err != nil {
		log.Fatalf("new http request with error: %s, %v", searchURL, err)
	}

	// set query params
	params := NewRepositoryListQuery(q, limit)
	req.URL.RawQuery = params.Encode()

	body, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer body.Close()

	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		log.Fatal(err)
	}

	doc.Find(".repo-list li ").Each(func(_ int, s *goquery.Selection) {
		repo := &Repository{
			Name:        strings.TrimSpace(s.Find("div.text-normal a").Text()),
			Description: strings.TrimSpace(s.Find("p").Text()),
		}

		s.Find("div.text-small .mr-3").Each(func(index int, s *goquery.Selection) {
			switch index {
			case 0:
				repo.Stars = strings.TrimSpace(s.Find("a").Text())
			case 2, 3:
				content, exists := s.Find("relative-time").Attr("datetime")
				if exists {
					repo.UpdatedAt = strings.TrimSpace(content)
					return
				}

				if index == 2 {
					repo.Protocol = strings.TrimSpace(s.Text())
				}
			}
		})

		repo.RepositoryOwner = client.LoadRepoOwner(repo.Name)
		items = append(items, repo)
	})

	return items
}

func NewRepositoryListQuery(q string, page int) url.Values {
	query := url.Values{}

	query.Set("q", q)
	query.Set("p", fmt.Sprintf("%d", page))

	query.Set("o", "desc")
	query.Set("s", "stars")
	query.Set("type", "Repositories")

	return query
}

type Repository struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	Stars           string `json:"stars"`
	Protocol        string `json:"protocol"`
	UpdatedAt       string `json:"updated_at"`
	RepositoryOwner `json:",inline"`
}

func (rep *Repository) IntStars() int {
	starStr := rep.Stars

	if !strings.Contains(starStr, "k") {
		starInt, err := strconv.Atoi(starStr)
		if err != nil {
			log.Fatal(err)
		}

		return starInt
	}

	starStr = strings.TrimSuffix(starStr, "k")
	starFloat, err := strconv.ParseFloat(starStr, 32)
	if err != nil {
		log.Fatal(err)
	}

	return int(starFloat * 1000)
}
