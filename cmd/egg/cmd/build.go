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
	"log"

	"github.com/icexin/eggos/cmd/egg/build"
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
		err := runBuild(args)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func runBuild(args []string) error {
	b := build.NewBuilder(build.Config{
		EggosVersion: eggosVersion,
		GoArgs:       args,
	})
	err := b.Build()
	if err != nil {
		return err
	}
	return nil
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
