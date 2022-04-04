#!/bin/bash

ifconfig wlo1 | grep -G "192.168.0.*" | awk '{print $2}' | tr -d '\n'