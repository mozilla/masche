#include "memory_access.h"
#include <stdlib.h>

void response_free(response_t *errors) {};

response_t *open_process_handle(pid_tt pid, process_handle_t *handle) {
    return NULL;
}

response_t *close_process_handle(process_handle_t process_handle) {
    return NULL;
}

response_t *get_next_readable_memory_region(process_handle_t handle,
        void *address, bool *region_available, memory_region_t *memory_region) {
    return NULL;
}

response_t *copy_process_memory(process_handle_t handle, void *start_address,
                                size_t bytes_to_read, void *buffer, size_t *bytes_read) {
    return NULL;
}
