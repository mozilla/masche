#include "list_libs_windows.h"

#include <windows.h>
#include <psapi.h>
#include <stdlib.h>

EnumProcessesResponse *getAllPids() {
  DWORD size = sizeof(DWORD) * 512;
  DWORD *aProcesses = NULL;
  DWORD cbNeeded, cProcesses;
  
  EnumProcessesResponse *res = calloc(1, sizeof(*res));
  
  // EnumProcesses modifies cbNeeded, setting it to the amount of bytes
  // written into aProcesses. Thus, we need to check if cbNeeded is equal
  // to size. In that case, it means that the array was filled completely and
  // we need to use a bigger array because probably we left elements out.
  do {
    size *= 2;
    aProcesses = realloc(aProcesses, size);
    
    BOOL success = EnumProcesses(aProcesses, size, &cbNeeded);
    if (!success) {
      res->error = GetLastError();
      free(aProcesses);
      return res;
    }
  }
  while(cbNeeded == size);
  
  res->error = 0;
  res->pids = aProcesses;
  res->length = cbNeeded/sizeof(DWORD);
  
  return res;
}

void EnumProcessesResponse_Free(EnumProcessesResponse *r) {
  if (r == NULL) return;
  free(r->pids);
  free(r);
}