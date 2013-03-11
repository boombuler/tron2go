package main

import (
    "encoding/json"
    "log"
)

type rawJson struct {
    data map[string]*json.RawMessage
}

func rawJSON(data []byte) *rawJson {
    res := &rawJson{}
    err := json.Unmarshal(data, &res.data)
    if err != nil {
        log.Printf("JSON-ERROR: %v", err.Error())
        return nil
    }

    return res
}

func (raw *rawJson)GetValue(idx string, result interface{}) bool {
    data, ok := raw.data[idx]
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