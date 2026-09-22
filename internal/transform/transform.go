package transform

import (
	"encoding/json"
	"fmt"
)

func Transform(body []byte) ([]byte, error) {
	var configs []map[string]any

	if err := json.Unmarshal(body, &configs); err != nil {
		return body, err
	}

	fmt.Printf("Xray configs received: %d\n", len(configs))

	for i, config := range configs {
		outbounds, ok := config["outbounds"]
		if !ok {
			fmt.Printf("config[%d]: no outbounds\n", i)
			continue
		}

		pretty, err := json.MarshalIndent(outbounds, "", "  ")
		if err != nil {
			fmt.Printf(
				"config[%d]: failed to marshal outbounds: %v\n",
				i,
				err,
			)
			continue
		}

		fmt.Printf(
			"\n========== CONFIG %d OUTBOUNDS ==========\n%s\n",
			i,
			string(pretty),
		)
	}

	return body, nil
}