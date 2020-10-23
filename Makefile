SHELL := /bin/bash
#TOOLPREFIX=i386-elf-
ifndef TOOLPREFIX
TOOLPREFIX := $(shell if i386-elf-objdump -i 2>&1 | grep '^elf32-i386$$' >/dev/null 2>&1; \
	then echo 'i386-elf-'; \
	elif objdump -i 2>&1 | grep 'elf32-i386' >/dev/null 2>&1; \
	then echo ''; \
	else echo "***" 1>&2; \
	echo "*** Error: Couldn't find an i386-*-elf version of GCC/binutils." 1>&2; \
	echo "*** Is the directory with i386-elf-gcc in your PATH?" 1>&2; \
	echo "*** If your i386-*-elf toolchain is installed with a command" 1>&2; \
	echo "*** prefix other than 'i386-elf-', set your TOOLPREFIX" 1>&2; \
	echo "*** environment variable to that prefix and run 'make' again." 1>&2; \
	echo "*** To turn off this error, run 'gmake TOOLPREFIX= ...'." 1>&2; \
	echo "***" 1>&2; exit 1; fi)
endif

CC = $(TOOLPREFIX)gcc
AS = $(TOOLPREFIX)gas
LD = $(TOOLPREFIX)ld
OBJCOPY = $(TOOLPREFIX)objcopy
OBJDUMP = $(TOOLPREFIX)objdump
CFLAGS = -fno-pic -static -fno-builtin -fno-strict-aliasing -O2 -Wall -ggdb -m32 -Werror -fno-omit-frame-pointer
CFLAGS += $(shell $(CC) -fno-stack-protector -E -x c /dev/null >/dev/null 2>&1 && echo -fno-stack-protector)
ASFLAGS = -m32 -gdwarf-2 -Wa,-divide
# FreeBSD ld wants ``elf_i386_fbsd''
LDFLAGS += -m $(shell $(LD) -V | grep elf_i386 2>/dev/null | head -n 1)

# Disable PIE when possible (for Ubuntu 16.10 toolchain)
ifneq ($(shell $(CC) -dumpspecs 2>/dev/null | grep -e '[^f]no-pie'),)
CFLAGS += -fno-pie -no-pie
endif
ifneq ($(shell $(CC) -dumpspecs 2>/dev/null | grep -e '[^f]nopie'),)
CFLAGS += -fno-pie -nopie
endif

GOVERSION=$(shell go version | awk '{print $$3}')

QEMU_OPT = -m 256M -no-reboot -serial mon:stdio \
	-netdev user,id=eth0,hostfwd=tcp::8080-:80,hostfwd=tcp::8081-:22 \
	-device e1000,netdev=eth0 \
	

QEMU_DEBUG_OPT = $(QEMU_OPT) -d int -D qemu.log \
	-object filter-dump,id=f1,netdev=eth0,file=qemu.pcap \
	-s -S

# TAGS = "gin sshd nes"
TAGS = "nes sshd"

.PHONY: all clean kernel.elf

all: multiboot.elf

kernel.elf:
	@if [[ ${GOVERSION} != go1.13* ]]; then echo "eggos only tested on go1.13.x"; exit 1; fi;
	GOOS=linux GOARCH=386 go build -o kernel.elf -tags $(TAGS) -ldflags '-E github.com/icexin/eggos/kernel.rt0 -T 0x100000' ./kmain

multiboot.elf: boot/multiboot_header.S boot/multiboot.c kernel.elf
	$(CC) $(CFLAGS) -fno-pic -O -nostdinc -I. -c boot/multiboot.c
	$(CC) $(CFLAGS) -fno-pic -nostdinc -I. -c boot/multiboot_header.S
	$(LD) $(LDFLAGS) -N -e _start -Ttext 0x3200000 -o multiboot.elf multiboot.o multiboot_header.o -b binary kernel.elf

qemu-debug:
	qemu-system-i386 $(QEMU_DEBUG_OPT) -kernel multiboot.elf # -nographic
	
qemu:
	qemu-system-x86_64 $(QEMU_OPT) -kernel multiboot.elf # -nographic

grub: multiboot.elf
	mount -t msdos /dev/disk2s1 osimg
	cp multiboot.elf osimg/grub/
	sync
	sleep 2
	umount osimg
	qemu-system-x86_64 $(QEMU_OPT) -M accel=hvf /dev/disk2

grub-debug: multiboot.elf
	mount -t msdos /dev/disk2s1 osimg
	cp multiboot.elf osimg/grub/
	sync
	sleep 2
	umount osimg
	qemu-system-i386 $(QEMU_DEBUG_OPT) /dev/disk2

clean:
	rm -f *.o *.log kernel.elf multiboot.elf qemu.pcap
