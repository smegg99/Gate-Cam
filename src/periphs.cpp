#include "periphs.h"
#include "common.h"
#include "display.h"
#include "config.h"
#include "network.h"

#ifdef ENABLE_REST
void enterRest() {
	DEBUG_PRINTLN("Preparing to enter rest mode...");

	if (fetchTaskHandle != NULL)
		vTaskSuspend(fetchTaskHandle);
	pushSolidColorFrame(TFT_BLACK);
	if (displayTaskHandle != NULL)
		pushSolidColorFrame(TFT_BLACK);
	vTaskSuspend(displayTaskHandle);
	if (periphTaskHandle != NULL)
		vTaskSuspend(periphTaskHandle);
	if (httpServerTaskHandle != NULL)
		vTaskSuspend(httpServerTaskHandle);

	esp_sleep_enable_ext0_wakeup(static_cast<gpio_num_t>(ENCODER_SWITCH_PIN), 0);
	esp_light_sleep_start();

	DEBUG_PRINTLN("Woke up from rest mode");

	reconnectToWiFi();

	if (fetchTaskHandle != NULL)
		vTaskResume(fetchTaskHandle);
	if (displayTaskHandle != NULL)
		vTaskResume(displayTaskHandle);
	if (periphTaskHandle != NULL)
		vTaskResume(periphTaskHandle);
	if (httpServerTaskHandle != NULL)
		vTaskResume(httpServerTaskHandle);

	resetStimulusTime();
}
#endif

#ifdef ENABLE_DEEP_SLEEP
void enterDeepSleep() {
	DEBUG_PRINTLN("Preparing to enter deep sleep...");

	tft.writecommand(0x10);
	delay(120);

	pushSolidColorFrame(TFT_BLACK);

	if (fetchTaskHandle != NULL)
		vTaskDelete(fetchTaskHandle);
	if (displayTaskHandle != NULL)
		vTaskDelete(displayTaskHandle);
	if (periphTaskHandle != NULL)
		vTaskDelete(periphTaskHandle);
	if (httpServerTaskHandle != NULL)
		vTaskDelete(httpServerTaskHandle);

	if (frameReadySemaphore != NULL)
		vSemaphoreDelete(frameReadySemaphore);
	if (displayUpdateSemaphore != NULL)
		vSemaphoreDelete(displayUpdateSemaphore);
	if (bufferMutex != NULL)
		vSemaphoreDelete(bufferMutex);

#ifndef DISABLE_NETWORKING
	if (wifiClient.connected())
		wifiClient.stop();
	if (WiFi.status() == WL_CONNECTED)
		WiFi.disconnect();
#endif

	DEBUG_PRINTLN("Deep sleep preparation complete. Entering deep sleep...");

	esp_sleep_enable_ext0_wakeup(static_cast<gpio_num_t>(ENCODER_SWITCH_PIN), 0);
	esp_deep_sleep_start();

	DEBUG_PRINTLN("Woke up from deep sleep");
}
#endif

#if defined(ENABLE_DEEP_SLEEP) || defined(ENABLE_REST)

void resetStimulusTime() {
	DEBUG_PRINTLN("Resetting stimulus time...");
	lastStimulusTime = millis();
}

bool shouldGoToSleep() {
	if (millis() - lastStimulusTime >= DEEP_SLEEP_INTERVAL) {
		DEBUG_PRINTLN("No stimulus detected for the past " + String(DEEP_SLEEP_INTERVAL) + " ms. Entering power conserving mode...");
		return true;
	}
	else {
		unsigned long timeLeft = DEEP_SLEEP_INTERVAL - (millis() - lastStimulusTime);
		DEBUG_PRINTLN("Time left before entering power conserving mode: " + String(timeLeft) + " ms.");
		return false;
	}
}
#endif

