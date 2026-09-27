// extension-sign signs an independently built package without exposing its key.
package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"blora.dev/panel/internal/extensions"
)

func main() {
	input := flag.String("package", "", "input package JSON")
	key := flag.String("key", "", "Ed25519 PKCS#8 PEM private key file")
	output := flag.String("out", "", "new signed package file (must not exist)")
	flag.Parse()
	if err := run(*input, *key, *output); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(input, keyPath, output string) error {
	if input == "" || keyPath == "" || output == "" {
		return errors.New("--package, --key and --out are required")
	}
	keyBytes, err := readBounded(keyPath, 16<<10)
	if err != nil {
		return err
	}
	block, rest := pem.Decode(keyBytes)
	if block == nil || block.Type != "PRIVATE KEY" || len(bytes.TrimSpace(rest)) != 0 {
		return errors.New("expected one PKCS#8 private key PEM block")
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return errors.New("invalid PKCS#8 private key")
	}
	private, ok := parsed.(ed25519.PrivateKey)
	if !ok {
		return errors.New("key must be Ed25519")
	}
	data, err := readBounded(input, 16<<20)
	if err != nil {
		return err
	}
	var envelope struct {
		Manifest  extensions.Manifest `json:"manifest"`
		SHA256    string              `json:"sha256"`
		Payload   []byte              `json:"payload"`
		Signature []byte              `json:"signature,omitempty"`
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&envelope); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return errors.New("package has trailing JSON")
	}
	h := sha256.Sum256(envelope.Payload)
	if !strings.EqualFold(hex.EncodeToString(h[:]), envelope.SHA256) {
		return errors.New("package payload digest mismatch")
	}
	if _, err := extensions.DecodeBundle(envelope.Payload); err != nil {
		return err
	}
	p := extensions.Package{Manifest: envelope.Manifest, SHA256: envelope.SHA256, Payload: envelope.Payload}
	envelope.Signature = ed25519.Sign(private, extensions.SignatureMessage(p))
	encoded, err := json.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return err
	}
	file, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	_, writeErr := file.Write(append(encoded, '\n'))
	if writeErr == nil {
		writeErr = file.Sync()
	}
	closeErr := file.Close()
	if writeErr != nil {
		_ = os.Remove(output)
		return writeErr
	}
	if closeErr != nil {
		_ = os.Remove(output)
		return closeErr
	}
	fmt.Printf("Signed package: %s\nPublic key (Master --extensions-public-key): %s\n", output, hex.EncodeToString(private.Public().(ed25519.PublicKey)))
	return nil
}

func readBounded(path string, limit int64) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, errors.New("input exceeds size limit")
	}
	return data, nil
}
