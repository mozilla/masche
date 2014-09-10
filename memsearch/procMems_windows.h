#include <windows.h>
#include <Processthreadsapi.h>

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