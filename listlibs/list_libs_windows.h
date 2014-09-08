#ifndef _LIST_LIBS_WINDOWS_H_
#define _LIST_LIBS_WINDOWS_H_

#include <windows.h>
#include <psapi.h>

typedef struct t_EnumProcessesResponse {
	DWORD error;
	DWORD *pids;
	DWORD length;
} EnumProcessesResponse;

typedef struct t_ModuleInfo {
	char *filename;
	MODULEINFO info;
} ModuleInfo;

typedef struct t_EnumProcessModulesResponse {
	DWORD error;
	DWORD length;
	ModuleInfo *modules;
} EnumProcessModulesResponse;

EnumProcessesResponse *getAllPids();
EnumProcessModulesResponse *getModules(DWORD pid);

void EnumProcessesResponse_Free(EnumProcessesResponse *r);
void EnumProcessModulesResponse_Free(EnumProcessModulesResponse *r);

#endif // _LIST_LIBS_WINDOWS_H_