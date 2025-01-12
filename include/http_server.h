#ifndef HTTP_SERVER_H
#define HTTP_SERVER_H

#include <WebServer.h>

void handleStatus();
void handleControl();
void httpServerTask(void* parameter);

#endif