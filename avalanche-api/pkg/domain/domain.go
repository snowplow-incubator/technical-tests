package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/snowplow-incubator/avalanche-api/pkg/errors"
)

type Profile struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

type Attribute struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
}

const EventIDField = "eventId"

func ExtractEventID(event map[string]string, eventIndex int) (string, error) {
	eventID := event[EventIDField]
	if eventID == "" {
		return "", errors.NewValidationError(
			EventIDField,
			nil,
			fmt.Sprintf("missing eventId at index %d", eventIndex),
		)
	}
	return eventID, nil
}

func ExtractAttributes(event map[string]string, extractKeys []string) map[string]string {
	nameValuePairs := make(map[string]string)
	for _, key := range extractKeys {
		if value, exists := event[key]; exists && value != "" {
			nameValuePairs[key] = value
		}
	}
	return nameValuePairs
}
