# Hospital Middleware System

A Go-based middleware service for hospital staff to search patient information across multiple hospital systems with proper authentication and authorization.

## Features

- **Staff Management**:
  - Create staff accounts with hospital association
  - JWT-based authentication
- **Patient Search**:
  - Search across local database and external hospital APIs
  - Hospital-restricted access to patient data
- **Security**:
  - Password hashing
  - Role-based access control
- **Infrastructure**:
  - Dockerized PostgreSQL database
  - Nginx reverse proxy
  - Ready for production deployment

## Tech Stack

- **Backend**: Go 1.24.2
- **Framework**: Gin
- **Database**: PostgreSQL
- **Authentication**: JWT
- **Containerization**: Docker
- **Web Server**: Nginx

## API Endpoints

### Staff APIs
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/staff/create` | POST | Create new staff account |
| `/staff/login` | POST | Staff login (returns JWT token) |

### Patient APIs (Requires Authentication)
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/patient/search` | POST | Search patients (local + external systems) |

## Database Schema

Entities:
---------
[HOSPITAL]
- ID (PK)
- Name (unique)

[STAFF]
- ID (PK)
- Username (unique)
- Password
- HospitalID (FK → HOSPITAL)

[PATIENT]
- ID (PK)
- FirstNameTH
- MiddleNameTH
- LastNameTH
- FirstNameEN
- MiddleNameEN
- LastNameEN
- DateOfBirth
- NationalID
- PassportID
- PhoneNumber
- Email
- Gender
- HospitalID (FK → HOSPITAL)

Relationships:
--------------
1. HOSPITAL (1) → (N) STAFF
2. HOSPITAL (1) → (N) PATIENT

## Setup Instructions

### Prerequisites
- Docker
- Docker Compose

### Installation
1. Clone the repository:
   ```bash
   git clone [https://github.com/sainisahil1/agnos-hospital-middleware.git]
   cd hospital-middleware

2. Start the services:
    ```bash
    docker-compose up -d

3. The application will be available at:

Backend: http://localhost:8080

Through Nginx: http://localhost