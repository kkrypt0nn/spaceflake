package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/kkrypt0nn/spaceflake"
)

func main() {
	if flagDecompose {
		decomposeSpaceflake()
		return
	}
	generateSpaceflake()
}

func decomposeSpaceflake() {
	id, err := strconv.ParseUint(sfDecompose, 10, 64)
	if err != nil {
		log.Fatalf("invalid Spaceflake ID (%s): %v", sfDecompose, err)
	}

	decomposed := spaceflake.Decompose(id, spaceflake.EPOCH)
	timestamp := int64(decomposed["time"])
	date := time.UnixMilli(timestamp).Format(time.RFC1123)

	fmt.Printf(
		"Node: %d\nWorker: %d\nSequence: %d\nTime: %d (%s)\n",
		decomposed["nodeID"],
		decomposed["workerID"],
		decomposed["sequence"],
		timestamp,
		date,
	)
}

func generateSpaceflake() {
	settings := spaceflake.NewGeneratorSettings()
	settings.NodeID = uint64(flagNode)
	settings.WorkerID = uint64(flagWorker)

	sf, err := spaceflake.Generate(settings)
	if err != nil {
		log.Fatalf("failed to generate Spaceflake: %v", err)
	}

	fmt.Println(sf.StringID())
}
