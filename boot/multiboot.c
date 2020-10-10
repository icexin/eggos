// Boot loader.
//
// Part of the boot block, along with bootasm.S, which calls bootmain().
// bootasm.S has put the processor into protected 32-bit mode.
// bootmain() loads an ELF kernel image from the disk starting at
// sector 1 and then jumps to the kernel entry routine.

typedef unsigned int uint;
typedef unsigned short ushort;
typedef unsigned char uchar;

#include "elf.h"
#include "multiboot.h"

extern char _binary_kernel_elf_start[];

void readseg(uchar *, uint, uint);
void memcpy(char *, char *, int);

static inline void
stosb(void *addr, int data, int cnt)
{
    asm volatile("cld; rep stosb"
                 : "=D"(addr), "=c"(cnt)
                 : "0"(addr), "1"(cnt), "a"(data)
                 : "memory", "cc");
}

void multibootmain(unsigned long magic, multiboot_info_t *mbi)
{
    struct elfhdr *elf;
    struct proghdr *ph, *eph;
    void (*entry)(unsigned long, multiboot_info_t *);
    uchar *pa;

    elf = (struct elfhdr *)(_binary_kernel_elf_start);

    // Is this an ELF executable?
    if (elf->magic != ELF_MAGIC)
        return; // let bootasm.S handle error

    // Load each program segment (ignores ph flags).
    ph = (struct proghdr *)((uchar *)elf + elf->phoff);
    eph = ph + elf->phnum;
    for (; ph < eph; ph++)
    {
        pa = (uchar *)ph->paddr;
        readseg(pa, ph->filesz, ph->off);
        if (ph->memsz > ph->filesz)
            stosb(pa + ph->filesz, 0, ph->memsz - ph->filesz);
    }

    // Call the entry point from the ELF header.
    // Does not return!
    entry = (void (*)(unsigned long, multiboot_info_t *))(elf->entry);
    entry(magic, mbi);
}

// Read 'count' bytes at 'offset' from kernel into physical address 'pa'.
void readseg(uchar *pa, uint count, uint offset)
{
    memcpy((char *)pa, (char *)(_binary_kernel_elf_start + offset), count);
}

void memcpy(char *dst, char *src, int count)
{
    int i = 0;
    for (; i < count; i++)
    {
        *dst++ = *src++;
    }
}
