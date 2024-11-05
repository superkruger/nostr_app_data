package compress

import (
	"math/rand"
	"testing"
)

const (
	jsonChars = "{},:[]/-+'!@#$%^&*()?=_|;\"\\\b\f\n\r\t abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

func RandomJSONBytes(length int) []byte {
	sizeChars := len(jsonChars)
	buffer := make([]byte, length)
	for i := range buffer {
		buffer[i] = jsonChars[rand.Intn(sizeChars)]
	}
	return buffer
}

func RandomJSON(length int) string {
	return string(RandomJSONBytes(length))
}

func TestEncodeDecode(t *testing.T) {
	jsonString := RandomJSON(1000)
	encoded, err := Encode(&jsonString)
	if err != nil {
		t.Errorf("problem encoding the string: %v", err)
	}

	t.Logf("Original size: %v, Encoded size: %v", len(jsonString), len(*encoded))

	decoded, err := Decode(encoded)
	if err != nil {
		t.Errorf("problem decoding the string: %v", err)
	}

	if *decoded != jsonString {
		t.Errorf("expected the decoded is equal to the original string, \nOriginal:\n %v \nEncoded:\n %v\nDecoded:\n%v\n", jsonString, *encoded, *decoded)
	}
}

func TestEncodeBytesDecode(t *testing.T) {
	jsonBytes := RandomJSONBytes(1000)
	encoded, err := EncodeBytes(jsonBytes)
	if err != nil {
		t.Errorf("problem encoding the string: %v", err)
	}

	t.Logf("Original size: %v, Encoded size: %v", len(jsonBytes), len(encoded))

	encodedString := string(encoded)
	decoded, err := Decode(&encodedString)
	if err != nil {
		t.Errorf("problem decoding the string: %v", err)
	}

	if *decoded != string(jsonBytes) {
		t.Errorf("expected the decoded is equal to the original string, \nOriginal:\n %v \nEncoded:\n %v\nDecoded:\n%v\n", string(jsonBytes), encoded, *decoded)
	}
}
