package tools

import (
	"strconv"

	"chain_worker/src/chain_worker"
)

type StringToInt struct {
	chain_worker.ChainTool
}

func NewStringToInt() *StringToInt {
	s := &StringToInt{}
	s.ChainTool = chain_worker.NewBaseTool[string, int](s)
	return s
}

func (sti *StringToInt) Do(data string) (int, error) {
	return strconv.Atoi(data)
}

func (*StringToInt) Name() string {
	return "StringToInt"
}
