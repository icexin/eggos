typedef unsigned short uint16;
typedef unsigned int uint32;
typedef unsigned long uint64;

uint16 (*screen)[25][80] = (uint16(*)[25][80])(0xb8000);
void puts(int line, char *str);

void boot64main(uint32 gomain, uint32 magic, uint32 mbinfo)
{
    void (*gomain_entry)(uint32, uint32);
    gomain_entry = (void(*)(uint32, uint32))(gomain);
    gomain_entry(magic, mbinfo);
    puts(2, "hello world");
    for (;;)
    {
    }
}

void puts(int line, char *str)
{
    char *p = str;
    int i = 0;
    for (; *p != 0; p++)
    {
        (*screen)[line][i] = (uint16)(*p | 0x700);
        i++;
    }
}