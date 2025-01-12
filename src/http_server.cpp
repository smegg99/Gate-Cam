#include "http_server.h"
#include "common.h"
#include "config.h"
#include <ArduinoJson.h>

void handleStatus() {
	JsonDocument jsonDoc;

	jsonDoc["wifi_status"] = (WiFi.status() == WL_CONNECTED) ? "connected" : "disconnected";
	jsonDoc["ip_address"] = WiFi.localIP().toString();

	if (xSemaphoreTake(bufferMutex, pdMS_TO_TICKS(100))) {
		jsonDoc["current_camera_id"] = currentCameraID;
		jsonDoc["stream_available"] = streamAvailable;
		xSemaphoreGive(bufferMutex);
	}
	else {
		jsonDoc["current_camera_id"] = "unavailable";
		jsonDoc["stream_available"] = "unavailable";
	}

	String response;
	serializeJson(jsonDoc, response);
	server.send(200, "application/json", response);
}

void handleControl() {
	if (server.hasArg("plain") == false) {
		server.send(400, "application/json", "{\"error\":\"Bad Request\"}");
		return;
	}

	String body = server.arg("plain");

	JsonDocument jsonDoc;
	DeserializationError error = deserializeJson(jsonDoc, body);

	if (error) {
		server.send(400, "application/json", "{\"error\":\"Invalid JSON\"}");
		return;
	}

	bool relayState = false;
	if (jsonDoc["relay"].is<bool>()) {
		relayState = jsonDoc["relay"];
		digitalWrite(RELAY_PIN, relayState ? HIGH : LOW);
	}

	bool buzzerState = false;
	if (jsonDoc["buzzer"].is<bool>()) {
		buzzerState = jsonDoc["buzzer"];
		digitalWrite(BUZZER_PIN, buzzerState ? HIGH : LOW);
	}

	if (jsonDoc["camera_id"].is<int>()) {
		int newCameraID = jsonDoc["camera_id"];
		if (newCameraID >= 0 && newCameraID < MAX_CAMERAS) {
			if (xSemaphoreTake(bufferMutex, pdMS_TO_TICKS(100))) {
				currentCameraID = newCameraID;
				streamAvailable = false;
				showStreamFlag = false;
				cameraIdChanged = true;
				DEBUG_PRINTF("Camera ID set to %d via API.\n", currentCameraID);

				if (!displayUpdatePending) {
					displayUpdatePending = true;
					xSemaphoreGive(displayUpdateSemaphore);
				}

				xSemaphoreGive(bufferMutex);
			}
		}
		else {
			server.send(400, "application/json", "{\"error\":\"Invalid camera_id\"}");
			return;
		}
	}

	server.send(200, "application/json", "{\"status\":\"OK\"}");
}

void httpServerTask(void* parameter) {
	server.on("/status", HTTP_GET, handleStatus);
	server.on("/control", HTTP_POST, handleControl);

	server.onNotFound([] () {
		server.send(404, "application/json", "{\"error\":\"Not Found\"}");
		});

	server.begin();
	DEBUG_PRINTLN("HTTP server started.");

	while (true) {
		server.handleClient();
		vTaskDelay(pdMS_TO_TICKS(1));
	}
}