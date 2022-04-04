#!/bin/bash

(
  ifconfig wlo1 2> /dev/null
  ifconfig wlp1s0 2> /dev/null
) | grep -G "192.168.0.*" | awk '{print $2}' | tr -d '\n'
