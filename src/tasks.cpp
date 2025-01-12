#include "tasks.h"
#include "common.h"
#include "config.h"
#include "display.h"
#include "network.h"
#include "periphs.h"
#include <HTTPClient.h>

void fetchTask(void* parameter) {
	TickType_t lastWakeTime = xTaskGetTickCount();
	String fetchURL;

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

		uint8_t cameraId;

		if (xSemaphoreTake(bufferMutex, portMAX_DELAY)) {
			cameraId = currentCameraID;
			xSemaphoreGive(bufferMutex);
		}

		fetchURL = buildFetchURL(cameraId);

		bool success = false;
		http.begin(wifiClient, fetchURL);
		http.setReuse(true);
		int httpCode = http.GET();

		if (httpCode == HTTP_CODE_OK) {
			int contentLength = http.getSize();
			if (contentLength == FRAME_SIZE) {
				size_t bytesRead = 0;
				WiFiClient* stream = http.getStreamPtr();
				while (stream->available() && bytesRead < FRAME_SIZE) {
					size_t chunk = stream->readBytes(fetchBuffer + bytesRead, FRAME_SIZE - bytesRead);
					if (chunk == 0) break;
					bytesRead += chunk;
				}

				if (bytesRead == FRAME_SIZE) {
					success = true;
					uint8_t fetchedCameraID = cameraId;
					if (xSemaphoreTake(bufferMutex, portMAX_DELAY)) {
						if (fetchedCameraID == currentCameraID) {
							streamAvailable = true;
							showStreamFlag = true;
							streamDisplayStartTime = millis();
							if (!displayUpdatePending) {
								displayUpdatePending = true;
								xSemaphoreGive(displayUpdateSemaphore);
							}
						}
						else {
							DEBUG_PRINTF("Camera ID changed during fetch. Discarding frame for Camera ID: %d\n", fetchedCameraID);
						}
						xSemaphoreGive(bufferMutex);
					}
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
				if (fetchBuffer != nullptr && displayBuffer != nullptr) {
					uint8_t fetchedCameraID = cameraId;
					if (fetchedCameraID == currentCameraID) {
						uint8_t* temp = displayBuffer;
						displayBuffer = fetchBuffer;
						fetchBuffer = temp;

						if (!frameReady) {
							frameReady = true;
							xSemaphoreGive(frameReadySemaphore);
						}
					}
				}
				xSemaphoreGive(bufferMutex);
			}
		}
		else {
			if (xSemaphoreTake(bufferMutex, portMAX_DELAY)) {
				streamAvailable = false;
				showStreamFlag = false;

				if (!displayUpdatePending) {
					displayUpdatePending = true;
					xSemaphoreGive(displayUpdateSemaphore);
				}

				xSemaphoreGive(bufferMutex);
			}
		}

		vTaskDelayUntil(&lastWakeTime, fetchInterval);
	}
}

void displayTask(void* parameter) {
	while (true) {
		if (xSemaphoreTake(displayUpdateSemaphore, portMAX_DELAY)) {
			tft.startWrite();

			if (xSemaphoreTake(bufferMutex, portMAX_DELAY)) {
				if (cameraIdChanged) {
					pushSolidColorFrame(TFT_BLACK);
					displayCameraStatus();
					cameraIdChanged = false;
				}
				else if (showStreamFlag && (millis() - streamDisplayStartTime) < CAMERA_ID_DISPLAY_TIME) {
					pushFrame();
				}
				else if (showStreamFlag && (millis() - streamDisplayStartTime) >= CAMERA_ID_DISPLAY_TIME) {
					showStreamFlag = false;
					pushSolidColorFrame(TFT_BLACK);
					displayCameraStatus();
				}
				else if (!showStreamFlag && streamAvailable) {
					pushFrame();
				}
				else if (!showStreamFlag && !streamAvailable) {
					displayCameraStatus();
				}

				xSemaphoreGive(bufferMutex);
			}

			tft.endWrite();

			if (xSemaphoreTake(bufferMutex, portMAX_DELAY)) {
				displayUpdatePending = false;
				xSemaphoreGive(bufferMutex);
			}
		}
	}
}

void periphTask(void* parameter) {
	while (true) {
		handleEncoder();
		handleBuzzer();
		vTaskDelay(pdMS_TO_TICKS(1));
	}
}