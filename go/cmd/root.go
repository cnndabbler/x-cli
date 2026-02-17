package cmd

const version = "0.1.0"

// CLI is the top-level command structure for kong.
type CLI struct {
	JSON     bool `short:"j" help:"JSON output"`
	Plain    bool `short:"p" help:"TSV output for piping"`
	Markdown bool `name:"markdown" short:"m" help:"Markdown output"`
	Verbose  bool `short:"v" help:"Verbose mode"`

	Tweet     TweetCmd     `cmd:"" help:"Tweet operations"`
	User      UserCmd      `cmd:"" help:"User operations"`
	Me        MeCmd        `cmd:"" help:"Self operations (authenticated user)"`
	Like      LikeCmd      `cmd:"" help:"Like a tweet"`
	Unlike    UnlikeCmd    `cmd:"" help:"Unlike a tweet"`
	Retweet   RetweetCmd   `cmd:"" help:"Retweet a tweet"`
	Unretweet UnretweetCmd `cmd:"" help:"Undo a retweet"`
	Watch     WatchCmd     `cmd:"" help:"Watch accounts for new tweets"`
	Version   VersionCmd   `cmd:"" help:"Print version"`
}

// Globals holds shared state passed to all command Run methods.
type Globals struct {
	JSON     bool
	Plain    bool
	Markdown bool
	Verbose  bool
}

// OutputMode returns the active output format name.
func (g *Globals) OutputMode() string {
	switch {
	case g.JSON:
		return "json"
	case g.Plain:
		return "plain"
	case g.Markdown:
		return "markdown"
	default:
		return "human"
	}
}
