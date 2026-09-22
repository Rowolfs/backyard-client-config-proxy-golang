package transform

import (
	"encoding/json"
	"fmt"
)

func Transform(body []byte) ([]byte, error) {
	var config map[string]any

	if err := json.Unmarshal(body, &config); err != nil {
		return body, nil
	}

	fmt.Printf("Xray config received, fields=%d\n", len(config))

	// Здесь позже будем менять конфиг.

	return body, nil
}