#include "config.h"

#ifndef PERIPHS_H
#define PERIPHS_H

#ifdef ENABLE_REST
void enterRest();
#endif

#ifdef ENABLE_DEEP_SLEEP
void enterDeepSleep();
#endif

#if defined(ENABLE_DEEP_SLEEP) || defined(ENABLE_REST)
bool shouldGoToSleep();
void resetStimulusTime();
#endif

void handleEncoder();
void handleBuzzer();

#endif
