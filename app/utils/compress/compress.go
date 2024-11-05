package compress

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
)

// Encode will GZip a string and Base64 encode it.
// This is usefull for making large JSON content fit in messages.
func Encode(stringToCompressAndEncode *string) (*string, error) {
	var compressBuffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressBuffer)
	if _, err := gzipWriter.Write([]byte(*stringToCompressAndEncode)); err != nil {
		log.WithField("stringToCompressAndEncode", stringToCompressAndEncode).Error("Could not Gzip.")
		return nil, err
	}
	if err := gzipWriter.Close(); err != nil {
		log.Error("Could not close Gzip.")
		return nil, err
	}
	compressedAndEncoded := base64.StdEncoding.EncodeToString(compressBuffer.Bytes())
	return &compressedAndEncoded, nil
}

// EncodeBytes will GZip and Base64 encode the bytes.
func EncodeBytes(toEncode []byte) ([]byte, error) {
	var compressBuffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressBuffer)
	if _, err := gzipWriter.Write(toEncode); err != nil {
		log.WithField("toEncode", string(toEncode)).Error("problem gzipping")
		return nil, err
	}
	if err := gzipWriter.Close(); err != nil {
		log.WithField("toEncode", string(toEncode)).Error("problem closing gzip")
		return nil, err
	}
	src := compressBuffer.Bytes()
	buf := make([]byte, base64.StdEncoding.EncodedLen(len(src)))
	base64.StdEncoding.Encode(buf, src)
	return buf, nil
}

// Decode will base64 decode a string and unzip it.
// This is usefull for making large JSON content fit in messages.
func Decode(stringToDecodeAndExpand *string) (*string, error) {
	stringToExpand, err := base64.StdEncoding.DecodeString(*stringToDecodeAndExpand)
	if err != nil {
		log.WithField("stringToDecodeAndExpand", stringToDecodeAndExpand).Error("Could not Decode Base64.")
		return nil, err
	}
	stringBytes := bytes.NewReader(stringToExpand)
	gzipReader, err := gzip.NewReader(stringBytes)
	if err != nil {
		log.WithField("stringBytes", stringBytes).Error("Could not unzip the string.")
		return nil, err
	}
	decodedAndExpanded, err := ioutil.ReadAll(gzipReader)
	if err != nil {
		log.WithField("stringBytes", stringBytes).Error("Could not unzip bytes.")
		return nil, err
	}
	decodedAndExpandedString := string(decodedAndExpanded)
	return &decodedAndExpandedString, nil
}
