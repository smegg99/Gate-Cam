#include "display.h"
#include "common.h"
#include "config.h"

void initDisplay() {
	DEBUG_PRINTLN("Initializing display...");
	tft.init();
	tft.setRotation(3);
	pushSolidColorFrame(TFT_BLACK);
	DEBUG_PRINTLN("Display initialized.");
}

void pushFrame() {
	tft.startWrite();
	tft.setAddrWindow(0, 0, TFT_WIDTH, TFT_HEIGHT);

#ifdef USE_RGB565_FRAMES
	size_t totalPixels = TFT_WIDTH * TFT_HEIGHT;
	size_t pixelsPushed = 0;

	while (pixelsPushed < totalPixels) {
		size_t pixelsToPush = min(CHUNK_SIZE, totalPixels - pixelsPushed);
		tft.pushPixels((uint16_t*)(displayBuffer + pixelsPushed * 2), pixelsToPush);
		pixelsPushed += pixelsToPush;
		vTaskDelay(pdMS_TO_TICKS(1));
	}
#else
	size_t totalPixels = FRAME_SIZE;
	size_t pixelsProcessed = 0;

	while (pixelsProcessed < totalPixels) {
		size_t pixelsToProcess = min(CHUNK_SIZE, totalPixels - pixelsProcessed);
		for (size_t i = 0; i < pixelsToProcess; i++) {
			uint16_t colorData = pgm_read_word_near(grayToRGB565Table + displayBuffer[pixelsProcessed + i]);
			tft.writeColor(colorData, 1);
		}
		pixelsProcessed += pixelsToProcess;
		vTaskDelay(pdMS_TO_TICKS(1));
	}
#endif

	tft.endWrite();
}

void pushSolidColorFrame(uint16_t color) {
	tft.startWrite();
	tft.setAddrWindow(0, 0, TFT_WIDTH, TFT_HEIGHT);

	size_t totalPixels = TFT_WIDTH * TFT_HEIGHT;
	size_t pixelsPushed = 0;

	uint16_t* solidColorBuffer = new uint16_t[CHUNK_SIZE];
	for (size_t i = 0; i < CHUNK_SIZE; i++) {
		solidColorBuffer[i] = color;
	}

	while (pixelsPushed < totalPixels) {
		size_t pixelsToPush = min(CHUNK_SIZE, totalPixels - pixelsPushed);
		tft.pushPixels(solidColorBuffer, pixelsToPush);
		pixelsPushed += pixelsToPush;
		vTaskDelay(pdMS_TO_TICKS(1));
	}
	delete[] solidColorBuffer;

	tft.endWrite();
}

void displayCameraStatus() {
	tft.setTextSize(1);
	tft.setTextColor(TFT_WHITE, TFT_BLACK);

	String camIdStr = "Cam ID: " + String(currentCameraID);

	int16_t x1, y1;
	uint16_t w, h;
	w = tft.textWidth(camIdStr);
	h = tft.fontHeight();
	int16_t x = (TFT_WIDTH - w) / 2;
	int16_t y = (TFT_HEIGHT - h) / 3;

	tft.setCursor(x, y);
	tft.println(camIdStr);

	tft.setTextSize(1);
	String statusStr = streamAvailable ? "online" : "offline";
	uint16_t statusColor = streamAvailable ? TFT_GREEN : TFT_RED;

	tft.setTextColor(statusColor, TFT_BLACK);
	w = tft.textWidth(statusStr);
	h = tft.fontHeight();
	int16_t statusX = (TFT_WIDTH - w) / 2;
	int16_t statusY = y + h + 10;

	tft.setCursor(statusX, statusY);
	tft.println(statusStr);
}
