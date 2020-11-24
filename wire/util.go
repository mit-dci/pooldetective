package wire

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	binser "github.com/kelindar/binary"
)

var aesGcm cipher.AEAD
var encryptTraffic bool

func msgFromFrames(frames [][]byte) (PoolDetectiveMsg, error) {
	if len(frames) != 2 {
		return nil, fmt.Errorf("Expected two or more frames, got %d", len(frames))
	}
	var err error
	if encryptTraffic {
		frames[0], err = decrypt(frames[0])
		if err != nil {
			return nil, err
		}
		frames[1], err = decrypt(frames[1])
		if err != nil {
			return nil, err
		}
	}

	var mt MessageType
	err = binser.Unmarshal(frames[0], &mt)
	if err != nil {
		return nil, err
	}
	msg := NewMessage(mt)
	err = binser.Unmarshal(frames[1], msg)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func msgToFrames(msg PoolDetectiveMsg) (frame [][]byte, err error) {
	frame = make([][]byte, 2)
	frame[0], err = binser.Marshal(GetMessageType(msg))
	if err != nil {
		return
	}
	frame[1], err = binser.Marshal(msg)
	if err != nil {
		return
	}
	if encryptTraffic {
		frame[0], err = encrypt(frame[0])
		if err != nil {
			return
		}
		frame[1], err = encrypt(frame[1])
		if err != nil {
			return
		}
	}
	return
}

func encrypt(b []byte) ([]byte, error) {

	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := aesGcm.Seal(nil, nonce, b, nil)

	result := make([]byte, 12+len(ciphertext))
	copy(result, nonce)
	copy(result[12:], ciphertext)
	return result, nil
}

func decrypt(b []byte) ([]byte, error) {
	if len(b) <= 12 {
		return nil, fmt.Errorf("Nonce missing or unencrypted data received")
	}
	nonce := b[0:12]
	return aesGcm.Open(nil, nonce, b[12:], nil)
}

func init() {
	encryptTraffic = os.Getenv("COMM_ENC") == "1"
	if encryptTraffic {
		key, err := hex.DecodeString(os.Getenv("COMM_ENC_KEY"))
		if err != nil {
			panic(err.Error())
		}

		block, err := aes.NewCipher(key)
		if err != nil {
			panic(err.Error())
		}

		aesGcm, err = cipher.NewGCM(block)
		if err != nil {
			panic(err.Error())
		}
	}
}
