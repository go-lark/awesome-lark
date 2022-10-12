// Package main .
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

var githubClient *github.Client

func main() {
	// init github client
	initGitHubClient()
	// load template
	tpl, err := loadTemplate()
	if err != nil {
		log.Fatalln(err)
		return
	}
	// load meta
	meta, err := loadMetaJSON()
	if err != nil {
		log.Fatalln(err)
		return
	}
	pageData := PageData{
		Meta: meta,
		Year: time.Now().Year(),
	}
	// build and write
	text, err := buildREADME(tpl, pageData)
	if err != nil {
		log.Fatalln(err)
		return
	}
	if err = writeREADME(text); err != nil {
		log.Fatalln(err)
	}
}

func initGitHubClient() {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	githubClient = github.NewClient(tc)
}

func getGitHubRepo(repoURL string) (*github.Repository, error) {
	u, _ := url.Parse(repoURL)
	path := strings.Split(u.Path, "/")
	if len(path) < 3 {
		return nil, fmt.Errorf("invalid repo url: %s", repoURL)
	}

	owner := path[1]
	repo := path[2]

	res, _, err := githubClient.Repositories.Get(context.Background(), owner, repo)
	return res, err
}

// PageData .
type PageData struct {
	Meta []*Meta
	Year int
}

// Meta json
type Meta struct {
	Category      string          `json:"category"`
	Children      []*Meta         `json:"children"`
	Repos         []string        `json:"repos"`
	ReposWithInfo []*RepoWithInfo `json:"-"`
}

// RepoWithInfo .
type RepoWithInfo struct {
	Name            string
	URL             string
	Description     string
	StargazersCount int
}

func loadTemplate() (string, error) {
	tpl, err := ioutil.ReadFile("README.md.template")
	if err != nil {
		return "", err
	}

	return string(tpl), nil
}

func loadMetaJSON() ([]*Meta, error) {
	bs, err := ioutil.ReadFile("meta.json")
	if err != nil {
		return nil, err
	}
	var metas []*Meta
	if err = json.Unmarshal(bs, &metas); err != nil {
		return nil, err
	}

	var loadMetaRepos func(meta *Meta) error
	loadMetaRepos = func(meta *Meta) error {
		for _, v := range meta.Children {
			if err = loadMetaRepos(v); err != nil {
				return err
			}
		}
		for _, repoURL := range meta.Repos {
			repo, err := getGitHubRepo(repoURL)
			if err != nil {
				return err
			}
			meta.ReposWithInfo = append(meta.ReposWithInfo, &RepoWithInfo{
				Name:            repo.GetFullName(),
				URL:             repo.GetHTMLURL(),
				Description:     getRepoDesc(meta.Category, repo.GetFullName(), repo.GetDescription()),
				StargazersCount: repo.GetStargazersCount(),
			})
		}
		return nil
	}

	for _, v := range metas {
		if err = loadMetaRepos(v); err != nil {
			return nil, err
		}
	}
	return metas, err
}

func buildREADME(tpl string, pageData PageData) (string, error) {
	t, err := template.New("").Parse(tpl)
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	err = t.Execute(buf, pageData)
	return buf.String(), err
}

func writeREADME(text string) error {
	return ioutil.WriteFile("README.md", []byte(text), 0o644)
}

func getRepoDesc(category, fullname, desc string) string {
	if desc != "" {
		return desc
	}
	if strings.Contains(fullname, "larksuite/") {
		return fmt.Sprintf("Larksuite official %s SDK", category)
	}
	return ""
}
