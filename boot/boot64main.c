typedef unsigned short uint16;
typedef unsigned int uint32;
typedef unsigned long uint64;

typedef void (*go_entry_t)(uint32, uint32);

void boot64main(uint32 gomain, uint32 magic, uint32 mbinfo)
{
    go_entry_t go_entry;
    go_entry = (go_entry_t)((uint64)gomain);
    go_entry(magic, mbinfo);
    for (;;)
    {
    }
}
