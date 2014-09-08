#include "list_libs_windows.h"

#include <stdlib.h>
#include <string.h>
#include <tchar.h>
#include <stdio.h>

// getModules retrieves all the modules for a process with their info.
// it calls GetModuleFilenameEx and GetModuleInformation on the module.
// Caller must call EnumProcessModulesResponse_Free even if there's an error.
EnumProcessModulesResponse *getModules(DWORD pid) {
    HMODULE *aMods = NULL;
    ModuleInfo *modsInfo = NULL;
    DWORD size = 512 * sizeof(HMODULE);
    DWORD cbNeeded, mCount = 0;
    DWORD i;
    
    EnumProcessModulesResponse *res = calloc(1, sizeof *res);
    HANDLE hProcess = OpenProcess(PROCESS_QUERY_INFORMATION |
                                  PROCESS_VM_READ,
                                  FALSE, pid);
    if (hProcess == NULL) {
        res->error = GetLastError();
        return res;
    };
    
    // Allocate a buffer large enough to carry all the modules,
    // there's no way to know the size beforehand, so if the array is full
    // (cbNeeded == size), we double its size and refill it again.
    do {
        size *= 2;
        aMods = realloc(aMods, size);
        
        BOOL success = EnumProcessModulesEx(hProcess, aMods, 
                                            size, &cbNeeded, 
                                            LIST_MODULES_ALL);
        if (!success) {
            res->error = GetLastError();
            goto closeHandle;
        }
    } while (cbNeeded == size);
    
    
    // Try to get module's filename and information for each of the
    // modules retrieved by EnumProcessModulesEx. If there's an error, 
    // we abort and cleanup everything.
    mCount =  cbNeeded / sizeof (HMODULE);
    modsInfo = calloc(mCount, sizeof *modsInfo);
    for (i = 0; i < mCount; i++) {
        TCHAR buf[MAX_PATH + 1];
        BOOL success = GetModuleFileNameEx(hProcess, aMods[i], buf, 
                                            sizeof(buf)/sizeof(TCHAR));
        if (!success) {
            res->error = GetLastError();
            goto cleanup;
        }
        buf[MAX_PATH] = '\0';
        
        modsInfo[i].filename = (char *) _tcsdup(buf);  
        // is there safer way to convert from TCHAR * to char *?
        
        MODULEINFO info;
        success = GetModuleInformation(hProcess, aMods[i], &info, sizeof(info));
        if (!success) {
            res->error = GetLastError();
            goto cleanup;
        }
        modsInfo[i].info = info;
    }
    
    res->modules = modsInfo;
    res->length = mCount;

closeHandle:
    if (!CloseHandle(hProcess)) {
        res->error = GetLastError();
        goto cleanup;
    }
    free(aMods);

    return res;
    
cleanup:
    for (i = 0; i < mCount; i += 1) {
        free(modsInfo[i].filename);
    }

    free(modsInfo);
    res->modules = NULL;
    res->length = 0;

    free(aMods);
    
    CloseHandle(hProcess);
    return res;
}

EnumProcessesResponse *getAllPids() {
    DWORD size = sizeof(DWORD) * 512;
    DWORD *aProcesses = NULL;
    DWORD cbNeeded;

    EnumProcessesResponse *res = calloc(1, sizeof *res);

    // EnumProcesses modifies cbNeeded, setting it to the amount of bytes
    // written into aProcesses. Thus, we need to check if cbNeeded is equal
    // to size. In that case, it means that the array was filled completely and
    // we need to use a bigger array because probably we left elements out.
    do {
        size *= 2;
        aProcesses = realloc(aProcesses, size);

        BOOL success = EnumProcesses(aProcesses, size, &cbNeeded);
        if (!success) {
            res->error = GetLastError();
            free(aProcesses);
            return res;
        }
    }
    while(cbNeeded == size);

    res->error = 0;
    res->pids = aProcesses;
    res->length = cbNeeded/sizeof(DWORD);

    return res;
}

void EnumProcessesResponse_Free(EnumProcessesResponse *r) {
    if (r == NULL) return;
    free(r->pids);
    free(r);
}

void EnumProcessModulesResponse_Free(EnumProcessModulesResponse *r) {
    DWORD i;
    if (r == NULL) return;
    for (i = 0; i < r->length; i++) {
        free(r->modules[i].filename);
    }
    free(r->modules);
    free(r);
}