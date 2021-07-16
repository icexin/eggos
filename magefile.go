//+build mage

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var (
	TOOLPREFIX = detectToolPrefix()
	CC         = TOOLPREFIX + "gcc"
	LD         = TOOLPREFIX + "ld"

	CFLAGS  = initCflags()
	LDFLAGS = initLdflags()
)

var (
	GOTAGS    = "nes phy"
	GOGCFLAGS = ""
)

var (
	QEMU64 = "qemu-system-x86_64"
	QEMU32 = "qemu-system-i386"

	QEMU_OPT       = initQemuOpt()
	QEMU_DEBUG_OPT = initQemuDebugOpt()
)

// Kernel target build the elf kernel for eggos, generate kernel.elf
func Kernel() error {
	detectGoVersion()
	env := map[string]string{
		"GOOS":   "linux",
		"GOARCH": "386",
	}
	goLdflags := "-E github.com/icexin/eggos/kernel.rt0 -T 0x100000"
	return sh.RunWithV(env, "go", "build", "-o", "kernel.elf", "-tags", GOTAGS,
		"-ldflags", goLdflags, "-gcflags", GOGCFLAGS, "./kmain")
}

// Multiboot target build Multiboot specification compatible elf format, generate multiboot.elf
func Multiboot() error {
	mg.Deps(Kernel)
	compileCfile("boot/multiboot.c")
	compileCfile("boot/multiboot_header.S")
	ldflags := "-Ttext 0x3200000 -o multiboot.elf multiboot.o multiboot_header.o -b binary kernel.elf"
	ldArgs := append([]string{}, LDFLAGS...)
	ldArgs = append(ldArgs, strings.Fields(ldflags)...)
	return sh.RunV(LD, ldArgs...)
}

// Qemu run multiboot.elf on qemu.
// If env QEMU_ACCEL is set，QEMU acceleration will be enabled.
// If env QEMU_GRAPHIC is set QEMU will run in graphic mode.
// Use Crtl+a c to switch console, and type `quit`
func Qemu() error {
	mg.Deps(Multiboot)

	detectQemu()
	args := append([]string{}, QEMU_OPT...)
	args = append(args, "-kernel", "multiboot.elf")
	return sh.RunV(QEMU64, args...)
}

// QemuDebug run multiboot.elf in debug mode.
// Monitor GDB connection on port 1234
func QemuDebug() error {
	GOGCFLAGS += " -N -l"
	mg.Deps(Multiboot)

	detectQemu()
	args := append([]string{}, QEMU_DEBUG_OPT...)
	args = append(args, "-kernel", "multiboot.elf")
	return sh.RunV(QEMU32, args...)
}

// Iso generate eggos.iso, which can be used with qemu -cdrom option.
func Iso() error {
	mg.Deps(Multiboot)

	tmpdir, err := ioutil.TempDir("", "eggos-iso")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpdir)
	grubDir := filepath.Join(tmpdir, "boot", "grub")
	os.MkdirAll(grubDir, 0755)
	sh.Copy(
		filepath.Join(grubDir, "grub.cfg"),
		filepath.Join("boot", "grub.cfg"),
	)
	sh.Copy(
		filepath.Join(tmpdir, "boot", "multiboot.elf"),
		"multiboot.elf",
	)
	return sh.RunV("grub-mkrescue", "-o", "eggos.iso", tmpdir)
}

// Graphic run eggos.iso on qemu, which vbe is enabled.
func Graphic() error {
	detectQemu()

	mg.Deps(Iso)
	args := append([]string{}, QEMU_OPT...)
	args = append(args, "-cdrom", "eggos.iso")
	return sh.RunV(QEMU64, args...)
}

// GraphicDebug run eggos.iso on qemu in debug mode.
func GraphicDebug() error {
	detectQemu()

	GOGCFLAGS += " -N -l"
	mg.Deps(Iso)
	args := append([]string{}, QEMU_DEBUG_OPT...)
	args = append(args, "-cdrom", "eggos.iso")
	return sh.RunV(QEMU32, args...)
}

