package main

import (
	"encoding/pem"
	"errors"
	"log"
	"os"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/pelletier/go-toml/v2"
)

// KeyManager is a struct that contains the private and public keys for the node.
// The private key is used to set the node's identity. Without it, the relay node
// would generate a new identity every time it is started.
type KeyManager struct {
	PrivateKey string
	PublicKey  string

	PrivKey crypto.PrivKey
	PubKey  crypto.PubKey
}

// initKeyManager creates a new key manager if one doesn't exist but returns the existing one if it does
func (km *KeyManager) InitKeyManager() {
	// Check if key file exists. If it does, load it. If it doesn't, create it
	b, err := os.ReadFile("key.toml")

	// If the key file cannot be found, there will be an attempt to create a new one
	if err != nil {
		log.Println("Key file not found. Creating new key file.")
		newKeyManager := CreateKeys()
		km = &newKeyManager
		km.SaveKeys()
		return
	}

	// Parse the toml file
	e := toml.Unmarshal(b, km)
	if e != nil {
		panic(e)
	}

	// Convert the keys to crypto.PrivKey and crypto.PubKey
	km.PrivKey, err = LoadPrivkey(km.PrivateKey)
	if err != nil {
		panic(err)
	}

	km.PubKey, err = LoadPubKey(km.PublicKey)
	if err != nil {
		panic(err)
	}

}

// CreateKeys creates a new key pair and saves it to the key.toml file
func CreateKeys() KeyManager {
	// Generate a new key pair
	priv, pub, err := crypto.GenerateKeyPair(crypto.Ed25519, -1)
	if err != nil {
		panic(err)
	}

	// Convert the keys to strings
	privKey := ExportPrivkeyAsPEMStr(priv)
	pubKey := ExportPubKeyAsPEMStr(pub)

	// Save the keys to the key.toml file
	keyManager := KeyManager{
		PrivateKey: privKey,
		PublicKey:  pubKey,
	}

	return keyManager
}

// ExportPubKeyAsPEMStr converts a public key to a string in PEM format
func ExportPubKeyAsPEMStr(pubkey crypto.PubKey) string {
	key, _ := crypto.MarshalPublicKey(pubkey)
	pubKeyPem := string(pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: key,
		},
	))
	return pubKeyPem
}

// ExportPrivkeyAsPEMStr converts a private key to a string in PEM format
func ExportPrivkeyAsPEMStr(Privkey crypto.PrivKey) string {
	key, _ := crypto.MarshalPrivateKey(Privkey)

	PrivkeyPem := string(pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: key,
		},
	))
	return PrivkeyPem

}

// SaveKeys saves the keys to the key.toml file
func (km *KeyManager) SaveKeys() {
	b, err := toml.Marshal(km)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile("key.toml", b, 0644)
	if err != nil {
		panic(err)
	}

	log.Println("Key file saved.")
}

// LoadPrivkey converts a private key in PEM format to a crypto.Privkey object
func LoadPrivkey(Privkey string) (crypto.PrivKey, error) {
	block, _ := pem.Decode([]byte(Privkey))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	priv, err := crypto.UnmarshalPrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return priv, nil
}

// LoadPubKey converts a public key in PEM format to a crypto.PubKey object
func LoadPubKey(pubPEM string) (crypto.PubKey, error) {
	block, _ := pem.Decode([]byte(pubPEM))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	pub, err := crypto.UnmarshalPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return pub, nil
}
