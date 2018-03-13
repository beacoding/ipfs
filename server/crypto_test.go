package server

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"proj2_f5w9a_h6v9a_q7w9a_r8u8_w1c0b/serverpb"
	"testing"
)

func marshal(t *testing.T, a interface{}) string {
	data, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestLoadOrGenerateCert(t *testing.T) {
	dir, err := ioutil.TempDir("", "ipfs-server-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	c := serverpb.NodeConfig{
		Path: dir,
	}

	s, err := New(c)
	if err != nil {
		t.Fatal(err)
	}
	if s.key == nil {
		t.Fatal("expected key to be non-nil")
	}
	if s.cert == nil {
		t.Fatal("expected cert to be non-nil")
	}
	if s.certPublic == "" {
		t.Fatal("expected certPublic to be non-nil")
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}

	s2, err := New(c)
	if err != nil {
		t.Fatal(err)
	}
	if s.certPublic != s2.certPublic {
		t.Fatal("certPublic wasn't correctly restored")
	}

	want := marshal(t, s.key)
	got := marshal(t, s2.key)
	if want != got {
		t.Fatalf("key wasn't correctly restored; got %q; want %q", got, want)
	}

	if marshal(t, s.key.Public()) != marshal(t, s2.key.Public()) {
		t.Fatal("public key wasn't correctly restored")
	}

	if marshal(t, s.cert) != marshal(t, s2.cert) {
		t.Fatal("cert wasn't correctly restored")
	}

	if err := s2.Close(); err != nil {
		t.Fatal(err)
	}
}
