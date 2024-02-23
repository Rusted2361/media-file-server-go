#!/bin/bash

# Retrieve input from the user
read -p 'What is the cluster name you setup on the Storagechain.io website? ' VarClusterName
VarClusterName="\"${VarClusterName}\""

read -p 'Enter your StorageChain.io email address: ' VarEmail
VarEmail="\"${VarEmail}\""

read -p 'Please enter your Password from the StorageChain.io website: ' VarPassword
VarPassword="\"${VarPassword}\""

read -p 'Enter your StorageChain.io node ID from the StorageChain.io website: ' VarNodeID
VarNodeID="\"${VarNodeID}\""

echo
echo "Thank you! Sit back, relax and let's build you a node. This could take up to 20 min..."

# Remove existing files and directories
rm -rfv ~/storagechainnode-linux
rm -rfv ~/storagechainnode-linux.zip

# Delete all pm2 processes
pm2 delete all

# Update package lists
sudo apt update

# Install necessary packages
sudo apt install -y curl nano ufw unzip

# Install pm2 globally
curl -sL https://install.pm2.io | bash

# Allow required ports through firewall
ufw allow 22,3008,4001,5001,8080,9094,9095,9096/tcp

# Download and unzip the childnode-linux package
curl -o ~/storagechainnode-linux.zip https://api.storagechain.io/api/file/download/vizJCgVFNfRXAHGWziitBRkT055IS4yD
unzip -o ~/storagechainnode-linux.zip -d ~/

CONFIG_FILE="~/storagechainnode-linux/startup.sh"
eval CONFIG_FILE="$CONFIG_FILE"

# Replace values in the config file
sed -i "s/CLUSTER_NAME=.*/CLUSTER_NAME=${VarClusterName}/" "${CONFIG_FILE}"
sed -i "s/EMAIL=.*/EMAIL=${VarEmail}/" "${CONFIG_FILE}"
sed -i "s/PASSWORD=.*/PASSWORD=${VarPassword}/" "${CONFIG_FILE}"
sed -i "s/NODE_ID=.*/NODE_ID=${VarNodeID}/" "${CONFIG_FILE}"

echo
echo "Config file ${CONFIG_FILE} after variable replacement:"
cat "${CONFIG_FILE}"

# Change directory to childnode-linux and run the main script
cd ~/storagechainnode-linux
./main.sh