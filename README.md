# Payslip Generation System

## Introduction

This project is a Fiber-based API for a payslip generation system. It allows administrators to manage attendance periods and run payroll, and employees to submit attendance, overtime, reimbursement requests, and view their payslips. The system features JWT-based authentication, role-based authorization, structured logging, and detailed audit trails.

## Features

*   **Admin Functionalities:**
    *   Secure login for administrators.
    *   Creation of attendance periods.
    *   Payroll processing for specified periods, calculating salaries, overtime, and reimbursements.
    *   Summary view of generated payslips for a period.
*   **Employee Functionalities:**
    *   Secure login for employees.
    *   Submission of daily attendance.
    *   Submission of overtime records.
    *   Submission of reimbursement requests.
    *   Viewing personal payslips for specific periods.
*   **Technical Features:**
    *   JWT-based authentication (Bearer Token).
    *   Role-based authorization (admin, employee).
    *   Structured JSON logging using Zap.
    *   Detailed audit logging for key actions.
    *   Automated API documentation via Swagger/OpenAPI.
    *   Configuration management using environment variables (`.env` files).
    *   PostgreSQL database with GORM as ORM.
    *   Unit and integration testing.

## Technology Stack

*   **Backend:** Go (Golang)
*   **Framework:** Fiber (v2)
*   **Database:** PostgreSQL
*   **ORM:** GORM
*   **Authentication:** JWT (github.com/golang-jwt/jwt/v5)
*   **Logging:** Zap (go.uber.org/zap)
*   **API Documentation:** Swag (github.com/swaggo/swag)
*   **Password Hashing:** bcrypt (golang.org/x/crypto/bcrypt)
*   **Fake Data (Seeding):** go-faker (github.com/go-faker/faker/v4)
*   **Decimal Calculations:** shopspring/decimal (github.com/shopspring/decimal)

## Prerequisites

*   Go: Version 1.21 or higher (due to `log/slog` usage, though Zap is primary here). Check `go.mod` for the exact version.
*   PostgreSQL: Version 12 or higher recommended.
*   Docker and Docker Compose (Optional, for running PostgreSQL in a container).
*   `swag` CLI tool (for regenerating API docs): `go install github.com/swaggo/swag/cmd/swag@latest`

## Getting Started

### 1. Installation

Clone the repository:
```bash
git clone https://github.com/your-repo/payslip-generator.git
cd payslip-generator
```

Install Go dependencies:
```bash
go mod tidy
# or
go mod download
```

### 2. Database Setup

*   Ensure PostgreSQL server is running and accessible.
*   Create a database, for example, `payslip_db` for development and `payslip_test_db` for testing.
    ```sql
    CREATE DATABASE payslip_db;
    CREATE DATABASE payslip_test_db;
    -- Optionally create a dedicated user
    CREATE USER youruser WITH PASSWORD 'yourpassword';
    ALTER DATABASE payslip_db OWNER TO youruser;
    ALTER DATABASE payslip_test_db OWNER TO youruser;
    ```
*   The application automatically runs database migrations on startup using GORM's `AutoMigrate`.

### 3. Configuration

*   Copy the example environment file:
    ```bash
    cp .env.example .env
    ```
*   Edit `.env` and provide your actual database credentials and JWT secret.
*   Key environment variables:
    *   `PORT`: Port for the application server (e.g., 8080).
    *   `APP_ENV`: Application environment (`development`, `production`, or `test`).
    *   `JWT_SECRET`: A strong, random string for signing JWTs.
    *   `DB_HOST`: Database host (e.g., `localhost`).
    *   `DB_USER`: Database username.
    *   `DB_PASSWORD`: Database password.
    *   `DB_NAME`: Database name (e.g., `payslip_db`).
    *   `DB_PORT`: Database port (e.g., `5432`).
    *   `DB_SSLMODE`: `disable`, `require`, etc.
    *   `DB_TIMEZONE`: (e.g., `UTC`).

### 4. Running the Application

```bash
go run cmd/server/main.go
```
The application will start, typically on the port specified in `.env` (default 8080).
You should see log messages indicating database connection and server startup.

### 5. Running Tests

*   Tests run in the `test` environment and require a separate test database.
*   Copy `.env.example` to `.env.test` and configure it for your test database:
    ```bash
    cp .env.example .env.test
    # Edit .env.test with test database details (e.g., DB_NAME=payslip_test_db)
    # Ensure APP_ENV=test is set in .env.test
    ```
