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

	"github.com/google/shlex"
	"github.com/icexin/eggos/cmd/egg/assets"
	"github.com/icexin/eggos/cmd/egg/build"
	"github.com/spf13/cobra"
)

const (
	qemu64 = "qemu-system-x86_64"
)

var (
	ports []string
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run <kernel>",
	Short: "run running a eggos kernel in qemu",
	Run: func(cmd *cobra.Command, args []string) {
		err := runKernel(args)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func runKernel(args []string) error {
	base, err := ioutil.TempDir("", "eggos-run")
	if err != nil {
		return err
	}
	defer os.RemoveAll(base)

	var kernelFile string

	if len(args) == 0 || args[0] == "" {
		kernelFile = filepath.Join(base, "kernel.elf")

		b := build.NewBuilder(build.Config{
			GoRoot:       goroot,
			Basedir:      base,
			BuildTest:    false,
			EggosVersion: eggosVersion,
			GoArgs: []string{
				"-o", kernelFile,
			},
		})
		if err := b.Build(); err != nil {
			return fmt.Errorf("error building kernel: %s", err)
		}
	} else {
		kernelFile = args[0]
	}

	var runArgs []string

	ext := filepath.Ext(kernelFile)
	switch ext {
	case ".elf", "":
		loaderFile := filepath.Join(base, "loader.elf")
		mustLoaderFile(loaderFile)
		runArgs = append(runArgs, "-kernel", loaderFile)
		runArgs = append(runArgs, "-initrd", kernelFile)
	case ".iso":
		runArgs = append(runArgs, "-cdrom", kernelFile)
	}

	var qemuArgs []string
	if qemuArgs, err = shlex.Split(os.Getenv("QEMU_OPTS")); err != nil {
		return fmt.Errorf("error parsing QEMU_OPTS: %s", err)
	}

	runArgs = append(runArgs, "-m", "256M", "-no-reboot", "-serial", "mon:stdio")
	runArgs = append(runArgs, "-netdev", "user,id=eth0"+portMapingArgs())
	runArgs = append(runArgs, "-device", "e1000,netdev=eth0")
	runArgs = append(runArgs, "-device", "isa-debug-exit")
	runArgs = append(runArgs, qemuArgs...)

	cmd := exec.Command(qemu64, runArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err == nil {
		return nil
	}
	switch e := err.(type) {
	case *exec.ExitError:
		code := e.ExitCode()
		if code == 0 || code == 1 {
			return nil
		}
		return err
	default:
		return err
	}
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
	runCmd.Flags().StringSliceVarP(&ports, "port", "p", nil, "port mapping from host to kernel, format $host_port:$kernel_port")
}
