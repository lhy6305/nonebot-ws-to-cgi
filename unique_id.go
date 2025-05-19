package main

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"time"
)

var (
	is_random_function_tested bool = false
)

func gen_unique_id() string {
	if !is_random_function_tested {
		test_random_function()
		is_random_function_tested = true
	}
	timestamp := time.Now().UnixNano()
	tsBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(tsBytes, uint64(timestamp))

	randBytes := make([]byte, 2)
	if _, err := rand.Read(randBytes); err != nil {
		randBytes[0] = 0
		randBytes[1] = 0
	}

	combined := append(tsBytes, randBytes...)

	return hex.EncodeToString(combined)
}

func test_random_function() {
	randBytes := make([]byte, 16)
	if _, err := rand.Read(randBytes); err != nil {
		custom_log("Warn", "rand.Read() failed in testing: %v")
	}
}
