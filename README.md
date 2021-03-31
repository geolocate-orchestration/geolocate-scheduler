# aida-scheduler

[![Test](https://github.com/aida-dos/aida-scheduler/actions/workflows/test.yml/badge.svg?branch=develop)](https://github.com/aida-dos/aida-scheduler/actions/workflows/test.yml)

## Development

```
ksync watch

# If config not created:
ksync create --selector=component=aida-scheduler --reload=false --local-read-only=true $(pwd) /code

keti aida-scheduler-55dd5b4747-l6l6z -- sh
  cd /code
  go run main.go
```
