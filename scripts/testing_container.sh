#!/usr/bin/env bash
set -e

COMPOSE="docker compose -f docker-compose.yml -f docker-compose.test.yml"

case "$1" in
  build)
    $COMPOSE build test
    ;;
  run)
    $COMPOSE up -d postgres redis rabbitmq test
    $COMPOSE exec test sh
    ;;
  stop)
    $COMPOSE down
    ;;
  *)
    echo "Usage: $0 {build|run|stop}"
    exit 1
    ;;
esac