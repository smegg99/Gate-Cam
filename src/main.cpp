#define USE_RGB565_FRAMES
// #define ENABLE_SERIAL

#ifdef ENABLE_SERIAL
#define DEBUG_PRINT(x) Serial.print(x)
#define DEBUG_PRINTLN(x) Serial.println(x)
#define DEBUG_PRINTF(...) Serial.printf(__VA_ARGS__)
#else
#define DEBUG_PRINT(x)
#define DEBUG_PRINTLN(x)
#define DEBUG_PRINTF(...)
#endif

#define SET_FPS 60

#include "User_Setup.h"
#include "secrets.h"
#include <Arduino.h>
#include <TFT_eSPI.h>
#include <WiFi.h>
#include <HTTPClient.h>
#include <esp32-hal-cpu.h>

// Constants and Sizes
const char* host = "192.168.1.42";
const int port = 2138;

#ifdef USE_RGB565_FRAMES
const char* path = "/api/v1/camera/gate/raw_color_frame";
const size_t FRAME_SIZE = TFT_WIDTH * TFT_HEIGHT * 2; // RGB565: 2 bytes per pixel
#else
const char* path = "/api/v1/camera/gate/raw_grayscale_frame";
const size_t FRAME_SIZE = TFT_WIDTH * TFT_HEIGHT;     // Grayscale: 1 byte per pixel
#endif

const String FETCH_FRAME_URL = String("http://") + host + ":" + String(port) + path;

// Frame Buffers
uint8_t* frameBuffer1 = nullptr;
uint8_t* frameBuffer2 = nullptr;
uint8_t* displayBuffer = frameBuffer1;
uint8_t* fetchBuffer = frameBuffer2;

SemaphoreHandle_t bufferMutex;
SemaphoreHandle_t frameReadySemaphore;

TFT_eSPI tft = TFT_eSPI();
WiFiClient wifiClient;
HTTPClient http;

#ifdef USE_RGB565_FRAMES
#else
uint16_t grayToRGB565Table[256] PROGMEM;
void initGrayToRGB565Table() {
  for (int i = 0; i < 256; i++) {
    grayToRGB565Table[i] = tft.color565(i, i, i);
  }
}
#endif

void connectToWiFi() {
  DEBUG_PRINT("Connecting to WiFi");
  WiFi.begin(WIFI_SSID, WIFI_PASSWORD);

  int retryCount = 0;
  while (WiFi.status() != WL_CONNECTED && retryCount < 10) {
    delay(1000);
    DEBUG_PRINT(".");
    retryCount++;
  }

  if (WiFi.status() == WL_CONNECTED) {
    DEBUG_PRINTLN("\nConnected to WiFi");
    DEBUG_PRINT("IP Address: ");
    DEBUG_PRINTLN(WiFi.localIP());
  }
  else {
    DEBUG_PRINTLN("\nFailed to connect to WiFi");
    ESP.restart();
  }
}

void fetchTask(void* parameter) {
  TickType_t lastWakeTime = xTaskGetTickCount();
#ifdef SET_FPS
  const TickType_t fetchInterval = pdMS_TO_TICKS(1000 / SET_FPS);
#else
  const TickType_t fetchInterval = pdMS_TO_TICKS(35); // Default to ~28.57 FPS
#endif

  while (true) {
    if (WiFi.status() != WL_CONNECTED) {
      DEBUG_PRINTLN("WiFi disconnected. Reconnecting...");
      connectToWiFi();
    }

    bool success = false;
    http.begin(wifiClient, FETCH_FRAME_URL);
    http.setReuse(true);
    int httpCode = http.GET();

    if (httpCode == HTTP_CODE_OK) {
      int contentLength = http.getSize();
      if (contentLength == FRAME_SIZE) {
        size_t bytesRead = 0;
        WiFiClient* stream = http.getStreamPtr();
        while (stream->available() && bytesRead < FRAME_SIZE) {
          bytesRead += stream->readBytes(fetchBuffer + bytesRead, FRAME_SIZE - bytesRead);
        }

        if (bytesRead == FRAME_SIZE) {
          success = true;
          DEBUG_PRINTLN("Frame fetched successfully");
        }
        else {
          DEBUG_PRINTF("Incomplete frame received: %d bytes\n", bytesRead);
        }
      }
      else {
        DEBUG_PRINTF("Unexpected frame size: expected %d, got %d\n", FRAME_SIZE, contentLength);
      }
    }
    else {
      DEBUG_PRINTF("HTTP GET failed, code: %d\n", httpCode);
    }

    http.end();

    if (success) {
      if (xSemaphoreTake(bufferMutex, portMAX_DELAY)) {
        if (displayBuffer != NULL && fetchBuffer != NULL) {
          uint8_t* temp = displayBuffer;
          displayBuffer = fetchBuffer;
          fetchBuffer = temp;
        }
        xSemaphoreGive(bufferMutex);
      }
      xSemaphoreGive(frameReadySemaphore);
    }

    vTaskDelayUntil(&lastWakeTime, fetchInterval);
  }
}

