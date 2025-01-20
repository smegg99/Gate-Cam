#ifndef CONFIG_H
#define CONFIG_H

#define USE_RGB565_FRAMES
//#define ENABLE_SERIAL
#define ENABLE_BUZZER
//#define ENABLE_OTA // Didn't work reliably with this exact ESP32 board
//#define ENABLE_DEEP_SLEEP // Some displays may not support it, the backlight may stay on

// A lighter way to conserve power, but not as deep as deep sleep,
// tells the ESP32 to go into light sleep mode and pushes a black screen
// NOTE: When using this, remember to wake the device up with the encoder so the API can be called
//#define ENABLE_REST
//#define DISABLE_NETWORKING // Used for testing the sleeping functionality

#define RESTART_PERIODICALLY // Used to counteract memory leaks which may occur over time due to my poor programming skills

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

#define SET_FPS                60
#define MAX_CAMERAS            16
#define COUNTS_PER_DETENT      2
#define MAX_STEPS_PER_LOOP     5
#define MIN_PROCESS_INTERVAL   50
#define CAMERA_ID_DISPLAY_TIME 500
#define WIFI_SCREEN_LIFESPAN   3000
#define DEBOUNCE_DELAY         50

#endif