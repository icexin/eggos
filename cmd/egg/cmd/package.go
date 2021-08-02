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
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/icexin/eggos/cmd/egg/assets"
	"github.com/icexin/eggos/cmd/egg/build"
	"github.com/spf13/cobra"
)

const (
	grubDockerImage = "fanbingxin/grub:0.2.0"
)

var (
	packageFormat     string
	packageOutFile    string
	packageKernelFile string

	withoutDocker bool
	keepTmpdir    bool
)

// packageCmd represents the package command
var packageCmd = &cobra.Command{
	Use:   "package",
	Short: "package kernel to release format, eg iso",
	Run: func(cmd *cobra.Command, args []string) {
		err := runPackage()
		if err != nil {
			log.Fatal(err)
		}
	},
}

func runPackage() error {
	base, err := ioutil.TempDir("", "eggos-package")
	if err != nil {
		return err
	}
	if !keepTmpdir {
		defer os.RemoveAll(base)
	} else {
		log.Println(base)
	}

	isoBase := filepath.Join(base, "iso")

	err = os.MkdirAll(isoBase, 0755)
	if err != nil {
		return err
	}

	err = extractBootDir(isoBase)
	if err != nil {
		return err
	}

	kfile, err := getKernelFile(base)
	if err != nil {
		return err
	}

	err = copyfile(
		filepath.Join(isoBase, "boot", "kernel.elf"),
		kfile,
	)
	if err != nil {
		return err
	}

	tmpOutFile := filepath.Join(base, "eggos.iso")
	err = mkiso(tmpOutFile, isoBase, base)
	if err != nil {
		return err
	}

	if packageOutFile == "" {
		packageOutFile = "eggos.iso"
	}
	return copyfile(packageOutFile, tmpOutFile)
}

func mkiso(outfile, isobase, moutbase string) error {
	var stderr bytes.Buffer
	if withoutDocker {
		cmd := exec.Command("grub-mkrescue", "-o", outfile, isobase)
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			log.Print(stderr.String())
		}
		return err
	}
	cmd := exec.Command(
		"docker", "run", "--rm",
		"-v", moutbase+":"+moutbase,
		"-w", moutbase,
		grubDockerImage,
		"grub-mkrescue", "-o", outfile, isobase,
	)
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Print(stderr.String())
	}
	return err
}

func extractBootDir(base string) error {
	bootfs := assets.Boot
	err := fs.WalkDir(assets.Boot, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		outFullPath := filepath.Join(base, path)
		if d.IsDir() {
			err = os.MkdirAll(outFullPath, 0755)
			if err != nil {
				return err
			}
			return nil
		}

		content, err := bootfs.ReadFile(path)
		if err != nil {
			return err
		}
		err = os.WriteFile(outFullPath, content, 0644)
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

func getKernelFile(base string) (string, error) {
	if packageKernelFile != "" {
		return packageKernelFile, nil
	}
	outputFile := filepath.Join(base, "kernel.elf")
	b := build.NewBuilder(build.Config{
		Basedir: base,
		GoArgs: []string{
			"-o", outputFile,
		},
	})
	err := b.Build()
	if err != nil {
		return "", err
	}

	return outputFile, nil
}

func copyfile(dst, src string) error {
	content, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, content, 0644)
}

func init() {
	rootCmd.AddCommand(packageCmd)

	packageCmd.Flags().StringVarP(&packageFormat, "format", "f", "iso", "package format, values `iso`")
	packageCmd.Flags().StringVarP(&packageKernelFile, "kernel", "k", "", "the kernel file, if empty current package will be built as kernel")
	packageCmd.Flags().StringVarP(&packageOutFile, "output", "o", "eggos.iso", "file name of output")
	packageCmd.Flags().BoolVar(&keepTmpdir, "keep-tmp", false, "keep temp dir, for debugging")
	packageCmd.Flags().BoolVarP(&withoutDocker, "without-docker", "d", false, "using docker for grub tools")
}
