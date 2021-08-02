package assets

import (
	_ "embed"
)

//go:embed multiboot.elf
var KernelLoader string
