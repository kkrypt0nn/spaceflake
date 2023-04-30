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
	// EPOCH is the base epoch for the Spaceflake generation. First second of 2015, year when I got my username "Krypton"
	EPOCH = 1420070400000
	// MAX5BITS is the maximum value for a 5 bits number
	MAX5BITS = 31
	// MAX12BITS is the maximum value for a 12 bits number
	MAX12BITS = 4095
	// MAX41BITS is the maximum value for a 41 bits number
	MAX41BITS = 2199023255551
)

// Spaceflake represents a Spaceflake
type Spaceflake struct {
	baseEpoch uint64
	binaryID  string
	id        uint64
}

// Time returns the time in milliseconds since the base epoch from the Spaceflake
func (s *Spaceflake) Time() uint64 {
	return (s.id >> 22) + s.baseEpoch
}

// NodeID returns the node ID from the Spaceflake
func (s *Spaceflake) NodeID() uint64 {
	return (s.id & 0x3E0000) >> 17
}

// WorkerID returns the worker ID from the Spaceflake
func (s *Spaceflake) WorkerID() uint64 {
	return (s.id & 0x1F000) >> 12
}

// Sequence returns the sequence number from the Spaceflake
func (s *Spaceflake) Sequence() uint64 {
	return s.id & 0xFFF
}

// ID returns the Spaceflake ID
func (s *Spaceflake) ID() uint64 {
	return s.id
}

// BinaryID returns the Spaceflake ID in binary format
func (s *Spaceflake) BinaryID() string {
	return s.binaryID
}

// StringID returns the Spaceflake ID as a string
func (s *Spaceflake) StringID() string {
	return fmt.Sprintf("%d", s.id)
}

// Decompose returns all the parts of the Spaceflake
func (s *Spaceflake) Decompose() map[string]uint64 {
	return map[string]uint64{
		"id":       s.ID(),
		"nodeID":   s.NodeID(),
		"sequence": s.Sequence(),
		"time":     s.Time(),
		"workerID": s.WorkerID(),
	}
}

// DecomposeBinary returns all the parts of the Spaceflake in binary format
func (s *Spaceflake) DecomposeBinary() map[string]string {
	return map[string]string{
		"id":       stringPadLeft(decimalBinary(s.ID()), 64, "0"),
		"nodeID":   stringPadLeft(decimalBinary(s.NodeID()), 5, "0"),
		"sequence": stringPadLeft(decimalBinary(s.Sequence()), 12, "0"),
		"time":     stringPadLeft(decimalBinary(s.Time()), 41, "0"),
		"workerD":  stringPadLeft(decimalBinary(s.WorkerID()), 5, "0"),
	}
}

// Node is a node in the Spaceflake Network that can hold workers
type Node struct {
	// ID is the ID of the node
	ID uint64
	// Workers is the list of workers in the node
	workers []*Worker
}

// NewNode creates a new node in the Spaceflake Network
func NewNode(nodeID uint64) *Node {
	return &Node{
		ID:      nodeID,
		workers: make([]*Worker, 0),
	}
}

