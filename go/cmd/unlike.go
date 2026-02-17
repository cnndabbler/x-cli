package cmd

import (
	"x-cli/internal/output"
)

type UnlikeCmd struct {
	IDOrURL string `arg:"" help:"Tweet ID or URL"`
}

func (c *UnlikeCmd) Run(g *Globals) error {
	tid, err := parseTweetID(c.IDOrURL)
	if err != nil {
		return err
	}
	client := newClient()
	data, err := client.UnlikeTweet(tid)
	if err != nil {
		return err
	}
	output.Format(data, g.OutputMode(), "Unliked", g.Verbose)
	return nil
}
