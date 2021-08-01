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
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	output string
)

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:                "build",
	Short:              "build like go build command",
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		buildPkg(false, args)
	},
}

func patchEggosPkg() (string, string) {
	base, err := ioutil.TempDir("", "eggos-build")
	if err != nil {
		panic(err)
	}
	initFile := filepath.Join(base, "eggosinit.go")
	os.WriteFile(initFile, []byte(`package main
	import _ "github.com/icexin/eggos"
	`), 0644)

	overlayFile := filepath.Join(base, "overlay.json")
	overlay := map[string]map[string]string{
		"Replace": {
			"eggosinit.go": initFile,
		},
	}
	out, _ := json.Marshal(overlay)
	os.WriteFile(overlayFile, out, 0644)
	return base, overlayFile
}

func buildPkg(testbuild bool, args []string) {
	var buildArgs []string
	ldflags := "-E github.com/icexin/eggos/kernel.rt0 -T 0x100000"
	if !testbuild {
		buildArgs = append(buildArgs, "build")
	} else {
		buildArgs = append(buildArgs, "test", "-c")
	}
	buildArgs = append(buildArgs, "-ldflags", ldflags)
	buildArgs = append(buildArgs, args...)

	tmpdir, overlayFile := patchEggosPkg()
	defer os.RemoveAll(tmpdir)
	buildArgs = append(buildArgs, "-overlay", overlayFile)

	env := append([]string{}, os.Environ()...)
	env = append(env, []string{
		"GOOS=linux",
		"GOARCH=amd64",
		"CGO_ENABLED=0",
	}...)

	cmd := exec.Command("go", buildArgs...)
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err == nil {
		return
	}
	exiterr := err.(*exec.ExitError)
	os.Exit(exiterr.ExitCode())
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
