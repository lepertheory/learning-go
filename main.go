package main

import (
	"os"

	gol "lepertheory.net/getopt-long"
)

func pointer[T any](val T) *T {
	return &val
}

func main() {
	getopt := gol.GetOpt{
		Options: []gol.Option{
			{
				Name: pointer("help"),
				Short: pointer("h"),
				Required: false,
				Arg: gol.ArgNotAllowed,
			},
		},
		Arguments: os.Args,
	}
	getopt.Process()
	println(getopt.Results[getopt.Options[0]].SetCount)
}
