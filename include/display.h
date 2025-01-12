#ifndef DISPLAY_H
#define DISPLAY_H

#include <TFT_eSPI.h>

void initDisplay();
void pushFrame();
void pushSolidColorFrame(uint16_t color);
void displayCameraStatus();

#endif