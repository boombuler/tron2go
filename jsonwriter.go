package main

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
)

type JsonWritable interface {
	ToJson() []byte
}

type JsonWriter struct {
	buffer bytes.Buffer
}

func (w *JsonWriter) StartObj() *JsonWriter {
	w.buffer.WriteByte('{')
	return w
}

func (w *JsonWriter) StartArray() *JsonWriter {
	w.buffer.WriteByte('[')
	return w
}

func (w *JsonWriter) WriteIdent(name string) *JsonWriter {
	w.buffer.WriteByte('"')
	w.buffer.WriteString(name)
	w.buffer.WriteString("\":")
	return w
}

func (w *JsonWriter) WriteStr(name, val string) *JsonWriter {
	val = strings.Replace(val, "\"", "\\\"", -1)
	w.WriteIdent(name)
	w.buffer.WriteByte('"')
	json.HTMLEscape(&w.buffer, []byte(val))
	w.buffer.WriteByte('"')
	return w
}

func (w *JsonWriter) WriteInt(name string, val int) *JsonWriter {
	w.WriteIdent(name)
	w.buffer.WriteString(strconv.Itoa(val))
	return w
}

func (w *JsonWriter) Next() *JsonWriter {
	w.buffer.WriteByte(',')
	return w
}

func (w *JsonWriter) EndArray() *JsonWriter {
	w.buffer.WriteByte(']')
	return w
}

func (w *JsonWriter) EndObj() *JsonWriter {
	w.buffer.WriteByte('}')
	return w
}

func (w *JsonWriter) Write(obj JsonWritable) *JsonWriter {
	w.buffer.Write(obj.ToJson())
	return w
}

func (w *JsonWriter) Flush() []byte {
	return w.buffer.Bytes()
}
