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
			// First test, just does it work.
			{
				Name: pointer("help"),
				Short: pointer("h"),
				Required: false,
				Arg: gol.ArgNotAllowed,
			},
			// Second test, can we do two?
			{
				Name: pointer("fart"),
				Short: pointer("f"),
				Required: false,
				Arg: gol.ArgNotAllowed,
			},
		},
		Arguments: os.Args,
	}
	getopt.Process()
	println(getopt.Results[getopt.Options[0]].SetCount)
}
