openssl req -x509 -newkey rsa:2048 -nodes -days 365 \
  -keyout server.key -out server.crt
