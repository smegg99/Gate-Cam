#ifndef TASKS_H
#define TASKS_H

void fetchTask(void* parameter);
void displayTask(void* parameter);
void periphTask(void* parameter);

#if defined(ENABLE_DEEP_SLEEP) || defined(ENABLE_REST)
void powerConservingModeTask(void* parameter);
#endif

#ifdef RESTART_PERIODICALLY
void autoRestartTask(void* parameter);
#endif

#endif