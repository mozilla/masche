#ifndef RESPONSE_H
#define RESPONSE_H

#include <stddef.h>

/**
 * This struct represents an error.
 *
 * error_number is the error as returned by the OS, 0 for no error.
 * description is a malloc'ed null-terminated string.
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
 * Creates a new response without any error.
 **/
response_t *response_create();

/**
 * Releases the resources used by an error response_t, including all error_t's
 * resources.
 **/
void response_free(response_t *errors);

/**
 * Sets a response's fatal error.
 *
 * description is a malloc'ed null-terminated string.
 * NOTE: The response MUST NOT have a fatal error already set.
 **/
void response_set_fatal_error(response_t *response, int error_number,
        char *description);

/**
 * Adds a soft error to a response.
 *
 * description is a malloc'ed null-terminated string.
 **/
void response_add_soft_error(response_t *response, int error_number,
        char *description);

#endif /* RESPONSE_H */

