package output

import (
	"encoding/json"
	"fmt"
	"strings"
)

// OutputPlain outputs data as TSV for piping.
func OutputPlain(data any, verbose bool) {
	m, ok := data.(map[string]any)
	if !ok {
		fmt.Println(data)
		return
	}

	inner := m["data"]
	if inner == nil {
		inner = m
	}

	switch v := inner.(type) {
	case []any:
		plainList(v, verbose)
	case map[string]any:
		plainDict(v, verbose)
	default:
		fmt.Println(v)
	}
}

func plainDict(d map[string]any, verbose bool) {
	skip := map[string]bool{
		"public_metrics": true, "entities": true, "edit_history_tweet_ids": true,
		"attachments": true, "referenced_tweets": true, "profile_image_url": true,
	}
	for k, v := range d {
		if !verbose && skip[k] {
			continue
		}
		switch val := v.(type) {
		case map[string]any, []any:
			b, _ := json.Marshal(val)
			fmt.Printf("%s\t%s\n", k, string(b))
		default:
			fmt.Printf("%s\t%v\n", k, v)
		}
	}
}

func plainList(items []any, verbose bool) {
	if len(items) == 0 {
		return
	}
	first, ok := items[0].(map[string]any)
	if !ok {
		for _, item := range items {
			fmt.Println(item)
		}
		return
	}

	// Pick columns
	var keys []string
	if _, hasUsername := first["username"]; hasUsername {
		for _, k := range []string{"username", "name", "description"} {
			if _, exists := first[k]; exists {
				keys = append(keys, k)
			}
		}
	} else {
		for _, k := range []string{"id", "author_id", "text", "created_at"} {
			if _, exists := first[k]; exists {
				keys = append(keys, k)
			}
		}
	}
	if verbose || len(keys) == 0 {
		keys = nil
		for k := range first {
			keys = append(keys, k)
		}
	}

	// Header
	fmt.Println(strings.Join(keys, "\t"))

	// Rows
	for _, item := range items {
		row, _ := item.(map[string]any)
		if row == nil {
			continue
		}
		vals := make([]string, len(keys))
		for i, k := range keys {
			v := row[k]
			switch val := v.(type) {
			case map[string]any, []any:
				b, _ := json.Marshal(val)
				vals[i] = string(b)
			case nil:
				vals[i] = ""
			default:
				vals[i] = fmt.Sprintf("%v", v)
			}
		}
		fmt.Println(strings.Join(vals, "\t"))
	}
}
