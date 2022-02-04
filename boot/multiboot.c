// According to the multiboot specification,
// the multiboot header must appear in the first 8192 bytes of the kernel image,
// and the go image is often megabytes in size.
//
// Therefore, we first write the elf loader in C language,
// and then load the kernel image in go language.
//
// https://www.gnu.org/software/grub/manual/multiboot/multiboot.html#OS-image-format

typedef unsigned char uint8;
typedef unsigned short uint16;
typedef unsigned int uint32;
typedef unsigned long long uint64;

#include "elf.h"
#include "multiboot.h"

extern char _binary_boot64_elf_start[];

void memcpy(char *dst, char *src, int count);
void memset(char *addr, char data, int cnt);
uint64 loadelf(char *image);
uint64 loadKernelElf(multiboot_info_t *info);
typedef void (*boot64_entry_t)(uint32, uint32, uint32);

void multibootmain(unsigned long magic, multiboot_info_t *mbi)
{
    uint64 entry_addr = 0;
    boot64_entry_t boot64_entry;

    entry_addr = loadelf(_binary_boot64_elf_start);
    if (entry_addr == 0)
    {
        return;
    }
    boot64_entry = (boot64_entry_t)((uint32)entry_addr);

    entry_addr = loadKernelElf(mbi);
    if (entry_addr == 0)
    {
        return;
    }
    boot64_entry((uint32)entry_addr, (uint32)magic, (uint32)mbi);
}

uint64 loadelf(char *image)
{
    struct elfhdr *elf;
    struct proghdr *ph, *eph;
    char *pa;

    elf = (struct elfhdr *)(image);

    // Is this an ELF executable?
    if (elf->magic != ELF_MAGIC)
        return 0;

    // Load each program segment (ignores ph flags).
    ph = (struct proghdr *)((uint8 *)elf + elf->phoff);
    eph = ph + elf->phnum;
    for (; ph < eph; ph++)
    {
        pa = (char *)(uint32)(ph->paddr);
        memcpy(pa, image + ph->off, ph->filesz);
        if (ph->memsz > ph->filesz)
        {
            memset((char *)(pa + ph->filesz), 0, ph->memsz - ph->filesz);
        }
    }
    return elf->entry;
}

uint64 loadKernelElf(multiboot_info_t *info)
{
    if (info->mods_count < 1)
    {
        return 0;
    }
    multiboot_module_t *mod = (multiboot_module_t *)(info->mods_addr);
    char *new_addr = (char *)(100 << 20); // 100 MB
    memcpy(new_addr, (char *)(mod->mod_start), mod->mod_end - mod->mod_start + 1);
    return loadelf(new_addr);
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
