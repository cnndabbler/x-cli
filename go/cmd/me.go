package cmd

import (
	"x-cli/internal/output"
)

type MeCmd struct {
	Mentions   MeMentionsCmd   `cmd:"" help:"Fetch your recent mentions"`
	Bookmarks  MeBookmarksCmd  `cmd:"" help:"Fetch your bookmarks"`
	Bookmark   MeBookmarkCmd   `cmd:"" help:"Bookmark a tweet"`
	Unbookmark MeUnbookmarkCmd `cmd:"" help:"Remove a bookmark"`
}

// --- mentions ---

type MeMentionsCmd struct {
	MaxResults int `name:"max" default:"10" help:"Max results (5-100)"`
}

func (c *MeMentionsCmd) Run(g *Globals) error {
	client := newClient()
	data, err := client.GetMentions(c.MaxResults)
	if err != nil {
		return err
	}
	output.Format(data, g.OutputMode(), "Mentions", g.Verbose)
	return nil
}

// --- bookmarks ---

type MeBookmarksCmd struct {
	MaxResults int `name:"max" default:"10" help:"Max results (1-100)"`
}

func (c *MeBookmarksCmd) Run(g *Globals) error {
	client := newClient()
	data, err := client.GetBookmarks(c.MaxResults)
	if err != nil {
		return err
	}
	output.Format(data, g.OutputMode(), "Bookmarks", g.Verbose)
	return nil
}

// --- bookmark ---

type MeBookmarkCmd struct {
	IDOrURL string `arg:"" help:"Tweet ID or URL"`
}

func (c *MeBookmarkCmd) Run(g *Globals) error {
	tid, err := parseTweetID(c.IDOrURL)
	if err != nil {
		return err
	}
	client := newClient()
	data, err := client.BookmarkTweet(tid)
	if err != nil {
		return err
	}
	output.Format(data, g.OutputMode(), "Bookmarked", g.Verbose)
	return nil
}

// --- unbookmark ---

type MeUnbookmarkCmd struct {
	IDOrURL string `arg:"" help:"Tweet ID or URL"`
}

func (c *MeUnbookmarkCmd) Run(g *Globals) error {
	tid, err := parseTweetID(c.IDOrURL)
	if err != nil {
		return err
	}
	client := newClient()
	data, err := client.UnbookmarkTweet(tid)
	if err != nil {
		return err
	}
	output.Format(data, g.OutputMode(), "Unbookmarked", g.Verbose)
	return nil
}
