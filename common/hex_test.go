package common

import (
	"strconv"
	"testing"
	"time"
)

func TestFormat(t *testing.T) {
	m := time.Duration(3000000000000)
	v := strconv.FormatFloat(m.Seconds(), 'f', -1, 64)
	t.Logf("v: %s", v)
	hd := []uint8{0x12, 0x23, 0x34, 0x45}
	t.Logf("hd: %x", hd)
}
