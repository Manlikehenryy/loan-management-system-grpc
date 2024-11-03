# Microservices-Based Loan Management System

A microservices-based Loan Management System API built with gRPC, Golang, Gin, and MongoDB. The system employs JWT for secure authentication, with tokens stored in cookies. Key features include modular microservices for managing user authentication, loan applications, and a digital wallet for handling transactions and balances. This architecture enhances scalability, security, and ease of management across loan and wallet services.


## Clone the repository

git clone [https://github.com/henry-mbamalu/loan-management-system-grpc.git](https://github.com/henry-mbamalu/loan-management-system-grpc.git)

## After cloning, navigate to the root directory

## Navigate to the apiGateway folder

    cd apiGateway

## Create a .env file, copy the content of example.env to the .env file

    cp example.env .env

## Install Dependencies

    go get ./...

## Run the app

    go run main.go


## Open a new terminal, navigate to the loanService folder

    cd loanService

## Create a .env file, copy the content of example.env to the .env file

    cp example.env .env

## Install Dependencies

    go get ./...

## Run the app

    go run main.go

## Open a new terminal, navigate to the userService folder

    cd userService

## Create a .env file, copy the content of example.env to the .env file

    cp example.env .env

## Install Dependencies

    go get ./...

## Run the app

    go run main.go

## Open a new terminal, navigate to the walletService folder

    cd walletService

## Create a .env file, copy the content of example.env to the .env file

    cp example.env .env

## Install Dependencies

    go get ./...

## Run the app

    go run main.go
.

# REST API

## Signup

### Request

`POST /api/register`

     http://localhost:50054/api/register

     {
	"firstName": "name",
    "lastName": "name",
    "username": "user",
    "password": "user"
    }

### Response

    HTTP/1.1 201 CREATED
    Status: 201 CREATED
    Content-Type: application/json

    {
    "message": "User registered successfully!"
    }

## Login

### Request

`POST /api/login`

    http://localhost:50054/api/login

    {
    "username": "user",
    "password": "user"
    }

### Response

    HTTP/1.1 200 OK
    Status: 200 OK
    Content-Type: application/json
   

    {
    "message": "Logged in successfully"
    }

## Apply for a loan

### Request

`POST /api/loan/apply-loan`

    http://localhost:50054/api/loan/apply-loan

    token needs to be stored in cookies

    {
     "amount": 100000,
     "duration": 6
    }
### Response

    HTTP/1.1 201 Created
    Status: 201 Created
    Content-Type: application/json


    {
    "data": {
        "loanId": "67266b5d038812f286a83cfe"
    },
    "message": "Loan application submitted"
    }
## Approve a loan

### Request

`PUT /api/loan/approve-loan`

    http://localhost:50054/api/loan/approve-loan

    token needs to be stored in cookies

    {
    "loanId": "67266b5d038812f286a83cfe",
    "approvedAmount": 10000,
    "tenure": 4,
    "monthlyRepayment": 200,
    "effectiveDate": "2024-01-03",
    "expiryDate": "2024-02-04"
    }

### Response

    HTTP/1.1 200 OK
    Status: 200 OK
    Content-Type: application/json


    {
    "message": "Loan approved successfully"
    }

## Reject a loan

### Request

`PUT /api/loan/reject-loan`

    http://localhost:50054/api/loan/reject-loan

    token needs to be stored in cookies

    {
     "loanId": "67266b5d038812f286a83cfe"
    }

### Response

    HTTP/1.1 200 OK
    Status: 200 OK
    Content-Type: application/json


    {
     "message": "Loan rejected successfully"
    }
