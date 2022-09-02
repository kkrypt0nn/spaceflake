package spaceflake

import (
	"testing"
	"time"
)

// Probably not the best tests, but they test what I want to test

func TestSpaceflakeGenerateUnique(t *testing.T) {
	spaceflakes := map[uint64]*Spaceflake{}
	settings := NewGeneratorSettings()

	for i := 0; i < 1000; i++ {
		sf, err := Generate(settings)
		if err != nil {
			t.Error(err)
		}
		if spaceflakes[sf.ID()] != nil {
			t.Error("Failed! A spaceflake has been generated twice")
			return
		}
		spaceflakes[sf.ID()] = sf
		// When using random there is a chance that the sequence will be twice the same due to Go's speed, hence using a worker is better. We wait a millisecond to make sure it's different.
		time.Sleep(1 * time.Millisecond)
	}

	t.Log("Success! All spaceflakes are unique")
}

func TestSpaceflakeWorkerUnique(t *testing.T) {
	spaceflakes := map[uint64]*Spaceflake{}
	node := NewNode(1)
	worker := node.NewWorker()

	for i := 0; i < 1000; i++ {
		sf, err := worker.GenerateSpaceflake()
		if err != nil {
			t.Error(err)
		}
		if spaceflakes[sf.ID()] != nil {
			t.Error("Failed! A spaceflake has been generated twice")
			return
		}
		spaceflakes[sf.ID()] = sf
	}

	t.Log("Success! All spaceflakes are unique")
}

type result struct {
	spaceflake *Spaceflake
	err        error
}

func generate(w *Worker, channel chan *result) {
	sf, err := w.GenerateSpaceflake()
	res := &result{sf, err}
	channel <- res
}

func TestSpaceflakeWorkerGoroutineUnique(t *testing.T) {
	spaceflakes := map[uint64]*Spaceflake{}
	node := NewNode(1)
	worker := node.NewWorker()

	for i := 0; i < 1000; i++ {
		// That's not a use case for a Goroutine, but I thought let's see if it works with Goroutines
		channel := make(chan *result)
		go generate(worker, channel)
		result := <-channel
		if result.err != nil {
			t.Error(result.err)
		}
		if spaceflakes[result.spaceflake.ID()] != nil {
			t.Error("Failed! A spaceflake has been generated twice")
			return
		}
		spaceflakes[result.spaceflake.ID()] = result.spaceflake
	}

	t.Log("Success! All spaceflakes are unique")
}

func TestSameTimeStampDifferentBaseEpoch(t *testing.T) {
	settings := NewGeneratorSettings()
	sf1, err := Generate(settings) // Default epoch
	if err != nil {
		t.Error(err)
	}
	// When using random there is a chance that the sequence will be twice the same due to Go's speed, hence using a worker is better. We wait a millisecond to make sure it's different.
	time.Sleep(1 * time.Millisecond)
	settings.BaseEpoch = 1640995200000 // Saturday, January 1, 2022 12:00:00 AM GMT
	sf2, err := Generate(settings)
	if err != nil {
		t.Error(err)
	}
	if (sf1.Time() == sf2.Time()-1) || (sf1.Time() == sf2.Time()-2) { // Need to do this because we waited between the generation of the two spaceflakes
		t.Log("Success! Generated same timestamp for different base epoch")
		return
	}

	t.Error("Failed! Generated different timestamps for different base epoch")
}
