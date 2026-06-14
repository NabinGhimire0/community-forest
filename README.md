# Sri Panchakanya Community Forest Management System

A React + Go + PostgreSQL application for digitizing the member register and managing the operational and financial records of **Community-Forest**.

## Project structure

```text
admin/                 React, Vite and Tailwind frontend
forest-management/     Go, Gin, GORM and PostgreSQL backend
docker-compose.yml     Local PostgreSQL service
```

## Implemented workflows

### Role-aware application and profiles

- Admin: organization settings, fiscal years, stock/rates, member register, approvals, cash collection, reports, committee accounts and audit operations.
- Staff: operational/member data entry and historical-balance drafts; no cash entry or organization configuration.
- Member: personal dashboard, complete household profile, family-member details, own requests, yearly Gasti fees, payments and verified ledger.
- The application name, logo, address and committee data are loaded from the organization settings API.
- Protected requests re-check the account's current database role/status, so a disabled account loses access immediately.

### Annual Gasti / membership fee

- Every fiscal year has one configured Gasti/membership fee.
- Activating a fiscal year automatically creates one verified `membership_fee` ledger entry for every eligible active member.
- A database-level partial unique index and idempotent service logic prevent the same member from receiving the fee twice for the same fiscal year.
- A member created during an active fiscal year receives that year's fee automatically when a fee setting exists.
- Changing the fee for the active year updates only completely unpaid, system-generated entries; paid or partially paid financial history is preserved.
- The member dashboard shows each fiscal year's total, paid and remaining fee separately.
- Members can pay each unpaid yearly fee through eSewa UAT. Past verified Gasti balances can use the same eSewa flow.

### Historical physical-register digitization

Inside a member's Fee Details and Sales Details, admin/staff can record:

- past fiscal year;
- past Gasti fee or timber/firewood/other-sale balance;
- physical register/receipt reference;
- scanned image or PDF evidence;
- draft/verified status.

Staff entries remain drafts. Admin can verify, reverse an unpaid incorrect entry, record a full or partial cash settlement, and download a receipt. Payments update the exact historical ledger entry atomically.

### Fiscal-year stock rollover

Activating a fiscal year:

- deactivates the previous active year;
- carries forward available stock (`remaining - reserved`) into missing target-year stock rows;
- carries forward resource rates and the membership fee into missing target-year rows;
- assigns the new year's Gasti fee once to eligible members;
- never overwrites values already entered for the new fiscal year;
- can safely be retried without duplicate stock, rate, fee-setting or member-fee rows.

Approved resource requests reserve stock. When the request becomes fully paid, its reserved quantity is deducted from remaining stock. Existing approved requests are backfilled into reservation totals during startup.

### Active fiscal-year defaults and member request restriction

- Every fiscal-year selector now defaults to the currently active fiscal year after the fiscal-year list loads.
- This applies to request, expense, fine, payment, transaction, report, export, historical-register, rate and stock screens.
- Administrators and staff can still deliberately change a selector when reviewing or entering historical records.
- A member's resource-request form shows only the active fiscal year and locks that selection.
- The backend independently verifies the active fiscal year, so changing the browser request payload cannot create a member request for a previous or future fiscal year.
- When no active fiscal year exists, the member cannot open a usable request form and receives a clear validation message.

### Payment rules

- Members can pay only their own approved resource requests and verified unpaid Gasti-fee ledger entries through eSewa ePay v2.
- eSewa payments are accepted only after callback-signature validation and an independent status-API check.
- eSewa payment covers the full outstanding amount of the selected request or yearly fee.
- Only an administrator can enter cash payments.
- Cash supports partial payment for approved requests and verified historical/annual fee balances.
- Admin/staff cannot manually mark an online payment as successful.
- Every paid payment has its own downloadable receipt.

### Committee login accounts and admin handover

- While adding a committee head, admin can optionally create a system login.
- Supported operational roles are `admin` and `staff`.
- Phone and password are required when creating credentials.
- If the phone already belongs to a registered member, the same user account is promoted to the selected admin/staff role and receives the new password instead of creating a duplicate login.
- Committee activation/deactivation also activates/deactivates its linked login.
- The environment-seeded administrator is treated as a one-time bootstrap account.
- When the first real committee administrator is created, all active bootstrap administrators are deactivated. They are retained in the database for audit history rather than hard-deleted.
- Re-running the seed command does not reactivate the bootstrap account while a real active administrator exists.

## Requirements

- Node.js 20+
- Go 1.25 or newer (matches `forest-management/go.mod`; use a currently patched release)
- PostgreSQL 14+, or Docker Desktop

## Run locally

### 1. Start PostgreSQL

From the project root:

```bash
docker compose up -d postgres
```

### 2. Configure and run the backend

```bash
cd forest-management
cp .env.example .env
```

PowerShell:

```powershell
Copy-Item .env.example .env
```

Set at least `DB_PASSWORD`, `DATA_ENCRYPTION_KEY`, `SEED_ADMIN_PHONE`, and `SEED_ADMIN_PASSWORD`. The encryption key must be a base64-encoded 32-byte value. For example:

```bash
openssl rand -base64 32
```

```bash
go run ./cmd/seed
go run ./cmd/api
```

Health check:

```text
http://localhost:8080/api/health
```

For development, `RUN_AUTO_MIGRATE=true` can apply the schema automatically. For production, keep it disabled and run the controlled migration command before starting the API:

```bash
go run ./cmd/migrate
```

### 3. Run the frontend

```bash
cd admin
cp .env.example .env
npm ci
npm run dev
```

Open:

```text
http://localhost:5173
```

## eSewa UAT testing

For localhost-only frontend/backend, the browser can return to `http://localhost:8080`. If eSewa cannot reach your callback, expose the backend using a public HTTPS URL and set:

```env
PUBLIC_BACKEND_URL=https://your-backend-public-url
FRONTEND_URL=https://your-frontend-url
CORS_ORIGINS=https://your-frontend-url
```

The included environment example uses the eSewa UAT form/status endpoints and `EPAYTEST`.

UAT login values:

```text
eSewa ID: 9711111111, 9711111112, 9711111113, or 9711111114
Password: Nepal@123
OTP token: 123456
MPIN for application testing: 1122
```

Never reuse test credentials or the UAT secret in production. The backend verifies the signed success response and also checks the status API before settling stock or the selected ledger. Reaching the failure URL alone does not mark a payment failed; the backend reconciles its status first.

## Validation commands

Frontend:

```bash
cd admin
npm ci
npm run lint
npm run build
```

Backend:

```bash
cd forest-management
go test ./...
```

The packaging environment used for this delivery could not download Go 1.25 or Go modules, so run the backend command on a connected machine with Go 1.25 before deployment.