// NewWorker creates a new worker in the node
func (n *Node) NewWorker() *Worker {
	w := Worker{
		BaseEpoch: EPOCH,
		ID:        uint64(len(n.workers) + 1),
		Node:      n,
		Sequence:  0,
		increment: 0,
		mutex:     &sync.Mutex{},
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

// BulkGenerateSpaceflakes will generate the amount of Spaceflakes specified and auto scale the worker IDs
func (n *Node) BulkGenerateSpaceflakes(amount int) ([]*Spaceflake, error) {
	node := NewNode(1)
	worker := node.NewWorker()
	spaceflakes := make([]*Spaceflake, amount)
	for i := 1; i <= amount; i++ {
		if i%((MAX12BITS*MAX5BITS)+1) == 0 {
			time.Sleep(1 * time.Millisecond)
			node.workers = make([]*Worker, 0)
			worker = node.NewWorker()
		} else if i%MAX12BITS == 0 && i%(MAX12BITS*MAX5BITS) != 0 {
			newWorker := node.NewWorker()
			worker = newWorker
		}
		spaceflake, err := generateSpaceflakeOnNodeAndWorker(worker, worker.Node)
		if err != nil {
			return nil, err
		}
		spaceflakes[i-1] = spaceflake
	}
	return spaceflakes, nil
}

// Worker is a worker in the Spaceflake Network that can generate Spaceflakes
type Worker struct {
	// BaseEpoch is the epoch that will be used for the first 41 bits when generating a Spaceflake
	BaseEpoch uint64
	// Node is the node that the worker belongs to
	Node *Node
	// Sequence is the last 12 bits, usually an incremented number but can be anything. If set to 0, it will be the incremented number.
	Sequence uint64
	// ID is the worker ID that the Spaceflake generator will use for the next 5 bits
	ID uint64

	increment uint64
	mutex     *sync.Mutex
}

// GenerateSpaceflake generates a Spaceflake
func (w *Worker) GenerateSpaceflake() (*Spaceflake, error) {
	if w.Node == nil {
		return nil, fmt.Errorf("node is not set")
	}
	if w.Node.ID > MAX5BITS {
		return nil, fmt.Errorf("node ID must be less than or equals to %d", MAX5BITS)
	}
	if w.ID > MAX5BITS {
		return nil, fmt.Errorf("worker ID must be less than or equals to %d", MAX5BITS)
	}
	if w.Sequence > MAX12BITS {
		return nil, fmt.Errorf("sequence must be less than or equals to %d", MAX12BITS)
	}
	if w.BaseEpoch > uint64(time.Now().UnixNano()/int64(time.Millisecond)) {
		return nil, fmt.Errorf("base epoch must be less than or equals to current epoch time")
	}
	// We reset the increment to 0 if it's greater or equals to 4095, which is the max sequence number
	if w.increment >= MAX12BITS {
		w.increment = 0
	}

	spaceflake := new(Spaceflake)
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.increment++

	milliseconds := uint64(math.Floor(microTime() * 1000))
	milliseconds -= w.BaseEpoch

	base := stringPadLeft(decimalBinary(milliseconds), 41, "0")
	nodeID := stringPadLeft(decimalBinary(w.Node.ID), 5, "0")
	workerID := stringPadLeft(decimalBinary(w.ID), 5, "0")
	actualSequence := w.Sequence
	if w.Sequence == 0 {
		actualSequence = w.increment
	}
	sequence := stringPadLeft(decimalBinary(actualSequence), 12, "0")
	id, _ := binaryDecimal("0" + base + nodeID + workerID + sequence)

	spaceflake.baseEpoch = w.BaseEpoch
	spaceflake.binaryID = "0" + base + nodeID + workerID + sequence
	spaceflake.id = id

	return spaceflake, nil
}

// GenerateSpaceflakeAt generates a Spaceflake at a specific time
func (w *Worker) GenerateSpaceflakeAt(at time.Time) (*Spaceflake, error) {
	if w.Node == nil {
		return nil, fmt.Errorf("node is not set")
	}
	if w.Node.ID > MAX5BITS {
		return nil, fmt.Errorf("node ID must be less than or equals to %d", MAX5BITS)
	}
	if w.ID > MAX5BITS {
		return nil, fmt.Errorf("worker ID must be less than or equals to %d", MAX5BITS)
	}
	if w.Sequence > MAX12BITS {
		return nil, fmt.Errorf("sequence must be less than or equals to %d", MAX12BITS)
	}
	if w.BaseEpoch > uint64(time.Now().UnixNano()/int64(time.Millisecond)) {
		return nil, fmt.Errorf("base epoch must be less than or equals to current epoch time")
	}
	if w.BaseEpoch > uint64(at.UnixNano()/int64(time.Millisecond)) {
		return nil, fmt.Errorf("base epoch must be less than or equals to the time you want to generate the Spaceflake at")
	}
	// We reset the increment to 0 if it's greater than 4095, which is the max sequence number
	if w.increment >= MAX12BITS {
		w.increment = 0
	}

	spaceflake := new(Spaceflake)
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.increment++

	microSeconds := float64(at.Nanosecond()) / 1000000000
	microTime := float64(at.Unix()) + microSeconds

	milliseconds := uint64(math.Floor(microTime * 1000))
	milliseconds -= w.BaseEpoch

	base := stringPadLeft(decimalBinary(milliseconds), 41, "0")
	nodeID := stringPadLeft(decimalBinary(w.Node.ID), 5, "0")
	workerID := stringPadLeft(decimalBinary(w.ID), 5, "0")
	actualSequence := w.Sequence
	if w.Sequence == 0 {
		actualSequence = w.increment
	}
	sequence := stringPadLeft(decimalBinary(actualSequence), 12, "0")
	id, _ := binaryDecimal("0" + base + nodeID + workerID + sequence)

	spaceflake.baseEpoch = w.BaseEpoch
	spaceflake.binaryID = "0" + base + nodeID + workerID + sequence
	spaceflake.id = id

	return spaceflake, nil
}

// BulkGenerateSpaceflakes will generate the amount of Spaceflakes specified and scale nothing, it will wait one millisecond if the sequence is over 4095
func (w *Worker) BulkGenerateSpaceflakes(amount int) ([]*Spaceflake, error) {
	worker := w
	spaceflakes := make([]*Spaceflake, amount)
	for i := 1; i <= amount; i++ {
		if i%(MAX12BITS+1) == 0 {
			time.Sleep(1 * time.Millisecond)
		}
		spaceflake, err := generateSpaceflakeOnNodeAndWorker(worker, worker.Node)
		if err != nil {
			return nil, err
		}
		spaceflakes[i-1] = spaceflake
	}
	return spaceflakes, nil
}

// BulkGeneratorSettings is a struct that contains the settings for the bulk Spaceflake generator
type BulkGeneratorSettings struct {
	// Amount is the amount of Spaceflakes to generate
	Amount int
	// BaseEpoch is the base epoch that the Spaceflake generator will use
	BaseEpoch uint64
}

// NewBulkGeneratorSettings is used for the settings of the BulkGenerate function below
func NewBulkGeneratorSettings(amount int) BulkGeneratorSettings {
	return BulkGeneratorSettings{
		Amount:    amount,
		BaseEpoch: EPOCH,
	}
}

// Bulkgenerate will generate the amount of Spaceflakes specified and auto scale the node and worker IDs
func BulkGenerate(s BulkGeneratorSettings) ([]*Spaceflake, error) {
	node := NewNode(1)
	worker := node.NewWorker()
	worker.BaseEpoch = s.BaseEpoch
	spaceflakes := make([]*Spaceflake, s.Amount)
	for i := 1; i <= s.Amount; i++ {
		if i%(MAX12BITS*MAX5BITS*MAX5BITS) == 0 {
			time.Sleep(1 * time.Millisecond)
			newNode := NewNode(1)
			newWorker := newNode.NewWorker()
			newWorker.BaseEpoch = s.BaseEpoch
			node = newNode
			worker = newWorker
		} else if len(node.workers)%MAX5BITS == 0 && i%(MAX5BITS*MAX12BITS) == 0 {
			newNode := NewNode(node.ID + 1)
			newWorker := newNode.NewWorker()
			newWorker.BaseEpoch = s.BaseEpoch
			node = newNode
			worker = newWorker
		} else if i%MAX12BITS == 0 {
			newWorker := worker.Node.NewWorker()
			newWorker.BaseEpoch = s.BaseEpoch
			worker = newWorker
		}
		spaceflake, err := generateSpaceflakeOnNodeAndWorker(worker, worker.Node)
		if err != nil {
			return nil, err
		}
		spaceflakes[i-1] = spaceflake
	}
	return spaceflakes, nil
}

// GeneratorSettings is a struct that contains the settings for the Spaceflake generator
type GeneratorSettings struct {
	// BaseEpoch is the epoch that the Spaceflake generator will use for the first 41 bits
	BaseEpoch uint64
	// NodeID is the node ID that the Spaceflake generator will use for the next 5 bits
	NodeID uint64
	// Sequence is the last 12 bits, usually an incremented number but can be anything. If set to 0, it will be random.
	Sequence uint64
	// WorkerID is the worker ID that the Spaceflake generator will use for the next 5 bits
	WorkerID uint64
}

// NewGeneratorSettings is used for the settings of the Generate function below
func NewGeneratorSettings() GeneratorSettings {
	return GeneratorSettings{
		BaseEpoch: EPOCH,
		NodeID:    0,
		WorkerID:  0,
	}
}

// Generate only a Spaceflake ID when you need to generate one without worker and node objects. Fastest way of creating a Spaceflake.
func Generate(s GeneratorSettings) (*Spaceflake, error) {
	if s.NodeID > MAX5BITS {
		return nil, fmt.Errorf("node ID must be less than or equals to %d", MAX5BITS)
	}
	if s.WorkerID > MAX5BITS {
		return nil, fmt.Errorf("worker ID must be less than or equals to %d", MAX5BITS)
	}
	if s.Sequence > MAX12BITS {
		return nil, fmt.Errorf("sequence must be less than or equals to %d", MAX12BITS)
	}
	if s.BaseEpoch > uint64(time.Now().UnixNano()/int64(time.Millisecond)) {
		return nil, fmt.Errorf("base epoch must be less than or equals to current epoch time")
	}

	spaceflake := new(Spaceflake)

	milliseconds := uint64(math.Floor(microTime() * 1000))
	milliseconds -= s.BaseEpoch

	base := stringPadLeft(decimalBinary(milliseconds), 41, "0")
	nodeID := stringPadLeft(decimalBinary(s.NodeID), 5, "0")
	workerID := stringPadLeft(decimalBinary(s.WorkerID), 5, "0")
	if s.Sequence == 0 {
		s.Sequence = uint64(random(0, MAX12BITS))
	}
	sequence := stringPadLeft(decimalBinary(s.Sequence), 12, "0")
	id, _ := binaryDecimal("0" + base + nodeID + workerID + sequence)

	spaceflake.baseEpoch = s.BaseEpoch
	spaceflake.binaryID = "0" + base + nodeID + workerID + sequence
	spaceflake.id = id

	return spaceflake, nil
}

// GenerateAt generates a Spaceflake at a specific time
func GenerateAt(s GeneratorSettings, at time.Time) (*Spaceflake, error) {
	if s.NodeID > MAX5BITS {
		return nil, fmt.Errorf("node ID must be less than or equals to %d", MAX5BITS)
	}
	if s.WorkerID > MAX5BITS {
		return nil, fmt.Errorf("worker ID must be less than or equals to %d", MAX5BITS)
	}
	if s.Sequence > MAX12BITS {
		return nil, fmt.Errorf("sequence must be less than or equals to %d", MAX12BITS)
	}
	if s.BaseEpoch > uint64(time.Now().UnixNano()/int64(time.Millisecond)) {
		return nil, fmt.Errorf("base epoch must be less than or equals to current epoch time")
	}
	if s.BaseEpoch > uint64(at.UnixNano()/int64(time.Millisecond)) {
		return nil, fmt.Errorf("base epoch must be less than or equals to the time you want to generate the Spaceflake at")
	}

	spaceflake := new(Spaceflake)

	microSeconds := float64(at.Nanosecond()) / 1000000000
	microTime := float64(at.Unix()) + microSeconds

	milliseconds := uint64(math.Floor(microTime * 1000))
	milliseconds -= s.BaseEpoch

	base := stringPadLeft(decimalBinary(milliseconds), 41, "0")
	nodeID := stringPadLeft(decimalBinary(s.NodeID), 5, "0")
	workerID := stringPadLeft(decimalBinary(s.WorkerID), 5, "0")
	if s.Sequence == 0 {
		s.Sequence = uint64(random(0, MAX12BITS))
	}
	sequence := stringPadLeft(decimalBinary(s.Sequence), 12, "0")
	id, _ := binaryDecimal("0" + base + nodeID + workerID + sequence)

	spaceflake.baseEpoch = s.BaseEpoch
	spaceflake.binaryID = "0" + base + nodeID + workerID + sequence
	spaceflake.id = id

	return spaceflake, nil
}

// ParseTime returns the time in milliseconds since the base epoch from the Spaceflake ID
func ParseTime(spaceflakeID, baseEpoch uint64) uint64 {
	return (spaceflakeID >> 22) + baseEpoch
}

// ParseNodeID returns the node ID from the Spaceflake ID
func ParseNodeID(spaceflakeID uint64) uint64 {
	return (spaceflakeID & 0x3E0000) >> 17
}

// ParseWorkerID returns the worker ID from the Spaceflake ID
func ParseWorkerID(spaceflakeID uint64) uint64 {
	return (spaceflakeID & 0x1F000) >> 12
}

// ParseSequence returns the sequence number from the Spaceflake ID
func ParseSequence(spaceflakeID uint64) uint64 {
	return spaceflakeID & 0xFFF
}

// Decompose returns all the parts of the Spaceflake ID
func Decompose(spaceflakeID, baseEpoch uint64) map[string]uint64 {
	return map[string]uint64{
		"id":       spaceflakeID,
		"nodeID":   ParseNodeID(spaceflakeID),
		"sequence": ParseSequence(spaceflakeID),
		"time":     ParseTime(spaceflakeID, EPOCH),
		"workerID": ParseWorkerID(spaceflakeID),
	}
}

// DecomposeBinary returns all the parts of the Spaceflake ID in binary format
func DecomposeBinary(spaceflakeID, baseEpoch uint64) map[string]string {
	return map[string]string{
		"binaryID": stringPadLeft(decimalBinary(spaceflakeID), 64, "0"),
		"nodeID":   stringPadLeft(decimalBinary(ParseNodeID(spaceflakeID)), 5, "0"),
		"sequence": stringPadLeft(decimalBinary(ParseSequence(spaceflakeID)), 12, "0"),
		"time":     stringPadLeft(decimalBinary(ParseTime(spaceflakeID, baseEpoch)), 41, "0"),
		"workerID": stringPadLeft(decimalBinary(ParseWorkerID(spaceflakeID)), 5, "0"),
	}
}

// --- Helper Functions ---

// generateSpaceflakeOnNodeAndWorker generates a Spaceflake on a specific node and worker
func generateSpaceflakeOnNodeAndWorker(w *Worker, n *Node) (*Spaceflake, error) {
	if w.Node == nil {
		return nil, fmt.Errorf("node is not set")
	}
	if w.Node.ID > MAX5BITS {
		return nil, fmt.Errorf("node ID must be less than or equals to %d", MAX5BITS)
	}
	if w.ID > MAX5BITS {
		return nil, fmt.Errorf("worker ID must be less than or equals to %d", MAX5BITS)
	}
	if w.Sequence > MAX12BITS {
		return nil, fmt.Errorf("sequence must be less than or equals to %d", MAX12BITS)
	}
	if w.BaseEpoch > uint64(time.Now().UnixNano()/int64(time.Millisecond)) {
		return nil, fmt.Errorf("base epoch must be less than or equals to current epoch time")
	}
	// We reset the increment to 0 if it's greater than 4095, which is the max sequence number
	if w.increment >= MAX12BITS {
		w.increment = 0
	}

	spaceflake := new(Spaceflake)
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.increment++

	milliseconds := uint64(math.Floor(microTime() * 1000))
	milliseconds -= w.BaseEpoch

	base := stringPadLeft(decimalBinary(milliseconds), 41, "0")
	nodeID := stringPadLeft(decimalBinary(w.Node.ID), 5, "0")
	workerID := stringPadLeft(decimalBinary(w.ID), 5, "0")
	actualSequence := w.Sequence
	if w.Sequence == 0 {
		actualSequence = w.increment
	}
	sequence := stringPadLeft(decimalBinary(actualSequence), 12, "0")
	id, _ := binaryDecimal("0" + base + nodeID + workerID + sequence)

	spaceflake.baseEpoch = w.BaseEpoch
	spaceflake.binaryID = "0" + base + nodeID + workerID + sequence
	spaceflake.id = id

	return spaceflake, nil
}

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
