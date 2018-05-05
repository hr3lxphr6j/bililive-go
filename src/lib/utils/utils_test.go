package utils

import "testing"

func TestReplaceIllegalChar(t *testing.T) {
	t.Log(ReplaceIllegalChar(":123.mp4"))
}
