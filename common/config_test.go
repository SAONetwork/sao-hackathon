package common

import (
	"testing"
)

func TestConfig(t *testing.T) {
	_, err := GetConfig("./test.toml")
	if err != nil {
		t.Fatal("failed to decode config file", err)
	}
}
