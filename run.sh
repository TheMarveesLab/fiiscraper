#!/bin/sh

if ! test -f /app/db/data.json; then
  echo "File not exists. Scraper running..."
  /app/scraper > /app/db/data.json
fi

json-server /app/db/data.json