void displayTask(void* parameter) {
  while (true) {
    if (xSemaphoreTake(frameReadySemaphore, portMAX_DELAY)) {
      DEBUG_PRINTLN("Displaying fetched frame...");
      tft.startWrite();
      tft.setAddrWindow(0, 0, TFT_WIDTH, TFT_HEIGHT);

#ifdef USE_RGB565_FRAMES
      tft.pushPixels((uint16_t*)displayBuffer, TFT_WIDTH * TFT_HEIGHT);
#else
      for (size_t i = 0; i < FRAME_SIZE; i++) {
        uint16_t colorData = grayToRGB565Table[displayBuffer[i]];
        tft.writeColor(colorData, 1);
      }
#endif

      tft.endWrite();
      DEBUG_PRINTLN("Frame displayed successfully");
    }
  }
}

void testDisplay() {
  DEBUG_PRINTLN("Testing display...");
  tft.fillScreen(TFT_BLACK);
  for (int y = 0; y < TFT_HEIGHT; y++) {
    for (int x = 0; x < TFT_WIDTH; x++) {
      uint16_t color = tft.color565((x * 256) / TFT_WIDTH, (y * 256) / TFT_HEIGHT, ((x + y) * 256) / (TFT_WIDTH + TFT_HEIGHT));
      tft.drawPixel(x, y, color);
    }
  }
  DEBUG_PRINTLN("Display test complete.");
}

void initDisplay() {
  DEBUG_PRINTLN("Initializing display...");
  tft.init();
  tft.setRotation(1);
  tft.fillScreen(TFT_BLACK);
  DEBUG_PRINTLN("Display initialized.");
}

void setupBuffers() {
  frameBuffer1 = (uint8_t*)malloc(FRAME_SIZE);
  frameBuffer2 = (uint8_t*)malloc(FRAME_SIZE);

  if (frameBuffer1 == NULL || frameBuffer2 == NULL) {
    DEBUG_PRINTLN("Failed to allocate frame buffers.");
    while (true);
  }

  displayBuffer = frameBuffer1;
  fetchBuffer = frameBuffer2;

  DEBUG_PRINTF("Frame buffers allocated: %d bytes each\n", FRAME_SIZE);
}

void setup() {
  setCpuFrequencyMhz(240);

#if defined(ENABLE_SERIAL)
  Serial.begin(115200);
  delay(1000);
#endif

  setupBuffers();

  DEBUG_PRINTLN("CPU Frequency: " + String(getCpuFrequencyMhz()) + " MHz");

#ifndef USE_RGB565_FRAMES
  initGrayToRGB565Table();
#endif
  initDisplay();
  testDisplay();

  connectToWiFi();

  bufferMutex = xSemaphoreCreateMutex();
  if (bufferMutex == NULL) {
    DEBUG_PRINTLN("Failed to create mutex");
    while (1);
  }

  frameReadySemaphore = xSemaphoreCreateBinary();
  if (frameReadySemaphore == NULL) {
    DEBUG_PRINTLN("Failed to create semaphore");
    while (1);
  }

  xTaskCreatePinnedToCore(fetchTask, "Fetch Task", 4096, NULL, 1, NULL, 1);
  xTaskCreatePinnedToCore(displayTask, "Display Task", 2048, NULL, 2, NULL, 0);
}

void loop() {
  vTaskDelay(pdMS_TO_TICKS(100));
}
