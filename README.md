# Go Application
requirements:
go get go.mongodb.org/mongo-driver
go get gopkg.in/gomail.v2




**Description:**  
This is a basic Go (Golang) application built with the **Gin** web framework. It serves as a backend API template for building and deploying services. It supports simple CRUD operations and serves as a foundation for further extensions.

---
### Run go in nodemon
npm install -g nodemon
nodemon --exec go run  main.go

### buil go server
go build main.go


 Response

json
    { "success": true, "message": "data"}

    { "success": false, "message": "error response "}

## Routes Available
## Table of Contents

1. [Installation](#installation)
2. [Project Structure](#project-structure)
3. [Backend Setup](#backend-setup)
4. [API Endpoints](#api-endpoints)
   - [GET /](#1-get-)
   - [GET /api/generate-otp](#2-get-apigenerate-otp)
   - [POST /api/create](#3-post-apicreate)
5. [Usage](#usage)
6. [Running the Application](#running-the-application)
7. [Environment Variables](#environment-variables)
8. [Contributing](#contributing)
9. [License](#license)

---

## Installation

### Prerequisites

Make sure you have the following installed:

- Go (1.18+)
- Git
- [Gin framework](https://github.com/gin-gonic/gin)

### Clone the Repository

First, clone the repository to your local machine:

```bash
git clone https://github.com/yourusername/go-application.git
cd go-application
