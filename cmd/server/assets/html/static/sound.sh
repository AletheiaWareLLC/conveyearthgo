#!/bin/bash

angle=0.25

x=$(echo "scale=2; 240 * c(${angle})" | bc -l)
y=$(echo "scale=2; 240 * s(${angle})" | bc -l)
echo -e "\t\t<path d=\"M ${x},$(echo "scale=2; 145-${y}" | bc) Q 240,145 ${x},$(echo "scale=2; 145+${y}" | bc)\" stroke-width=\"2\" />"
x=$(echo "scale=2; 312 * c(${angle})" | bc -l)
y=$(echo "scale=2; 312 * s(${angle})" | bc -l)
echo -e "\t\t<path d=\"M ${x},$(echo "scale=2; 145-${y}" | bc) Q 312,145 ${x},$(echo "scale=2; 145+${y}" | bc)\" stroke-width=\"6\" />"
x=$(echo "scale=2; 384 * c(${angle})" | bc -l)
y=$(echo "scale=2; 384 * s(${angle})" | bc -l)
echo -e "\t\t<path d=\"M ${x},$(echo "scale=2; 145-${y}" | bc) Q 384,145 ${x},$(echo "scale=2; 145+${y}" | bc)\" stroke-width=\"10\" />"

x=$(echo "scale=2; 276 * c(${angle})" | bc -l)
y=$(echo "scale=2; 276 * s(${angle})" | bc -l)
echo -e "\t\t<path d=\"M ${x},$(echo "scale=2; 145-${y}" | bc) Q 276,145 ${x},$(echo "scale=2; 145+${y}" | bc)\" stroke-width=\"4\" />"
x=$(echo "scale=2; 348 * c(${angle})" | bc -l)
y=$(echo "scale=2; 348 * s(${angle})" | bc -l)
echo -e "\t\t<path d=\"M ${x},$(echo "scale=2; 145-${y}" | bc) Q 348,145 ${x},$(echo "scale=2; 145+${y}" | bc)\" stroke-width=\"8\" />"
x=$(echo "scale=2; 420 * c(${angle})" | bc -l)
y=$(echo "scale=2; 420 * s(${angle})" | bc -l)
echo -e "\t\t<path d=\"M ${x},$(echo "scale=2; 145-${y}" | bc) Q 420,145 ${x},$(echo "scale=2; 145+${y}" | bc)\" stroke-width=\"12\" />"