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

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

var githubClient *github.Client

func main() {
	InitGithubClient()

	metas, err := LoadMetaJson()
	if err != nil {
		log.Fatalln(err)
		return
	}

	text, err := BuildReadme(metas)
	if err != nil {
		log.Fatalln(err)
		return
	}

	if err = WriteReadme(text); err != nil {
		log.Fatalln(err)
	}
}

// github

func InitGithubClient() {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	githubClient = github.NewClient(tc)
}

func GetGitHubRepo(repoURL string) (*github.Repository, error) {
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

// meta json

type Meta struct {
	Category      string          `json:"category"`
	Children      []*Meta         `json:"children"`
	Repos         []string        `json:"repos"`
	ReposWithInfo []*RepoWithInfo `json:"-"`
}

type RepoWithInfo struct {
	Name            string
	URL             string
	Description     string
	StargazersCount int
}

func LoadMetaJson() ([]*Meta, error) {
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
			repo, err := GetGitHubRepo(repoURL)
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

// readme and template

func BuildReadme(metas []*Meta) (string, error) {
	t, err := template.New("").Parse(readmeTemplate)
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	err = t.Execute(buf, metas)
	return buf.String(), err
}

func WriteReadme(text string) error {
	return ioutil.WriteFile("README.md", []byte(text), 0644)
}

var readmeTemplate = `# awesome-lark [![Awesome](https://github.com/sindresorhus/awesome/raw/main/media/badge.svg)](https://github.com/sindresorhus/awesome)

A curated list of awesome Feishu/Lark APIs, libraries, and resources.

## Platforms

- [飞书开放平台](https://open.feishu.cn/)
- [LARK Developer](https://open.larksuite.com/)

## Libraries

{{ range . }}### {{ .Category }}
{{ if gt (len .Children) 0 }}{{ range .Children }}#### {{ .Category }}
{{ range .ReposWithInfo }}- [{{ .Name }}(★{{ .StargazersCount}})]({{ .URL }}): {{ .Description }}
{{ end }}
{{ end }}
{{ else }}
{{ range .ReposWithInfo }}- [{{ .Name }}(★{{ .StargazersCount}})]({{ .URL }}): {{ .Description }}
{{ end }}
{{ end }}
{{ end }}

## Bots

- [Chatopera 飞书 Custom App](https://github.com/chatopera/chatopera.feishu): 通过 Feishu 开放平台和 Chatopera 机器人平台上线企业聊天机器人服务。
- [giphy-bot](https://github.com/go-lark/examples/tree/main/giphy-bot): A giphy bot which is built with go-lark/gin/gorm as an real world example.

## Tools

- [Card Builder](https://open.feishu.cn/tool/cardbuilder)

## Resources

- [开放平台文档](https://open.feishu.cn/document/home/index)
- [Development documentation](https://open.larksuite.com/document/home/index)
- [消息卡片设计规范](https://open.feishu.cn/document/ukTMukTMukTM/ugDOwYjL4gDM24CO4AjN)

## Contributing

Pull Request are welcomed.

## License

Copyright (c) David Zhang, 2021. Licensed under CC0 1.0 Universal.
`

func getRepoDesc(category, fullname, desc string) string {
	if desc != "" {
		return desc
	}
	if strings.Contains(fullname, "larksuite/") {
		return fmt.Sprintf("Larksuite official %s SDK", category)
	}
	return ""
}
