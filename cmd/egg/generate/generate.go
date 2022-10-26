package generate

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jspc/eggos/cmd/egg/util"
)

const (
	ImportFile = "zz_load_eggos.go"
)

type Generator struct {
	gobin   string
	basedir string
}

func NewGenerator(basedir string) (g Generator, err error) {
	g = Generator{
		gobin:   util.GoBin(),
		basedir: basedir,
	}

	return
}

func (g Generator) Generate() error {
	return g.writeImportFile(g.eggosImportFile())
}

func (g *Generator) eggosImportFile() string {
	return filepath.Join(g.basedir, ImportFile)
}

func (b *Generator) currentPkgName() string {
	out, err := exec.Command(b.gobin, "list", "-f", `{{.Name}}`).CombinedOutput()
	if err != nil {
		log.Panicf("get current package name:%s", out)
	}

	return strings.TrimSpace(string(out))
}

func (g *Generator) writeImportFile(fname string) error {
	pkgname := g.currentPkgName()
	var rawFile bytes.Buffer

	err := eggosImportTpl.Execute(&rawFile, map[string]interface{}{
		"name": pkgname,
	})

	if err != nil {
		return err
	}

	return os.WriteFile(fname, rawFile.Bytes(), 0644)
}
