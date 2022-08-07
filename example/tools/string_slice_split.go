package tools

import (
	"chain_worker/src/chain_worker"
)

type StringSliceSplit struct {
	chain_worker.ChainTool
}

func NewStringSliceSplit() *StringSliceSplit {
	s := &StringSliceSplit{}
	s.ChainTool = chain_worker.NewBaseGeneratorTool[string, string](s)
	return s
}

func (s *StringSliceSplit) Generate(data []string) func() (string, bool) {
	var i int
	length := len(data)

	return func() (stringData string, ok bool) {
		ok = i < length
		if !ok {
			return
		}
		stringData = s.do(data[i])
		i++
		return
	}
}

func (s *StringSliceSplit) do(data string) string {
	return data
}

func (*StringSliceSplit) Name() string {
	return "StringSliceSplit"
}
