Media File Server
This repository contains the implementation of a media file server with various functionalities. Below, you'll find an overview of the implemented modules, APIs, and their use cases.

Modules & Use Cases
1. File Controller
Implemented APIs:
getstatus API:

Description: Returns true if the process 1 (ipfs-cluster-runner) is running.
Endpoint: /api/file/node/status
Example URL: http://39.62.5.88:3008/api/file/node/status
play video API:

Description: Play video when access token passed in URL.
Endpoint: /api/file/view/access-play/:access-token
Example URL: http://39.62.5.88:3008/api/file/view/access-play/unRyi9hfZVOye6KiNXCFg1eO91SJufs5
getaccessfile API:

Description: Access static file (image, doc, etc.) when access token passed in URL.
Endpoint: /api/file/view/access/:access-key
Example URL: http://39.62.5.88:3008/api/file/view/access/<access-key-of-static-file>
downloadfile API:

Description: Access & download static file (image, doc, etc.) when access token passed in URL.
Endpoint: /api/file/download/:access-key
Example URL: http://39.62.5.88:3008/api/file/download/<access-key-of-static-file>
Test Cases:
Play Video Test:

Should play the video when the access token is passed in the URL.
Access Static File Test:

Should access static file (image, doc, etc.) when the access token is passed in the URL.
Download Static File Test:

Should access & download static file (image, doc, etc.) when the access token is passed in the URL.
2. Recursive Functions
Implemented Functions:
Heart Beat (Check Health):

Description: Log health check after every 5 seconds.
Endpoint: /api/file/health-check
Save Node Details:

Description: Save node details (ipaddress of machine, ipfs_id, cluster_id) when the node starts.
Endpoint: /api/file/save-node-details
Cron Job in video.service.ts (Delete Junk Video Data):

Description: Delete video junk data after every 6 hours.
Usage
Clone the repository and run the application. Make API requests using the provided endpoints.

Brief description of your project.

## Building Binaries

To build binaries for Windows, Linux, and macOS, you can use the provided `build.sh` script. Follow the steps below:

1. Clone the repository:

   ```bash
   git clone https://github.com/your-username/your-repository.git
   cd your-repository
   ./build.sh

Feel free to contribute or report issues!