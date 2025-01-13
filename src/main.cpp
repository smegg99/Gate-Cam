#include <Arduino.h>
#include "config.h"
#include "common.h"
#include "display.h"
#include "network.h"
#include "http_server.h"
#include "tasks.h"
#include "periphs.h"
#include "user_setup.h"

void setup() {
  // Even when ENABLE_SERIAL is not defined, it still has to get initialised
  // otherwise it will throw an error, some libraries depend on it I think.
  Serial.begin(115200);
  while (!Serial) { delay(10); }

  delay(1000);

  setCpuFrequencyMhz(240);
  DEBUG_PRINTLN("CPU Frequency: " + String(getCpuFrequencyMhz()) + " MHz");

  setupBuffers();

#ifndef USE_RGB565_FRAMES
  initGrayToRGB565Table();
#endif

  initDisplay();
  delay(2000);

  bufferMutex = xSemaphoreCreateMutex();
  if (bufferMutex == NULL) {
    DEBUG_PRINTLN("Failed to create buffer mutex");
    while (1) { delay(1000); }
  }

  frameReadySemaphore = xSemaphoreCreateBinary();
  if (frameReadySemaphore == NULL) {
    DEBUG_PRINTLN("Failed to create frame ready semaphore");
    while (1) { delay(1000); }
  }

  displayUpdateSemaphore = xSemaphoreCreateBinary();
  if (displayUpdateSemaphore == NULL) {
    DEBUG_PRINTLN("Failed to create display update semaphore");
    while (1) { delay(1000); }
  }

  connectToWiFi();

#ifndef DISABLE_NETWORKING
  xTaskCreatePinnedToCore(fetchTask, "Fetch Task", 8192, NULL, 2, &fetchTaskHandle, 1);
#endif

  xTaskCreatePinnedToCore(displayTask, "Display Task", 8192, NULL, 1, &displayTaskHandle, 0);

#ifdef ENABLE_OTA
  xTaskCreatePinnedToCore(otaTask, "OTA Task", 8192, NULL, 3, NULL, 1);
#endif

  xTaskCreatePinnedToCore(periphTask, "Peripheral Task", 8192, NULL, 1, &periphTaskHandle, 1);

#ifndef DISABLE_NETWORKING
  xTaskCreatePinnedToCore(httpServerTask, "HTTP Server Task", 8192, NULL, 2, &httpServerTaskHandle, 1);
#endif

#if defined(ENABLE_DEEP_SLEEP) || defined(ENABLE_REST)
  xTaskCreatePinnedToCore(powerConservingModeTask, "Power Conserving Mode Task", 8192, NULL, 1, NULL, 1);
#endif

  encoder.attachHalfQuad(ENCODER_PIN_CLK, ENCODER_PIN_DT);
  encoder.clearCount();

#ifdef ENABLE_BUZZER
  pinMode(BUZZER_PIN, OUTPUT);
  digitalWrite(BUZZER_PIN, LOW);
#endif

  pinMode(RELAY_PIN, OUTPUT);
  digitalWrite(RELAY_PIN, LOW);

  pinMode(ENCODER_SWITCH_PIN, INPUT_PULLUP);
}

void loop() {
  vTaskDelay(pdMS_TO_TICKS(1));
}
