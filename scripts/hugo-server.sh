#!/bin/bash

cd apps/site || {
  echo "error: can't find 'apps/site'"
  exit 2
}

# load the site into the browser
open http://localhost:1313/

hugo server

exit 0
