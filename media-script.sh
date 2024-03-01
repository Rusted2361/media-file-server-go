#!/bin/bash

#run important services
systemctl restart ipfs.service
systemctl restart ipfs-cluster.service
systemctl restart s3-integration-svc

# Delete all PM2 processes
pm2 delete all

#remove local changes
git stash

# Switch to main branch
git switch main

#pull recent changes
git pull

# Remove files
rm -rf file-server-lin file-server-mac file-server-win.exe

# Run build script
./build.sh

# Start PM2 process for main
pm2 start ./file-server-lin --name Media-Server-Main-V2

#remove local changes
git stash

# Switch to test-main branch
git switch test-main

#get latest changes
git pull

# Remove files
rm -rf file-server-lin file-server-mac file-server-win.exe

# Run build script
./build.sh

# Start PM2 process for staging
pm2 start ./file-server-lin --name Media-Server-Staging-V2

#remove local changes
git stash

# Switch to dev branch
git switch dev

#get latest changes
git pull

# Remove files
rm -rf file-server-lin file-server-mac file-server-win.exe

# Run build script
./build.sh

# Start PM2 process for staging
pm2 start ./file-server-lin --name Media-Server-Dev-V2
