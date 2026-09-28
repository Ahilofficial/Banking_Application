# Production-Grade Banking Microservices in Go & Fiber v3

A distributed, production-grade Banking Application built with **Go** and **Fiber v3**, implementing high-throughput REST microservices, fine-grained Role-Based Access Control (RBAC), two-level authorization, ACID money transfers with row-level locking, double-entry bookkeeping ledgers, distributed request tracing, and idempotency guarantees.

---

## 🏛️ System Architecture

```
                       ┌──────────────────────────────┐
                       │     Client Applications      │
                       │ Web / Mobile / Admin / ATM   │
                       └──────────────┬───────────────┘
                                      │
                                      ▼
               ┌──────────────────────────────────────────────┐
               │         Fiber v3 API Gateway (8000)          │
               │                                              │
               │  ├── Request ID Tracing (X-Request-ID)       │
               │  ├── Rate Limiter (Sliding Window)           │
               │  ├── CORS & Panic Recovery                   │
               │  ├── Authentication & Session Validation     │
               │  └── Downstream Reverse Proxy / Dispatcher   │
               └──────────────┬───────────────────────────────┘
                              │
       ┌──────────────┬───────┴──────┬───────────────┬────────────────┐
       ▼              ▼              ▼               ▼                ▼
┌─────────────┐┌─────────────┐┌─────────────┐┌───────────────┐┌───────────────┐
│Auth Service ││Customer Svc ││Account Svc  ││Transaction Svc││ Audit Service │
│ (Port 8001) ││ (Port 8002) ││ (Port 8003) ││  (Port 8004)  ││  (Port 8005)  │
└──────┬──────┘└──────┬──────┘└──────┬──────┘└───────┬───────┘└───────┬───────┘
       │              │              │               │                │
       └──────────────┴──────────────┼───────────────┴────────────────┘
                                     ▼
                      ┌─────────────────────────────┐
                      │    Database Layer (GORM)    │
                      │   MySQL 8.0 (InnoDB Engine) │
                      └─────────────────────────────┘
```

---

## 🚀 Key Architectural Principles

### 1. Two-Level Authorization
Banking security enforces two distinct defense layers:
- **Level 1 — RBAC (Role-Based Access Control)**:
  - Validates whether the caller possesses the coarse permission to perform the operation (e.g. `account:view`, `transaction:create`).
  - Roles (`SUPER_ADMIN`, `BANK_ADMIN`, `BRANCH_MANAGER`, `BANK_EMPLOYEE`, `TELLER`, `CUSTOMER`) map dynamically to granular permissions stored in the database.
- **Level 2 — Resource / Business Authorization**:
  - Validates whether the caller is authorized to operate on the **specific resource**.
  - Standard customers can **only view/operate their own customer profiles and accounts**.
  - A customer possessing `account:view` cannot query or debit another customer's account.

### 2. ACID Money Transfer & Deadlock Prevention
- All money transfers execute inside a database transaction:
  ```go
  BEGIN TRANSACTION;
  -- Deadlock prevention: sort account IDs in ascending order
  SELECT * FROM accounts WHERE id = min(from_id, to_id) FOR UPDATE;
  SELECT * FROM accounts WHERE id = max(from_id, to_id) FOR UPDATE;
  -- Balance & status checks
  -- Debit source & Credit destination
  -- Create transaction record
  -- Create double-entry ledger entries (DEBIT & CREDIT)
  COMMIT;
  ```

### 3. Double-Entry Bookkeeping Ledger
- Account balances are supported by an immutable double-entry ledger (`ledger_entries`):
  - Transfer ₹10,000:
    - **DEBIT**: Account A for ₹10,000 (`BalanceAfter: ₹40,000`)
    - **CREDIT**: Account B for ₹10,000 (`BalanceAfter: ₹20,000`)
  - Guaranteed Zero-Sum: $\sum \text{Credits} = \sum \text{Debits}$.

### 4. Zero Floating-Point Precision Error
- Currency is strictly handled using **integer minor units** (`int64`, e.g., Paise or Cents):
  - ₹100.50 is stored as `10050` paise.
  - Arithmetic avoids IEEE 754 floating-point rounding errors.

### 5. Idempotency Guarantees
- Financial mutations accept an `Idempotency-Key` HTTP header.
- Cached responses ensure network retries never double-charge or duplicate transfers.

### 6. Distributed Audit Trail
- High-priority operations emit immutable audit records to the `audit-service`:
  - Tracks `user_id`, `action`, `resource_type`, `resource_id`, `old_value`, `new_value`, `ip_address`, and `request_id`.

---

## 📁 Repository Structure

