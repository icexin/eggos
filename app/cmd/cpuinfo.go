package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/icexin/eggos/app"
	"github.com/klauspost/cpuid"
)

type cpuinfo struct {
	*cpuid.CPUInfo
	Features string
}

func cpuinfoMain(ctx *app.Context) error {
	info := &cpuinfo{
		CPUInfo:  &cpuid.CPU,
		Features: cpuid.CPU.Features.String(),
	}
	buf, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return err
	}
	fmt.Fprintf(ctx.Stdout, "%s\n", buf)
	return nil
}

func init() {
	app.Register("cpuinfo", cpuinfoMain)
}
