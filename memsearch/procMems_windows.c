// This program lists the memory regions of a pid (right now it is hardcoded).
// Basically it calls OpenProcess and then calls VirtualQueryEx to iterate
// over all memory regions. VirtualQueryEx takes an address and gives you the
// next region from that address and its size.
//
// The idea is to call ReadProcessMemory in each of these regions to search of a specific string. (TBD)
#include <stdlib.h>
#include <stdio.h>
#include <string.h>

#include "procMems_windows.h"

static void printInfo(MEMORY_BASIC_INFORMATION info);

static void printInfo(MEMORY_BASIC_INFORMATION info) {
    printf("%p %p %d %llu %d %d %d\n", info.BaseAddress, info.AllocationBase,
           info.AllocationProtect, info.RegionSize,
           info.State, info.Protect, info.Type);
}

void PrintMemory(HANDLE hndl, PVOID addr, SIZE_T size) {
    char buf[size + 1];
    SIZE_T out;
    printf("Trying to read %d bytes from addr %p\n", size, addr);
    BOOL res = ReadProcessMemory(hndl, addr, buf, size, &out);
    if (!res) {
        fprintf(stderr, "ReadProcessMemory failed; error: %d", GetLastError());
        return;
    }

    for (int i = 0; i < size; i++) {
        printf("%x ", buf[i]);
    }
    printf("\n");
}

BOOL FindInRange(HANDLE hndl, MEMORY_BASIC_INFORMATION m, char needle[],
                 int n) {
    // Read a buffer of size n * 2, so we can search the whole buffer two times,
    // with just one call to readProcessMemory
    SIZE_T bufSize = n * 2;
    SIZE_T outn;
    PVOID addr = m.BaseAddress;
    char *buf = malloc(sizeof (char) * bufSize);
    SIZE_T bytesRead = 0;
    do {
        if (bytesRead + bufSize >= m.RegionSize) bufSize = m.RegionSize -
                    bytesRead; //TODO: Check Bounds
        BOOL res = ReadProcessMemory(hndl, addr, buf, n * 2, &outn);
        if (!res) {
            free(buf);
            return FALSE;
        }

        // Search the needle in the haystack (inefficient solution n^2)
        for (int i = 0; i < outn - n; i++) { // TODO: Check Bounds
            if (memcmp(buf + i, needle, n) == 0) {
                printf("Found needle at address: %p\n", addr + i);
                free(buf);
                return TRUE;
            }
        }

        addr += outn;
        bytesRead += outn;
    } while (bufSize == 2 * n);

    free(buf);
    return FALSE;
}

MemoryInformation *GetMemoryInformation(DWORD pid) {
    MemoryInformation *res = calloc(1, sizeof * res);

    HANDLE hndl = OpenProcess(PROCESS_QUERY_INFORMATION | PROCESS_VM_READ,
                              FALSE,
                              pid);


    if (hndl == NULL) {
        res->error = GetLastError();
        return res;
    }
    res->hndl = hndl;

    int bufSize = 1024;
    MEMORY_BASIC_INFORMATION *info = calloc(bufSize, sizeof * info);
    LPCVOID addr = 0x0;
    int i = 0;
    while (TRUE) {

        // The entries may not fit in our initial array, that's why we double its size.
        if (i == bufSize) {
            bufSize *= 2;
            realloc(info, bufSize * sizeof * info);
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

/*
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
*/