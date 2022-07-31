#!/bin/bash
for f in `ls assets/*.in`; do go run main.go < $f > assets/`basename $f .in`.out; done
