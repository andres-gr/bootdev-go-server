package main

import (
	"encoding/json"

	"github.com/google/uuid"
)

type event string

const (
	userUpgraded event = "user.upgraded"
)

type PolkaWebhook struct {
	Event event           `json:"event"`
	Data  json.RawMessage `json:"data"`
}

type UserUpgradedData struct {
	UserID uuid.UUID `json:"user_id"`
}
