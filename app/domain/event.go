package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"

	"github.com/superkruger/nostr_app_data/app/utils"
)

const (
	EventTypeEOSE  = "EOSE"
	EventTypeEvent = "EVENT"
)

type Subscriber struct {
	ID     string `json:"id"`
	ConnID string `json:"connId"`
}

type ForwardEvent struct {
	Subscribers []Subscriber `json:"subscribers"`
	Event       string       `json:"event"`
}

type Event struct {
	ID        string `json:"id" bson:"id"`
	PubKey    string `json:"pubkey" bson:"pubkey"`
	Kind      int    `json:"kind" bson:"kind"`
	Tags      Tags   `json:"tags" bson:"tags"`
	CreatedAt int    `json:"created_at" bson:"created_at"`
	Content   string `json:"content" bson:"content"`
	Sig       string `json:"sig" bson:"sig"`
}

type DBEvent struct {
	ID        string    `bson:"id"`
	PubKey    string    `bson:"pubkey"`
	Kind      int       `bson:"kind"`
	Tags      DBTags    `bson:"tags"`
	CreatedAt int       `bson:"created_at"`
	Content   string    `bson:"content"`
	Checksum  int       `bson:"checksum"`
	Sig       string    `bson:"sig"`
	ExpireAt  time.Time `bson:"expire_at,omitempty"`
}

func (evt *Event) Serialize() []byte {
	// the serialization process is just putting everything into a JSON array
	// so the order is kept. See NIP-01
	dst := make([]byte, 0)

	// the header portion is easy to serialize
	// [0,"pubkey",created_at,kind,[
	dst = append(dst, []byte(
		fmt.Sprintf(
			"[0,\"%s\",%d,%d,",
			evt.PubKey,
			evt.CreatedAt,
			evt.Kind,
		))...)

	// tags
	dst = evt.Tags.marshalTo(dst)
	dst = append(dst, ',')

	// content needs to be escaped in general as it is user generated.
	dst = utils.EscapeString(dst, evt.Content)
	dst = append(dst, ']')

	return dst
}

// CheckSignature checks if the signature is valid for the id
// (which is a hash of the serialized event content).
// returns an error if the signature itself is invalid.
func (evt *Event) CheckSignature() (bool, error) {
	// read and check pubkey
	pk, err := hex.DecodeString(evt.PubKey)
	if err != nil {
		return false, fmt.Errorf("event pubkey '%s' is invalid hex: %w", evt.PubKey, err)
	}

	pubkey, err := schnorr.ParsePubKey(pk)
	if err != nil {
		return false, fmt.Errorf("event has invalid pubkey '%s': %w", evt.PubKey, err)
	}

	// read signature
	s, err := hex.DecodeString(evt.Sig)
	if err != nil {
		return false, fmt.Errorf("signature '%s' is invalid hex: %w", evt.Sig, err)
	}
	sig, err := schnorr.ParseSignature(s)
	if err != nil {
		return false, fmt.Errorf("failed to parse signature: %w", err)
	}

	// check signature
	hash := sha256.Sum256(evt.Serialize())
	return sig.Verify(hash[:], pubkey), nil
}

// Sign signs an event with a given privateKey.
func (evt *Event) Sign(secretKey string, signOpts ...schnorr.SignOption) error {
	s, err := hex.DecodeString(secretKey)
	if err != nil {
		return fmt.Errorf("Sign called with invalid secret key '%s': %w", secretKey, err)
	}

	if evt.Tags == nil {
		evt.Tags = make(Tags, 0)
	}

	sk, pk := btcec.PrivKeyFromBytes(s)
	pkBytes := pk.SerializeCompressed()
	evt.PubKey = hex.EncodeToString(pkBytes[1:])

	h := sha256.Sum256(evt.Serialize())
	sig, err := schnorr.Sign(sk, h[:], signOpts...)
	if err != nil {
		return err
	}

	evt.ID = hex.EncodeToString(h[:])
	evt.Sig = hex.EncodeToString(sig.Serialize())

	return nil
}

func (evt *Event) ToDB() DBEvent {
	return DBEvent{
		ID:        evt.ID,
		PubKey:    evt.PubKey,
		Kind:      evt.Kind,
		Tags:      evt.Tags.toDB(),
		CreatedAt: evt.CreatedAt,
		Content:   evt.Content,
		Checksum:  int(crc32.Checksum([]byte(evt.Content), crc32.IEEETable)),
		Sig:       evt.Sig,
	}
}

func (evt *DBEvent) IsRegular() bool {
	return (evt.Kind >= 1000 && evt.Kind < 10000) || (evt.Kind >= 4 && evt.Kind < 45) || evt.Kind == 1 || evt.Kind == 2
}

func (evt *DBEvent) IsReplaceable() bool {
	return (evt.Kind >= 10000 && evt.Kind < 20000) || evt.Kind == 0 || evt.Kind == 3
}

func (evt *DBEvent) IsAddressable() bool {
	dTag, ok := evt.Tags["d"]
	return evt.Kind >= 30000 && evt.Kind < 40000 && ok && len(dTag.Values) > 0
}

func (evt *DBEvent) ToJson() Event {
	return Event{
		ID:        evt.ID,
		PubKey:    evt.PubKey,
		Kind:      evt.Kind,
		Tags:      evt.Tags.toJson(),
		CreatedAt: evt.CreatedAt,
		Content:   evt.Content,
		Sig:       evt.Sig,
	}
}
