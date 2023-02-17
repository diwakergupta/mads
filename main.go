package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/google/go-github/v49/github"
	"golang.org/x/oauth2"
)

// Create a map from week (by date) to a map of contributors (identified by id)
type StatsMap map[time.Time]map[int64]struct{}

type config struct {
	ctx    context.Context
	client *github.Client
}

// Sentinel value for hash-maps
var sentinel = struct{}{}

// The 2 structs below are to parse EC's toml -> json
type Url struct {
	Url string `json:"url"`
}
type Ecosystem struct {
	GithubOrganizations []string `json:"github_organizations"`
	Repo                []Url    `json:"repo"`
}

func printStats(weeklyStats StatsMap) {
	// Get the dates, sort them and use that to print the map
	keys := make([]time.Time, len(weeklyStats))
	i := 0
	for k := range weeklyStats {
		keys[i] = k
		i++
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Before(keys[j])
	})

	var m time.Month
	var y int = 0
	var reset bool
	devs := make(map[int64]struct{})
	for _, k := range keys {
		if y == 0 {
			y = k.Year()
			m = k.Month()
		}
		// Compare years. If same, compare months.
		reset = false
		if y == k.Year() {
			if m < k.Month() { // month has advanced, reset
				reset = true
			}
		} else { // year has advanced, reset
			reset = true
		}
		if reset {
			fmt.Printf("%d-%d, %d\n", y, m, len(devs))
			devs = make(map[int64]struct{})
			y = k.Year()
			m = k.Month()
		}

		// Add all devs for this week to the monthly set
		for id := range weeklyStats[k] {
			devs[id] = sentinel
		}
	}
	fmt.Printf("%d-%d, %d\n", y, m, len(devs))
}

func collectRepos(cfg config, file string) map[string]struct{} {
	jsonData, err := os.ReadFile(file)
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}

	var e Ecosystem
	if err := json.Unmarshal(jsonData, &e); err != nil {
		log.Fatal("Error while unmarshaling json: ", err)
	}

	repos := make(map[string]struct{})
	for _, u := range e.Repo {
		repos[u.Url] = sentinel
	}
	log.Println("Number of repos before org scan ", len(repos))

	opts := github.RepositoryListOptions{Visibility: "public", ListOptions: github.ListOptions{PerPage: 200}}
	for _, o := range e.GithubOrganizations {
		parts := strings.Split(o, "/")
		org := parts[len(parts)-1]

		log.Println("Fetching public repositories for Github org", org)
		repositories, _, err := cfg.client.Repositories.List(cfg.ctx, org, &opts)
		if err != nil {
			log.Println("Warning getting repos for org ", org, err)
		}
		log.Println("Found", len(repositories), "public repos in org ", org)
		for _, r := range repositories {
			repos[r.GetHTMLURL()] = sentinel
		}
	}
	log.Println("Number of repos after org scan ", len(repos))

	return repos
}

func getStats(ctx context.Context, client *github.Client, org, repo string) ([]*github.ContributorStats, error) {
	stats, _, err := client.Repositories.ListContributorsStats(ctx, org, repo)
	if err != nil {
		if rateErr, ok := err.(*github.RateLimitError); ok {
			handleRateLimit(rateErr)
			return getStats(ctx, client, org, repo)
		}
		if _, ok := err.(*github.AcceptedError); ok {
			return getStats(ctx, client, org, repo)
		}
	}
	return stats, err
}

func handleRateLimit(err *github.RateLimitError) {
	s := err.Rate.Reset.UTC().Sub(time.Now().UTC())
	if s < 0 {
		s = 5 * time.Second
	}
	log.Printf("hit rate limit, waiting %v", s)
	time.Sleep(s)
}

func processRepos(cfg config, repoMap map[string]struct{}) {
	stats := make(StatsMap)

	for k := range repoMap {
		parts := strings.Split(k, "/")
		if len(parts) < 2 {
			log.Println("Skipping", k)
			continue
		}
		owner := parts[len(parts)-2]
		repo := parts[len(parts)-1]
		log.Println("Processing owner/repo", owner, repo)

		weeklyStats, err := getStats(cfg.ctx, cfg.client, owner, repo)
		if err != nil {
			fmt.Println(err)
		}

		for _, c := range weeklyStats {
			for _, w := range c.Weeks {
				if w.GetCommits() > 0 { // Only count positive commits
					// Initialize a contributor map for this week if none exists
					if _, exists := stats[w.Week.Time]; !exists {
						stats[w.Week.Time] = make(map[int64]struct{})
					}
					// Add this developer for this week
					stats[w.Week.Time][c.GetAuthor().GetID()] = sentinel
				}
			}
		}
	}

	printStats(stats)
}

func main() {
	var token = flag.String("token", "", "Github API token")
	var file = flag.String("file", "stacks.json", "JSON file to process")
	flag.Parse()

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: *token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	cfg := config{ctx, client}

	processRepos(cfg, collectRepos(cfg, *file))
}
