package output

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

var (
	bold    = color.New(color.Bold)
	dim     = color.New(color.Faint)
	blue    = color.New(color.FgBlue)
	green   = color.New(color.FgGreen)
	yellow  = color.New(color.FgYellow)
	boldStr = color.New(color.Bold).SprintFunc()
	dimStr  = color.New(color.Faint).SprintFunc()
)

// OutputHuman outputs data with colored terminal formatting.
func OutputHuman(data any, title string, verbose bool) {
	m, ok := data.(map[string]any)
	if !ok {
		fmt.Println(data)
		return
	}

	inner := m["data"]
	includes, _ := m["includes"].(map[string]any)
	meta, _ := m["meta"].(map[string]any)
	if inner == nil {
		inner = m
	}
	if includes == nil {
		includes = map[string]any{}
	}
	if meta == nil {
		meta = map[string]any{}
	}

	switch v := inner.(type) {
	case []any:
		humanList(v, includes, title, verbose)
	case map[string]any:
		humanSingle(v, includes, title, verbose)
	default:
		fmt.Println(v)
	}

	if verbose {
		if nt, ok := meta["next_token"].(string); ok && nt != "" {
			dim.Fprintf(color.Error, "Next page: --next-token %s\n", nt)
		}
	}
}

func humanSingle(item, includes map[string]any, title string, verbose bool) {
	if _, ok := item["username"]; ok {
		humanUser(item, verbose)
	} else {
		humanTweet(item, includes, title, verbose)
	}
}

func humanTweet(tweet, includes map[string]any, title string, verbose bool) {
	author := resolveAuthor(tweet["author_id"], includes)
	text, _ := tweet["text"].(string)
	tweetID, _ := tweet["id"].(string)

	if note, ok := tweet["note_tweet"].(map[string]any); ok {
		if nt, ok := note["text"].(string); ok && nt != "" {
			text = nt
		}
	}

	panelTitle := title
	if panelTitle == "" {
		panelTitle = "Tweet " + tweetID
	}

	// Top border
	borderLen := max(len(panelTitle)+4, 60)
	blue.Println(strings.Repeat("─", borderLen))
	blue.Printf(" %s\n", panelTitle)
	blue.Println(strings.Repeat("─", borderLen))

	fmt.Printf(" %s", boldStr(author))
	if verbose {
		if created, ok := tweet["created_at"].(string); ok && created != "" {
			fmt.Printf("  %s", dimStr(created))
		}
	}
	fmt.Println()
	fmt.Println()
	fmt.Printf(" %s\n", text)

	if article, ok := tweet["article"].(map[string]any); ok {
		artTitle, _ := article["title"].(string)
		artBody, _ := article["plain_text"].(string)
		if artTitle != "" {
			fmt.Println()
			bold.Printf(" Article: %s\n", artTitle)
		}
		if artBody != "" {
			preview := artBody
			if len(preview) > 500 {
				preview = preview[:500] + "..."
			}
			fmt.Println()
			fmt.Printf(" %s\n", preview)
			fmt.Println()
			dim.Printf(" (%s chars total)\n", formatNum(len(artBody)))
		}
	}

	if verbose {
		if metrics, ok := tweet["public_metrics"].(map[string]any); ok {
			var parts []string
			for k, v := range metrics {
				label := strings.ReplaceAll(k, "_count", "")
				label = strings.ReplaceAll(label, "_", " ")
				parts = append(parts, fmt.Sprintf("%s: %s", label, formatMetricValue(v)))
			}
			fmt.Println()
			dim.Printf(" %s\n", strings.Join(parts, " | "))
		}
	}

	blue.Println(strings.Repeat("─", borderLen))
}

func humanUser(user map[string]any, verbose bool) {
	name, _ := user["name"].(string)
	username, _ := user["username"].(string)
	desc, _ := user["description"].(string)

	borderLen := max(len(username)+6, 60)
	green.Println(strings.Repeat("─", borderLen))
	green.Printf(" @%s\n", username)
	green.Println(strings.Repeat("─", borderLen))

	fmt.Printf(" %s @%s", boldStr(name), username)
	if verified, _ := user["verified"].(bool); verified {
		blue.Print(" verified")
	}
	fmt.Println()
	if desc != "" {
		fmt.Printf(" %s\n", desc)
	}

	if verbose {
		if loc, ok := user["location"].(string); ok && loc != "" {
			dim.Printf(" Location: %s\n", loc)
		}
		if created, ok := user["created_at"].(string); ok && created != "" {
			dim.Printf(" Joined: %s\n", created)
		}
	}

	if metrics, ok := user["public_metrics"].(map[string]any); ok {
		var parts []string
		for k, v := range metrics {
			label := strings.ReplaceAll(k, "_count", "")
			label = strings.ReplaceAll(label, "_", " ")
			parts = append(parts, fmt.Sprintf("%s: %s", label, formatMetricValue(v)))
		}
		fmt.Println()
		fmt.Printf(" %s\n", strings.Join(parts, " | "))
	}

	green.Println(strings.Repeat("─", borderLen))
}

func humanList(items []any, includes map[string]any, title string, verbose bool) {
	if len(items) == 0 {
		return
	}

	first, _ := items[0].(map[string]any)
	if first != nil {
		if _, ok := first["username"]; ok {
			humanUserTable(items, title, verbose)
			return
		}
	}

	for _, item := range items {
		if m, ok := item.(map[string]any); ok {
			humanTweet(m, includes, "", verbose)
		}
	}
}

func humanUserTable(users []any, title string, verbose bool) {
	if title != "" {
		bold.Println(title)
		fmt.Println()
	}

	// Collect rows for alignment
	type row struct {
		username, name, followers, desc string
	}
	var rows []row
	for _, u := range users {
		user, _ := u.(map[string]any)
		if user == nil {
			continue
		}
		uname, _ := user["username"].(string)
		name, _ := user["name"].(string)
		metrics, _ := user["public_metrics"].(map[string]any)
		followers := "0"
		if metrics != nil {
			if fc, ok := metrics["followers_count"]; ok {
				followers = formatMetricValue(fc)
			}
		}
		d, _ := user["description"].(string)
		if len(d) > 50 {
			d = d[:50]
		}
		rows = append(rows, row{username: "@" + uname, name: name, followers: followers, desc: d})
	}

	// Print header
	header := fmt.Sprintf("%-20s %-25s %12s", "USERNAME", "NAME", "FOLLOWERS")
	if verbose {
		header += fmt.Sprintf("  %-50s", "DESCRIPTION")
	}
	bold.Println(header)
	fmt.Println(strings.Repeat("-", len(header)))

	for _, r := range rows {
		line := fmt.Sprintf("%-20s %-25s %12s", r.username, r.name, r.followers)
		if verbose {
			line += fmt.Sprintf("  %-50s", r.desc)
		}
		fmt.Println(line)
	}
}

// resolveAuthor is shared between human and markdown formatters.
func resolveAuthor(authorID any, includes map[string]any) string {
	aid, ok := authorID.(string)
	if !ok || aid == "" {
		return "?"
	}
	users, _ := includes["users"].([]any)
	for _, u := range users {
		um, _ := u.(map[string]any)
		if um == nil {
			continue
		}
		if id, _ := um["id"].(string); id == aid {
			if uname, _ := um["username"].(string); uname != "" {
				return "@" + uname
			}
		}
	}
	return aid
}
