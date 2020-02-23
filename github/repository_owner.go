package github

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type RepositoryOwner struct {
	Location       string `json:"location"`
	Country        string `json:"country"`
	Email          string `json:"email"`
	IsOrganization bool   `json:"organization"`
}

func (owner *RepositoryOwner) ParseCountry() {
	splits := strings.Split(owner.Location, ",")
	owner.Country = splits[len(splits)-1]
}

func (client *Client) LoadRepoOwner(repoName string) RepositoryOwner {
	ownerName := strings.Split(repoName, "/")[0]
	ownerURL := fmt.Sprintf("%s/%s", host, ownerName)

	req, err := client.NewRequest(http.MethodGet, ownerURL, nil)
	if err != nil {
		log.Fatalf("new http request with error: %s, %v", ownerURL, err)
	}

	body, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer body.Close()

	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		log.Fatal(err)
	}

	cardInfo := doc.Find("ul.vcard-details")
	// check if organization
	if len(cardInfo.Nodes) == 0 {
		return client.loadOrganization(doc)
	}

	return client.loadNormalUser(cardInfo)
}

func (client *Client) loadOrganization(doc *goquery.Document) RepositoryOwner {
	owner := RepositoryOwner{
		IsOrganization: true,
	}

	doc.Find(".TableObject-item ul li").Each(func(_ int, s *goquery.Selection) {
		if len(s.Find("svg.octicon-location").Nodes) > 0 {
			owner.Location = strings.TrimSpace(s.Text())
			owner.ParseCountry()
		}

		if len(s.Find("svg.octicon-mail").Nodes) > 0 {
			owner.Email = strings.TrimSpace(s.Text())
		}
	})

	return owner
}

func (client *Client) loadNormalUser(s *goquery.Selection) RepositoryOwner {
	owner := RepositoryOwner{}

	s.Find("li").Each(func(_ int, s *goquery.Selection) {
		itemprop, exists := s.Attr("itemprop")
		if !exists {
			return
		}

		switch itemprop {
		case "homeLocation":
			owner.Location = strings.TrimSpace(s.Text())
			owner.ParseCountry()
		case "email":
			owner.Email = s.Find("a").Text()
		}
	})

	return owner
}
