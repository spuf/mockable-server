package control

import (
	"encoding/json"
	"fmt"
	"time"
)

type DelayDuration struct {
	time.Duration
}

func (d DelayDuration) MarshalJSON() ([]byte, error) {
	if d.Duration == 0 {
		return json.Marshal(0)
	}

	return json.Marshal(d.String())
}

func (d *DelayDuration) UnmarshalJSON(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	var err error
	var stringValue string
	switch value := v.(type) {
	case float64:
		stringValue = fmt.Sprintf("%fs", value)
	case string:
		stringValue = value
	case nil:
		stringValue = "0"
	default:
		return fmt.Errorf("%w: delay %s must be null, numeric, or string", ErrValidation, data)
	}

	d.Duration, err = time.ParseDuration(stringValue)
	if err != nil {
		return err
	}

	return nil
}
