#!/bin/bash

source .envrc.secret
reflex -s -- sh -c "invalidate-devserver && go run ."
