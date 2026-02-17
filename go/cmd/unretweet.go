package cmd

import (
	"x-cli/internal/output"
)

type UnretweetCmd struct {
	IDOrURL string `arg:"" help:"Tweet ID or URL"`
}

func (c *UnretweetCmd) Run(g *Globals) error {
	tid, err := parseTweetID(c.IDOrURL)
	if err != nil {
		return err
	}
	client := newClient()
	data, err := client.Unretweet(tid)
	if err != nil {
		return err
	}
	output.Format(data, g.OutputMode(), "Unretweeted", g.Verbose)
	return nil
}
