package utils

import (
	"path/filepath"
	"testing"

	"github.com/mapprotocol/atlas/cmd/node"
	"github.com/mapprotocol/atlas/consensus/istanbul"
)

func TestIstanbulDBPathDefaults(t *testing.T) {
	cfg := istanbul.DefaultConfig
	for name, path := range map[string]string{
		"RoundStateDBPath":         cfg.RoundStateDBPath,
		"ValidatorEnodeDBPath":     cfg.ValidatorEnodeDBPath,
		"VersionCertificateDBPath": cfg.VersionCertificateDBPath,
	} {
		if path == "" {
			t.Errorf("default %s must not be empty, otherwise the database is not persisted", name)
		}
	}
	if cfg.ReplicaStateDBPath != "" {
		t.Errorf("default ReplicaStateDBPath must stay empty (in memory), have %q", cfg.ReplicaStateDBPath)
	}
}

func TestResolveDBPath(t *testing.T) {
	dir := t.TempDir()
	stack, err := node.New(&node.Config{Name: "atlas", DataDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	defer stack.Close()

	if have, want := resolveDBPath(stack, "roundstates"), filepath.Join(dir, "atlas", "roundstates"); have != want {
		t.Errorf("relative path: have %q, want %q", have, want)
	}
	if have := resolveDBPath(stack, ""); have != "" {
		t.Errorf("empty path must stay empty, have %q", have)
	}
	abs := filepath.Join(dir, "custom")
	if have := resolveDBPath(stack, abs); have != abs {
		t.Errorf("absolute path: have %q, want %q", have, abs)
	}
	// Resolving twice is a no-op.
	once := resolveDBPath(stack, "roundstates")
	if have := resolveDBPath(stack, once); have != once {
		t.Errorf("resolve twice: have %q, want %q", have, once)
	}

	// Without a data directory the databases stay in memory.
	ephemeral, err := node.New(&node.Config{Name: "atlas"})
	if err != nil {
		t.Fatal(err)
	}
	defer ephemeral.Close()
	if have := resolveDBPath(ephemeral, "roundstates"); have != "" {
		t.Errorf("no datadir: have %q, want empty", have)
	}
}
