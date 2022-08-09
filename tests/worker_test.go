package tests

import (
	"context"
	"sort"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/goiste/chain_worker"
	"github.com/goiste/chain_worker/example/tools"
)

func TestWorker_Run(t *testing.T) {
	strSlice := []string{"3", "5", "8"}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stringToInt := tools.NewStringToInt()
	intToString := tools.NewIntToString()
	intMultiplier := tools.NewIntMultiplier(3)
	stringToStringHolder := tools.NewStringToStringHolder()

	wrk := chain_worker.New(strSlice)
	wrk.Subscribe(chain_worker.InputName, stringToInt, stringToStringHolder)
	wrk.Subscribe(stringToInt.Name(), intMultiplier, intToString)
	wrk.Subscribe(intMultiplier.Name(), intToString)
	wrk.Subscribe(intToString.Name(), stringToStringHolder)
	wrk.SetOutput(map[string]func() interface{}{
		stringToStringHolder.Name(): func() interface{} {
			return new(tools.StringHolder)
		},
	})

	res, errs := wrk.Run(ctx)
	require.Empty(t, errs)

	expected := []string{"3", "3", "5", "5", "8", "8", "9", "15", "24"}
	actual := make([]string, 9)

	for j := range res {
		actual[j] = res[j].(*tools.StringHolder).Str
	}

	sort.Slice(actual, func(i, j int) bool {
		ii, _ := strconv.Atoi(actual[i])
		jj, _ := strconv.Atoi(actual[j])
		return ii < jj
	})

	require.Equal(t, expected, actual)
}
