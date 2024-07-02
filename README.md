# User Wallet and Money Transfer System

This project is a Golang backend system that maintains user wallets and performs money transfers between users within the platform. It consists of two microservices: `user-service` and `transactions-service`.

## Overview

- **User-service**:
    - `CreateUser` API (HTTP POST)
    - `Balance` API (HTTP GET)
- **Transactions-service**:
    - `Add Money` API (HTTP POST)
    - `Transfer Money` API (HTTP POST)

## Prerequisites

- **Golang**: Install Golang from the [official website](https://golang.org/dl/).
- **Docker**: Install Docker from the [official website](https://www.docker.com/products/docker-desktop).
- **Docker Compose**: Docker Compose is usually included with Docker Desktop. Verify installation with `docker-compose --version`.
- **Kafka**: Kafka will be set up using Docker Compose.
- **NATS**: NATS will be set up using Docker Compose.
- **PostgreSQL**: PostgreSQL will be set up using Docker Compose.

## Setup

### Step 1: Clone the Repository

```bash
git clone <repository_url>
cd <repository_directory>
```

### Step 2: Docker Setup
Ensure Docker and Docker Compose are installed on your system and Docker daemon is running.

Use the following commands to start the services:
```bash
docker-compose up -d
```
This will start the following containers:

- **PostgreSQL**
- **Kafka**
- **Zookeeper**
- **NATS Server**
- **Users-service**
- **Transactions-service**


## API Endpoints

### User-service

#### 1. Create User
This endpoint creates a new user in the `users` table and pushes a "user-created" event into Kafka.
- **Endpoint**: `:8080/createUser`
- **Method**: `POST`
- **Request Body**:
  ```json
  {
    "email": "user@example.com"
  }
  ```
- **Response Body**:
  ```json
  {
    "user_id": "uuid",
    "email": "user@example.com",
    "created_at": "timestamp"
  }
  ```

#### 2. Get Balance
This endpoint fetches the user's latest balance by making a service-to-service call using NATS to the `transactions-service`.
- **Endpoint**: `:8080/getBalance`
- **Method**: `GET`
- **Request Body**:
  ```json
  {
    "email": "user@example.com"
  }
  ```
- **Response Body**:
  ```json
  {
    "email": "user@example.com",
    "balance": 100.0
  }
  ```


### Transactions-service

#### 1. Add Money
This endpoint credits money to a user's account and updates the balance in the `user` table. It also records the transaction.
- **Endpoint**: `:8081/addMoney`
- **Method**: `POST`
- **Request Body**:
  ```json
  {
    "user_id": "uuid",
    "amount": 100.0
  }
  ```
- **Response Body**:
  ```json
  {
    "updated_balance": 200.0
  }
  ```


#### 2. Transfer Money
This endpoint transfers funds from one user to another and records the transactions for both users.
- **Endpoint**: `:8081/transferMoney`
- **Method**: `POST`
- **Request Body**:
  ```json
  {
    "from_user_id": "uuid",
    "to_user_id": "uuid",
    "amount": 50.0
  }
  ```

## Concurrency and Idempotency

The system ensures concurrency safety and idempotency by using a `sync.Map` to manage user-specific locks. This prevents race conditions during balance updates and ensures the integrity of financial transactions.

## Database Schema

### User-service `users` table:
- `user_id`: UUID (Primary Key)
- `email`: STRING (Unique)
- `created_at`: TIMESTAMP

### Transactions-service `users` table:
- `user_id`: UUID (Primary Key)
- `balance`: NUMERIC
- `created_at`: TIMESTAMP

### Transactions-service `transactions` table:
- `id`: UUID (Primary Key)
- `user_id`: UUID (Foreign Key)
- `balance_change`: NUMERIC
- `old_balance`: NUMERIC
- `new_balance`: NUMERIC
- `timestamp`: TIMESTAMP