```
banking-microservices/
├── services/
│   ├── gateway/                  # Fiber v3 API Gateway (Port 8000)
│   ├── auth-service/             # RBAC, JWT, Sessions (Port 8001)
│   ├── customer-service/         # Customer Profiles & KYC (Port 8002)
│   ├── account-service/          # Bank Accounts & Balances (Port 8003)
│   ├── transaction-service/      # Transfers, Ledger, Row Locking (Port 8004)
│   └── audit-service/            # Immutable Audit Trail (Port 8005)
│
├── pkg/
│   ├── client/                   # Inter-service HTTP clients
│   └── common/
│       ├── config/               # Environment & service config
│       ├── database/             # GORM MySQL & Pure-Go SQLite connector
│       ├── errors/               # Centralized Fiber v3 error handler
│       ├── middleware/           # Auth, RBAC, RequestID, RateLimiter
│       ├── money/                # Minor-unit integer money math
│       ├── response/             # Standard API envelope
│       └── security/             # JWT Claims & Password Hashing
│
├── scripts/
│   ├── seed.go                   # Database seeder (Actors, Roles, Accounts)
│   └── run_all.go                # Concurrent microservices launcher
│
├── tests/
│   └── integration_test.go       # Comprehensive E2E integration test suite
│
├── docker-compose.yml            # Multi-container orchestration (MySQL, Redis, Services)
├── Dockerfile                    # Multi-stage container build
└── Makefile                      # Build, test, and run automation
```

---

## 👥 Banking Actor & Permission Matrix

| Role | Permissions |
| :--- | :--- |
| **SUPER_ADMIN** | All system permissions (`*`), `user:manage`, `role:manage`, `audit:view` |
| **BANK_ADMIN** | `customer:*`, `account:*`, `transaction:view`, `transaction:approve`, `ledger:view`, `audit:view` |
| **BRANCH_MANAGER** | `customer:view`, `customer:update`, `account:*`, `transaction:view`, `transaction:approve`, `ledger:view` |
| **BANK_EMPLOYEE** | `customer:create`, `customer:view`, `account:create`, `account:view`, `transaction:view` |
| **TELLER** | `customer:view`, `account:view`, `transaction:create`, `transaction:view` |
| **CUSTOMER** | `customer:view` (self), `account:view` (self), `transaction:create` (own accounts), `transaction:view` |

---

## ⚡ Quick Start

### 1. Run with Local MySQL
Make sure your MySQL server is running (defaults to `127.0.0.1:3306`, credentials can be configured in `.env`).

```bash
# 1. Seed initial roles, permissions, actors, and test accounts on MySQL
go run scripts/seed.go

# 2. Start all microservices + API Gateway concurrently
go run scripts/run_all.go
```

The database `banking` and all necessary tables, constraints, roles, and initial accounts will be created automatically.

### 3. Import Postman Collection
The project includes a ready-to-import Postman Collection v2.1: [`postman.json`](file:///c:/Users/ahilc/Downloads/Ready%20Assist/Banking%20Application/postman.json).
1. Open Postman -> Click **Import** -> Select `postman.json`.
2. Run **1.1 Login - Alice** or **1.3 Login - Admin**: tokens are **automatically saved** into collection variables!
3. Execute transfers, verify Level 2 authorization rules, inspect the double-entry ledger, and review audit logs directly.

### 4. Run Full Automated Integration Test Suite
```bash
go test -v ./tests
```

---

## 🧪 Seeded Test Accounts

| Actor | Email | Password | Role | Initial Account | Balance |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Admin** | `admin@bank.com` | `Admin@123` | `SUPER_ADMIN` | - | - |
| **Teller** | `teller@bank.com` | `Teller@123` | `TELLER` | - | - |
| **Alice** | `alice@bank.com` | `Alice@123` | `CUSTOMER` | `AC-MAIN-1000000001` | **₹50,000.00** |
| **Bob** | `bob@bank.com` | `Bob@123` | `CUSTOMER` | `AC-MAIN-1000000002` | **₹10,000.00** |

---

## 📡 API Reference & Examples

### 1. Authenticate (Login)
```bash
curl -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "alice@bank.com", "password": "Alice@123"}'
```

### 2. View Accounts (Level 2 Authorization)
```bash
curl -X GET http://localhost:8000/api/v1/accounts \
  -H "Authorization: Bearer <ALICE_ACCESS_TOKEN>"
```

### 3. Money Transfer with Idempotency Key
```bash
curl -X POST http://localhost:8000/api/v1/transactions/transfer \
  -H "Authorization: Bearer <ALICE_ACCESS_TOKEN>" \
  -H "Idempotency-Key: a1b2c3d4-e5f6-7890" \
  -H "Content-Type: application/json" \
  -d '{
    "from_account_id": 1,
    "to_account_id": 2,
    "amount": 10000.00,
    "currency": "INR",
    "description": "Monthly rent transfer"
  }'
```

### 4. Inspect Double-Entry Ledger (Admin / Auditor)
```bash
curl -X GET http://localhost:8000/api/v1/transactions/ledger \
  -H "Authorization: Bearer <ADMIN_ACCESS_TOKEN>"
```

### 5. Inspect Audit Trail (Admin)
```bash
curl -X GET http://localhost:8000/api/v1/audit/logs \
  -H "Authorization: Bearer <ADMIN_ACCESS_TOKEN>"
```

---

## 🛡️ Error Handling Specification

All error responses strictly follow a uniform banking error format:

```json
{
  "success": false,
  "error": {
    "code": "INSUFFICIENT_FUNDS",
    "message": "Insufficient balance. Available: 40000.00, Requested: 100000.00"
  },
  "request_id": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d"
}
```
