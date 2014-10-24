#include <stdlib.h>
#include <assert.h>
#include <stdio.h>

#include <mach/mach_vm.h>

#include "memaccess.h"

/**
 * Creates a new error_t for a given kern_return_t.
 **/
static error_t *error_create_from_kern_return_t(kern_return_t error_number) {
    error_t *error = malloc(sizeof(*error));
    error->error_number = (int) error_number;
    error->description = strdup(mach_error_string(error_number));
    return error;
}

/**
 * Frees an error_t.
 **/
static void error_free(error_t *error) {
    if (error == NULL) {
        return;
    }

    free(error->description);
    free(error);
}

/**
 * Creates a new response without any error.
 **/
static response_t *response_create() {
    return calloc(1, sizeof(response_t));
}

void response_free(response_t *response) {
    if (response == NULL) {
        return;
    }

    error_free(response->fatal_error);
    if (response->soft_errors != NULL) {
        for (size_t i = 0; i < response->soft_errors_count; i++) {
            free(response->soft_errors[i].description);
        }
        free(response->soft_errors);
    }

    free(response);
}

/**
 * Sets a response's fatal error.
 *
 * The response MUST NOT have a fatal error already set.
 **/
static void response_set_fatal_error(response_t *response,
        kern_return_t error_number) {
    assert(response->fatal_error == NULL);
    response->fatal_error = error_create_from_kern_return_t(error_number);
}

/**
 * Adds a soft error to a response.
 *
 * description is a heap-allocated null-terminated string.
 **/
static void response_add_soft_error(response_t *response, int error_number,
        char *description) {

#define SOFT_ERRORS_INITIAL_CAPACITY 2
#define SOFT_ERRORS_REALLOCATION_FACTOR 2

    if (response->soft_errors_capacity == 0) {
        response->soft_errors_count = 0;
        response->soft_errors_capacity = SOFT_ERRORS_INITIAL_CAPACITY;
        response->soft_errors = calloc(SOFT_ERRORS_INITIAL_CAPACITY,
                sizeof(*response->soft_errors));
    }

    if (response->soft_errors_count == response->soft_errors_capacity) {
        response->soft_errors_capacity *= SOFT_ERRORS_REALLOCATION_FACTOR;
        response->soft_errors = realloc(response->soft_errors,
                response->soft_errors_capacity *
                sizeof(*response->soft_errors));
    }

    response->soft_errors[response->soft_errors_count].error_number =
        error_number;
    response->soft_errors[response->soft_errors_count].description =
        description;
    response->soft_errors_count++;
}

response_t *open_process_handle(pid_tt pid, process_handle_t *handle) {
    task_t task;
    kern_return_t kret;
    response_t *response = response_create();

    kret = task_for_pid(mach_task_self(), pid, &task);
    if (kret != KERN_SUCCESS) {
        response_set_fatal_error(response, kret);
    } else {
        *handle = task;
    }

    return response;
}

response_t *close_process_handle(process_handle_t process_handle) {
    kern_return_t kret;
    response_t *response = response_create();

    kret = mach_port_deallocate(mach_task_self(), process_handle);
    if (kret != KERN_SUCCESS) {
        response_set_fatal_error(response, kret);
    }

    return response;
}

response_t *get_next_readable_memory_region(process_handle_t handle,
        memory_address_t address, bool *region_available,
        memory_region_t *memory_region) {
    response_t *response = response_create();

    kern_return_t kret;
    struct vm_region_submap_info_64 info;
    mach_msg_type_number_t info_count = 0;
    mach_vm_address_t addr = address;
    mach_vm_size_t size = 0;
    uint32_t depth = 0;
    *region_available = false;

    for (;;) {
        info_count = VM_REGION_SUBMAP_INFO_COUNT_64;
        kret = mach_vm_region_recurse(handle, &addr, &size, &depth,
                (vm_region_recurse_info_t)&info, &info_count);

        if (kret == KERN_INVALID_ADDRESS) {
            break;
        }

        if (kret != KERN_SUCCESS) {
            response_set_fatal_error(response, kret);
            return response;
        }

        if(info.is_submap) {
            depth += 1;
            continue;
        }

        if ((info.protection & VM_PROT_READ) != VM_PROT_READ) {
            char *description = NULL;
            asprintf(
                &description,
                "memory unreadable: %llx-%llx",
                addr,
                addr + size - 1
            );
            response_add_soft_error(response, -1, description);

            if (*region_available) {
                return response;
            }
        } else {
            if (!(*region_available)) {

                // Sometimes a previous region is returned that doesn't contain,
                // address. This would lead to an infinite loop while using
                // the regions, getting every time the same one. To avoid this
                // we ask for the region 1 byte after address.
                if (addr + size <= address) {
                    char *description = NULL;
                    asprintf(
                        &description,
                        "wrong region obtained, expected it to contain %llx, "
                        "but got: %llx-%llx",
                        address,
                        addr,
                        addr + size - 1
                    );
                    response_add_soft_error(response, -1, description);

                    addr = address + 1;
                    continue;
                }

                *region_available = true;
                memory_region->start_address = addr;
                memory_region->length = size;
            } else {
                memory_address_t limit_address = memory_region->start_address +
                    memory_region->length;

                if (limit_address < addr) {
                    return response;
                }

                mach_vm_size_t overlaped_bytes = limit_address - addr;
                memory_region->length += size - overlaped_bytes;
            }
        }

        addr += size;
    }

    return response;
}

response_t *copy_process_memory(process_handle_t handle,
        memory_address_t start_address, size_t bytes_to_read, void *buffer,
        size_t *bytes_read) {

    response_t *response = response_create();

    mach_vm_size_t read;
    kern_return_t kret = mach_vm_read_overwrite(handle, start_address,
            bytes_to_read, (mach_vm_address_t) buffer, &read);

    if (kret != KERN_SUCCESS) {
        response_set_fatal_error(response, kret);
        return response;
    }

    *bytes_read = read;
    return response;
}

