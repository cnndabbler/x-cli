package cmd

import (
	"x-cli/internal/output"
)

type RetweetCmd struct {
	IDOrURL string `arg:"" help:"Tweet ID or URL"`
}

func (c *RetweetCmd) Run(g *Globals) error {
	tid, err := parseTweetID(c.IDOrURL)
	if err != nil {
		return err
	}
	client := newClient()
	data, err := client.Retweet(tid)
	if err != nil {
		return err
	}
	output.Format(data, g.OutputMode(), "Retweeted", g.Verbose)
	return nil
}
