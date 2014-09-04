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

int main(int argc, char **argv) {
    if (argc != 2) {
        printf("Usage: %s <pid>", argv[0]);
        return 0;
    }

    DWORD pid = strtol(argv[1], NULL, 10);

    HANDLE hndl = OpenProcess(PROCESS_QUERY_INFORMATION | PROCESS_VM_READ,
            FALSE,
            pid);

    if (hndl == NULL) {
        fprintf(stderr, "Error! %d", GetLastError());
        exit(EXIT_FAILURE);
    }

    MEMORY_BASIC_INFORMATION info;
    LPCVOID addr = 0x0;
    while (TRUE) {
        SIZE_T res = VirtualQueryEx(hndl,
                addr,
                &info,
                sizeof(info));

        if (res == 0) {
            DWORD err = GetLastError();
            if (err == ERROR_INVALID_PARAMETER) { 
                // This means that the address we are using is invalid, i.e: no more addresses left!
                break;
            }
            fprintf(stderr, "Error! %d", GetLastError());
            exit(EXIT_FAILURE);
        }

        printInfo(info);
        addr += info.RegionSize;
    }

    return 0;
}
