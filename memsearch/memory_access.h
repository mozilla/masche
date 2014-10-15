#ifndef __MEMORY_ACCESS__
#define __MEMORY_ACCESS__

/**
 * This header defines a common interface for reading processes' memory.
 *
 *  The basic workflow excluding memory freeing is
 *      1. Open a process handle.
 *      2. Ask for memory region containing address 0 to get the first region.
 *      3. Read whatever you want from the region.
 *      4. Ask for the next region, if available goto 3.
 *      5. Close the process.
 **/

#include <stdbool.h>
#include <stdint.h>

/**
 * Process ID type.
 **/
typedef uint32_t pid_tt;

#ifdef _WIN32

#include <windows.h>

/**
 * Windows specific process handle.
 *
 * This is a workaround for issue 8921: https://code.google.com/p/go/issues/detail?id=8921
 * Go gc panics if process_handle_t is HANDLE (which is PVOID).
 * For now we have to use unsigned long long to hold the HANDLE object.
 **/
typedef unsigned long long process_handle_t;

#endif /* _WIN32 */

#ifdef __MACH__

#include <mach/mach.h>

/**
 * Mac specific process handle.
 **/
typedef task_t process_handle_t;

#endif /* __MACH__ */

/**
 * A type representing a memory address used to represent addresses in the
 * inspected process.
 *
 * NOTE: This is necessary because Go doesn't allow us to have an unsafe pointer
 * with an address that is not mapped in the current process.
 *
 * Portability note: C99 defines unsigned long long to be at least 64 bits long.
 **/
typedef unsigned long long memory_address_t;

/**
 * This struct represents an error.
 *
 * error_number is the error as returned by the OS, 0 for no error.
 * description is a null-terminated string.
 **/
typedef struct {
    int error_number;
    char *description;
} error_t;

/**
 * This struct represents the error releated parts of a response to a function
 * call.
 *
 * fatal_error may point to an error_t that made the operation fail or be NULL.
 * soft_errors may be an array of non-fatal errors or be NULL.
 * soft_errors_count is the number errors in soft_errors (if no array, a 0).
 * soft_errors_capaciy is the syze of the soft_errors array (if no array, a 0).
 **/
typedef struct {
    error_t *fatal_error;
    error_t *soft_errors;
    size_t soft_errors_count;
    size_t soft_errors_capacity;
} response_t;

/**
 * This struct represents a region of readable contiguos memory of a process.
 *
 * No readable memory can be available right next to this region, it's maximal
 * in its upper bound.
 *
 * Note that this region is not necessary equivalent to the OS's region, if any.
 **/
typedef struct {
    memory_address_t start_address;
    size_t length;
} memory_region_t;

/**
 * Releases the resources used by an error response_t, including all error_t's
 * resources.
 **/
void response_free(response_t *errors);

/**
 * Creates a handle for a given process based on its pid.
 *
 * If a fatal error ocurres the handle must not be used, but it must be closed
 * anyway to ensure that all resources are freed.
 **/
response_t *open_process_handle(pid_tt pid, process_handle_t *handle);

/**
 * Closes a specific process handle, freen all its resources.
 *
 * The process_handle_t must not be used after calling this function.
 **/
response_t *close_process_handle(process_handle_t process_handle);

/**
 * Returns a memory region containing address, or the next readable region
 * after address in case it's not readable.
 *
 * If there is no region to return region_available will be false. Otherwise
 * it will be true, and the region will be returned in memory_region.
 **/
response_t *get_next_readable_memory_region(process_handle_t handle,
        memory_address_t address, bool *region_available,
        memory_region_t *memory_region);


/**
 * Copies a chunk of memory from the process' address space to the buffer.
 *
 * Note that start_address is the address as seen by the process.
 *
 * If no fatal error ocurred the buffer will be populated with bytes_read bytes.
 *
 * It's caller responsibility to provide a big enough buffer.
 **/
response_t *copy_process_memory(process_handle_t handle,
                                memory_address_t start_address,
                                size_t bytes_to_read, void *buffer, size_t *bytes_read);

#endif /* __MEMORY_ACCESS__ */

