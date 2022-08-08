package tools

import (
	"github.com/goiste/chain_worker/src/chain_worker"
)

type StringHolder struct {
	Str string
}

type StringToStringHolder struct {
	chain_worker.ChainTool
}

func NewStringToStringHolder() *StringToStringHolder {
	s := &StringToStringHolder{}
	s.ChainTool = chain_worker.NewBaseTool[string, StringHolder](s)
	return s
}

func (*StringToStringHolder) Do(data string) (StringHolder, error) {
	return StringHolder{Str: data}, nil
}

func (*StringToStringHolder) Name() string {
	return "StringToStringHolder"
}
