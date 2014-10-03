#ifndef __MEMORY_ACCESS__
#define __MEMORY_ACCESS__

//TODO: File level documentations.
//TODO: Define the different process_handle_t

#ifdef _WIN32
#include <windows.h>
/**
 * Windows specific process handle.
 * handle represent the handle to an process object
 * pid is the process id.
 **/
typedef struct {
    HANDLE handle;
    DWORD pid;
} process_handle_t;
#endif

#define pid_t uint_t

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
 * soft_errors_count is the number errors in soft_errors (if nor array, a 0).
 **/
typedef struct {
    error_t *fatal_error;
    error_t *soft_errors;
    size_t soft_errors_count;
} response_errors_t;

// response_errors_free releases the resources used by the given errors.
void response_errors_free(response_errors_t *errors);

/**
 * This struct represents a region of readable contiguos memory of a process.
 *
 * No readable memory can be contiguos to this region, it's maximal in its
 * upper bound.
 *
 * Note that this region is not necessary equivalent to the OS's region.
 **/
typedef struct {
    void *start_address;
    size_t length;
} memory_region_t;

/**
 * Response for an open_process_handle request, with the handle and any possible
 * error.
 **/
typedef struct {
    response_errors_t *errors;
    process_handle_t process_handle;
} process_handle_response_t;

/**
 * open_process_handle creates a handle for a given process based on its pid.
 * this handle is used for interacting with all the other functions in this
 * library.
 * This function returns a process_handle_response_t *, which contains
 * the error in case there was an error (NULL otherwise), and a pointer to
 * a process_handle, which may be invalid if the error is non-NULL.
 * In order to release all the resources allocated by this function,
 * if there's an error, the caller must call response_errors_free with
 * the errors, otherwise, the caller must call close_process_handle with
 * the process_handle.
 **/
process_handle_response_t open_process_handle(pid_t pid);

/**
 * close_process_handle closes a specific process handle.
 * The given handle should not be used anymore after calling this function.
 * An error will be returned upon failure, or NULL if the function succeeded.
 * Note that some errors might mean that the resources have already been freed
 * (for example, if the process dies before calling this function).
 **/
error_t *close_process_handle(process_handle_t process_handle);

/**
 * Response for a memory_region request (get_next_readable_memory_region
 * for example), with the region information and any possible error.
 */
typedef struct {
    response_errors_t *errors;
    memory_region_t region;
} memory_region_response_t;

// memory_region_response_free releases the resources used by the given region.
void memory_region_response_free(memory_region_response *region);

/**
 * get_next_readable_memory_region returns a memory_region that is readable
 * and starts at an address greater or equal than start_address.
 * If no error ocurred, the error field will be NULL.
 * The response errors should be freed using response_errors_free.
 **/
memory_region_response_t get_next_readable_memory_region(
    process_handle_t handle, void *start_address);

// TODO: Add doc.
typedef struct {
    response_errors_t *errors;
    void *address;
} memory_search_response_t;

/**
 * find_next_occurrence searches the memory starting at address start_address
 * for the n-bytes given in buf.
 * If the errors are non-NULL, they should be freed using response_errors_free.
 **/
memory_search_response_t find_next_occurrence(
    process_handle_t handle, void *start_address, char buf[], int n);

// TODO: Add doc.
typedef struct {
    response_errors_t *errors;
    int size;
    char *data;
} memory_read_response_t

// TODO: Add doc.
memory_read_response_t *read_memory(process_handle_t handle,
                                    void *start_address, int size);

// TODO: Add doc.
void memory_read_response_free(memory_read_response_t *mem);


#endif /* __MEMORY_ACCESS__ */
