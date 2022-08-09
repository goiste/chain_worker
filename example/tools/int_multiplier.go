package tools

import (
	"github.com/goiste/chain_worker"
)

type IntMultiplier struct {
	multiplier int
	chain_worker.ChainTool
}

func NewIntMultiplier(multiplier int) *IntMultiplier {
	i := &IntMultiplier{
		multiplier: multiplier,
	}
	i.ChainTool = chain_worker.NewBaseTool[int, int](i)
	return i
}

func (im *IntMultiplier) Do(data int) (int, error) {
	return data * im.multiplier, nil
}

func (*IntMultiplier) Name() string {
	return "IntMultiplier"
}
