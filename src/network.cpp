#include "network.h"
#include "common.h"
#include "config.h"
#include "secrets.h"
#include "display.h"

#ifdef ENABLE_OTA
#include <ArduinoOTA.h>
#endif

void connectToWiFi() {
	pushSolidColorFrame(TFT_BLACK);

	tft.setTextSize(1);
	tft.setCursor(0, 0);

	tft.setTextColor(TFT_BLACK, TFT_YELLOW);
	tft.println("Connecting to WiFi:");

	tft.setTextColor(TFT_WHITE, TFT_BLACK);
	tft.println(WIFI_SSID);

	WiFi.disconnect(true);
	WiFi.config(INADDR_NONE, INADDR_NONE, INADDR_NONE);
	delay(1000);
	WiFi.begin(WIFI_SSID, WIFI_PASSWORD);

	tft.println("Connecting");

	while (WiFi.status() != WL_CONNECTED) {
		DEBUG_PRINT(".");
		tft.print(".");
		delay(1000);
	}

	WiFi.setHostname(HOSTNAME);

	tft.setTextColor(TFT_GREEN, TFT_BLACK);
	tft.println("\nConnected!");
	DEBUG_PRINTLN("Connected to WiFi");

	// Sadly, I didn't manage to setup OTA to work reliably with this exact
	// ESP32 board I used in this project, but it might work with other boards.
#ifdef ENABLE_OTA
	tft.setTextColor(TFT_WHITE, TFT_BLACK);
	tft.println("Starting OTA...");

	ArduinoOTA.setPort(OTA_PORT);
	ArduinoOTA.setPassword(OTA_PASSWORD);

	ArduinoOTA.onStart([] () {
		vTaskSuspendAll();
		String type = (ArduinoOTA.getCommand() == U_FLASH) ? "sketch" : "filesystem";
		DEBUG_PRINTLN("Start updating " + type);
		});

	ArduinoOTA.onEnd([] () {
		xTaskResumeAll();
		DEBUG_PRINTLN("OTA End");
		});

	ArduinoOTA.onProgress([] (unsigned int progress, unsigned int total) {
		DEBUG_PRINTF("Progress: %u%%\r", (progress / (total / 100)));
		});

	ArduinoOTA.onError([] (ota_error_t error) {
		DEBUG_PRINTF("Error[%u]: ", error);
		if (error == OTA_AUTH_ERROR) DEBUG_PRINTLN("Auth Failed");
		else if (error == OTA_BEGIN_ERROR) DEBUG_PRINTLN("Begin Failed");
		else if (error == OTA_CONNECT_ERROR) DEBUG_PRINTLN("Connect Failed");
		else if (error == OTA_RECEIVE_ERROR) DEBUG_PRINTLN("Receive Failed");
		else if (error == OTA_END_ERROR) DEBUG_PRINTLN("End Failed");
		});

	ArduinoOTA.begin();

	tft.setTextColor(TFT_GREEN, TFT_BLACK);
	tft.println("OTA Ready");
#endif

	tft.setTextColor(TFT_BLACK, TFT_ORANGE);
	tft.println("IP address:");

	tft.setTextColor(TFT_WHITE, TFT_BLACK);
	tft.println(WiFi.localIP());

	tft.setTextColor(TFT_BLACK, TFT_DARKCYAN);
	tft.printf("Waiting %d seconds...\n", WIFI_SCREEN_LIFESPAN / 1000);

	unsigned long startTime = millis();
	while (millis() - startTime < WIFI_SCREEN_LIFESPAN) {

#ifdef ENABLE_OTA
		ArduinoOTA.handle();
#endif

		delay(100);
	}

	pushSolidColorFrame(TFT_BLACK);

	if (!displayUpdatePending) {
		displayUpdatePending = true;
		xSemaphoreGive(displayUpdateSemaphore);
	}
}

#ifdef ENABLE_OTA
void otaTask(void* parameter) {
	while (true) {
		ArduinoOTA.handle();
		vTaskDelay(pdMS_TO_TICKS(1));
	}
}
#endif