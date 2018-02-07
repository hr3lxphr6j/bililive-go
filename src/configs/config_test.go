package configs

import (
	"testing"
)

func TestNewConfig(t *testing.T) {
	t.Log(NewConfigWithFile("../../config.yml"))
}
