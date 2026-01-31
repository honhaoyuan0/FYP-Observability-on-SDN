#!/bin/bash
docker compose down && docker compose up --build -d && ./bid_requests.sh