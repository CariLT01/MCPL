package data

import (
	_ "embed"
)

//go:embed bin/delta.bin
var DeltaBinary []byte
