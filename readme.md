# TLV TCP File Transfer

A Go-based TCP file transfer system using Type-Length-Value (TLV) protocol for reliable file operations over the network.

## Features

- File upload and download capabilities
- Progress tracking for file transfers
- Support for various file operations:
  - Upload files
  - Delete files
  - Archive files
  - Compress files
  - Read/Download files
- Concurrent connection handling
- Progress bar visualization for client operations
- Configurable timeouts for read and write operations

## Project Structure

```
.
├── client/                 # Client implementation
│   ├── cmmanager/         # Client connection manager
│   │   ├── cm_manager.go  # Client connection handling
│   │   └── command.go     # Command definitions
│   └── main.go            # Client entry point
└── server/                # Server implementation
    ├── cmmanager/         # Server connection manager
    │   ├── cm_manager.go  # Server connection handling
    │   ├── excuter.go     # File operation implementations
    │   └── parser.go      # TLV protocol parsing
    └── main.go            # Server entry point
```

## Protocol Specification

The TLV (Type-Length-Value) protocol is structured as follows:

1. **Type** (2 bytes): Command identifier
   - `0x0001`: Upload command
   - `0x0002`: Delete command
   - `0x0003`: Archive command
   - `0x0004`: Compress command
   - `0x0005`: Read command

2. **Length** (4 bytes): Length of the filename

3. **Value**: Variable length content
   - For filenames: UTF-8 encoded string
   - For file content: Binary data

## Available Commands

- `upload`: Upload a file to the server
- `delete`: Delete a file from the server
- `archive`: Archive a file on the server
- `compress`: Compress a file on the server
- `read`: Download a file from the server

## Getting Started

### Prerequisites

- Go 1.24.0 or higher

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd tlv_tcp
```

2. Install dependencies:
```bash
# For server
cd server
go mod download

# For client
cd ../client
go mod download
```

### Running the Server

```bash
cd server
go run main.go
```

The server will start listening on port 8000 by default.

### Running the Client

The client supports various commands with the following syntax:

```bash
cd client
go run main.go -cmd <command> -f <filename>
```

Examples:

```bash
# Upload a file
go run main.go -cmd upload -f ./path/to/file.txt

# Download a file
go run main.go -cmd read -f ./path/to/file.txt

# Delete a file
go run main.go -cmd delete -f ./path/to/file.txt
```

## Configuration

Both client and server components support configuration of:

- Read timeout
- Write timeout
- Custom message handlers
- Connection callbacks

### Server Configuration Example

```go
cmConfig := cmManager.ConnectionMangerConfig{
    ReadTimeout: 30 * time.Second,
    WriteTimeout: 60 * time.Second,
    OnConnect: func(conn *net.Conn) {
        // Handle new connection
    },
    OnMessage: func(conn *net.Conn, data []byte) {
        // Handle received message
    },
}
```

### Client Configuration Example

```go
config := cmManager.ClientConfig{
    ReadTimeout: 25 * time.Second,
    WriteTimeout: 10 * time.Second,
    OnMessage: func(cmd cmManager.Command, data []byte) {
        // Handle received message
    },
}
```

## Error Handling

The system implements robust error handling for:
- Network disconnections
- Incomplete file transfers
- Invalid commands
- File system errors
- Timeouts

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.


MIT License

Copyright (c) 2024 Yacine BENKAID ALI

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
