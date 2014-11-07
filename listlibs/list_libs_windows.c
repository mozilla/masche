#include "list_libs_windows.h"

#include <stdlib.h>
#include <stdbool.h>
#include <string.h>
#include <stdio.h>
#include <tchar.h>

// getAllProcesses returns all the active pids and their filenames.
// If there's a process that couldn't be accessed, its name
// will be NULL and its error field will be set.
EnumProcessesFullResponse *getAllProcesses() {
    EnumProcessesFullResponse *res = calloc(1, sizeof * res);

    EnumProcessesResponse *pids = getAllPids();
    if (pids->error != 0) {
        res->error = pids->error;
        EnumProcessesResponse_Free(pids);
        return res;
    }

    res->length = pids->length;

    ProcessInfo *pinfo = calloc(pids->length, sizeof * pinfo);
    res->processes = pinfo;

    DWORD bufSize = 128;
    char *buf = calloc(bufSize, sizeof * buf);
    for (DWORD i = 0; i < pids->length; i++) {
        pinfo[i].pid = pids->pids[i];

        HANDLE hProcess = OpenProcess( PROCESS_QUERY_INFORMATION,
                                       FALSE, pinfo[i].pid );
        if (hProcess == NULL) {
            pinfo[i].error = GetLastError();
            continue;
        }

        //TODO(mvanotti): GetProcessImageFileName returns the name in "Device Format"
        // ej. \Device\HardDiskVolume4\Windows\foo.exe We can use another function to get it
        // in "normal" format ( C:\Windows\foo.exe )
        DWORD res = GetProcessImageFileName(hProcess, buf, bufSize);
        if (res == 0) {
            pinfo[i].error = GetLastError();

            // This is the only error that we know how to handle:
            // use a bigger buffer.
            if (pinfo[i].error != ERROR_INSUFFICIENT_BUFFER) {
                CloseHandle(hProcess);
                continue;
            }
        }

        // We don't have a max_path for the filename, so we keep resizing our buffer
        // the whole pathname fits. It may or may not grow longer than 32k.
        // I'm doing it this way because I want to reuse the buffer for all processes.
        bool failed = false;
        while (pinfo[i].error == ERROR_INSUFFICIENT_BUFFER) {
            bufSize *= 2;
            buf = realloc(buf, bufSize);
            pinfo[i].error = 0;
            DWORD res = GetProcessImageFileName(hProcess, buf, bufSize);
            if (res == 0) {
                pinfo[i].error = GetLastError();
                if (pinfo[i].error == ERROR_INSUFFICIENT_BUFFER) {
                    continue;
                }
                failed = true;
                break;
            }
        }

        if (failed) {
            CloseHandle(hProcess);
            continue;
        }

        char *processName = _strdup(buf);
        pinfo[i].filename = processName;

        CloseHandle(hProcess);
    }

    free(buf);
    EnumProcessesResponse_Free(pids);
    return res;
}

// getModules retrieves all the modules for a process with their info.
// it calls GetModuleFilenameEx and GetModuleInformation on the module.
// Caller must call EnumProcessModulesResponse_Free even if there's an error.
EnumProcessModulesResponse *getModules(DWORD pid) {
    HMODULE *aMods = NULL;
    ModuleInfo *modsInfo = NULL;
    DWORD size = 512 * sizeof(HMODULE);
    DWORD cbNeeded, mCount = 0;
    DWORD i;

    EnumProcessModulesResponse *res = calloc(1, sizeof * res);
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
                                            0x03);
//TODO(mvanotti): Fix and replace the 0x03 to LIST_MODULES_ALL
        if (!success) {
            res->error = GetLastError();
            goto closeHandle;
        }
    } while (cbNeeded == size);


    // Try to get module's filename and information for each of the
    // modules retrieved by EnumProcessModulesEx. If there's an error,
    // we abort and cleanup everything.
    mCount =  cbNeeded / sizeof (HMODULE);
    modsInfo = calloc(mCount, sizeof * modsInfo);
    for (i = 0; i < mCount; i++) {
        TCHAR buf[MAX_PATH + 1];
        BOOL success = GetModuleFileNameEx(hProcess, aMods[i], buf,
                                           sizeof(buf) / sizeof(TCHAR));
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

    EnumProcessesResponse *res = calloc(1, sizeof * res);

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
    } while (cbNeeded == size);

    res->error = 0;
    res->pids = aProcesses;
    res->length = cbNeeded / sizeof(DWORD);

    return res;
}

void EnumProcessesResponse_Free(EnumProcessesResponse *r) {
    if (r == NULL) {
        return;
    }
    free(r->pids);
    free(r);
}

void EnumProcessesFullResponse_Free(EnumProcessesFullResponse *r) {
    if (r == NULL) {
        return;
    }
    for (DWORD i = 0; i < r->length; i++) {
        free(r->processes[i].filename);
    }
    free(r->processes);
    free(r);
}

void EnumProcessModulesResponse_Free(EnumProcessModulesResponse *r) {
    DWORD i;
    if (r == NULL) {
        return;
    }
    for (i = 0; i < r->length; i++) {
        free(r->modules[i].filename);
    }
    free(r->modules);
    free(r);
}