func Clean() {
	rmGlob("*.o")
	rmGlob("kernel.elf")
	rmGlob("multiboot.elf")
	rmGlob("qemu.log")
	rmGlob("qemu.pcap")
}

func detectToolPrefix() string {
	prefix := os.Getenv("TOOLPREFIX")
	if prefix != "" {
		return prefix
	}

	if hasOutput("elf32-i386", "x86_64-elf-objdump", "-i") {
		return "x86_64-elf-"
	}

	if hasOutput("elf32-i386", "i386-elf-objdump", "-i") {
		return "i386-elf-"
	}

	if hasOutput("elf32-i386", "objdump", "-i") {
		return ""
	}
	panic(`
	*** Error: Couldn't find an i386-*-elf or x86_64-*-elf version of GCC/binutils
	*** Is the directory with i386-elf-gcc or x86_64-elf-gcc in your PATH?
	*** If your i386/x86_64-*-elf toolchain is installed with a command
	*** prefix other than 'i386/x86_64-elf-', set your TOOLPREFIX
	*** environment variable to that prefix and run 'make' again.
	`)
}

func detectGoVersion() {
	if !hasCommand("go") {
		panic(`go command not found`)
	}

	if !hasOutput("go1.13", "go", "version") {
		panic(`eggos only tested on go1.13.x`)
	}
}

func detectQemu() {
	if !hasCommand(QEMU64) {
		panic(QEMU64 + ` command not found`)
	}
}

func accelArg() []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{"-M", "accel=hvf"}
	default:
		fmt.Printf("accel method not found")
		return nil
	}
}

func initCflags() []string {
	cflags := strings.Fields("-fno-pic -static -fno-builtin -fno-strict-aliasing -O2 -Wall -ggdb -m32 -Werror -fno-omit-frame-pointer -I. -nostdinc")
	if hasOutput("-fno-stack-protector", CC, "--help") {
		cflags = append(cflags, "-fno-stack-protector")
	}
	if hasOutput("[^f]no-pie", CC, "-dumpspecs") {
		cflags = append(cflags, "-fno-pie", "-no-pie")
	}
	if hasOutput("[^f]nopie", CC, "-dumpspecs") {
		cflags = append(cflags, "-fno-pie", "-nopie")
	}
	return cflags
}

func initLdflags() []string {
	ldflags := strings.Fields("-N -e _start -m elf_i386")
	return ldflags
}

func initQemuOpt() []string {
	opts := `
	-m 256M -no-reboot -serial mon:stdio
	-netdev user,id=eth0,hostfwd=tcp::8080-:80,hostfwd=tcp::8081-:22
	-device e1000,netdev=eth0
	`
	out := strings.Fields(opts)
	if os.Getenv("QEMU_ACCEL") != "" {
		out = append(out, accelArg()...)
	}
	if os.Getenv("QEMU_GRAPHIC") == "" {
		out = append(out, "-nographic")
	}
	return out
}

func initQemuDebugOpt() []string {
	opts := `
	-d int -D qemu.log
	-object filter-dump,id=f1,netdev=eth0,file=qemu.pcap
	-s -S
	`
	ret := append([]string{}, initQemuOpt()...)
	return append(ret, strings.Fields(opts)...)
}

func compileCfile(file string) {
	args := append([]string{}, CFLAGS...)
	args = append(args, "-c", file)
	err := sh.RunV(CC, args...)
	if err != nil {
		panic(err)
	}
}
func cmdOutput(cmd string, args ...string) ([]byte, error) {
	return exec.Command(cmd, args...).CombinedOutput()
}

func hasCommand(cmd string) bool {
	_, err := exec.LookPath(cmd)
	if err != nil {
		return false
	}
	return true
}

func hasOutput(regstr, cmd string, args ...string) bool {
	out, err := cmdOutput(cmd, args...)
	if err != nil {
		return false
	}
	match, err := regexp.Match(regstr, []byte(out))
	if err != nil {
		return false
	}
	return match
}

func rmGlob(patten string) error {
	match, err := filepath.Glob(patten)
	if err != nil {
		return err
	}
	for _, file := range match {
		err = os.Remove(file)
		if err != nil {
			return err
		}
	}
	return nil
}
