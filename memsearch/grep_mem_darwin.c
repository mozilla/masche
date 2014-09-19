//FIXME(alcuadrado): Document which permissions each function needs to work
//without root.

#include <stdio.h>
#include <stdlib.h>

//FIXME(alcuadrado): Are all these necessary?
#include <mach/vm_map.h>
#include <mach/message.h>
#include <mach/mach_vm.h>
#include <mach-o/dyld_images.h>

/**
 * This is the initial value will use for mach_vm_region_recurse, LLDB considers
 * (LLVM's debugger, and the current oficial OS X debugger) it enough for all
 * cases, so it's a good starting point.
 **/
#define INITIAL_VMMAP_DEPTH 1024

static void print_regions(int pid);
static void print_regions_with_submaps(int pid);
static task_t get_task(int pid);

enum mem_search_status {
    MEM_SEARCH_STATUS_FOUND,
    MEM_SEARCH_STATUS_NOT_FOUND,
    MEM_SEARCH_STATUS_PROCESS_NOT_FOUND
};

/**
 * Searches a sequence of bytes in a mach_region, returing true if found.
 * FIXME(alcuadrado): This should also return the address where found.
 **/
bool search_bytes_in_region(task_t task, mach_vm_address_t region_start,
        mach_vm_size_t region_size, char *bytes, size_t bytes_size) {

    vm_offset_t data;
    mach_msg_type_number_t data_size;
    kern_return_t kret;
    kret = mach_vm_read(task, region_start, region_size, &data, &data_size);

    if (kret != KERN_SUCCESS) {
        //FIXME(alcuadrado): Should we propagate this error?
        //FIXME(alcuadrado): If needed, change the protection.
        fprintf(stderr, "mach_vm_read failed. error: %x %s\n",
                    kret, mach_error_string(kret));
        return false;
    }

    //FIXME(alcuadrado): Implement the actual search here.

    vm_deallocate(mach_task_self(), data, data_size);
    return false;
}


/**
 * Searches for the first occurrence of sequence of bytes in a process' memory
 * starting at initial_address.
 *
 * This function returns a mem_search_status indicating if the bytes where
 * found or not, or if an error ocurred. If the sequence of bytes was found,
 * its initial address (in the process address space) will be available at
 * *found_at.
 **/
enum mem_search_status memory_search_from_address(int pid, char *bytes,
        size_t bytes_size, mach_vm_address_t initial_address,
        mach_vm_address_t *found_at) {

    task_t task;
    kern_return_t kret;

    kret = task_for_pid(mach_task_self(), pid, &task);
    if (kret != KERN_SUCCESS) {
        //FIXME(alcuadrado): How should we propagate this error?
        fprintf(stderr, "task_for_pid failed. error: %x %s\n",
                    kret, mach_error_string(kret));
        return MEM_SEARCH_STATUS_PROCESS_NOT_FOUND;
    }

    struct vm_region_submap_info_64 info;
    mach_msg_type_number_t info_count = 0;
    mach_vm_address_t addr = initial_address;
    mach_vm_size_t size = 0;
    uint32_t depth = INITIAL_VMMAP_DEPTH;

    while (true) {
        info_count = VM_REGION_SUBMAP_INFO_COUNT_64;
        kret = mach_vm_region_recurse(task,
            &addr, &size, &depth,
            (vm_region_recurse_info_t)&info,
            &info_count);

        if (kret == KERN_INVALID_ADDRESS) {
            break;
        }

        if (kret != KERN_SUCCESS) {
            //FIXME(alcuadrado): How should we propagate this error?
            //Or we can keep searching in the next page.
            fprintf(stderr, "vm_region failed. error: %x %s\n",
                    kret, mach_error_string(kret));
            break;
        }

        if(info.is_submap) {
            depth += 1;
            continue;
        }

        printf("Region starting at %p of %llu bytes\n", (void *) addr, size);

        search_bytes_in_region(task, addr, size, bytes, bytes_size);
        addr += size;
    }

    return MEM_SEARCH_STATUS_NOT_FOUND;
}

void try_to_find_and_print(int pid, char *bytes, size_t bytes_size) {
    mach_vm_address_t found_at;
    enum mem_search_status ret = memory_search_from_address(pid, bytes,
        bytes_size, 0, &found_at);

    switch (ret) {
        case MEM_SEARCH_STATUS_FOUND:
            printf("Found at: %llu\n", found_at);
            break;
        case MEM_SEARCH_STATUS_NOT_FOUND:
            printf("Not found\n");
            break;
        case MEM_SEARCH_STATUS_PROCESS_NOT_FOUND:
            printf("Process %d not found\n", pid);
            break;
    }
}

int main(int argc, char **argv) {
    if (argc != 2) {
        printf("Usage: %s <PID> \n", argv[0]);
        exit(1);
    }

    int pid = atoi(argv[1]);
    char deadbeef[] = {0xd, 0xe, 0xa, 0xd, 0xb, 0xe, 0xe, 0xf};
    try_to_find_and_print(pid, deadbeef, sizeof(deadbeef));

    return 0;
}
