package chain_worker

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"
)

// Message stores the handling data
type Message struct {
	ToolName string
	Done     bool
	data     string
}

// Encode encodes obj to Message data field
func (m *Message) Encode(obj interface{}) error {
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)

	err := e.Encode(obj)
	if err != nil {
		return fmt.Errorf("object encoding error: %w", err)
	}

	m.data = base64.StdEncoding.EncodeToString(b.Bytes())

	return nil
}

// Decode decodes Message data field to obj
//
// obj must be a pointer to the corresponding structure
func (m *Message) Decode(obj interface{}) error {
	by, err := base64.StdEncoding.DecodeString(m.data)
	if err != nil {
		return fmt.Errorf("base64 decoding error: %w", err)
	}

	b := bytes.Buffer{}
	b.Write(by)
	d := gob.NewDecoder(&b)

	err = d.Decode(obj)

	if err != nil {
		return fmt.Errorf("object decoding error: %w", err)
	}

	return nil
}
