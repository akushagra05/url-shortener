package snowflake

import (
	"encoding/base64"
	"errors"
	"sync"
	"time"
)

const (
	// Bit lengths
	timestampBits = 41
	machineIDBits = 10
	sequenceBits  = 12

	// Max values
	maxMachineID = -1 ^ (-1 << machineIDBits) // 1023
	maxSequence  = -1 ^ (-1 << sequenceBits)  // 4095

	// Bit shifts
	machineIDShift = sequenceBits
	timestampShift = sequenceBits + machineIDBits
)

var (
	ErrInvalidMachineID = errors.New("machine ID must be between 0 and 1023")
	ErrClockMovedBack   = errors.New("clock moved backwards")
)

// Generator generates unique IDs using the Snowflake algorithm
type Generator struct {
	mu           sync.Mutex
	epoch        int64 // Custom epoch timestamp in milliseconds
	machineID    int64 // 10 bits: 0-1023
	sequence     int64 // 12 bits: 0-4095
	lastTimestamp int64 // Last timestamp used
}

// NewGenerator creates a new Snowflake ID generator
func NewGenerator(machineID int64, epoch int64) (*Generator, error) {
	if machineID < 0 || machineID > maxMachineID {
		return nil, ErrInvalidMachineID
	}

	return &Generator{
		epoch:     epoch,
		machineID: machineID,
		sequence:  0,
		lastTimestamp: 0,
	}, nil
}

// Generate creates a new unique ID
func (g *Generator) Generate() (int64, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	timestamp := g.currentTimestamp()

	// Clock moved backwards
	if timestamp < g.lastTimestamp {
		return 0, ErrClockMovedBack
	}

	// Same millisecond: increment sequence
	if timestamp == g.lastTimestamp {
		g.sequence = (g.sequence + 1) & maxSequence
		
		// Sequence exhausted: wait for next millisecond
		if g.sequence == 0 {
			timestamp = g.waitNextMillis(timestamp)
		}
	} else {
		// New millisecond: reset sequence
		g.sequence = 0
	}

	g.lastTimestamp = timestamp

	// Construct the ID
	id := ((timestamp - g.epoch) << timestampShift) |
		(g.machineID << machineIDShift) |
		g.sequence

	return id, nil
}

// GenerateString generates a unique ID and returns it as a Base62 string
func (g *Generator) GenerateString() (string, error) {
	id, err := g.Generate()
	if err != nil {
		return "", err
	}
	return EncodeBase62(id), nil
}

// currentTimestamp returns current timestamp in milliseconds
func (g *Generator) currentTimestamp() int64 {
	return time.Now().UnixNano() / 1e6
}

// waitNextMillis waits until next millisecond
func (g *Generator) waitNextMillis(currentTimestamp int64) int64 {
	for currentTimestamp <= g.lastTimestamp {
		time.Sleep(time.Millisecond)
		currentTimestamp = g.currentTimestamp()
	}
	return currentTimestamp
}

// Base62 encoding constants
const base62Alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// EncodeBase62 encodes a number to Base62 string (7-8 characters for snowflake IDs)
func EncodeBase62(num int64) string {
	if num == 0 {
		return string(base62Alphabet[0])
	}

	encoded := make([]byte, 0, 11) // Max 11 chars for int64
	base := int64(len(base62Alphabet))

	for num > 0 {
		remainder := num % base
		encoded = append([]byte{base62Alphabet[remainder]}, encoded...)
		num = num / base
	}

	return string(encoded)
}

// DecodeBase62 decodes a Base62 string back to int64
func DecodeBase62(str string) (int64, error) {
	var num int64
	base := int64(len(base62Alphabet))

	for i := 0; i < len(str); i++ {
		char := str[i]
		var value int64 = -1

		switch {
		case char >= '0' && char <= '9':
			value = int64(char - '0')
		case char >= 'A' && char <= 'Z':
			value = int64(char-'A') + 10
		case char >= 'a' && char <= 'z':
			value = int64(char-'a') + 36
		}

		if value == -1 {
			return 0, errors.New("invalid character in Base62 string")
		}

		num = num*base + value
	}

	return num, nil
}

// EncodeBase64URL encodes bytes to URL-safe Base64 (alternative encoding)
func EncodeBase64URL(data []byte) string {
	return base64.URLEncoding.EncodeToString(data)
}

// ParseID extracts timestamp, machine ID, and sequence from a Snowflake ID
func ParseID(id int64) (timestamp int64, machineID int64, sequence int64) {
	sequence = id & maxSequence
	machineID = (id >> machineIDShift) & maxMachineID
	timestamp = id >> timestampShift
	return
}

// GetIDTimestamp extracts just the timestamp from a Snowflake ID
func GetIDTimestamp(id int64, epoch int64) time.Time {
	timestamp, _, _ := ParseID(id)
	ms := timestamp + epoch
	return time.Unix(ms/1000, (ms%1000)*1e6)
}
