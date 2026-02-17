package cmd

import (
	"encoding/json"
	"fmt"

	"x-cli/internal/api"
	"x-cli/internal/config"
	"x-cli/internal/output"
)

type TweetCmd struct {
	Post    TweetPostCmd    `cmd:"" help:"Post a tweet"`
	Get     TweetGetCmd     `cmd:"" help:"Fetch a tweet by ID or URL"`
	Delete  TweetDeleteCmd  `cmd:"" help:"Delete a tweet"`
	Reply   TweetReplyCmd   `cmd:"" help:"Reply to a tweet"`
	Quote   TweetQuoteCmd   `cmd:"" help:"Quote tweet"`
	Search  TweetSearchCmd  `cmd:"" help:"Search recent tweets"`
	Metrics TweetMetricsCmd `cmd:"" help:"Get tweet engagement metrics"`
	Article TweetArticleCmd `cmd:"" help:"Extract full article text from a tweet"`
}

func newClient() *api.Client {
	creds, err := config.LoadCredentials()
	if err != nil {
		output.Error(err.Error(), 1)
	}
	return api.NewClient(creds)
}

// --- post ---

type TweetPostCmd struct {
	Text         string `arg:"" help:"Tweet text"`
	Poll         string `help:"Comma-separated poll options"`
	PollDuration int    `name:"poll-duration" default:"1440" help:"Poll duration in minutes"`
}

func (c *TweetPostCmd) Run(g *Globals) error {
	client := newClient()
	var pollOptions []string
	if c.Poll != "" {
		for _, o := range splitComma(c.Poll) {
			pollOptions = append(pollOptions, o)
		}
	}
	data, err := client.PostTweet(c.Text, "", "", pollOptions, c.PollDuration)
	if err != nil {
		return err
	}
	output.Format(data, g.OutputMode(), "Posted", g.Verbose)
	return nil
}

// --- get ---

type TweetGetCmd struct {
	IDOrURL string `arg:"" help:"Tweet ID or URL"`
}

func (c *TweetGetCmd) Run(g *Globals) error {
	tid, err := parseTweetID(c.IDOrURL)
	if err != nil {
		return err
	}
	client := newClient()
	data, err := client.GetTweet(tid)
	if err != nil {
		return err
	}
	output.Format(data, g.OutputMode(), "Tweet "+tid, g.Verbose)
	return nil
}

// --- delete ---

type TweetDeleteCmd struct {
	IDOrURL string `arg:"" help:"Tweet ID or URL"`
}

func (c *TweetDeleteCmd) Run(g *Globals) error {
	tid, err := parseTweetID(c.IDOrURL)
	if err != nil {
		return err
	}
	client := newClient()
	data, err := client.DeleteTweet(tid)
	if err != nil {
		return err
	}
	output.Format(data, g.OutputMode(), "Deleted", g.Verbose)
	return nil
}

// --- reply ---

type TweetReplyCmd struct {
	IDOrURL string `arg:"" help:"Tweet ID or URL to reply to"`
	Text    string `arg:"" help:"Reply text"`
}

func (c *TweetReplyCmd) Run(g *Globals) error {
	tid, err := parseTweetID(c.IDOrURL)
	if err != nil {
		return err
	}
	client := newClient()
	data, err := client.PostTweet(c.Text, tid, "", nil, 0)
	if err != nil {
		return err
	}
	output.Format(data, g.OutputMode(), "Reply", g.Verbose)
	return nil
}

// --- quote ---

type TweetQuoteCmd struct {
	IDOrURL string `arg:"" help:"Tweet ID or URL to quote"`
	Text    string `arg:"" help:"Quote text"`
}

func (c *TweetQuoteCmd) Run(g *Globals) error {
	tid, err := parseTweetID(c.IDOrURL)
	if err != nil {
		return err
	}
	client := newClient()
	data, err := client.PostTweet(c.Text, "", tid, nil, 0)
	if err != nil {
		return err
	}
	output.Format(data, g.OutputMode(), "Quote", g.Verbose)
	return nil
}

// --- search ---

type TweetSearchCmd struct {
	Query      string `arg:"" help:"Search query"`
	MaxResults int    `name:"max" default:"10" help:"Max results (10-100)"`
}

func (c *TweetSearchCmd) Run(g *Globals) error {
	client := newClient()
	data, err := client.SearchTweets(c.Query, c.MaxResults)
	if err != nil {
		return err
	}
	output.Format(data, g.OutputMode(), "Search: "+c.Query, g.Verbose)
	return nil
}

// --- metrics ---

type TweetMetricsCmd struct {
	IDOrURL string `arg:"" help:"Tweet ID or URL"`
}

func (c *TweetMetricsCmd) Run(g *Globals) error {
	tid, err := parseTweetID(c.IDOrURL)
	if err != nil {
		return err
	}
	client := newClient()
	data, err := client.GetTweetMetrics(tid)
	if err != nil {
		return err
	}
	output.Format(data, g.OutputMode(), "Metrics "+tid, g.Verbose)
	return nil
}

// --- article ---

type TweetArticleCmd struct {
	IDOrURL string `arg:"" help:"Tweet ID or URL"`
}

func (c *TweetArticleCmd) Run(g *Globals) error {
	tid, err := parseTweetID(c.IDOrURL)
	if err != nil {
		return err
	}
	client := newClient()
	data, err := client.GetTweet(tid)
	if err != nil {
		return err
	}

	tweetData, _ := data["data"].(map[string]any)
	if tweetData == nil {
		return fmt.Errorf("tweet %s not found", tid)
	}
	article, _ := tweetData["article"].(map[string]any)
	if article == nil {
		return fmt.Errorf("tweet %s does not contain an article", tid)
	}

	if g.JSON {
		b, _ := json.MarshalIndent(article, "", "  ")
		fmt.Println(string(b))
	} else {
		title, _ := article["title"].(string)
		body, _ := article["plain_text"].(string)
		if title != "" {
			fmt.Printf("# %s\n\n", title)
		}
		fmt.Println(body)
	}
	return nil
}

// splitComma splits a string by comma and trims whitespace.
func splitComma(s string) []string {
	var result []string
	for _, part := range split(s, ",") {
		trimmed := trimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func split(s, sep string) []string {
	var result []string
	for {
		i := indexOf(s, sep)
		if i < 0 {
			result = append(result, s)
			break
		}
		result = append(result, s[:i])
		s = s[i+len(sep):]
	}
	return result
}

func indexOf(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func trimSpace(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t') {
		s = s[:len(s)-1]
	}
	return s
}
