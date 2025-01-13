#ifndef HTTP_SERVER_H
#define HTTP_SERVER_H

#ifndef DISABLE_NETWORKING
#include <WebServer.h>

void handleStatus();
void handleControl();
void httpServerTask(void* parameter);

#endif
#endif