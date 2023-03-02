package main

import (
	"fmt"

	"github.com/kkrypt0nn/spaceflake"
)

func main() {
	// Node **and** worker IDs will auto scale/increment. Will be the fastest, hence 2 millions Spaceflakes generated.
	settings := spaceflake.NewBulkGeneratorSettings(2_000_000)
	spaceflakes, err := spaceflake.BulkGenerate(settings)
	if err != nil {
		panic(err)
	}
	fmt.Println("Successfully generated", len(spaceflakes), "Spaceflakes") // Successfully generated 2000000 Spaceflakes
	fmt.Println(spaceflakes[1337331].Decompose())                          // map[id:<Spaceflake> nodeID:11 sequence:2363 time:<timestamp> workerID:17]

	// Worker IDs will auto scale/increment. Will be average speed compared to above and below, hence "only" 1 million Spaceflakes generated.
	nodeOne := spaceflake.NewNode(1)
	spaceflakes, err = nodeOne.BulkGenerateSpaceflakes(1_000_000)
	if err != nil {
		panic(err)
	}
	fmt.Println("Successfully generated", len(spaceflakes), "Spaceflakes") // Successfully generated 1000000 Spaceflakes
	fmt.Println(spaceflakes[7331].Decompose())                             // map[id:<Spaceflake> nodeID:1 sequence:3238 time:<timestamp> workerID:2]

	// Nothing will auto scale/increment. Will be the slowest, hence "only" 500 thousand Spaceflakes generated.
	nodeTwo := spaceflake.NewNode(2)
	worker := nodeTwo.NewWorker()
	spaceflakes, err = worker.BulkGenerateSpaceflakes(500_000)
	if err != nil {
		panic(err)
	}
	fmt.Println("Successfully generated", len(spaceflakes), "Spaceflakes") // Successfully generated 500000 Spaceflakes
	fmt.Println(spaceflakes[1337].Decompose())                             // map[id:<Spaceflake> nodeID:2 sequence:1338 time:<timestamp> workerID:1]
}
