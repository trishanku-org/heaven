#!/bin/bash

cat README.md | gh-md-toc - | sed 's/\*/-/g' | sed 's/   /  /g' | sed 's/^  //'