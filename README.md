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


## ER Diagram

```mermaid
erDiagram
    HOSPITAL ||--o{ STAFF : has
    HOSPITAL ||--o{ PATIENT : has
    STAFF ||--o| HOSPITAL : belongs_to
    PATIENT ||--o| HOSPITAL : belongs_to

    HOSPITAL {
        uint ID PK
        string Name "unique"
    }

    STAFF {
        uint ID PK
        string Username "unique"
        string Password
        uint HospitalID FK
    }

    PATIENT {
        uint ID PK
        string FirstNameTH
        string MiddleNameTH
        string LastNameTH
        string FirstNameEN
        string MiddleNameEN
        string LastNameEN
        string DateOfBirth
        string NationalID
        string PassportID
        string PhoneNumber
        string Email
        string Gender
        uint HospitalID FK
    }

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