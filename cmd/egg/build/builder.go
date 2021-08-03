package build

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

type Config struct {
	GoRoot       string
	Basedir      string
	BuildTest    bool
	EggosVersion string
	GoArgs       []string
}

type Builder struct {
	cfg     Config
	basedir string
}

func NewBuilder(cfg Config) *Builder {
	return &Builder{
		cfg: cfg,
	}
}

func (b *Builder) Build() error {
	if b.cfg.Basedir == "" {
		basedir, err := ioutil.TempDir("", "eggos-build")
		if err != nil {
			return err
		}
		b.basedir = basedir
		defer os.RemoveAll(basedir)
	} else {
		b.basedir = b.cfg.Basedir
	}

	err := b.buildPrepare()
	if err != nil {
		return err
	}

	return b.buildPkg()
}

func (b *Builder) gobin() string {
	if b.cfg.GoRoot == "" {
		return "go"
	}
	return filepath.Join(b.cfg.GoRoot, "bin", "go")
}

func (b *Builder) fixGoTags() {
	var idx int
	args := b.cfg.GoArgs
	for i, arg := range args {
		if arg == "-tags" {
			if i >= len(b.cfg.GoArgs)-1 {
				return
			}
			idx = i + 1
			break
		}
	}
	tags := args[idx]
	tags += " eggos"
	args[idx] = tags
}

func (b *Builder) buildPkg() error {
	var buildArgs []string
	ldflags := "-E github.com/icexin/eggos/kernel.rt0 -T 0x100000"
	if !b.cfg.BuildTest {
		buildArgs = append(buildArgs, "build")
	} else {
		buildArgs = append(buildArgs, "test", "-c")
	}
	b.fixGoTags()
	buildArgs = append(buildArgs, b.cfg.GoArgs...)
	buildArgs = append(buildArgs, "-ldflags", ldflags)
	buildArgs = append(buildArgs, "-overlay", b.overlayFile())

	env := append([]string{}, os.Environ()...)
	env = append(env, []string{
		"GOOS=linux",
		"GOARCH=amd64",
		"CGO_ENABLED=0",
	}...)

	cmd := exec.Command(b.gobin(), buildArgs...)
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err
}
