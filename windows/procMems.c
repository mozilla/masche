#include <windows.h>
#include <Processthreadsapi.h>
#include <stdlib.h>
#include <stdio.h>

void printInfo(MEMORY_BASIC_INFORMATION info) {
  printf("%p %p %d %llu %d %d %d\n", info.BaseAddress, info.AllocationBase,
                                    info.AllocationProtect, info.RegionSize,
                                    info.State, info.Protect, info.Type);
}

int main(void) {
  DWORD pid = 3308;
  HANDLE hndl = OpenProcess(PROCESS_QUERY_INFORMATION | PROCESS_VM_READ,
					FALSE,
					pid);
	
	if (hndl == NULL) {
	  fprintf(stderr, "Error! %d", GetLastError());
    exit(EXIT_FAILURE);
	}
	
	MEMORY_BASIC_INFORMATION info;
	LPCVOID addr = 0x0;
  while (TRUE) {
    SIZE_T res = VirtualQueryEx(hndl,
                                addr,
                                &info,
                                sizeof(info));
                                
    if (res == 0) {
      DWORD err = GetLastError();
      if (err == ERROR_INVALID_PARAMETER) { 
        // This means that the address we are using is invalid, i.e: no more addresses left!
        break;
      }
      fprintf(stderr, "Error! %d", GetLastError());
      exit(EXIT_FAILURE);
    }
    
    printInfo(info);
    addr += info.RegionSize;
  }
  
  return 0;
}
