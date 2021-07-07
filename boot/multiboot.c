// According to the multiboot specification,
// the multiboot header must appear in the first 8192 bytes of the kernel image,
// and the go image is often megabytes in size.
//
// Therefore, we first write the elf loader in C language,
// and then load the kernel image in go language.
//
// https://www.gnu.org/software/grub/manual/multiboot/multiboot.html#OS-image-format

typedef unsigned int uint;
typedef unsigned short ushort;
typedef unsigned char uchar;

#include "elf.h"
#include "multiboot.h"

extern char _binary_kernel_elf_start[];

void readseg(uchar *pa, uint count, uint offset);
void memcpy(char *dst, char *src, int count);
void memset(char *addr, char data, int cnt);

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
        {
            memset((char *)(pa + ph->filesz), 0, ph->memsz - ph->filesz);
        }
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

void memset(char *addr, char data, int count)
{
    int i = 0;
    for (; i < count; i++)
    {
        *addr++ = data;
    }
}