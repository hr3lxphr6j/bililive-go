package configs

import (
	"testing"
)

func TestNewConfig(t *testing.T) {
	t.Log(NewConfig("../../config.yml"))
}
