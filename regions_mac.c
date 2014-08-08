// This program lists the different mapped regions of a given pid.
// It does so in two different ways: using mach_vm_region and
// mach_vm_region_recurse. The latter recurses trough the different
// submaps of the process.
//
// In order to run it, use ./regions <pid> (with root permissions)
// 
// If you want to take a look at the source code, first look at print_regions.
//
// After getting all the regions one can see an arbitrary memory address
// within the regions with vm_read (or mach_vm_read).
//
// There are some things that I'm missing:
// * What exactly is a submap?
// * Is mach_vm_region_recurse working properly?
// * There is also mach_vm_region_recurse64 should we use?
// * We have mach_vm_read and vm_read, mach_vm_region and vm_region
//   which ones should we use? What are the differences?
// * Even though we see all the regions, there are some differences with
//   the ones reported by the vmmap program. Why is that?
#include <mach/vm_map.h>
#include <mach/message.h>
#include <mach/mach_vm.h>
#include <stdio.h>
#include <stdlib.h>
#include <mach-o/dyld_images.h>

static void print_regions(int pid);
static void print_regions_with_submaps(int pid);
static task_t get_task(int pid);

int main(int argc, char **argv) {
    if (argc != 2) {
        printf("Usage: %s <PID> \n", argv[0]);
        exit(1);
    }

    int pid = atoi(argv[1]);
    print_regions(pid);
    print_regions_with_submaps(pid);

    return 0;
}

static task_t get_task(int pid) {
    task_t task;
    // acquire task port
    kern_return_t kret = task_for_pid(mach_task_self(), pid, &task);
    if (kret != KERN_SUCCESS) {
        fprintf(stderr, "task_for_pid failed. pid: %d; error: %x %s\n",
                pid, kret, mach_error_string(kret));
        exit(1);
    }
    return task;
}

static void print_regions_with_submaps(int pid) {
    printf("print_regions_with_submaps\n");
    task_t task = get_task(pid);

    struct vm_region_submap_info_64 info;
    mach_msg_type_number_t info_count = 0;
    mach_vm_address_t addr = 0;
    mach_vm_size_t  size = 0;
    uint32_t depth = 0;
    kern_return_t kret;
 
    do {
        info_count = VM_REGION_SUBMAP_INFO_COUNT_64;
        kret = mach_vm_region_recurse(task,
                        &addr, &size, &depth,
                        (vm_region_recurse_info_t)&info,
                        &info_count);
        if (kret == KERN_INVALID_ADDRESS) {
            break;
        }
        if (kret != KERN_SUCCESS) {
            fprintf(stderr, "vm_region failed. error: %x %s\n",
                    kret, mach_error_string(kret));
            exit(1);
        }

        if(info.is_submap) {
            depth += 1;
        } else {
            printf("%16p - %16p\t%10llu\n", (void*) addr, (void *) (addr + size), size);
            addr += size;
        }
    } while (1);
}

static void print_regions(int pid) {
    task_t task = get_task(pid);

    // stuff for calling mach_vm_region
    mach_vm_address_t addr = 0; // start searching regions from address 0.
    mach_vm_size_t size = 0;
    vm_region_flavor_t flavor = VM_REGION_BASIC_INFO;
    vm_region_basic_info_data_t info;
    mach_msg_type_number_t info_count = 0;
    mach_port_t object_name; // this should be unused.

    printf("Address\t\t\t\tSize\n");

    // mach_vm_region will automatically update the addr and size variables,
    // addr will have the address of the beginning of the region and size will
    // have its size. To iterate over all regions, we add size to addr. 
    do {
        addr += size;
        info_count = VM_REGION_BASIC_INFO_COUNT_64;
        kern_return_t kret = mach_vm_region(task, &addr, &size, flavor, 
                (vm_region_info_t) &info, &info_count, &object_name);
        if (kret == KERN_INVALID_ADDRESS) {
            break;
        }
        if (kret != KERN_SUCCESS) {
            fprintf(stderr, "vm_region failed. error: %x %s\n",
                    kret, mach_error_string(kret));
            exit(1);
        }
        printf("%16p - %16p\t%10llu\n", (void*) addr, (void *) (addr + size), size);
    } while (1);
    
}
