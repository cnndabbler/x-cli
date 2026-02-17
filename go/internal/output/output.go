package output

import (
	"fmt"
	"os"
)

// Format routes data to the appropriate formatter.
func Format(data any, mode, title string, verbose bool) {
	switch mode {
	case "json":
		OutputJSON(data, verbose)
	case "plain":
		OutputPlain(data, verbose)
	case "markdown":
		OutputMarkdown(data, title, verbose)
	default:
		OutputHuman(data, title, verbose)
	}
}

// Error prints an error message to stderr and exits.
func Error(msg string, code int) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
	os.Exit(code)
}

// formatMetricValue formats a JSON number (float64) as an integer string,
// avoiding scientific notation for large numbers.
func formatMetricValue(v any) string {
	switch n := v.(type) {
	case float64:
		return fmt.Sprintf("%d", int64(n))
	case int:
		return fmt.Sprintf("%d", n)
	case int64:
		return fmt.Sprintf("%d", n)
	default:
		return fmt.Sprintf("%v", v)
	}
}
