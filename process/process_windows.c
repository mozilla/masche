#include "process.h"
#include "process_windows.h"

response_t *open_process_handle(pid_tt pid, process_handle_t *handle) {
    response_t *res = response_create();

    *handle = (uintptr_t) OpenProcess(PROCESS_QUERY_INFORMATION |
            PROCESS_VM_READ,
            FALSE,
            pid);

    if (*handle == 0) {
        res->fatal_error = error_create(GetLastError());
    }

    return res;
}

response_t *close_process_handle(process_handle_t process_handle) {
    //TODO(mvanotti): See which errors should be considered hard and which ones soft.
    response_t *res = response_create();
    BOOL success = CloseHandle((HANDLE) process_handle);
    if (!success) {
        res->fatal_error = error_create(GetLastError());
    }

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

response_t *GetProcessName(process_handle_t hndl, char **name) {
    response_t *res = response_create();
    DWORD bufSize = 128;
    char *buf = calloc(bufSize, sizeof * buf);
    //TODO(mvanotti): GetProcessImageFileName returns the name in "Device Format"
    // ej. \Device\HardDiskVolume4\Windows\foo.exe We can use another function to get it
    // in "normal" format ( C:\Windows\foo.exe )
    DWORD r;
    do {
        r = GetProcessImageFileName((HANDLE) hndl, buf, bufSize);
        if (r == 0) {
            DWORD error = GetLastError();
            // This is the only error that we know how to handle:
            // use a bigger buffer.
            if (error == ERROR_INSUFFICIENT_BUFFER) {
                bufSize *= 2;
                buf = realloc(buf, bufSize);
            } else {
                free(buf);
                res->fatal_error = error_create(error);
                return res;
            }
        }
    } while (r == 0);

    *name = _strdup(buf);
    free(buf);
    return res;
}
