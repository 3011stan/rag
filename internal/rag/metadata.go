package rag

import "encoding/json"

func marshalJSON(data map[string]interface{}) ([]byte, error) {
	if data == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(data)
}

func unmarshalJSON(data []byte) (map[string]interface{}, error) {
	if len(data) == 0 {
		return map[string]interface{}{}, nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	if result == nil {
		return map[string]interface{}{}, nil
	}

	return result, nil
}
