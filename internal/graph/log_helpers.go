package graph

import (
	"encoding/json"
	"fmt"
)

const graphLogPayloadLimit = 4000

func jsonForGraphLog(value any) string {
	body, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprintf("<json encode failed: %v>", err)
	}
	return truncateGraphLogValue(string(body))
}

func truncateGraphLogValue(value string) string {
	runes := []rune(value)
	if len(runes) <= graphLogPayloadLimit {
		return value
	}
	return string(runes[:graphLogPayloadLimit]) + "...(truncated)"
}
