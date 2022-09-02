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
	fmt.Println(sf.Decompose()) // map[id:<spaceflake> nodeId:0 sequence:<random> time:<timestamp> workerId:0]

	settings.NodeId = 5
	settings.WorkerId = 5
	settings.Sequence = 1337
	sf, err = spaceflake.Generate(settings)
	if err != nil {
		panic(err)
	}
	fmt.Println(sf.Decompose()) // map[id:<spaceflake> nodeId:5 sequence:1337 time:<timestamp> workerId:5]
}
