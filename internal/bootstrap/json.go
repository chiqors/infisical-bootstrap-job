package bootstrap

import "encoding/json"

func mustMarshal(value any) []byte {
	payload, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return payload
}

func unmarshalInto[T any](payload []byte, target *T) error {
	if len(payload) == 0 {
		return nil
	}
	return json.Unmarshal(payload, target)
}
