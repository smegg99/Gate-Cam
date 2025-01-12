#ifndef NETWORK_H
#define NETWORK_H

#ifdef ENABLE_OTA
#include <ArduinoOTA.h>
#endif

void connectToWiFi();

#ifdef ENABLE_OTA
void otaTask(void* parameter);
#endif

#endif