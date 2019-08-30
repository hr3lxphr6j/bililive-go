package configs

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	file := "../../config.yml"
	c, err := NewConfigWithFile("../../config.yml")
	assert.NoError(t, err)
	assert.Equal(t, file, c.file)

}

func TestTLS_Verify(t *testing.T) {
	var tls *TLS
	assert.NoError(t, tls.Verify())
	tls = new(TLS)
	assert.NoError(t, tls.Verify())
	tls.Enable = true
	assert.Error(t, tls.Verify())
}

func TestRPC_Verify(t *testing.T) {
	var rpc *RPC
	assert.NoError(t, rpc.Verify())
	rpc = new(RPC)
	rpc.Bind = "foo@bar"
	assert.NoError(t, rpc.Verify())
	rpc.Enable = true
	assert.Error(t, rpc.Verify())
}

func TestConfig_Verify(t *testing.T) {
	var cfg *Config
	assert.Error(t, cfg.Verify())
	cfg = &Config{
		Interval:   30,
		OutPutPath: os.TempDir(),
	}
	assert.NoError(t, cfg.Verify())
	cfg.Interval = 0
	assert.Error(t, cfg.Verify())
	cfg.Interval = 30
	cfg.OutPutPath = "foobar"
	assert.Error(t, cfg.Verify())
}
