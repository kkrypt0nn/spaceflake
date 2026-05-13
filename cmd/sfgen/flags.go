package main

import "flag"

var (
	flagDecompose bool
	sfDecompose   string

	flagNode   int
	flagWorker int
)

func init() {
	flag.BoolVar(&flagDecompose, "decompose", false, "whether to decompose the Spaceflake and not generate one")
	flag.IntVar(&flagNode, "node", 0, "the node id to use")
	flag.IntVar(&flagWorker, "worker", 0, "the worker id to use")

	flag.Parse()

	sfDecompose = flag.Arg(0)
}
