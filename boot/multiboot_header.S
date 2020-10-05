#define ASM_FILE        1
#include "multiboot.h"

/* The size of our stack (16KB). */
#define STACK_SIZE                      0x4000
#define MULTIBOOT_HEADER_FLAGS  MULTIBOOT_PAGE_ALIGN | MULTIBOOT_MEMORY_INFO | MULTIBOOT_VIDEO_MODE
// #define MULTIBOOT_HEADER_FLAGS  MULTIBOOT_PAGE_ALIGN | MULTIBOOT_MEMORY_INFO


.p2align 2
.text
.globl multiboot_header
multiboot_header:
  #define magic 0xe85250d6
  #define flags 0
  #define length (multiboot_header_end - multiboot_header)
  .long magic
  .long flags
  .long length
  .long (-magic-flags-length)
  .word 0
  .word 0
  .long 8
multiboot_header_end:


.align  4
.global multiboot1_header
multiboot1_header:
  .long MULTIBOOT_HEADER_MAGIC
  .long MULTIBOOT_HEADER_FLAGS
  .long -(MULTIBOOT_HEADER_MAGIC + MULTIBOOT_HEADER_FLAGS)
  .long   0
  .long   0
  .long   0
  .long   0
  .long   0 
  .long 0
  .long 640
  .long 480
  .long 32

.global _start
_start:
  jmp     multiboot_entry

multiboot_entry:
  /* Initialize the stack pointer. */
  movl    $(stack + STACK_SIZE), %esp

  /* Reset EFLAGS. */
  pushl   $0
  popf

  /* Push the pointer to the Multiboot information structure. */
  pushl   %ebx
  /* Push the magic value. */
  pushl   %eax

  /* Now enter the C main function... */
  call    multibootmain

  /* Our stack area. */
.comm   stack, STACK_SIZE
