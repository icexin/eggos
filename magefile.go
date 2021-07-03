//+build mage

package main

import (
	"fmt"
	"os"
	"os/exec"
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
	GOTAGS = ""
)

var (
	QEMU64 = "qemu-system-x86_64"
	QEMU32 = "qemu-system-i386"

	QEMU_OPT       = initQemuOpt()
	QEMU_DEBUG_OPT = initQemuDebugOpt()
)

func Kernel() error {
	detectGoVersion()
	env := map[string]string{
		"GOOS":   "linux",
		"GOARCH": "386",
	}
	goLdflags := "-E github.com/icexin/eggos/kernel.rt0 -T 0x100000"
	return sh.RunWithV(env, "go", "build", "-o", "kernel.elf", "-tags", GOTAGS,
		"-ldflags", goLdflags, "./kmain")
}

func Multiboot() error {
	mg.Deps(Kernel)
	compileCfile("boot/multiboot.c")
	compileCfile("boot/multiboot_header.S")
	ldflags := "-Ttext 0x3200000 -o multiboot.elf multiboot.o multiboot_header.o -b binary kernel.elf"
	ldArgs := append([]string{}, LDFLAGS...)
	ldArgs = append(ldArgs, strings.Fields(ldflags)...)
	return sh.RunV(LD, ldArgs...)
}

func Qemu(accel bool, graphic bool) error {
	mg.Deps(Multiboot)

	detectQemu()
	args := append([]string{}, QEMU_OPT...)
	if accel {
		args = append(args, accelArg()...)
	}
	if !graphic {
		args = append(args, "-nographic")
	}
	args = append(args, "-kernel", "multiboot.elf")
	return sh.RunV(QEMU64, args...)
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
	return strings.Fields(opts)
}

func initQemuDebugOpt() []string {
	opts := `
	-d int -D qemu.log
	-object filter-dump,id=f1,netdev=eth0,file=qemu.pcap
	-s -S
	`
	ret := append([]string{}, QEMU_OPT...)
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
