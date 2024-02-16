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
rm -rfv /home/storagechain/childnode-linux
rm -rfv /home/storagechain/childnode-linux.zip

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
curl -o /home/storagechain/childnode-linux.zip https://api.storagechain.io/api/file/download/3eTitILNWs3e2d5USz4YGe2525E15WQM
unzip -o /home/storagechain/childnode-linux.zip -d /home/storagechain/

CONFIG_FILE="/home/storagechain/childnode-linux/startup.sh"

# Replace values in the config file
sed -i "s/CLUSTER_NAME=.*/CLUSTER_NAME=${VarClusterName}/" "${CONFIG_FILE}"
sed -i "s/EMAIL=.*/EMAIL=${VarEmail}/" "${CONFIG_FILE}"
sed -i "s/PASSWORD=.*/PASSWORD=${VarPassword}/" "${CONFIG_FILE}"
sed -i "s/NODE_ID=.*/NODE_ID=${VarNodeID}/" "${CONFIG_FILE}"

echo
echo "Config file ${CONFIG_FILE} after variable replacement:"
cat "${CONFIG_FILE}"

# Change directory to childnode-linux and run the main script
cd /home/storagechain/childnode-linux
./main.sh