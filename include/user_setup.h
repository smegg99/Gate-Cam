// Remember to change the #include path in the TFT_eSPI User_Setup_Select.h to this file's path.
#ifndef USER_SETUP_H
#define USER_SETUP_H

#define ST7735_DRIVER
#define ST7735_BLACKTAB
#define TFT_WIDTH            160
#define TFT_HEIGHT           128
#define TFT_CS               4
#define TFT_DC               2
#define TFT_RST              13
#define TFT_MOSI             23
#define TFT_SCLK             18
#define SPI_FREQUENCY        27000000 // 27 MHz might be unstable, but works for me
#define SUPPORT_TRANSACTIONS

// Most of them are not needed for this project, but I like to keep them here for future reference
#define LOAD_GLCD
#define LOAD_FONT2
#define LOAD_FONT4
#define LOAD_FONT6
#define LOAD_FONT7
#define LOAD_FONT8
#define LOAD_GFXFF
#define SMOOTH_FONT

#define TOUCH_CS			 21 // Not used in this project, just to silence warnings

#endif