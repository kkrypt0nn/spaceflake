/*
MIT License

Copyright (c) 2022-Present Krypton

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package spaceflake

import (
	"testing"
	"time"
)

// Probably not the best tests, but they test what I want to test

func TestSpaceflakeBulkGeneration(t *testing.T) {
	spaceflakes := map[uint64]*Spaceflake{}
	settings := NewBulkGeneratorSettings(1000000) // Default epoch
	bulk, err := BulkGenerate(settings)
	if err != nil {
		t.Error(err)
		return
	}

	for _, sf := range bulk {
		if spaceflakes[sf.ID()] != nil {
			t.Errorf("Failed! Spaceflake ID %d already exists", sf.ID())
			return
		}
		spaceflakes[sf.ID()] = sf
	}

	t.Log("Success! All Spaceflakes are unique")
}

func TestSpaceflakeBulkGenerationNode(t *testing.T) {
	spaceflakes := map[uint64]*Spaceflake{}
	node := NewNode(1)
	bulk, err := node.BulkGenerateSpaceflakes(1000000)
	if err != nil {
		t.Error(err)
		return
	}

	for _, sf := range bulk {
		if sf.NodeID() != 1 {
			t.Error("Failed! Generated Spaceflake with wrong Node ID")
			return
		}
		if spaceflakes[sf.ID()] != nil {
			t.Errorf("Failed! Spaceflake ID %d already exists", sf.ID())
			return
		}
		spaceflakes[sf.ID()] = sf
	}

	t.Log("Success! All Spaceflakes are unique")
}

func TestSpaceflakeBulkGenerationWorker(t *testing.T) {
	spaceflakes := map[uint64]*Spaceflake{}
	node := NewNode(1)
	worker := node.NewWorker()
	bulk, err := worker.BulkGenerateSpaceflakes(1000000)
	if err != nil {
		t.Error(err)
		return
	}

	for _, sf := range bulk {
		if sf.WorkerID() != 1 {
			t.Error("Failed! Generated Spaceflake with wrong Worker ID")
			return
		}
		if spaceflakes[sf.ID()] != nil {
			t.Errorf("Failed! Spaceflake ID %d already exists", sf.ID())
			return
		}
		spaceflakes[sf.ID()] = sf
	}

	t.Log("Success! All Spaceflakes are unique")
}

func TestSpaceflakeAt(t *testing.T) {
	node := NewNode(1)
	worker := node.NewWorker()
	sf, err := worker.GenerateSpaceflakeAt(time.Date(2018, 7, 21, 13, 43, 32, 64*int(time.Millisecond), time.UTC))
	if err != nil {
		t.Error(err)
		return
	}
	if sf.Time() != 1532180612064 {
		t.Error("Failed! Time of Spaceflake is not set to Saturday, July 21, 2018 1:43:32.064 PM")
		return
	}
	t.Log("Success! Time of Spaceflake is correct")
}

func TestSpaceflakeInFuture(t *testing.T) {
	node := NewNode(1)
	worker := node.NewWorker()
	worker.BaseEpoch = 2662196938000 // Tuesday, May 12, 2054 11:08:58 AM GMT
	_, err := worker.GenerateSpaceflake()
	if err != nil {
		t.Log("Success! The generator did not allowed the Spaceflake generation")
		return
	}
	t.Error("Failed! A Spaceflake has been generated with a future base epoch")
}

func TestSpaceflakeWorkerUnique(t *testing.T) {
	spaceflakes := map[uint64]*Spaceflake{}
	node := NewNode(1)
	worker := node.NewWorker()

	for i := 0; i < 1000; i++ {
		sf, err := worker.GenerateSpaceflake()
		if err != nil {
			t.Error(err)
			return
		}
		if spaceflakes[sf.ID()] != nil {
			t.Error("Failed! A Spaceflake has been generated twice")
			return
		}
		// Here we don't need to sleep because the worker is using an incrementing sequence, and we don't generate more than 4095 spaceflakes per millisecond
		spaceflakes[sf.ID()] = sf
	}

	t.Log("Success! All Spaceflakes are unique")
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
			return
		}
		if spaceflakes[result.spaceflake.ID()] != nil {
			t.Errorf("Failed! Spaceflake ID %d already exists", result.spaceflake.ID())
			return
		}
		spaceflakes[result.spaceflake.ID()] = result.spaceflake
	}

	t.Log("Success! All Spaceflakes are unique")
}

func TestSameTimeStampDifferentBaseEpoch(t *testing.T) {
	node := NewNode(1)
	worker := node.NewWorker()
	sf1, err := worker.GenerateSpaceflake() // Default epoch
	if err != nil {
		t.Error(err)
		return
	}
	worker.BaseEpoch = 1640995200000 // Saturday, January 1, 2022 12:00:00 AM GMT
	sf2, err := worker.GenerateSpaceflake()
	if err != nil {
		t.Error(err)
		return
	}
	if sf1.Time() == sf2.Time() {
		t.Log("Success! Generated same timestamp for different base epoch")
		return
	}

	t.Error("Failed! Generated different timestamps for different base epoch")
}

func TestSpaceflakeGenerateUnique(t *testing.T) {
	spaceflakes := map[uint64]*Spaceflake{}
	settings := NewGeneratorSettings()

	for i := 0; i < 1000; i++ {
		sf, err := Generate(settings)
		if err != nil {
			t.Error(err)
			return
		}
		if spaceflakes[sf.ID()] != nil {
			t.Errorf("Failed! Spaceflake ID %d already exists", sf.ID())
			return
		}
		spaceflakes[sf.ID()] = sf
		// When using random there is a chance that the sequence will be twice the same due to Go's speed, hence using a worker is better. We wait a millisecond to make sure it's different.
		time.Sleep(1 * time.Millisecond)
	}

	t.Log("Success! All Spaceflakes are unique")
}
