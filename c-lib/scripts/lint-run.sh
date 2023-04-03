#!/bin/bash
#

docker run --rm     -e RUN_LOCAL=true     --env-file ".github/super-linter.env"     -v "$PWD":/tmp/lint github/super-linter:slim-v5 2> linter.txt
