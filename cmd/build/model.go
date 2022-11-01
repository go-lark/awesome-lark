package main

// PageData .
type PageData struct {
	Metadata []Metadata `yaml:"metadata"`
	Year     int
}

// Metadata .
type Metadata struct {
	Category  string     `yaml:"category"`
	Languages []Language `yaml:"languages"`
}

// Language .
type Language struct {
	Language      string          `yaml:"language"`
	Repos         []string        `yaml:"repos"`
	ReposWithInfo []*RepoWithInfo `yaml:"-"`
}

// RepoWithInfo .
type RepoWithInfo struct {
	Name            string
	URL             string
	Description     string
	StargazersCount int
}
