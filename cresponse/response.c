#include <stdlib.h>
#include <assert.h>

#include "response.h"

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

response_t *response_create() {
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

void response_set_fatal_error(response_t *response, int error_number,
        char *description) {
    assert(response->fatal_error == NULL);
}

void response_add_soft_error(response_t *response, int error_number,
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

