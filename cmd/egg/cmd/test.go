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
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/icexin/eggos/cmd/egg/build"
	"github.com/spf13/cobra"
)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "test likes go test but running in qemu",
	Run: func(cmd *cobra.Command, args []string) {
		err := runTest()
		if err != nil {
			log.Fatal(err)
		}
	},
}

func runTest() error {
	base, err := ioutil.TempDir("", "eggos-test")
	if err != nil {
		return err
	}
	defer os.RemoveAll(base)

	outfile := filepath.Join(base, "eggos.test.elf")

	b := build.NewBuilder(build.Config{
		Basedir:      base,
		BuildTest:    true,
		EggosVersion: eggosVersion,
		GoArgs: []string{
			"-o", outfile,
			"-vet=off",
		},
	})
	err = b.Build()
	if err != nil {
		return err
	}

	kernelFile = outfile
	runKernel()
	return nil
}

func init() {
	rootCmd.AddCommand(testCmd)
}
