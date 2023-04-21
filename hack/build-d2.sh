#!/bin/bash

for i in "$@"; do
    grep -R '^#! d2' "$i" | awk -F ':#!' '{ print "sh -c \"cd $(dirname " $1 ");" $2 "\"" }' | bash
done