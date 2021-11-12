package proxy_test

import (
	"os"
	"testing"

	"github.com/mapprotocol/atlas/consensus/istanbul/backend"
	"github.com/mapprotocol/atlas/consensus/istanbul/backend/backendtest"
)

func TestMain(m *testing.M) {
	backendtest.InitTestBackendFactory(backend.TestBackendFactory)
	code := m.Run()
	os.Exit(code)
}
