# Go Backend API Template

**Description:**  
This is a basic **Go (Golang)** application built using the **Gin** web framework. It serves as a backend API template suitable for building and deploying services. It supports simple CRUD operations and provides a foundation for further extensions.

---

## Development

### Run Go in Development Mode (with live reload)

You can use **nodemon** to automatically restart your Go server during development:

```bash
# Install nodemon globally
npm install -g nodemon

# Run the Go server with live reload
nodemon --exec "go run main.go"

Build the Go Server

To compile your Go application:
# Build the main.go file
go build main.go

Build for Production
For production deployment:

# Clean dependencies and build with optimizations
# Initialize Go modules (if not already initialized)
go mod init go_backend_api

# Tidy and download all dependencies
go mod tidy
go build -tags netgo -ldflags "-s -w" -o app.exe



Flags explanation:
-tags netgo → Forces the use of the Go net package implementation.
-ldflags '-s -w' → Strips debug information to reduce binary size.

Prerequisites
Ensure the following software is installed:
Go 1.18+
Git
Gin Framework

Clone the Repository
# Clone the repository
git clone https://github.com/vartrick98/go_backend_api.git

# Navigate into the project folder
cd go_backend_api

start.bat ```





## Go Backend API Documentation
# API Base URL 
http://serverurl:port/api/v1

# Endpoints
1. /create

POST: Create a single record

Request JSON:

{
    "table": "test",
    "data": {
        "name": "nodejs",
        "password": "vsvgdjdjdgdndeueoeddbd66h",
        "user_name": "VariTrick98"
    }
}


Response JSON:

{
    "success": true,
    "message": {
        "id": 5,
        "name": "nodejs",
        "password": "vsvgdjdjdgdndeueoeddbd66h",
        "user_name": "VariTrick98"
    }
}

2. /bulk-create

POST: Create multiple records in one request

Request JSON:

[
    {
        "table":"test",
        "data":{
            "name":"nodejs",
            "password":"vsvgdjdjdgdndeueoeddbd66h",
            "user_name":"VariTrick98"
        }
    },
    {
        "table":"test",
        "data":{
            "name":"mysql",
            "password":"vsvgdjdjdgdndeueoeddbd66h",
            "user_name":"VariTrick1988"
        }
    }
]


Response JSON:

{
    "success": true,
    "message": [
        {
            "success": true,
            "message": {
                "id": 7,
                "name": "nodejs"
            }
        },
        {
            "success": true,
            "message": {
                "id": 8,
                "name": "mysql"
            }
        }
    ]
}

3. /read

POST: Read a single record

Request JSON:

{
    "table": "test",
    "condition": { "name": "nodejs" }
}


Response JSON:

{
    "success": true,
    "message": {
        "id": 5,
        "name": "nodejs",
        "password": "vsvgdjdjdgdndeueoeddbd66h"
    }
}

4. /bulk-read

POST: Read multiple records

Request JSON:

[
    { "table": "test", "condition": { "name": "nodejs" } },
    { "table": "test", "condition": { "name": "mysql" } }
]


Response JSON:

{
    "success": true,
    "message": [
        { "success": true, "message": { "id": 5, "name": "nodejs" } },
        { "success": true, "message": { "id": 6, "name": "mysql" } }
    ]
}

5. /update

POST: Update a record

Request JSON:

{
    "table": "test",
    "data": { "password": "newpassword123" },
    "condition": { "name": "nodejs" }
}


Response JSON:

{
    "success": true,
    "message": { "affectedRows": 1 }
}

6. /bulk-update

POST: Update multiple records

Request JSON:

[
    {
        "table": "test",
        "data": { "password": "pass1" },
        "condition": { "name": "nodejs" }
    },
    {
        "table": "test",
        "data": { "password": "pass2" },
        "condition": { "name": "mysql" }
    }
]


Response JSON:

{
    "success": true,
    "message": [
        { "success": true, "message": { "affectedRows": 1 } },
        { "success": true, "message": { "affectedRows": 1 } }
    ]
}