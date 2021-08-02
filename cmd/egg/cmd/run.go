/*
Copyright © 2021 fanbingxin <fanbingxin.me@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/icexin/eggos/cmd/egg/assets"
	"github.com/spf13/cobra"
)

const (
	qemu64 = "qemu-system-x86_64"
)

var (
	kernelFile  string
	showgraphic bool
	envs        []string
	ports       []string
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run running a eggos kernel in qemu",
	Run: func(cmd *cobra.Command, args []string) {
		runKernel()
	},
}

func runKernel() {
	base, err := ioutil.TempDir("", "eggos-run")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(base)
	if kernelFile == "" {
		log.Fatal("missing kernel file")
	}

	var runArgs []string

	ext := filepath.Ext(kernelFile)
	switch ext {
	case ".elf":
		loaderFile := filepath.Join(base, "loader.elf")
		mustLoaderFile(loaderFile)
		runArgs = append(runArgs, "-kernel", loaderFile)
		runArgs = append(runArgs, "-initrd", kernelFile)
		runArgs = append(runArgs, "-append", strings.Join(envs, " "))
	case ".iso":
		runArgs = append(runArgs, "-cdrom", kernelFile)
	default:
		log.Fatalf("unsupported file ext:%s", ext)
	}

	runArgs = append(runArgs, "-m", "256M", "-no-reboot", "-serial", "mon:stdio")
	runArgs = append(runArgs, "-netdev", "user,id=eth0"+portMapingArgs())
	runArgs = append(runArgs, "-device", "e1000,netdev=eth0")
	runArgs = append(runArgs, "-device", "isa-debug-exit")

	if !showgraphic {
		runArgs = append(runArgs, "-nographic")
	}

	cmd := exec.Command(qemu64, runArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
	if err == nil {
		return
	}
	exiterr := err.(*exec.ExitError)
	os.Exit(exiterr.ExitCode())
}

func mustLoaderFile(fname string) {
	content, err := assets.Boot.ReadFile("boot/multiboot.elf")
	if err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile(fname, content, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func portMapingArgs() string {
	var ret []string
	for _, mapping := range ports {
		fs := strings.Split(mapping, ":")
		if len(fs) < 2 {
			continue
		}
		arg := fmt.Sprintf(",hostfwd=tcp::%s-:%s", fs[0], fs[1])
		ret = append(ret, arg)
	}
	return strings.Join(ret, "")
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringVarP(&kernelFile, "kernel", "k", "", "eggos kernel file, kernel.elf|eggos.iso")
	runCmd.Flags().BoolVarP(&showgraphic, "graphic", "g", false, "show qemu graphic window")
	runCmd.Flags().StringSliceVarP(&envs, "env", "e", nil, "env passed to kernel")
	runCmd.Flags().StringSliceVarP(&ports, "port", "p", nil, "port mapping from host to kernel, format $host_port:$kernel_port")
}
