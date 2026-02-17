package output

import (
	"fmt"
	"strings"
)

// OutputMarkdown outputs data as Markdown.
func OutputMarkdown(data any, title string, verbose bool) {
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
		mdList(v, includes, title, verbose)
	case map[string]any:
		mdSingle(v, includes, title, verbose)
	default:
		fmt.Println(v)
	}

	if verbose {
		if nt, ok := meta["next_token"].(string); ok && nt != "" {
			fmt.Printf("\n*Next page: `--next-token %s`*\n", nt)
		}
	}
}

func mdSingle(item, includes map[string]any, title string, verbose bool) {
	if _, ok := item["username"]; ok {
		mdUser(item, verbose)
	} else if isSimpleResponse(item) {
		mdAction(item, title)
	} else {
		mdTweet(item, includes, title, verbose)
	}
}

func isSimpleResponse(m map[string]any) bool {
	_, hasID := m["id"]
	_, hasText := m["text"]
	return !hasID && !hasText
}

func mdAction(item map[string]any, title string) {
	if title != "" {
		fmt.Printf("**%s**: ", title)
	}
	var parts []string
	for k, v := range item {
		parts = append(parts, fmt.Sprintf("%s=%v", k, v))
	}
	fmt.Println(strings.Join(parts, ", "))
}

func mdTweet(tweet, includes map[string]any, title string, verbose bool) {
	author := resolveAuthor(tweet["author_id"], includes)
	text, _ := tweet["text"].(string)
	tweetID, _ := tweet["id"].(string)

	if note, ok := tweet["note_tweet"].(map[string]any); ok {
		if nt, ok := note["text"].(string); ok && nt != "" {
			text = nt
		}
	}

	if title != "" {
		fmt.Printf("## %s\n\n", title)
	}

	fmt.Printf("**%s**\n", author)
	if verbose {
		if created, ok := tweet["created_at"].(string); ok && created != "" {
			fmt.Printf("*%s*\n", created)
		}
	}
	fmt.Printf("\n%s\n\n", text)

	if article, ok := tweet["article"].(map[string]any); ok {
		artTitle, _ := article["title"].(string)
		artBody, _ := article["plain_text"].(string)
		if artTitle != "" {
			fmt.Printf("### Article: %s\n\n", artTitle)
		}
		if artBody != "" {
			preview := artBody
			if len(preview) > 500 {
				preview = preview[:500] + "..."
			}
			fmt.Printf("%s\n\n", preview)
			fmt.Printf("*(%s chars total)*\n\n", formatNum(len(artBody)))
		}
	}

	if verbose {
		if metrics, ok := tweet["public_metrics"].(map[string]any); ok {
			var parts []string
			for k, v := range metrics {
				label := strings.ReplaceAll(k, "_count", "")
				parts = append(parts, fmt.Sprintf("%s: %s", label, formatMetricValue(v)))
			}
			fmt.Println(strings.Join(parts, " | "))
			fmt.Println()
		}
	}
	fmt.Printf("ID: `%s`\n", tweetID)
}

func mdUser(user map[string]any, verbose bool) {
	name, _ := user["name"].(string)
	username, _ := user["username"].(string)
	desc, _ := user["description"].(string)

	fmt.Printf("## %s (@%s)\n\n", name, username)
	if desc != "" {
		fmt.Printf("%s\n\n", desc)
	}

	if metrics, ok := user["public_metrics"].(map[string]any); ok {
		var parts []string
		for k, v := range metrics {
			label := strings.ReplaceAll(k, "_count", "")
			parts = append(parts, fmt.Sprintf("**%s**: %s", label, formatMetricValue(v)))
		}
		fmt.Println(strings.Join(parts, " | "))
		fmt.Println()
	}

	if verbose {
		if loc, ok := user["location"].(string); ok && loc != "" {
			fmt.Printf("Location: %s\n", loc)
		}
		if created, ok := user["created_at"].(string); ok && created != "" {
			fmt.Printf("Joined: %s\n", created)
		}
	}
}

func mdList(items []any, includes map[string]any, title string, verbose bool) {
	if len(items) == 0 {
		return
	}
	if title != "" {
		fmt.Printf("## %s\n\n", title)
	}

	first, _ := items[0].(map[string]any)
	if first != nil {
		if _, ok := first["username"]; ok {
			mdUserTable(items, verbose)
			return
		}
	}

	for i, item := range items {
		if i > 0 {
			fmt.Println("\n---\n")
		}
		if m, ok := item.(map[string]any); ok {
			mdTweet(m, includes, "", verbose)
		}
	}
}

func mdUserTable(users []any, verbose bool) {
	if verbose {
		fmt.Println("| Username | Name | Followers | Description |")
		fmt.Println("|----------|------|-----------|-------------|")
	} else {
		fmt.Println("| Username | Name | Followers |")
		fmt.Println("|----------|------|-----------|")
	}
	for _, u := range users {
		user, _ := u.(map[string]any)
		if user == nil {
			continue
		}
		username, _ := user["username"].(string)
		name, _ := user["name"].(string)
		metrics, _ := user["public_metrics"].(map[string]any)
		followers := "0"
		if metrics != nil {
			if fc, ok := metrics["followers_count"]; ok {
				followers = formatMetricValue(fc)
			}
		}
		if verbose {
			desc, _ := user["description"].(string)
			if len(desc) > 60 {
				desc = desc[:60]
			}
			desc = strings.ReplaceAll(desc, "|", "/")
			desc = strings.ReplaceAll(desc, "\n", " ")
			fmt.Printf("| @%s | %s | %s | %s |\n", username, name, followers, desc)
		} else {
			fmt.Printf("| @%s | %s | %s |\n", username, name, followers)
		}
	}
}

func formatNum(n int) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var result []byte
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, byte(c))
	}
	return string(result)
}
