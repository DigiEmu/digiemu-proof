package main

import (
"crypto"
"crypto/rsa"
"crypto/sha256"
"crypto/x509"
"encoding/base64"
"encoding/json"
"encoding/pem"
"errors"
"os"
"testing"
)

func TestTBNVerifySignature(t *testing.T) {
responsePath := "testdata/tbn_verify_response.local.json"
publicKeyPath := "testdata/tbn_public_key.local.json"

responseRaw, err := os.ReadFile(responsePath)
if err != nil {
t.Skipf("local TBN response not found, skipping signature verification: %v", err)
}

publicKeyRaw, err := os.ReadFile(publicKeyPath)
if err != nil {
t.Skipf("local TBN public key not found, skipping signature verification: %v", err)
}

var response map[string]any
if err := json.Unmarshal(responseRaw, &response); err != nil {
t.Fatalf("failed to parse TBN response JSON: %v", err)
}

signatureValue, ok := response["tbn_signature"].(string)
if !ok || signatureValue == "" {
t.Fatalf("missing or invalid tbn_signature")
}

signature, err := base64.StdEncoding.DecodeString(signatureValue)
if err != nil {
t.Fatalf("failed to base64-decode tbn_signature: %v", err)
}

delete(response, "tbn_signature")

canonical, err := json.Marshal(response)
if err != nil {
t.Fatalf("failed to canonicalize TBN response: %v", err)
}

publicKey, err := parseTBNPublicKey(publicKeyRaw)
if err != nil {
t.Fatalf("failed to parse TBN public key: %v", err)
}

digest := sha256.Sum256(canonical)

err = rsa.VerifyPSS(
publicKey,
crypto.SHA256,
digest[:],
signature,
&rsa.PSSOptions{
SaltLength: rsa.PSSSaltLengthAuto,
Hash:       crypto.SHA256,
},
)
if err != nil {
t.Fatalf("TBN RSA-PSS-SHA256 signature verification failed: %v\ncanonical payload:\n%s", err, string(canonical))
}

t.Logf("TBN RSA-PSS-SHA256 signature verified successfully")
}

func parseTBNPublicKey(raw []byte) (*rsa.PublicKey, error) {
var wrapped map[string]any
if err := json.Unmarshal(raw, &wrapped); err == nil {
for _, key := range []string{"public_key", "public_key_pem", "publicKey", "publicKeyPem", "pem", "key"} {
if value, ok := wrapped[key].(string); ok && value != "" {
raw = []byte(value)
break
}
}
}

block, _ := pem.Decode(raw)
if block == nil {
return nil, errors.New("failed to decode PEM public key")
}

parsed, err := x509.ParsePKIXPublicKey(block.Bytes)
if err == nil {
if rsaKey, ok := parsed.(*rsa.PublicKey); ok {
return rsaKey, nil
}
}

rsaKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
if err != nil {
return nil, err
}

return rsaKey, nil
}
