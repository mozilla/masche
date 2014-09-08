// This program lists the memory regions of a pid (right now it is hardcoded).
// Basically it calls OpenProcess and then calls VirtualQueryEx to iterate 
// over all memory regions. VirtualQueryEx takes an address and gives you the
// next region from that address and its size.
//
// The idea is to call ReadProcessMemory in each of these regions to search of a specific string. (TBD)
#include <windows.h>
#include <Processthreadsapi.h>
#include <stdlib.h>
#include <stdio.h>

void printInfo(MEMORY_BASIC_INFORMATION info) {
    printf("%p %p %d %llu %d %d %d\n", info.BaseAddress, info.AllocationBase,
            info.AllocationProtect, info.RegionSize,
            info.State, info.Protect, info.Type);
}

typedef struct t_MemoryInformation {
    DWORD error;
    MEMORY_BASIC_INFORMATION *info;
    DWORD length;
    HANDLE hndl;
} MemoryInformation;

MemoryInformation *getMemoryInformation(DWORD pid) {
    MemoryInformation *res = calloc(1, sizeof *res);

    HANDLE hndl = OpenProcess(PROCESS_QUERY_INFORMATION | PROCESS_VM_READ,
            FALSE,
            pid);


    if (hndl == NULL) {
        res->error = GetLastError();
        return res;
    }
    res->hndl = hndl;

    int bufSize = 1024;
    MEMORY_BASIC_INFORMATION *info = calloc(bufSize, sizeof *info);
    LPCVOID addr = 0x0;
    int i = 0;
    while (TRUE) {

        // The entries may not fit in our initial array, that's why we double its size.
        if (i == bufSize) {
            bufSize *= 2;
            realloc(info, bufSize * sizeof *info);
        }

        SIZE_T r = VirtualQueryEx(hndl,
                addr,
                &info[i],
                sizeof(*info));

        if (r == 0) {
            DWORD err = GetLastError();
            if (err == ERROR_INVALID_PARAMETER) { 
                // This means that the address we are using is invalid, i.e: no more addresses left!
                break;
            }
            res->error = err;
            free(info);
            return res;
        }
        addr += info[i].RegionSize;
        i++;
    }


    res->info = info;
    res->length = i;
    return res;
}

void MemoryInformation_Free(MemoryInformation *m) {
    if (m == NULL) return;
    if (m->hndl != 0) CloseHandle(m->hndl);
    free(m->info);
    free(m);
}

inline static BOOL isReadable(DWORD protection) {
    switch (protection) {
        case PAGE_EXECUTE_READ:
        case PAGE_EXECUTE_READWRITE:
        case PAGE_READONLY:
        case PAGE_READWRITE:
            return TRUE;
        default:
            return FALSE;
    }
}

int main(int argc, char **argv) {
    if (argc != 2) {
        printf("Usage: %s <pid>", argv[0]);
        return 0;
    }

    DWORD pid = strtol(argv[1], NULL, 10);
    MemoryInformation *m = getMemoryInformation(pid);
    if (m->error != 0) {
        fprintf(stderr, "Error! number: %d", m->error);
        exit(EXIT_FAILURE);
    }


    for (int i = 0; i < m->length; i++) {
        printInfo(m->info[i]);
    }
    printf("Entries: %d\n", m->length);


    MemoryInformation_Free(m);
    return 0;
}
