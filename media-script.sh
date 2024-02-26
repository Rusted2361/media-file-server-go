#!/bin/bash

# Delete all PM2 processes
pm2 delete all

# Switch to main branch
git switch main

# Remove files
rm -rf file-server-lin file-server-mac file-server-win.exe

# Run build script
./build.sh

# Start PM2 process for main
pm2 start ./file-server-lin --name Media-Server-Main-V2

# Switch to test-main branch
git switch test-main

# Remove files
rm -rf file-server-lin file-server-mac file-server-win.exe

# Run build script
./build.sh

# Start PM2 process for staging
pm2 start ./file-server-lin --name Media-Server-Staging-V2

# Switch to dev branch
git switch dev

# Remove files
rm -rf file-server-lin file-server-mac file-server-win.exe

# Run build script
./build.sh

# Start PM2 process for staging
pm2 start ./file-server-lin --name Media-Dev-Staging-V2
