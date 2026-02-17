package cmd

import "fmt"

type VersionCmd struct{}

func (c *VersionCmd) Run(g *Globals) error {
	fmt.Printf("x-cli %s\n", version)
	return nil
}
