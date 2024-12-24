#include <Arduino.h>

void setup() {
  pinMode(9, OUTPUT);
  Serial.begin(115200);
}

void loop() {
  digitalWrite(9, HIGH);
  delay(100);
  digitalWrite(9, LOW);
  delay(100);
  Serial.println("Hello World");
}