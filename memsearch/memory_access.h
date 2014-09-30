#ifndef __MEMORY_ACCESS__
#define __MEMORY_ACCESS__

//TODO: File level documentations.
//TODO: Define the different process_handle_t

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
    error_t error;
    process_handle_t process_handle;
} process_handle_response_t;

//TODO: Add doc.
process_handle_response_t *open_process_handle(pid_t pid);

//TODO: Add doc.
response_errors_t *close_process_handle(process_handle_t *process_handle);

//TODO: Add doc.
memory_region_response_t get_next_readable_memory_region(
    process_handle_t handle, void *start_address);

#endif /* __MEMORY_ACCESS__ */
