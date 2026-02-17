package output

import (
	"encoding/json"
	"fmt"
)

// OutputJSON outputs data as formatted JSON.
func OutputJSON(data any, verbose bool) {
	m, ok := data.(map[string]any)
	if ok && !verbose {
		if inner, exists := m["data"]; exists {
			printJSON(inner)
			return
		}
	}
	printJSON(data)
}

func printJSON(data any) {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Println("{}")
		return
	}
	fmt.Println(string(b))
}
