package spaceflake

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

const (
	// EPOCH is the base epoch for the spaceflake generation. First second of 2015, year when I got my username "Krypton"
	EPOCH = 1420070400000
	// MAX5BIT is the maximum value for a 5 bit number
	MAX5BIT = 31
	// MAX12BIT is the maximum value for a 12 bit number
	MAX12BIT = 4095
	// MAX41BIT is the maximum value for a 41 bit number
	MAX41BIT = 2199023255551
)

// Spaceflake represents a spaceflake
type Spaceflake struct {
	BaseEpoch uint64 `json:"baseEpoch"`
	BinaryId  string `json:"binaryId"`
	Id        uint64 `json:"id"`
	mutex     *sync.Mutex
}

// Time returns the time in milliseconds since the base epoch from the spaceflake
func (s *Spaceflake) Time() uint64 {
	return (s.Id >> 22) + s.BaseEpoch
}

// NodeID returns the node ID from the spaceflake
func (s *Spaceflake) NodeID() uint64 {
	return (s.Id & 0x3E0000) >> 17
}

// WorkerID returns the worker ID from the spaceflake
func (s *Spaceflake) WorkerID() uint64 {
	return (s.Id & 0x1F000) >> 12
}

// Sequence returns the sequence number from the spaceflake
func (s *Spaceflake) Sequence() uint64 {
	return s.Id & 0xFFF
}

// ID returns the spaceflake ID
func (s *Spaceflake) ID() uint64 {
	return s.Id
}

// BinaryID returns the spaceflake ID in binary format
func (s *Spaceflake) BinaryID() string {
	return s.BinaryId
}

// Decompose returns all the parts of the spaceflake
func (s *Spaceflake) Decompose() map[string]uint64 {
	return map[string]uint64{
		"id":       s.Id,
		"nodeId":   s.NodeID(),
		"sequence": s.Sequence(),
		"time":     s.Time(),
		"workerId": s.WorkerID(),
	}
}

// DecomposeBinary returns all the parts of the spaceflake in binary format
func (s *Spaceflake) DecomposeBinary() map[string]string {
	return map[string]string{
		"id":       stringPadLeft(decimalBinary(s.ID()), 64, "0"),
		"nodeId":   stringPadLeft(decimalBinary(s.NodeID()), 5, "0"),
		"sequence": stringPadLeft(decimalBinary(s.ID()), 12, "0"),
		"time":     stringPadLeft(decimalBinary(s.Time()), 41, "0"),
		"workerId": stringPadLeft(decimalBinary(s.WorkerID()), 5, "0"),
	}
}

// Node is a node in the spaceflake network that can hold workers
type Node struct {
	// NodeId is the ID of the node
	NodeId uint64 `json:"nodeId"`
	// Workers is the list of workers in the node
	workers []*Worker
}

// NewNode creates a new node in the spaceflake network
func NewNode(nodeId uint64) *Node {
	return &Node{
		NodeId:  nodeId,
		workers: make([]*Worker, 0),
	}
}

// NewWorker creates a new worker in the node
func (n *Node) NewWorker() *Worker {
	w := Worker{
		BaseEpoch: EPOCH,
		Node:      n,
		Sequence:  0,
		WorkerId:  uint64(len(n.workers) + 1),
		increment: 0,
	}
	n.workers = append(n.workers, &w)
	return &w
}

// RemoveWorker removes a worker from the node
func (n *Node) RemoveWorker(w *Worker) {
	for i, worker := range n.workers {
		if worker == w {
			n.workers = append(n.workers[:i], n.workers[i+1:]...)
			return
		}
	}
}

// GetWorkers returns the list of workers in the node
func (n *Node) GetWorkers() []*Worker {
	return n.workers
}

