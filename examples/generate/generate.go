package main

import (
	"fmt"

	"github.com/kkrypt0nn/spaceflake"
)

func main() {
	settings := spaceflake.NewGeneratorSettings()
	settings.BaseEpoch = 1640995200000 // Saturday, January 1, 2022 12:00:00 AM GMT

	sf, err := spaceflake.Generate(settings)
	if err != nil {
		panic(err)
	}
	fmt.Println(sf.Decompose()) // map[id:<Spaceflake> nodeID:0 sequence:<random> time:<timestamp> workerID:0]

	settings.NodeID = 5
	settings.WorkerID = 5
	settings.Sequence = 1337
	sf, err = spaceflake.Generate(settings)
	if err != nil {
		panic(err)
	}
	fmt.Println(sf.Decompose()) // map[id:<Spaceflake> nodeID:5 sequence:1337 time:<timestamp> workerID:5]
}