*   Run tests using:
    ```bash
    # Ensure APP_ENV is not globally set to something other than test, or tests might pick up wrong .env
    # The tests/main_test.go sets os.Setenv("APP_ENV", "test") internally.
    go test ./...
    ```

## API Usage

### API Documentation

Interactive API documentation (Swagger UI) is available when the application is running:
*   URL: `http://localhost:<PORT>/swagger/index.html` (replace `<PORT>` with the actual port).

### Authentication

*   Most endpoints require JWT authentication.
*   Obtain a token by logging in via `/admin/login` or `/employee/login`.
*   Include the token in the `Authorization` header as a Bearer token:
    `Authorization: Bearer <your_jwt_token>`

### Example API Calls

**Admin Login:**
```bash
curl -X POST "http://localhost:8080/api/v1/admin/login" \
-H "Content-Type: application/json" \
-d '{
    "username": "admin",
    "password": "adminpassword"
}'
# Response will contain a token if successful.
```

**Employee Submit Attendance (requires Bearer token):**
```bash
curl -X POST "http://localhost:8080/api/v1/employee/attendance" \
-H "Content-Type: application/json" \
-H "Authorization: Bearer <employee_jwt_token>"
# No request body needed for this specific endpoint.
```

## Software Architecture

*   **`cmd/server/main.go`**: Entry point of the application, initializes Fiber, database, middleware, and routes.
*   **`pkg/`**: Contains the core application logic.
    *   **`config`**: Configuration loading from environment variables.
    *   **`constants`**: Application-wide constants (e.g., context keys).
    *   **`controllers`**: HTTP handlers for API requests, parsing input, calling services/DB, formatting responses.
    *   **`database`**: Database connection, GORM setup, migration, and seeder.
    *   **`docs`**: Generated Swagger documentation files.
    *   **`middleware`**: Custom Fiber middleware (Request ID, Logger, Auth).
    *   **`models`**: GORM database models (structs representing DB tables).
    *   **`routes`**: API route definitions, grouping related endpoints.
    *   **`services`**: Business logic services (e.g., AuditService).
    *   **`utils`**: Utility functions (password hashing, JWT generation, date calculations, logger instance).
*   **`tests/`**: Integration tests for API endpoints. Unit tests are co-located with the packages they test (e.g., `pkg/utils/password_test.go`).

**Data Flow (Typical Request):**
1.  HTTP Request hits Fiber.
2.  Middleware (RequestID, Logger, DeserializeUser for auth) are processed.
3.  Request is routed to the appropriate controller handler based on path and method.
4.  Controller:
    *   Parses request payload/parameters.
    *   Validates input.
    *   Performs authorization checks (if not already handled by middleware).
    *   Interacts with database (via GORM directly or through services) to fetch/store data.
    *   Calls utility functions or services for specific tasks.
    *   Constructs and sends HTTP response.
5.  AuditService may be called by controllers to log significant actions.
6.  Logger middleware logs the completed request details.

## Database Schema

The database schema is defined by GORM models in `pkg/models/`:
*   `BaseModel`: Common fields (ID, CreatedAt, UpdatedAt, CreatedBy, UpdatedBy, IPAddress).
*   `Admin`: Administrator users.
*   `Employee`: Employee users and their salary.
*   `AttendancePeriod`: Defines payroll periods (start date, end date).
*   `AttendanceRecord`: Records employee check-in times for specific dates.
*   `OvertimeRecord`: Records employee overtime hours.
*   `ReimbursementRequest`: Tracks employee reimbursement claims.
*   `Payslip`: Stores generated payslip details for each employee per period.
*   `AuditLog`: Logs significant actions performed in the system.

Refer to the struct definitions in `pkg/models/` for detailed field information and GORM tags.

## Logging and Auditing

*   **Structured Logging:** Uses `go.uber.org/zap` for structured JSON logs. Logs include request details, errors, and application lifecycle events.
*   **Audit Logging:** The `AuditLog` table (`models.AuditLog`) records critical business actions (e.g., payroll run, attendance submission) providing a trail of who did what, when, and from where.

## Contributing

Contributions are welcome! Please follow these general guidelines:
1.  Fork the repository.
2.  Create a new branch for your feature or bug fix.
3.  Write clear and concise commit messages.
4.  Ensure your code is well-tested (add unit/integration tests).
5.  Update documentation (README, Swagger annotations) if necessary.
6.  Submit a pull request.

## License

This project is licensed under the MIT License. See the `LICENSE` file for details.
(Note: A `LICENSE` file should be created if one doesn't exist.)
