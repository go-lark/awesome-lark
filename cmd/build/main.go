// Package main .
package main

import (
	"bytes"
	"context"
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
	"gopkg.in/yaml.v2"
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
	var pageData PageData
	err = loadMetadata(&pageData)
	if err != nil {
		log.Fatalln(err)
		return
	}
	pageData.Year = time.Now().Year()
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

func getRepoInfo(repoURL string) (*github.Repository, error) {
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

func loadTemplate() (string, error) {
	tpl, err := ioutil.ReadFile("README.md.template")
	if err != nil {
		return "", err
	}

	return string(tpl), nil
}

func loadMetadata(pageData *PageData) error {
	b, err := ioutil.ReadFile("meta.yaml")
	if err != nil {
		return err
	}
	if err = yaml.Unmarshal(b, &pageData); err != nil {
		return err
	}
	return nil

	var loadRepoInfo func([]Metadata) error
	loadRepoInfo = func(md []Metadata) error {
		for i, cate := range md {
			for j, lang := range cate.Languages {
				for _, url := range lang.Repos {
					repo, err := getRepoInfo(url)
					if err != nil {
						return err
					}
					md[i].Languages[j].ReposWithInfo = append(
						md[i].Languages[j].ReposWithInfo,
						&RepoWithInfo{
							Name:            repo.GetFullName(),
							URL:             repo.GetHTMLURL(),
							Description:     getRepoDesc(lang.Language, repo.GetFullName(), repo.GetDescription()),
							StargazersCount: repo.GetStargazersCount(),
						},
					)
				}
			}
		}
		return nil
	}
	err = loadRepoInfo(pageData.Metadata)

	return err
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
