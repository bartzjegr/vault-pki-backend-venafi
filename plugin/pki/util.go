package pki

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/hashicorp/vault/logical"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

func sliceContains(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}

	_, ok := set[item]
	return ok
}

func getHexFormatted(buf []byte, sep string) string {
	var ret bytes.Buffer
	for _, cur := range buf {
		if ret.Len() > 0 {
			fmt.Fprintf(&ret, sep)
		}
		fmt.Fprintf(&ret, "%02x", cur)
	}
	return ret.String()
}

func normalizeSerial(serial string) string {
	return strings.Replace(strings.ToLower(serial), ":", "-", -1)
}

func createBackendWithStorage(t *testing.T) (*backend, logical.Storage) {
	config := logical.TestBackendConfig()
	config.StorageView = &logical.InmemStorage{}

	var err error
	b := Backend(config)
	err = b.Setup(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	return b, config.StorageView
}

func getPrivateKeyPEMBock(key interface{}) (*pem.Block, error) {
	switch k := key.(type) {
	case *rsa.PrivateKey:
		return &pem.Block{Type: PKCS1Block, Bytes: x509.MarshalPKCS1PrivateKey(k)}, nil
	case *ecdsa.PrivateKey:
		b, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			return nil, err
		}
		return &pem.Block{Type: ECBlock, Bytes: b}, nil
	default:
		return nil, fmt.Errorf("Unable to format Key")
	}
}

func randSeq(n int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	var letters = []rune("abcdefghijklmnopqrstuvwxyz1234567890")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

const (
	PKCS1Block string = "RSA PRIVATE KEY"
	PKCS8Block string = "PRIVATE KEY"
	ECBlock    string = "EC PRIVATE KEY"
)

type RunContext struct {
	TPPurl              string
	TPPuser             string
	TPPPassword         string
	TPPZone             string
	CloudUrl            string
	CloudAPIkey         string
	CloudZone           string
	TPPTestingEnabled   bool
	CloudTestingEnabled bool
	FakeTestingEnabled  bool
	TPPIssuerCN         string
	CloudIssuerCN       string
	FakeIssuerCN        string
}

func GetContext() *RunContext {

	c := RunContext{}

	c.TPPurl = os.Getenv("TPPURL")
	c.TPPuser = os.Getenv("TPPUSER")
	c.TPPPassword = os.Getenv("TPPPASSWORD")
	c.TPPZone = os.Getenv("TPPZONE")

	c.CloudUrl = os.Getenv("CLOUDURL")
	c.CloudAPIkey = os.Getenv("CLOUDAPIKEY")
	c.CloudZone = os.Getenv("CLOUDZONE")
	c.TPPTestingEnabled, _ = strconv.ParseBool(os.Getenv("VENAFI_TPP_TESTING"))
	c.CloudTestingEnabled, _ = strconv.ParseBool(os.Getenv("VENAFI_CLOUD_TESTING"))
	c.FakeTestingEnabled, _ = strconv.ParseBool(os.Getenv("VENAFI_FAKE_TESTING"))
	c.TPPIssuerCN = os.Getenv("TPP_ISSUER_CN")
	c.CloudIssuerCN = os.Getenv("CLOUD_ISSUER_CN")
	c.FakeIssuerCN = os.Getenv("FAKE_ISSUER_CN")

	return &c
}

func SameStringSlice(x, y []string) bool {
	if len(x) != len(y) {
		return false
	}
	// create a map of string -> int
	diff := make(map[string]int, len(x))
	for _, _x := range x {
		// 0 value for int is 0, so just increment a counter for the string
		diff[_x]++
	}
	for _, _y := range y {
		// If the string _y is not in diff bail out early
		if _, ok := diff[_y]; !ok {
			return false
		}
		diff[_y] -= 1
		if diff[_y] == 0 {
			delete(diff, _y)
		}
	}
	return len(diff) == 0
}
