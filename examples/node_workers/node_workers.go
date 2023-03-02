package main

import (
	"fmt"

	"github.com/kkrypt0nn/spaceflake"
)

func main() {
	node := spaceflake.NewNode(5)
	worker := node.NewWorker() // If BaseEpoch not changed, it will use the EPOCH constant
	sf, err := worker.GenerateSpaceflake()
	if err != nil {
		panic(err)
	}
	fmt.Println(sf.Decompose()) // map[id:<Spaceflake> nodeID:5 sequence:1 time:<timestamp> workerID:1]

	worker.ID = 5
	worker.Sequence = 1337
	sf, err = worker.GenerateSpaceflake()
	if err != nil {
		panic(err)
	}
	fmt.Println(sf.Decompose()) // map[id:<Spaceflake> nodeID:5 sequence:1337 time:<timestamp> workerID:5]

	node.ID = 2
	worker.Sequence = 0 // We reset to auto incremented sequence
	sf, err = worker.GenerateSpaceflake()
	if err != nil {
		panic(err)
	}
	fmt.Println(sf.Decompose()) // map[id:<Spaceflake> nodeID:2 sequence:3 time:<timestamp> workerID:5]
}
