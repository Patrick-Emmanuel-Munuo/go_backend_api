Go Backend API Documentation

Project: Go Backend API Template
Framework: Gin Web Framework

Author: Vartrick98
Language: Go (Golang)

Description:
This is a backend API template built with Go and Gin. It supports CRUD operations, bulk operations, authentication, and data encryption. It serves as a foundation for building scalable Go-based APIs.

Table of Contents

Development Setup

Build & Production

API Base URL

Endpoints

/create

/bulk-create

/read

/bulk-read

/update

/bulk-update

/delete

/bulk-delete

/search

/search-between

/count

/count-bulk

/backup

/query

/database-handle

Development Setup
Run Go in Development Mode
# Install nodemon globally
npm install -g nodemon

# Run the Go server with live reload
nodemon --exec "go run main.go"

Build Go Server
# Compile main.go
go build main.go

Build for Production
# Initialize Go modules
go mod init go_backend_api

# Download dependencies
go mod tidy

# Build optimized binary
go build -tags netgo -ldflags "-s -w" -o app.exe


Flags Explanation:

-tags netgo → Forces Go to use net package implementation

-ldflags "-s -w" → Removes debug info to reduce binary size

Prerequisites

Go 1.18+

Git

Gin Framework

Clone Repository
git clone https://github.com/vartrick98/go_backend_api.git
cd go_backend_api

API Base URL
http://localhost:PORT/api/v1

Endpoints
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