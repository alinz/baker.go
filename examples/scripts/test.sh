#!/bin/bash
for i in {1..50000}
do
   curl -H "Host: example1.com" http://localhost/sample1/hahaha
   sleep 0.1
done