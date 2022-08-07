package main

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	"chain_worker/example/tools"
	"chain_worker/src/chain_worker"
)

func main() {
	strSlice := [][]string{{"3", "5", "8"}}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stringToInt := tools.NewStringToInt()
	intToString := tools.NewIntToString()
	intMultiplier := tools.NewIntMultiplier(3)
	stringSliceSplit := tools.NewStringSliceSplit()
	stringToStringHolder := tools.NewStringToStringHolder()

	wrk := chain_worker.New(strSlice)

	wrk.Subscribe(chain_worker.InputName, stringSliceSplit)
	wrk.Subscribe(stringSliceSplit.Name(), stringToInt, stringToStringHolder)
	wrk.Subscribe(stringToInt.Name(), intMultiplier, intToString)
	wrk.Subscribe(intMultiplier.Name(), intToString)
	wrk.Subscribe(intToString.Name(), stringToStringHolder)

	wrk.SetOutput(map[string]func() interface{}{
		stringToStringHolder.Name(): func() interface{} {
			return new(tools.StringHolder)
		},
		intToString.Name(): func() interface{} {
			return new(string)
		},
	})

	res, errs := wrk.Run(ctx)
	if len(errs) > 0 {
		fmt.Println("Errors:")
		for _, e := range errs {
			fmt.Println(e.Error())
		}
	}

	results := make([]string, len(res))
	for i := range res {
		switch val := res[i].(type) {
		case *string:
			results[i] = *val
		case *tools.StringHolder:
			results[i] = val.Str
		default:
			results[i] = fmt.Sprintf("%#v", val)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		ii, _ := strconv.Atoi(results[i])
		jj, _ := strconv.Atoi(results[j])
		return ii < jj
	})

	fmt.Println("Results:")
	for _, r := range results {
		fmt.Printf("%#v\n", r)
	}
}