void handleEncoder() {
	int32_t delta = encoder.getCount();

	if (delta != 0) {
		unsigned long currentTime = millis();
		if (currentTime - lastEncoderChangeTime >= DEBOUNCE_DELAY && currentTime - lastProcessTime >= MIN_PROCESS_INTERVAL) {
			int32_t steps = delta / COUNTS_PER_DETENT;

			if (steps > MAX_STEPS_PER_LOOP) steps = MAX_STEPS_PER_LOOP;
			if (steps < -MAX_STEPS_PER_LOOP) steps = -MAX_STEPS_PER_LOOP;

			if (steps != 0) {
				bool idChanged = false;

				if (xSemaphoreTake(bufferMutex, portMAX_DELAY)) {
					if (steps > 0) {
						for (int i = 0; i < steps; i++) {
							if (currentCameraID < MAX_CAMERAS - 1) {
								currentCameraID++;
								DEBUG_PRINTLN("Encoder rotated clockwise. Incrementing Camera ID.");
#if defined(ENABLE_DEEP_SLEEP) || defined(ENABLE_REST)
								resetStimulusTime();
#endif
								idChanged = true;
							}
						}
					}
					else if (steps < 0) {
						for (int i = 0; i < (-steps); i++) {
							if (currentCameraID > 0) {
								currentCameraID--;
								DEBUG_PRINTLN("Encoder rotated counter-clockwise. Decrementing Camera ID.");
#if defined(ENABLE_DEEP_SLEEP) || defined(ENABLE_REST)
								resetStimulusTime();
#endif
								idChanged = true;
							}
						}
					}

					if (idChanged) {
						streamAvailable = false;
						showStreamFlag = false;
						cameraIdChanged = true;
						DEBUG_PRINTLN("Stream flags and cameraIdChanged reset due to Camera ID change.");

#if defined(ENABLE_DEEP_SLEEP) || defined(ENABLE_REST)
						resetStimulusTime();
#endif

						if (!displayUpdatePending) {
							displayUpdatePending = true;
							if (displayUpdateSemaphore != NULL)
								xSemaphoreGive(displayUpdateSemaphore);
						}
					}

					if (bufferMutex != NULL)
						xSemaphoreGive(bufferMutex);
				}

				encoder.clearCount();
				lastEncoderChangeTime = currentTime;
				lastProcessTime = currentTime;

				if (idChanged) {
					buzzerOn = true;
					buzzerOffTime = currentTime + BUZZER_DURATION;
					digitalWrite(BUZZER_PIN, HIGH);
				}
			}
		}
	}

	static bool lastSwitchState = HIGH;
	bool currentSwitchState = digitalRead(ENCODER_SWITCH_PIN);
	if (lastSwitchState == HIGH && currentSwitchState == LOW) {
		DEBUG_PRINTLN("Encoder switch pressed.");
#if defined(ENABLE_DEEP_SLEEP) || defined(ENABLE_REST)
		resetStimulusTime();
#endif

		if (xSemaphoreTake(bufferMutex, portMAX_DELAY)) {
			currentCameraID = 0;
			streamAvailable = false;
			showStreamFlag = false;
			cameraIdChanged = true;
			DEBUG_PRINTLN("Camera ID reset to 0 and stream flags set to false.");

			if (!displayUpdatePending) {
				displayUpdatePending = true;
				if (displayUpdateSemaphore != NULL)
					xSemaphoreGive(displayUpdateSemaphore);
			}

			if (bufferMutex != NULL)
				xSemaphoreGive(bufferMutex);
		}

		buzzerOn = true;
		buzzerOffTime = millis() + BUZZER_DURATION;
		digitalWrite(BUZZER_PIN, HIGH);
	}
	lastSwitchState = currentSwitchState;
}

void handleBuzzer() {

#if defined(ENABLE_DEEP_SLEEP) || defined(ENABLE_REST)
	if (buzzerOn) {
		resetStimulusTime();
	}
#endif

	if (buzzerOn && millis() >= buzzerOffTime) {
		digitalWrite(BUZZER_PIN, LOW);
		buzzerOn = false;
	}
}