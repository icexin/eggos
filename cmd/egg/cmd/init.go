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
	"bytes"
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"text/template"

	"github.com/spf13/cobra"
)

const (
	eggosModulePath = "github.com/icexin/eggos"
	eggosImportFile = "import_eggos.go"
)

var (
	eggosImportTpl = template.Must(template.New("eggos").Parse(`
	//+build eggos
	package {{.name}}
	import _ "github.com/icexin/eggos"
	`))
)

var (
	eggosVersion string
)

type gomodule struct {
	Module struct {
		Path string `json:"Path"`
	} `json:"Module"`
	Go      string `json:"Go"`
	Require []struct {
		Path    string `json:"Path"`
		Version string `json:"Version"`
	} `json:"Require"`
	Exclude interface{} `json:"Exclude"`
	Replace interface{} `json:"Replace"`
	Retract interface{} `json:"Retract"`
}

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "init a go project with eggos support",
	Run: func(cmd *cobra.Command, args []string) {
		err := initEggos()
		if err != nil {
			log.Fatal(err)
		}
	},
}

func initEggos() error {
	var err error
	if !hasImportFile() {
		log.Printf("%s not found, create one", eggosImportFile)
		err = addImportEggos()
		if err != nil {
			return err
		}
	}
	if !modHasEggos() {
		log.Printf("eggos not found in go.mod")
		err = editGoMod()
		if err != nil {
			return err
		}
	}
	return nil
}

func readGomodule() (*gomodule, error) {
	var buf bytes.Buffer
	cmd := exec.Command("go", "mod", "edit", "-json")
	cmd.Stdout = &buf
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	var mod gomodule
	err = json.Unmarshal(buf.Bytes(), &mod)
	if err != nil {
		return nil, err
	}
	return &mod, nil
}

func modHasEggos() bool {
	mods, err := readGomodule()
	if err != nil {
		panic(err)
	}
	for _, mod := range mods.Require {
		if mod.Path == eggosModulePath {
			return true
		}
	}
	return false
}

func editGoMod() error {
	getPath := eggosModulePath
	if eggosVersion != "" {
		getPath = getPath + "@" + eggosVersion
	}
	log.Printf("go get %s", getPath)
	env := []string{
		"GOOS=linux",
		"GOARCH=amd64",
	}
	env = append(env, os.Environ()...)
	cmd := exec.Command("go", "get", getPath)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func hasImportFile() bool {
	_, err := os.Stat(eggosImportFile)
	return err == nil
}

func currentPkgName() string {
	out, err := exec.Command("go", "list", "-f", `{{.Name}}`).CombinedOutput()
	if err != nil {
		panic(err)
	}
	return string(out)
}

func addImportEggos() error {
	pkgname := currentPkgName()
	var rawFile bytes.Buffer
	err := eggosImportTpl.Execute(&rawFile, map[string]interface{}{
		"name": pkgname,
	})
	if err != nil {
		return err
	}

	var out bytes.Buffer
	cmd := exec.Command("gofmt")
	cmd.Stdin = &rawFile
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return err
	}
	err = os.WriteFile(eggosImportFile, out.Bytes(), 0644)
	return err
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVarP(&eggosVersion, "version", "", "", "eggos version, match go get version")
}
