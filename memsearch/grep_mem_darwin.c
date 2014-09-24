//FIXME(alcuadrado): Document which permissions each function needs to work
//without root.

#include <stdio.h>
#include <stdlib.h>
#include <assert.h>

#include <mach/mach.h>
#include <mach/mach_vm.h>

/**
 * This struct represents an error.
 *
 * error_number is the error as returned by the OS, 0 for no error.
 * description is a null-terminated string.
 **/
struct error {
    int error_number;
    char *description;
};

/**
 * This struct represents the result of a memory search operation.
 *
 * error may point to a fatal error or be NULL if no such error ocurred.
 * soft_errors is a dynamic array containing all the non-fatal errors in order.
 * soft_errors_length is the amount of soft errors in soft_errors.
 * soft_errors_capacity is the number of errors that fit in soft_errors.
 * located_at is the address where the pattern was found, or NULL if not found.
 *
 * NOTE: This structs must be created and destroyed with
 *  new_mem_search_return and free_mem_search_return.
 **/
struct mem_search_return {
    struct error *error;
    struct error *soft_errors;
    unsigned int soft_errors_count;
    unsigned int soft_errors_capacity;
    void *located_at;
};

/**
 * Allocates a new mem_search_return in the heap and initializes it.
 **/
struct mem_search_return *new_mem_search_return(void) {
    return calloc(1, sizeof(struct mem_search_return));
}

/**
 * Frees a struct mem_search_return.
 **/
void free_mem_search_return(struct mem_search_return *ret) {
    if (ret == NULL) {
        return;
    }

    if (ret->error != NULL) {
        free(ret->error);
        ret->error = NULL;
    }

    assert((ret->soft_errors_capacity > 0) == (ret->soft_errors != NULL));
    if (ret->soft_errors != NULL) {
        free(ret->soft_errors);
        ret->soft_errors = NULL;
    }

    ret->located_at = NULL;
    free(ret);
}

/**
 * Sets a mem_search_return's error.
 **/
void mem_search_return_set_error(struct mem_search_return *ret,
        kern_return_t error_number) {
    if (error_number == KERN_SUCCESS) {
        return;
    }

    struct error *error = malloc(sizeof(struct error));
    error->error_number = error_number;
    error->description = mach_error_string(error_number);
    ret->error = error;
}

/**
 * Adds a soft error to a mem_search_return.
 **/
void mem_search_return_add_soft_error(struct mem_search_return *ret,
        kern_return_t error_number) {
    if (error_number == KERN_SUCCESS) {
        return;
    }

#define SOFT_ERRORS_INITIAL_CAPACITY 2
#define SOFT_ERRORS_REALLOCATION_FACTOR 2

    if (ret->soft_errors_capacity == 0) {
        ret->soft_errors = malloc(
                sizeof(struct error) * SOFT_ERRORS_INITIAL_CAPACITY);
        ret->soft_errors_capacity = SOFT_ERRORS_INITIAL_CAPACITY;

    } else if (ret->soft_errors_capacity == ret->soft_errors_count) {
        ret->soft_errors = realloc(ret->soft_errors,
                ret->soft_errors_capacity * SOFT_ERRORS_REALLOCATION_FACTOR);
        ret->soft_errors_capacity *= SOFT_ERRORS_REALLOCATION_FACTOR;
    }

    ret->soft_errors[ret->soft_errors_count].error_number = error_number;
    ret->soft_errors[ret->soft_errors_count].description =
        mach_error_string(error_number);
    ret->soft_errors_count++;
}


/**
 * Searches for the first occurrence of sequence of bytes in a process' memory
 * starting at initial_address.
 **/
struct mem_search_return * memory_search_from_address(int pid, char *bytes,
        size_t bytes_size, mach_vm_address_t initial_address) {
    task_t task;
    kern_return_t kret;
    struct mem_search_return *ret = new_mem_search_return();

    kret = task_for_pid(mach_task_self(), pid, &task);
    if (kret != KERN_SUCCESS) {
        mem_search_return_set_error(ret, kret);
        return ret;
    }

    struct vm_region_submap_info_64 info;
    mach_msg_type_number_t info_count = 0;
    mach_vm_address_t addr = initial_address;
    mach_vm_size_t size = 0;
    uint32_t depth = 0;

    for (;;) {
        info_count = VM_REGION_SUBMAP_INFO_COUNT_64;
        kret = mach_vm_region_recurse(task,
            &addr, &size, &depth,
            (vm_region_recurse_info_t)&info,
            &info_count);

        if (kret == KERN_INVALID_ADDRESS) {
            break;
        }

        if (kret != KERN_SUCCESS) {
            mem_search_return_set_error(ret, kret);
            return ret;
        }

        if(info.is_submap) {
            depth += 1;
            continue;
        }

        printf("Region starting at %p of %llu bytes\n", (void *) addr, size);

        //The actual search should be done here, but it's not a straight forward
        //by-region search, as we have to search in between regions.

        addr += size;
    }

    return ret;
}

/* #<{(|* */
/*  * Searches a sequence of bytes in a mach_region, returing true if found. */
/*  * FIXME(alcuadrado): This should also return the address where found. */
/*  *|)}># */
/* bool search_bytes_in_region(task_t task, mach_vm_address_t region_start, */
/*         mach_vm_size_t region_size, char *bytes, size_t bytes_size) { */
/*     vm_offset_t data; */
/*     mach_msg_type_number_t data_size; */
/*     kern_return_t kret; */
/*     kret = mach_vm_read(task, region_start, region_size, &data, &data_size); */
/*  */
/*     if (kret != KERN_SUCCESS) { */
/*         //FIXME(alcuadrado): Should we propagate this error? */
/*         //FIXME(alcuadrado): If needed, change the protection. */
/*         fprintf(stderr, "mach_vm_read failed. error: %x %s\n", */
/*                     kret, mach_error_string(kret)); */
/*         return false; */
/*     } */
/*  */
/*     //FIXME(alcuadrado): Implement the actual search here. */
/*  */
/*     vm_deallocate(mach_task_self(), data, data_size); */
/*     return false; */
/* } */
/*  */
/* void try_to_find_and_print(int pid, char *bytes, size_t bytes_size) { */
/*     mach_vm_address_t found_at; */
/*     enum mem_search_status ret = memory_search_from_address(pid, bytes, */
/*         bytes_size, 0, &found_at); */
/*  */
/*     switch (ret) { */
/*         case MEM_SEARCH_STATUS_FOUND: */
/*             printf("Found at: %llu\n", found_at); */
/*             break; */
/*         case MEM_SEARCH_STATUS_NOT_FOUND: */
/*             printf("Not found\n"); */
/*             break; */
/*         case MEM_SEARCH_STATUS_PROCESS_NOT_FOUND: */
/*             printf("Process %d not found\n", pid); */
/*             break; */
/*     } */
/* } */

int main(int argc, char **argv) {
    if (argc != 2) {
        printf("Usage: %s <PID> \n", argv[0]);
        exit(1);
    }

    int pid = atoi(argv[1]);
    char deadbeef[] = {0xd, 0xe, 0xa, 0xd, 0xb, 0xe, 0xe, 0xf};
    /* try_to_find_and_print(pid, deadbeef, sizeof(deadbeef)); */

    return 0;
}
