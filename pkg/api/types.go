package api

import (
	"encoding/json"
	"fmt"
	"time"
)

type Timestamp time.Time

func (timestamp Timestamp) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(timestamp).Format(time.RFC3339))
}

func (timestamp *Timestamp) UnmarshalJSON(data []byte) error {
	var timeString string
	if err := json.Unmarshal(data, &timeString); err != nil {
		return fmt.Errorf("unmarshal timestamp string: %w", err)
	}

	parsedTime, err := time.Parse(time.RFC3339, timeString)
	if err != nil {
		return fmt.Errorf("parse timestamp: %w", err)
	}

	*timestamp = Timestamp(parsedTime)
	return nil
}

type Duration time.Duration

func (duration Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(duration).String())
}

func (duration *Duration) UnmarshalJSON(data []byte) error {
	var durationString string
	if err := json.Unmarshal(data, &durationString); err != nil {
		return fmt.Errorf("unmarshal duration string: %w", err)
	}

	parsedDuration, err := time.ParseDuration(durationString)
	if err != nil {
		return fmt.Errorf("parse duration: %w", err)
	}

	*duration = Duration(parsedDuration)
	return nil
}
