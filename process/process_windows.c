#include "process.h"

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
