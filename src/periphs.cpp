#include "periphs.h"
#include "common.h"
#include "config.h"

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
								idChanged = true;
							}
						}
					}
					else if (steps < 0) {
						for (int i = 0; i < (-steps); i++) {
							if (currentCameraID > 0) {
								currentCameraID--;
								DEBUG_PRINTLN("Encoder rotated counter-clockwise. Decrementing Camera ID.");
								idChanged = true;
							}
						}
					}

					if (idChanged) {
						streamAvailable = false;
						showStreamFlag = false;
						cameraIdChanged = true;
						DEBUG_PRINTLN("Stream flags and cameraIdChanged reset due to Camera ID change.");

						if (!displayUpdatePending) {
							displayUpdatePending = true;
							xSemaphoreGive(displayUpdateSemaphore);
						}
					}

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

		if (xSemaphoreTake(bufferMutex, portMAX_DELAY)) {
			currentCameraID = 0;
			streamAvailable = false;
			showStreamFlag = false;
			cameraIdChanged = true;
			DEBUG_PRINTLN("Camera ID reset to 0 and stream flags set to false.");

			if (!displayUpdatePending) {
				displayUpdatePending = true;
				xSemaphoreGive(displayUpdateSemaphore);
			}

			xSemaphoreGive(bufferMutex);
		}

		buzzerOn = true;
		buzzerOffTime = millis() + BUZZER_DURATION;
		digitalWrite(BUZZER_PIN, HIGH);
	}
	lastSwitchState = currentSwitchState;
}

void handleBuzzer() {
	if (buzzerOn && millis() >= buzzerOffTime) {
		digitalWrite(BUZZER_PIN, LOW);
		buzzerOn = false;
	}
}