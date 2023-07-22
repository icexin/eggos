package build

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Config struct {
	WorkDir      string
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

	goversion, err := b.goVersion()
	if err != nil {
		return err
	}
	if !strings.Contains(goversion, "go1.16") {
		return fmt.Errorf("eggos requires go 1.16 but found %s", strings.TrimSpace(goversion))
	}

	err = b.buildPrepare()
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

func (b *Builder) fixGoTags() bool {
	args := b.cfg.GoArgs
	for i, arg := range args {
		if arg == "-tags" {
			if i >= len(b.cfg.GoArgs)-1 {
				return false
			}
			idx := i + 1
			tags := args[idx]
			tags += " eggos"
			args[idx] = tags
			return true
		}
	}
	return false
}

func (b *Builder) goVersion() (string, error) {
	cmd := exec.Command(b.gobin(), "version")
	cmd.Env = os.Environ()
	stdout, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error executing: go version: %w", err)
	}
	return string(stdout), nil
}

func (b *Builder) buildPkg() error {
	var buildArgs []string
	ldflags := "-E github.com/icexin/eggos/kernel.rt0 -T 0x100000"
	if !b.cfg.BuildTest {
		buildArgs = append(buildArgs, "build")
	} else {
		buildArgs = append(buildArgs, "test", "-c")
	}
	hasGoTags := b.fixGoTags()
	if !hasGoTags {
		buildArgs = append(buildArgs, "-tags", "eggos")
	}
	buildArgs = append(buildArgs, "-ldflags", ldflags)
	buildArgs = append(buildArgs, "-overlay", b.overlayFile())
	buildArgs = append(buildArgs, b.cfg.GoArgs...)

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
	if b.cfg.WorkDir != "" {
		cmd.Dir = b.cfg.WorkDir
	}
	err := cmd.Run()
	return err
}
