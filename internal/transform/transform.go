package transform

import (
	"context"
	"encoding/json"
	"log"

	"backyard-client-config-proxy-golang/internal/remnawave"
)

type Transformer struct {
	remnawave *remnawave.Client
}

func New(
	remnawaveClient *remnawave.Client,
) *Transformer {
	return &Transformer{
		remnawave: remnawaveClient,
	}
}

func (t *Transformer) Transform(
	ctx context.Context,
	subscriptionID string,
	body []byte,
) ([]byte, error) {

	log.Printf(
		"[transform] called subscriptionID=%s body=%d bytes",
		subscriptionID,
		len(body),
	)

	var configs []map[string]any

	if err := json.Unmarshal(body, &configs); err != nil {
		log.Printf("[transform] JSON parse failed: %v", err)
		return body, nil
	}

	log.Printf("[transform] parsed %d Xray configs", len(configs))

	// ============================================================
	// Admin Internal Squad
	// ============================================================

	if subscriptionID != "" {
		user, err := t.remnawave.GetUserByShortUUID(
			ctx,
			subscriptionID,
		)

		if err != nil {
			// Fail-safe:
			// если не смогли определить пользователя,
			// НЕ выдаём все конфиги, а продолжаем обычную фильтрацию.
			log.Printf(
				"[transform] failed to get user shortUuid=%q: %v",
				subscriptionID,
				err,
			)
		} else {
			log.Printf(
				"[transform] user=%q squads=%v",
				user.Username,
				user.InternalSquadNames(),
			)

			if user.HasInternalSquad("ADMINS") {
				log.Printf(
					"[transform] user=%q is in admins squad, returning all %d configs",
					user.Username,
					len(configs),
				)

				return body, nil
			}
		}
	}

	// ============================================================
	// Default filtering
	// ============================================================

	log.Printf("[transform] requesting hosts from Remnawave API")

	hosts, err := t.remnawave.GetHosts(ctx)
	if err != nil {
		log.Printf("[transform] Remnawave API ERROR: %v", err)
		return body, nil
	}

	log.Printf("[transform] Remnawave API returned %d hosts", len(hosts))

	for i := len(configs) - 1; i >= 0; i-- {
		config := configs[i]

		remark, ok := config["remarks"].(string)
		if !ok || remark == "" {
			continue
		}

		host, err := remnawave.FindHostByRemark(hosts, remark)
		if err != nil {
			log.Printf(
				"[transform] config[%d] remark=%q host not found: %v",
				i,
				remark,
				err,
			)
			continue
		}

		tags := host.TagNames()

		log.Printf(
			"[transform] config[%d] remark=%q tags=%v",
			i,
			remark,
			tags,
		)

		// PUBLIC или NODES оставляем.
		if HasTag(tags, "PUBLIC") || HasTag(tags, "NODES") {
			continue
		}

		log.Printf(
			"[transform] removing config[%d] remark=%q",
			i,
			remark,
		)

		configs = append(
			configs[:i],
			configs[i+1:]...,
		)
	}

	newBody, err := json.Marshal(configs)
	if err != nil {
		return body, err
	}

	return newBody, nil
}

func HasTag(
	tags []string,
	target string,
) bool {
	for _, tag := range tags {
		if tag == target {
			return true
		}
	}

	return false
}
