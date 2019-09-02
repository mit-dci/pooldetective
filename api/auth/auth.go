package auth

import (
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"os"
	"strings"

	"github.com/btcsuite/fastsha256"
)

var knownApiKeyHashes []string

func Auth(handler http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKeyBytes := make([]byte, 32)
		apiKeyHash := fastsha256.Sum256([]byte(r.Header.Get("X-Api-Key")))
		copy(apiKeyBytes, apiKeyHash[:])
		apiKeyHex := hex.EncodeToString(apiKeyBytes)

		validAPIKey := false
		for _, k := range knownApiKeyHashes {
			if subtle.ConstantTimeCompare([]byte(apiKeyHex), []byte(k)) == 1 {
				validAPIKey = true
			}
		}

		if !validAPIKey {
			w.WriteHeader(401)
			w.Write([]byte("Unauthorised.\n"))
			return
		}

		handler(w, r)
	})
}

func init() {
	knownApiKeyHashes = strings.Split(os.Getenv("APIKEYS"), "|")
}