// Worker is a worker in the spaceflake network that can generate spaceflakes
type Worker struct {
	// BaseEpoch is the epoch that will be used for the first 41 bits when generating a spaceflake
	BaseEpoch uint64 `json:"baseEpoch"`
	// Node is the node that the worker belongs to
	Node *Node `json:"node"`
	// Sequence is the last 12 bits, usually an incremented number but can be anything. If set to 0, it will be the incremented number.
	Sequence uint64 `json:"sequence"`
	// WorkerId is the worker ID that the spaceflake generator will use for the next 5 bits
	WorkerId uint64 `json:"workerId"`

	increment uint64
}

// GenerateSpaceflake generates a spaceflake
func (w *Worker) GenerateSpaceflake() (*Spaceflake, error) {
	if w.Node == nil {
		return nil, fmt.Errorf("node is not set")
	}
	if w.Node.NodeId > MAX5BIT {
		return nil, fmt.Errorf("node ID must be less than or equals to %d", MAX5BIT)
	}
	if w.WorkerId > MAX5BIT {
		return nil, fmt.Errorf("worker ID must be less than or equals to %d", MAX5BIT)
	}
	if w.Sequence > MAX12BIT {
		return nil, fmt.Errorf("sequence must be less than or equals to %d", MAX12BIT)
	}
	// We reset the increment to 0 if it's greater than 4095, which is the max sequence number
	if w.increment > MAX12BIT {
		w.increment = 0
	}

	spaceflake := new(Spaceflake)
	spaceflake.mutex = new(sync.Mutex)
	spaceflake.mutex.Lock()
	w.increment++
	defer spaceflake.mutex.Unlock()

	milliseconds := uint64(math.Floor(microTime() * 1000))
	milliseconds -= w.BaseEpoch

	base := stringPadLeft(decimalBinary(milliseconds), 41, "0")
	nodeId := stringPadLeft(decimalBinary(w.Node.NodeId), 5, "0")
	workerId := stringPadLeft(decimalBinary(w.WorkerId), 5, "0")
	actualSequence := w.Sequence
	if w.Sequence == 0 {
		actualSequence = w.increment
	}
	sequence := stringPadLeft(decimalBinary(actualSequence), 12, "0")
	id, _ := binaryDecimal("0" + base + nodeId + workerId + sequence)

	spaceflake.BaseEpoch = w.BaseEpoch
	spaceflake.BinaryId = "0" + base + nodeId + workerId + sequence
	spaceflake.Id = id

	return spaceflake, nil
}

// GeneratorSettings is a struct that contains the settings for the spaceflake generator
type GeneratorSettings struct {
	// BaseEpoch is the epoch that the spaceflake generator will use for the first 41 bits
	BaseEpoch uint64
	// NodeId is the node ID that the spaceflake generator will use for the next 5 bits
	NodeId uint64
	// Sequence is the last 12 bits, usually an incremented number but can be anything. If set to 0, it will be random.
	Sequence uint64
	// WorkerId is the worker ID that the spaceflake generator will use for the next 5 bits
	WorkerId uint64
}

// NewGeneratorSettings is used for the settings of the Generate function below
func NewGeneratorSettings() GeneratorSettings {
	return GeneratorSettings{
		BaseEpoch: EPOCH,
		NodeId:    0,
		WorkerId:  0,
	}
}

// Generate only a spaceflake ID when you need to generate one without worker and node objects. Fastest way of creating a spaceflake.
func Generate(s GeneratorSettings) (*Spaceflake, error) {
	if s.NodeId > MAX5BIT {
		return nil, fmt.Errorf("node ID must be less than or equals to %d", MAX5BIT)
	}
	if s.WorkerId > MAX5BIT {
		return nil, fmt.Errorf("worker ID must be less than or equals to %d", MAX5BIT)
	}
	if s.Sequence > MAX12BIT {
		return nil, fmt.Errorf("sequence must be less than or equals to %d", MAX12BIT)
	}

	spaceflake := new(Spaceflake)
	spaceflake.mutex = new(sync.Mutex)
	spaceflake.mutex.Lock()
	defer spaceflake.mutex.Unlock()

	milliseconds := uint64(math.Floor(microTime() * 1000))
	milliseconds -= s.BaseEpoch

	base := stringPadLeft(decimalBinary(milliseconds), 41, "0")
	nodeId := stringPadLeft(decimalBinary(s.NodeId), 5, "0")
	workerId := stringPadLeft(decimalBinary(s.WorkerId), 5, "0")
	if s.Sequence == 0 {
		s.Sequence = uint64(random(0, MAX12BIT))
	}
	sequence := stringPadLeft(decimalBinary(s.Sequence), 12, "0")
	id, _ := binaryDecimal("0" + base + nodeId + workerId + sequence)

	spaceflake.BaseEpoch = s.BaseEpoch
	spaceflake.BinaryId = "0" + base + nodeId + workerId + sequence
	spaceflake.Id = id

	return spaceflake, nil
}

