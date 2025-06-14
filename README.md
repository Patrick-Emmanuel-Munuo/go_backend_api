# SSL window power shell output 
# Step 1: Ensure the output directory exists
New-Item -ItemType Directory -Path .\configurations -Force
# Step 2: Generate the certificate
$cert = New-SelfSignedCertificate -Subject 'CN=169.254.110.224' -DnsName '169.254.110.224' -CertStoreLocation 'Cert:\LocalMachine\My' -KeyLength 2048 -NotAfter (Get-Date).AddYears(2) -KeyExportPolicy Exportable

# Step 3: Secure the password
$pwd = ConvertTo-SecureString -String 'Variable98@' -Force -AsPlainText

# Step 4: Export the PFX file
Export-PfxCertificate -Cert "Cert:\LocalMachine\My\$($cert.Thumbprint)" -FilePath .\configurations\server.pfx -Password $pwd

# generate .pem files
openssl pkcs12 -in configurations\server.pfx -clcerts -nokeys -out configurations\cert.pem -passin pass:Variable98@
openssl pkcs12 -in configurations\server.pfx -nocerts -nodes -out configurations\key.pem -passin pass:Variable98@








**Description:**  
This is a basic Go (Golang) application built with the **Gin** web framework. It serves as a backend API template for building and deploying services. It supports simple CRUD operations and serves as a foundation for further extensions.

---
### Run go in nodemon
npm install -g nodemon
nodemon --exec go run main.go

### buil go server
go build main.go

#buld for deployment
go mod tidy && go build -tags netgo -ldflags '-s -w' -o app


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
