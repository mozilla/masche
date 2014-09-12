#include <windows.h>
#include <Processthreadsapi.h>

//TODO(mvanotti): MEMORY_BASIC_INFORMATION should be switched to its 64 bit or 32 bit version.

typedef struct t_MemoryInformation {
    DWORD error;
    MEMORY_BASIC_INFORMATION *info;
    DWORD length;
    HANDLE hndl;
} MemoryInformation;

MemoryInformation *GetMemoryInformation(DWORD pid);
void MemoryInformation_Free(MemoryInformation *m);
inline static BOOL isReadable(DWORD protection);
BOOL FindInRange(HANDLE hndl, MEMORY_BASIC_INFORMATION m, char buf[], int n);
void PrintMemory(HANDLE hndl, PVOID addr, SIZE_T size);