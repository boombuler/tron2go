package main

import (
	"encoding/json"
	"log"
)

type rawJson map[string]*json.RawMessage

func newRawJSON(data []byte) rawJson {
	res := make(rawJson)
	err := json.Unmarshal(data, &res)
	if err != nil {
		log.Printf("JSON-ERROR: %v", err.Error())
		return nil
	}

	return res
}

func (raw rawJson) getValue(idx string, result interface{}) bool {
	data, ok := raw[idx]
	if !ok {
		log.Printf("JSON-ERROR: Message does not contain index %v", idx)
		return false
	}
	err := json.Unmarshal([]byte(*data), result)
	if err != nil {
		log.Printf("JSON-ERROR: %v", err.Error())
		return false
	}
	return true
}
