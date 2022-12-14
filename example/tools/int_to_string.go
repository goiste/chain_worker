package tools

import (
	"fmt"

	"github.com/goiste/chain_worker"
)

type IntToString struct {
	chain_worker.ChainTool
}

func NewIntToString() *IntToString {
	i := &IntToString{}
	i.ChainTool = chain_worker.NewBaseTool[int, string](i)
	return i
}

func (*IntToString) Do(data int) (string, error) {
	return fmt.Sprintf("%d", data), nil
}

func (*IntToString) Name() string {
	return "IntToString"
}
