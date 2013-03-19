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
	buffer        bytes.Buffer
	needSeparator []bool
}

func (w *JsonWriter) writeIdent(name string) *JsonWriter {
	w.buffer.WriteByte('"')
	w.buffer.WriteString(name)
	w.buffer.WriteString("\":")
	return w
}

func (w *JsonWriter) checkSeperator() {
	if len(w.needSeparator) > 0 {
		if w.needSeparator[len(w.needSeparator)-1] {
			w.buffer.WriteByte(',')
		} else {
			w.needSeparator[len(w.needSeparator)-1] = true
		}
	}
}

func (w *JsonWriter) StartObj(name string) *JsonWriter {
	w.checkSeperator()
	if name != "" {
		w.writeIdent(name)
	}
	w.buffer.WriteByte('{')
	w.needSeparator = append(w.needSeparator, false)
	return w
}

func (w *JsonWriter) StartArray(name string) *JsonWriter {
	w.checkSeperator()
	if name != "" {
		w.writeIdent(name)
	}
	w.buffer.WriteByte('[')
	w.needSeparator = append(w.needSeparator, false)
	return w
}

func (w *JsonWriter) WriteStr(name, val string) *JsonWriter {
	w.checkSeperator()
	val = strings.Replace(val, "\"", "\\\"", -1)
	w.writeIdent(name)
	w.buffer.WriteByte('"')
	json.HTMLEscape(&w.buffer, []byte(val))
	w.buffer.WriteByte('"')
	return w
}

func (w *JsonWriter) WriteInt(name string, val int) *JsonWriter {
	w.checkSeperator()
	w.writeIdent(name)
	w.buffer.WriteString(strconv.Itoa(val))
	return w
}

func (w *JsonWriter) EndArray() *JsonWriter {
	w.buffer.WriteByte(']')
	w.needSeparator = w.needSeparator[1:]
	return w
}

func (w *JsonWriter) EndObj() *JsonWriter {
	w.buffer.WriteByte('}')
	w.needSeparator = w.needSeparator[1:]
	return w
}

func (w *JsonWriter) Write(name string, obj JsonWritable) *JsonWriter {
	w.checkSeperator()
	if name != "" {
		w.writeIdent(name)
	}
	w.buffer.Write(obj.ToJson())
	return w
}

func (w *JsonWriter) Flush() []byte {
	return w.buffer.Bytes()
}
