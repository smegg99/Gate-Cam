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

const char*  host            = "192.168.1.42";
const int    port            = 3001;
const char*  path            = "/api/v1/cameras/balls/raw_grayscale_frame";

const size_t FRAME_SIZE      = TFT_WIDTH * TFT_HEIGHT;
const String FETCH_FRAME_URL = String("http://") + host + ":" + String(port) + path;


uint8_t  frameBuffer1[FRAME_SIZE];
uint8_t  frameBuffer2[FRAME_SIZE];
uint8_t* displayBuffer = frameBuffer1;
uint8_t* fetchBuffer   = frameBuffer2;


uint16_t grayToRGB565Table[256];

SemaphoreHandle_t bufferMutex;
SemaphoreHandle_t frameReadySemaphore;


TFT_eSPI tft = TFT_eSPI();
WiFiClient wifiClient;
HTTPClient http;

void initGrayToRGB565Table() {
  for (int i = 0; i < 256; i++) {
    grayToRGB565Table[i] = tft.color565(i, i, i);
  }
}

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
  } else {
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
        } else {
          DEBUG_PRINTF("Incomplete frame received: %d bytes\n", bytesRead);
        }
      } else {
        DEBUG_PRINTF("Unexpected frame size: expected %d, got %d\n", FRAME_SIZE, contentLength);
      }
    } else {
      DEBUG_PRINTF("HTTP GET failed, code: %d\n", httpCode);
    }

    http.end();

    if (success) {
      if (xSemaphoreTake(bufferMutex, portMAX_DELAY)) {
        uint8_t* temp = displayBuffer;
        displayBuffer = fetchBuffer;
        fetchBuffer = temp;
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

      for (size_t i = 0; i < FRAME_SIZE; i++) {
        uint16_t colorData = grayToRGB565Table[displayBuffer[i]];
        tft.writeColor(colorData, 1);
      }

      tft.endWrite();
      DEBUG_PRINTLN("Frame displayed successfully");
    }
  }
}

void setup() {
  setCpuFrequencyMhz(240);

  #if defined (ENABLE_SERIAL)
    Serial.begin(115200);
    delay(1000);
  #endif

  DEBUG_PRINTLN("CPU Frequency: " + String(getCpuFrequencyMhz()) + " MHz");

  initGrayToRGB565Table();

  DEBUG_PRINTLN("Initializing display...");
  tft.init();
  tft.setRotation(1);
  tft.fillScreen(TFT_BLACK);
  DEBUG_PRINTLN("Display initialized.");

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

  xTaskCreatePinnedToCore(fetchTask, "Fetch Task", 8192, NULL, 1, NULL, 1);
  xTaskCreatePinnedToCore(displayTask, "Display Task", 8192, NULL, 2, NULL, 0);
}

void loop() {}