// ParseTime returns the time in milliseconds since the base epoch from the spaceflake ID
func ParseTime(spaceflakeId, baseEpoch uint64) uint64 {
	return (spaceflakeId >> 22) + baseEpoch
}

// ParseNodeID returns the node ID from the spaceflake ID
func ParseNodeID(spaceflakeId uint64) uint64 {
	return (spaceflakeId & 0x3E0000) >> 17
}

// ParseWorkerID returns the worker ID from the spaceflake ID
func ParseWorkerID(spaceflakeId uint64) uint64 {
	return (spaceflakeId & 0x1F000) >> 12
}

// ParseSequence returns the sequence number from the spaceflake ID
func ParseSequence(spaceflakeId uint64) uint64 {
	return spaceflakeId & 0xFFF
}

// Decompose returns all the parts of the spaceflake ID
func Decompose(spaceflakeId, baseEpoch uint64) map[string]uint64 {
	return map[string]uint64{
		"id":       spaceflakeId,
		"nodeId":   ParseNodeID(spaceflakeId),
		"sequence": ParseSequence(spaceflakeId),
		"time":     ParseTime(spaceflakeId, EPOCH),
		"workerId": ParseWorkerID(spaceflakeId),
	}
}

// DecomposeBinary returns all the parts of the spaceflake ID in binary format
func DecomposeBinary(spaceflakeId, baseEpoch uint64) map[string]string {
	return map[string]string{
		"binaryId": stringPadLeft(decimalBinary(spaceflakeId), 64, "0"),
		"nodeId":   stringPadLeft(decimalBinary(ParseNodeID(spaceflakeId)), 5, "0"),
		"sequence": stringPadLeft(decimalBinary(ParseSequence(spaceflakeId)), 12, "0"),
		"time":     stringPadLeft(decimalBinary(ParseTime(spaceflakeId, baseEpoch)), 41, "0"),
		"workerId": stringPadLeft(decimalBinary(ParseWorkerID(spaceflakeId)), 5, "0"),
	}
}

// --- Helper Functions ---

// microTime returns the current time in microseconds
func microTime() float64 {
	location, _ := time.LoadLocation("UTC")
	now := time.Now().In(location)
	microSeconds := float64(now.Nanosecond()) / 1000000000
	return float64(now.Unix()) + microSeconds
}

// decimalBinary returns the binary representation of a decimal number
func decimalBinary(input uint64) string {
	return strconv.FormatUint(input, 2)
}

// binaryDecimal returns the decimal representation of a binary number
func binaryDecimal(str string) (uint64, error) {
	i, err := strconv.ParseUint(str, 2, 0)
	if err != nil {
		return 0, err
	}
	return i, nil
}

// stringPadLeft returns a string with a given length and padded with a given character from the left
func stringPadLeft(input string, padLength int, padString string) string {
	output := padString
	for padLength > len(output) {
		output += output
	}
	if len(input) >= padLength {
		return input
	}
	return output[:padLength-len(input)] + input
}

// random returns a "random" number between min and max
func random(min, max int64) int64 {
	location, _ := time.LoadLocation("UTC")
	r := rand.New(rand.NewSource(time.Now().In(location).UnixNano()))
	return r.Int63n(max-min+1) + min
}
