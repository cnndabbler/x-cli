package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/fatih/color"

	"x-cli/internal/api"
	"x-cli/internal/output"
)

type WatchCmd struct {
	Usernames []string `arg:"" help:"Usernames to watch"`
	Interval  int      `name:"interval" short:"i" default:"60" help:"Poll interval in seconds"`
	Filters   []string `name:"filter" short:"f" help:"Only show tweets containing keyword (repeatable)"`
	Notify    bool     `name:"notify" short:"n" help:"Desktop notifications for new tweets"`
	MaxTweets int      `name:"max" default:"0" help:"Stop after N tweets (0=unlimited)"`
}

type watchTarget struct {
	username   string
	userID     string
	lastSeenID string
}

type watchStats struct {
	tweetsSeen int
	polls      int
	perUser    map[string]int
}

func (c *WatchCmd) Run(g *Globals) error {
	client := newClient()
	dim := color.New(color.Faint)
	boldGreen := color.New(color.Bold, color.FgGreen)
	yellow := color.New(color.FgYellow)
	red := color.New(color.FgRed)
	bold := color.New(color.Bold)

	// Resolve usernames to user IDs
	var targets []watchTarget
	for _, username := range c.Usernames {
		uname := stripAt(username)
		userData, err := client.GetUser(uname)
		if err != nil {
			return fmt.Errorf("could not find user @%s: %w", uname, err)
		}
		d, _ := userData["data"].(map[string]any)
		uid, _ := d["id"].(string)
		targets = append(targets, watchTarget{username: uname, userID: uid})
	}

	// Seed last_seen_id
	dim.Fprintln(os.Stderr, "Initializing watch...")
	for i := range targets {
		seedLastSeen(client, &targets[i])
		dim.Fprintf(os.Stderr, "  Tracking @%s (id=%s)\n", targets[i].username, targets[i].userID)
	}

	var usernames []string
	for _, t := range targets {
		usernames = append(usernames, "@"+t.username)
	}
	boldGreen.Fprintf(os.Stderr, "Watching %s", strings.Join(usernames, ", "))
	fmt.Fprintf(os.Stderr, " (every %ds, Ctrl+C to stop)\n", c.Interval)
	if len(c.Filters) > 0 {
		dim.Fprintf(os.Stderr, "Filters: %s\n", strings.Join(c.Filters, ", "))
	}
	fmt.Fprintln(os.Stderr)

	stats := &watchStats{perUser: make(map[string]int)}

	// Handle Ctrl+C
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	for {
		for i := range targets {
			t := &targets[i]
			data, err := client.GetTimeline(t.userID, 10, t.lastSeenID)
			if err != nil {
				if rle, ok := err.(*api.RateLimitError); ok {
					wait := 60
					if ts, err2 := parseUnixTimestamp(rle.ResetAt); err2 == nil {
						w := int(ts-time.Now().Unix()) + 5
						if w > 0 {
							wait = w
						}
					}
					yellow.Fprintf(os.Stderr, "Rate limited. Waiting %ds...\n", wait)
					sleepWithSignal(time.Duration(wait)*time.Second, sigCh, func() {
						printSummary(stats, targets, bold)
					})
					continue
				}
				red.Fprintf(os.Stderr, "Error for @%s: %v\n", t.username, err)
				continue
			}

			tweets, _ := data["data"].([]any)
			if len(tweets) == 0 {
				continue
			}
			includes, _ := data["includes"].(map[string]any)
			if includes == nil {
				includes = map[string]any{}
			}

			// Process oldest-first for chronological output
			for j := len(tweets) - 1; j >= 0; j-- {
				tweet, _ := tweets[j].(map[string]any)
				if tweet == nil {
					continue
				}

				text, _ := tweet["text"].(string)
				if note, ok := tweet["note_tweet"].(map[string]any); ok {
					if nt, _ := note["text"].(string); nt != "" {
						text = nt
					}
				}

				if !matchesFilters(text, c.Filters) {
					continue
				}

				payload := map[string]any{
					"data":     tweet,
					"includes": includes,
				}
				output.Format(payload, g.OutputMode(), "@"+t.username, g.Verbose)

				if c.Notify {
					notify(t.username, text)
				}

				stats.tweetsSeen++
				stats.perUser[t.username]++

				if c.MaxTweets > 0 && stats.tweetsSeen >= c.MaxTweets {
					bold.Fprintf(os.Stderr, "\nReached --max %d tweets.\n", c.MaxTweets)
					printSummary(stats, targets, bold)
					return nil
				}
			}

			// Update cursor to newest tweet
			if first, ok := tweets[0].(map[string]any); ok {
				if id, _ := first["id"].(string); id != "" {
					t.lastSeenID = id
				}
			}
		}

		stats.polls++

		// Sleep with signal handling
		done := sleepWithSignal(time.Duration(c.Interval)*time.Second, sigCh, func() {
			printSummary(stats, targets, bold)
		})
		if done {
			return nil
		}
	}
}

func seedLastSeen(client *api.Client, target *watchTarget) {
	data, err := client.GetTimeline(target.userID, 5, "")
	if err != nil {
		return
	}
	tweets, _ := data["data"].([]any)
	if len(tweets) > 0 {
		if first, ok := tweets[0].(map[string]any); ok {
			target.lastSeenID, _ = first["id"].(string)
		}
	}
}

func matchesFilters(text string, filters []string) bool {
	if len(filters) == 0 {
		return true
	}
	lower := strings.ToLower(text)
	for _, f := range filters {
		if strings.Contains(lower, strings.ToLower(f)) {
			return true
		}
	}
	return false
}

func notify(username, text string) {
	// Terminal bell fallback (cross-platform)
	fmt.Print("\a")
}

func printSummary(stats *watchStats, targets []watchTarget, bold *color.Color) {
	fmt.Fprintln(os.Stderr)
	bold.Fprintln(os.Stderr, "Watch session summary")
	fmt.Fprintf(os.Stderr, "  Polls: %d\n", stats.polls)
	fmt.Fprintf(os.Stderr, "  Tweets seen: %d\n", stats.tweetsSeen)
	for _, t := range targets {
		count := stats.perUser[t.username]
		fmt.Fprintf(os.Stderr, "  @%s: %d new tweets\n", t.username, count)
	}
}

func parseUnixTimestamp(s string) (int64, error) {
	var ts int64
	_, err := fmt.Sscanf(s, "%d", &ts)
	return ts, err
}

// sleepWithSignal sleeps for the given duration but returns early on signal.
// Returns true if a signal was received (and onSignal was called).
func sleepWithSignal(d time.Duration, sigCh chan os.Signal, onSignal func()) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-timer.C:
		return false
	case <-sigCh:
		onSignal()
		return true
	}
}
