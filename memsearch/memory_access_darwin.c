#include <stdlib.h>
#include <assert.h>

#include <mach/mach_vm.h>

#include "memory_access.h"

/**
 * Creates a new error_t for a given error number.
 **/
static error_t *error_create(kern_return_t error_number) {
    error_t *error = malloc(sizeof(*error));
    error->error_number = (int) error_number;
    error->description = mach_error_string(error_number);
    return error;
}

/**
 * Frees an error_t.
 **/
static void error_free(error_t *error) {
    if (error == NULL) {
        return;
    }

    //as descriptions are staticly allocated we just need to free the error_t
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

    //as error descriptions are staticaly allocated we can free the entire
    //array of errors at once.
    if (response->soft_errors != NULL) {
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
    response->fatal_error = error_create(error_number);
}

/**
 * Adds a soft error to a response.
 **/
static void response_add_soft_error(response_t *response,
        kern_return_t error_number) {

#define SOFT_ERRORS_INITIAL_CAPACITY 2
#define SOFT_ERRORS_REALLOCATION_FACTOR 2

    if (response->soft_errors_capacity == 0) {
        response->soft_errors_count = 0;
        response->soft_errors_capacity = SOFT_ERRORS_INITIAL_CAPACITY;
        response->soft_errors = calloc(SOFT_ERRORS_INITIAL_CAPACITY,
                sizeof(error_t));
    }

    if (response->soft_errors_count == response->soft_errors_capacity) {
        response->soft_errors_capacity *= SOFT_ERRORS_REALLOCATION_FACTOR;
        response->soft_errors = realloc(response->soft_errors,
                response->soft_errors_capacity);
    }

    response->soft_errors[response->soft_errors_count].error_number =
        error_number;
    response->soft_errors[response->soft_errors_count].description =
        mach_error_string(error_number);
    response->soft_errors_count++;
}

response_t *open_process_handle(pid_t pid, process_handle_t *handle) {
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
        void *address, bool *region_available, memory_region_t *memory_region) {
    response_t *response = response_create();

    kern_return_t kret;
    struct vm_region_submap_info_64 info;
    mach_msg_type_number_t info_count = 0;
    mach_vm_address_t addr = (mach_vm_address_t) address;
    mach_vm_size_t size = 0;
    uint32_t depth = 0;
    *region_available = false;

    for (;;) {
        info_count = VM_REGION_SUBMAP_INFO_COUNT_64;
        kret = mach_vm_region_recurse(handle,
            &addr, &size, &depth,
            (vm_region_recurse_info_t)&info,
            &info_count);

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

        if ((info.protection & VM_PROT_READ) == VM_PROT_READ) {
            //TODO(alcuadrado): Add a soft error here.

            if (*region_available) {
                return response;
            }
        } else {
            if (!(*region_available)) {
                *region_available = true;
                memory_region->start_address = (void *) addr;
                memory_region->length = size;
            } else {
                memory_region->length += size;
            }
        }

        addr += size;
    }

    return response;
}

response_t *copy_process_memory(process_handle_t handle, void *start_address,
        size_t bytes_to_read, void *buffer, size_t *bytes_read) {
    response_t *response = response_create();


    //TODO(alcuadrado): Implement this function.

    return response;
}

