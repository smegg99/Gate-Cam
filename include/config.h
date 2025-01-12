#ifndef CONFIG_H
#define CONFIG_H

#define USE_RGB565_FRAMES
// #define ENABLE_SERIAL
#define ENABLE_BUZZER
// #define ENABLE_OTA

#ifdef ENABLE_SERIAL
#define DEBUG_PRINT(x)    Serial.print(x)
#define DEBUG_PRINTLN(x)  Serial.println(x)
#define DEBUG_PRINTF(...) Serial.printf(__VA_ARGS__)
#else
#define DEBUG_PRINT(x)
#define DEBUG_PRINTLN(x)
#define DEBUG_PRINTF(...)
#endif

#define BUZZER_PIN         17
#define RELAY_PIN          16

#define ENCODER_PIN_CLK    0
#define ENCODER_PIN_DT     26
#define ENCODER_SWITCH_PIN 25

#define SET_FPS                40
#define MAX_CAMERAS            16
#define COUNTS_PER_DETENT      2
#define MAX_STEPS_PER_LOOP     5
#define MIN_PROCESS_INTERVAL   50
#define CAMERA_ID_DISPLAY_TIME 500
#define WIFI_SCREEN_LIFESPAN   3000
#define DEBOUNCE_DELAY         50

#endif