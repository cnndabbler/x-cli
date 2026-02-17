package cmd

import (
	"x-cli/internal/output"
)

type LikeCmd struct {
	IDOrURL string `arg:"" help:"Tweet ID or URL"`
}

func (c *LikeCmd) Run(g *Globals) error {
	tid, err := parseTweetID(c.IDOrURL)
	if err != nil {
		return err
	}
	client := newClient()
	data, err := client.LikeTweet(tid)
	if err != nil {
		return err
	}
	output.Format(data, g.OutputMode(), "Liked", g.Verbose)
	return nil
}
