#!/bin/bash

[ -d data/beta ] || {
  echo "error: must run from backend"
  exit 2
}

go run ./cmd/cli create game || {
  echo "error: 'create game' failed"
  exit 2
}
go run ./cmd/cli create cluster || {
  echo "error: 'create cluster' failed"
  exit 2
}
go run ./cmd/cli create homeworld || {
  echo "error: 'create homeworld' failed"
  exit 2
}
go run ./cmd/cli create empire --name "First Empire" || {
  echo "error: 'create empire' failed"
  exit 2
}
go run ./cmd/cli show magic-link --empire 1 || {
  echo "error: 'show magic-link' failed"
  exit 2
}
go run ./cmd/cli create empire --name "Second Empire" || {
  echo "error: 'create empire' failed"
  exit 2
}
go run ./cmd/cli show magic-link --empire 2 || {
  echo "error: 'show magic-link' failed"
  exit 2
}
go run ./cmd/cli create empire --name "Third Empire" || {
  echo "error: 'create empire' failed"
  exit 2
}
go run ./cmd/cli show magic-link --empire 3 || {
  echo "error: 'show magic-link' failed"
  exit 2
}
