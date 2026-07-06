package redis_client

import (
	"encoding/json"
)

type ClickEvent struct {
	ShortCode string `json:"short_code"`
	UserAgent string `json:"user_agent"`
	IPAddress string `json:"ip_address"`
}

func PushClickEvent(event ClickEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return RDB.LPush(Ctx, "click_events", data).Err()
}
