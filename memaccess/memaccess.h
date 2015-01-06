#ifndef MEMACCES_H
#define MEMACCES_H

#include <stdbool.h>
#include <stdint.h>

#include "../cresponse/response.h"

/**
 * Process ID type.
 **/
typedef uint32_t pid_tt;

#ifdef _WIN32

#include <windows.h>

/**
 * Windows specific process handle.
 *
 * NOTE: We use uintptr_t instead of HANDLE because Go doesn't allow
 * pointers with invalid values. Windows' HANDLE is a PVOID internally and
 * sometimes it is used as an integer.
 **/
typedef uintptr_t process_handle_t;

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
 * NOTE: This is necessary because Go doesn't allow us to have an unsafe.Pointer
 * with an address that is not mapped in the current process.
 **/
typedef uintptr_t memory_address_t;

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
 * after address in case addresss is not in a readable region.
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
 * If no fatal error ocurred the buffer will be populated with bytes_read bytes.
 * It's caller's responsibility to provide a big enough buffer.
 **/
response_t *copy_process_memory(process_handle_t handle,
        memory_address_t start_address, size_t bytes_to_read, void *buffer,
        size_t *bytes_read);

#endif /* MEMACCES_H */

