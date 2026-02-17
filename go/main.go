package main

import (
	"github.com/alecthomas/kong"
	"x-cli/cmd"
)

func main() {
	var cli cmd.CLI
	ctx := kong.Parse(&cli,
		kong.Name("x"),
		kong.Description("CLI for X/Twitter API v2"),
		kong.UsageOnError(),
	)
	err := ctx.Run(&cmd.Globals{
		JSON:     cli.JSON,
		Plain:    cli.Plain,
		Markdown: cli.Markdown,
		Verbose:  cli.Verbose,
	})
	ctx.FatalIfErrorf(err)
}
