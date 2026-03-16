package nutshell

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"filippo.io/edwards25519"
	"golang.org/x/crypto/nacl/box"
)

const (
	MagicEncrypted = "NUT\x02" // encrypted bundle magic
	nonceSize      = 24
	keySize        = 32
)

// EncryptBundle encrypts a .nut bundle for a specific recipient.
//
// Format: "NUT\x02" (4) + ephemeralPub (32) + nonce (24) + ciphertext
func EncryptBundle(plainPath, outPath string, recipientEd25519Pub ed25519.PublicKey) error {
	plain, err := os.ReadFile(plainPath)
	if err != nil {
		return fmt.Errorf("reading bundle: %w", err)
	}

	// Convert recipient Ed25519 public key to X25519
	recipientX25519, err := ed25519PubToX25519(recipientEd25519Pub)
	if err != nil {
		return fmt.Errorf("converting recipient key: %w", err)
	}

	// Generate ephemeral X25519 keypair for one-time use
	ephPub, ephPriv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return fmt.Errorf("generating ephemeral key: %w", err)
	}

	// Random nonce
	var nonce [nonceSize]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return fmt.Errorf("generating nonce: %w", err)
	}

	// Encrypt with NaCl box (X25519 + XSalsa20-Poly1305)
	var recipientKey [keySize]byte
	copy(recipientKey[:], recipientX25519)
	ciphertext := box.Seal(nil, plain, &nonce, &recipientKey, ephPriv)

	// Write: magic + ephemeral pub + nonce + ciphertext
	out, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("creating output: %w", err)
	}
	defer out.Close()

	if _, err := out.Write([]byte(MagicEncrypted)); err != nil {
		return err
	}
	if _, err := out.Write(ephPub[:]); err != nil {
		return err
	}
	if _, err := out.Write(nonce[:]); err != nil {
		return err
	}
	if _, err := out.Write(ciphertext); err != nil {
		return err
	}
	return nil
}

// DecryptBundle decrypts an encrypted .nut bundle using the recipient's Ed25519 private key.
// Returns the decrypted plaintext bytes (a valid NUT\x01 bundle).
func DecryptBundle(encPath string, privKey ed25519.PrivateKey) ([]byte, error) {
	data, err := os.ReadFile(encPath)
	if err != nil {
		return nil, fmt.Errorf("reading encrypted bundle: %w", err)
	}

	minSize := 4 + keySize + nonceSize + box.Overhead
	if len(data) < minSize {
		return nil, errors.New("encrypted bundle too short")
	}
	if string(data[:4]) != MagicEncrypted {
		return nil, errors.New("not an encrypted nutshell bundle")
	}

	// Parse header
	var ephPub [keySize]byte
	copy(ephPub[:], data[4:4+keySize])
	var nonce [nonceSize]byte
	copy(nonce[:], data[4+keySize:4+keySize+nonceSize])
	ciphertext := data[4+keySize+nonceSize:]

	// Convert Ed25519 private key to X25519
	ourX25519 := ed25519PrivToX25519(privKey)
	var ourKey [keySize]byte
	copy(ourKey[:], ourX25519)

	// Decrypt
	plain, ok := box.Open(nil, ciphertext, &nonce, &ephPub, &ourKey)
	if !ok {
		return nil, errors.New("decryption failed — wrong key or corrupted bundle")
	}

	return plain, nil
}

// IsEncryptedBundle checks if a file starts with the encrypted magic bytes.
func IsEncryptedBundle(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	magic := make([]byte, 4)
	if _, err := io.ReadFull(f, magic); err != nil {
		return false
	}
	return string(magic) == MagicEncrypted
}

// LoadClawNetIdentity loads the Ed25519 private key from ClawNet's default data directory.
func LoadClawNetIdentity() (ed25519.PrivateKey, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	keyPath := filepath.Join(home, ".openclaw", "clawnet", "identity.key")
	return LoadIdentityKey(keyPath)
}

// LoadIdentityKey reads a libp2p-marshaled Ed25519 private key file.
func LoadIdentityKey(path string) (ed25519.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading identity key: %w", err)
	}
	return extractEd25519Key(data)
}

// extractEd25519Key tries to parse an Ed25519 private key from various formats.
func extractEd25519Key(data []byte) (ed25519.PrivateKey, error) {
	// Format 1: Raw 64-byte Ed25519 key (seed + public)
	if len(data) == ed25519.PrivateKeySize {
		return ed25519.PrivateKey(data), nil
	}
	// Format 2: Raw 32-byte seed
	if len(data) == ed25519.SeedSize {
		return ed25519.NewKeyFromSeed(data), nil
	}
	// Format 3: libp2p protobuf wrapper — scan for 64-byte Ed25519 key inside
	if len(data) > ed25519.PrivateKeySize {
		for i := 0; i <= len(data)-ed25519.PrivateKeySize; i++ {
			candidate := data[i : i+ed25519.PrivateKeySize]
			seed := candidate[:ed25519.SeedSize]
			full := ed25519.NewKeyFromSeed(seed)
			if ed25519.PublicKey(full[32:]).Equal(ed25519.PublicKey(candidate[32:])) {
				return full, nil
			}
		}
	}
	return nil, fmt.Errorf("unrecognized key format (size=%d)", len(data))
}

// ParsePeerPubKey parses a hex-encoded Ed25519 public key (64 hex chars = 32 bytes).
func ParsePeerPubKey(s string) (ed25519.PublicKey, error) {
	decoded, err := hex.DecodeString(s)
	if err != nil || len(decoded) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("expected 64-char hex Ed25519 public key, got %q", s)
	}
	return ed25519.PublicKey(decoded), nil
}

// ed25519PubToX25519 converts an Ed25519 public key to X25519 using filippo.io/edwards25519.
func ed25519PubToX25519(pub ed25519.PublicKey) ([]byte, error) {
	p, err := new(edwards25519.Point).SetBytes(pub)
	if err != nil {
		return nil, fmt.Errorf("invalid Ed25519 public key: %w", err)
	}
	return p.BytesMontgomery(), nil
}

// ed25519PrivToX25519 converts an Ed25519 private key to X25519 by hashing the seed.
func ed25519PrivToX25519(priv ed25519.PrivateKey) []byte {
	h := sha512.Sum512(priv.Seed())
	h[0] &= 248
	h[31] &= 127
	h[31] |= 64
	return h[:32]
}
