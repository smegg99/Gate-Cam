#include "common.h"
#include "config.h"

const char* API_HOST = "192.168.1.50";
const int API_PORT = 2138;

#ifdef ENABLE_OTA
const uint16_t OTA_PORT = 3232;
#endif

#ifdef USE_RGB565_FRAMES
const char* pathFormat = "/api/v1/camera/%d/raw_color_frame";
const size_t FRAME_SIZE = TFT_WIDTH * TFT_HEIGHT * 2;
const size_t CHUNK_SIZE = 1024;
#else
const char* pathFormat = "/api/v1/camera/%d/raw_grayscale_frame";
const size_t FRAME_SIZE = TFT_WIDTH * TFT_HEIGHT;
const size_t CHUNK_SIZE = 1024;
#endif

#if defined(ENABLE_DEEP_SLEEP) || defined(ENABLE_REST)
const uint64_t DEEP_SLEEP_INTERVAL = 180000;
#endif

uint8_t* frameBuffer1 = nullptr;
uint8_t* frameBuffer2 = nullptr;
uint8_t* displayBuffer = nullptr;
uint8_t* fetchBuffer = nullptr;

SemaphoreHandle_t bufferMutex = NULL;
SemaphoreHandle_t frameReadySemaphore = NULL;
SemaphoreHandle_t displayUpdateSemaphore = NULL;

TaskHandle_t fetchTaskHandle = NULL;
TaskHandle_t displayTaskHandle = NULL;
TaskHandle_t periphTaskHandle = NULL;
TaskHandle_t httpServerTaskHandle = NULL;

#ifdef RESTART_PERIODICALLY
TaskHandle_t autoRestartTaskHandle;

// Timezone for Warsaw
const char* ntpServer = "pl.pool.ntp.org";
const long gmtOffset_sec = 3600;       // UTC+1
const int daylightOffset_sec = 3600;   // Daylight saving time
const int targetHour = 0;
const int targetMinute = 0;
#endif

volatile bool displayUpdatePending = false;
volatile bool cameraIdChanged = false;
volatile bool frameReady = false;
volatile bool streamAvailable = false;

volatile uint8_t currentCameraID = 0;

#if defined(ENABLE_DEEP_SLEEP) || defined(ENABLE_REST)
unsigned long lastStimulusTime = 0;
#endif

unsigned long lastEncoderChangeTime = 0;
unsigned long lastProcessTime = 0;

volatile bool showStreamFlag = false;
unsigned long streamDisplayStartTime = 0;

DisplayState currentDisplayState = SHOW_CAMERA_STATUS;

TFT_eSPI tft = TFT_eSPI();

#ifndef DISABLE_NETWORKING
WiFiClient wifiClient;
HTTPClient http;

WebServer server(80);
#endif

ESP32Encoder encoder;
volatile bool buzzerOn = false;
unsigned long buzzerOffTime = 0;
const unsigned long BUZZER_DURATION = 20;

#ifndef USE_RGB565_FRAMES
uint16_t grayToRGB565Table[256];
void initGrayToRGB565Table() {
	for (int i = 0; i < 256; i++) {
		grayToRGB565Table[i] = tft.color565(i, i, i);
	}
}
#endif

String buildFetchURL(uint8_t cameraId) {
	char path[50];
	snprintf(path, sizeof(path), pathFormat, cameraId);
	return String("http://") + API_HOST + ":" + String(API_PORT) + path;
}

void setupBuffers() {
	frameBuffer1 = (uint8_t*)malloc(FRAME_SIZE);
	frameBuffer2 = (uint8_t*)malloc(FRAME_SIZE);

	if (frameBuffer1 == NULL || frameBuffer2 == NULL) {
		DEBUG_PRINTLN("Failed to allocate frame buffers.");
		while (true) { delay(1000); }
	}

	memset(frameBuffer1, 0, FRAME_SIZE);
	memset(frameBuffer2, 0, FRAME_SIZE);

	displayBuffer = frameBuffer1;
	fetchBuffer = frameBuffer2;

	DEBUG_PRINTF("Frame buffers allocated and initialized: %d bytes each\n", FRAME_SIZE);
}
