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
		log.Println("JSON-ERROR: ", err.Error())
		return nil
	}

	return res
}

func (raw rawJson) getValue(idx string, result interface{}) bool {
	defer recover()
	data, ok := raw[idx]
	if !ok {
		log.Println("JSON-ERROR: Message does not contain index ", idx)
		return false
	}
	if data == nil {
		log.Println("JSON-ERROR value is null")
		return false
	}

	err := json.Unmarshal([]byte(*data), result)
	if err != nil {
		log.Println("JSON-ERROR: ", err.Error())
		return false
	}
	return true
}
