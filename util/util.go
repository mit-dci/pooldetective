package util

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// NewJobID Returns a new stratum job id for internal use
func NewJobID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// NewClientID returns a new ID for internal references to client connections
func NewClientID() string {
	// Can use the same? Probably.
	return NewJobID()
}

func GetLoglevelFromEnv() int {
	logLevel := int(0)
	logLevel64, err := strconv.ParseInt(os.Getenv("LOGLEVEL"), 10, 32)
	if err == nil {
		logLevel = int(logLevel64)
	}
	return logLevel
}

func DeserializeJsonMap(jsonMap map[string]interface{}, target interface{}) error {
	j, err := json.Marshal(jsonMap)
	if err != nil {
		return fmt.Errorf("Error serializing jsonMap back to json: %s", err.Error())
	}
	return json.Unmarshal(j, target)
}

func ReverseByteArray(b []byte) []byte {
	for i := len(b)/2 - 1; i >= 0; i-- {
		opp := len(b) - 1 - i
		b[i], b[opp] = b[opp], b[i]
	}
	return b
}

func DecodeStratumHash(hash []byte) (*chainhash.Hash, error) {
	if len(hash) < 32 {
		return nil, errors.New("Hash should be 32 bytes")
	}

	// Decode the stratum endianness
	newHash := make([]byte, 0)
	for i := 28; i >= 0; i -= 4 {
		newHash = append(newHash, hash[i:i+4]...)
	}

	// Reverse it in order to make it a native chainhash format
	// (we store all of our stuff in this)
	newHash = ReverseByteArray(newHash)

	return chainhash.NewHash(newHash)
}

func RevHashBytes(hash []byte) []byte {
	if len(hash) < 32 {
		return hash
	}
	newHash := make([]byte, 0)
	for i := 28; i >= 0; i -= 4 {
		newHash = append(newHash, hash[i:i+4]...)
	}
	return newHash
}

func RevHash(hash string) string {
	hashBytes, _ := hex.DecodeString(hash)
	return hex.EncodeToString(RevHashBytes(hashBytes))
}

func HashToString(h []byte) string {
	hsh, err := chainhash.NewHash(h)
	if err != nil {
		return ""
	}
	return hsh.String()
}
