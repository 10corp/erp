# Architecture & System Design Specification

## 1. Core Philosophy & Evolution Beyond Legacy / Monolithic ERPs

### 1.1 The Pitfalls of Legacy Monoliths (Odoo, ERPNext)
Traditional enterprise ERPs (Odoo, ERPNext, SAP, NetSuite) suffer from:
1. **Bloated Monoliths:** Heavy ORM layers, dynamic interpreted languages (Python/PHP), and server-side rendering resulting in multi-second page loads.
2. **Weak Ledger Immutability:** Mutable database records that lead to accounting discrepancies and difficult audit trails.
3. **Complex Infrastructure Footprint:** Dependency on multiple external brokers (Redis, RabbitMQ, Celery/Sidekiq, Elasticsearch) just to handle basic queues and caching.

### 1.2 Lessons from Hexya (`hexya-erp/hexya`)
[Hexya](https://github.com/hexya-erp/hexya) proved that Go is exceptionally well-suited for ERP performance. However, Hexya inherited Odoo's core architectural flaw: **trying to replicate Odoo's monolithic, XML-driven, server-side UI paradigm inside Go**. This resulted in:
- Tight coupling between business logic and UI definitions.
- Heavy XML parsing and complex reflection-based ORMs trying to mimic Python's dynamic classes.
- Slower frontend evolution and loss of modern frontend ecosystem capabilities.

### 1.3 The Headless Go + React SPA Advantage
Our architecture breaks completely free from the monolithic approach:
- **True Headless Decoupling:** Go focuses 100% on high-speed data validation, state machines, and API delivery (REST / JSON-RPC + SSE). It has zero UI rendering overhead.
- **Blazing React SPA UX:** Zero full-page reloads. Transitions happen in single-digit milliseconds. Optimistic updates via TanStack Query and virtualized 60 FPS scrolling via AG Grid.
- **Fault Isolation & Independent Team Velocity:** Go backend developers and React frontend developers work with clear API contracts. A backend bug doesn't crash the UI shell, and frontend iterations ship without recompiling the Go binary.
- **Seamless Omnichannel & Mobile (React Native):** Because the API is pure headless JSON, the exact same backend endpoints power the React Web SPA, warehouse barcode scanners, and a React Native mobile application with 90%+ shared business logic and types.
- **Compiled, Ultra-Low-Latency Core:** Go compiles to a single static binary with ultra-low memory footprint (~20-50MB), native concurrent goroutines, and sub-millisecond response times.
- **Strict Double-Entry Immutability:** Inspired by Tryton, general ledger entries are **append-only**. Cancellations and corrections generate explicit reversal contra entries.
- **Zero-Float Accounting:** Pure fixed-point arithmetic (`NUMERIC(19, 6)` in PostgreSQL and `shopspring/decimal` in Go).
- **Single Source of Truth Storage:** Leveraging PostgreSQL 16+ native features—table partitioning, `FOR UPDATE SKIP LOCKED` for task queues, and materialized views—eliminating the need for separate Redis/RabbitMQ infrastructure during core operations.
- **Dense, Keyboard-First UX:** React SPA with Ant Design density presets, AG Grid for handling 100,000+ data rows with 60 FPS scrolling, context drawers instead of full-page navigation, and global `Command+K` shortcuts.

---

## 2. The Hybrid Architectural Synthesis: Hexya + Tryton + iDempiere + ERPNext

Our system is engineered by synthesizing the best elements of the world's top open-source ERP systems, eliminating their individual flaws:

```
+-----------------------------------------------------------------------------------------+
|                                    OUR HYBRID ENGINE                                    |
+-----------------------------------------------------------------------------------------+
| 1. Hexya Engine Layer:                                                                  |
|    - High-throughput Go routing and sub-millisecond execution.                          |
|    - Bitmask RBAC Security (Read, Write, Unlink, Post).                                 |
|    - Goroutine Background Worker loops & concurrent SSE dispatch.                       |
|                                                                                         |
| 2. Tryton Financial & Double-Entry Stock Foundation:                                    |
|    - Immutable append-only general ledgers (account_account, account_move, account_move_line)|
|    - Double-Entry Stock Movements (stock_location, stock_move):                         |
|      Moves inventory symmetrically: Supplier Location -> Warehouse -> Customer Location |
|    - Universal Legal Entity: party.party & multi-tax identifiers (BIN, TIN, VAT).       |
|    - Recursive Multi-Branch Tree Hierarchy (company.company / vw_company_hierarchy).    |
|    - Zero-Float Policy: Pure shopspring/decimal & NUMERIC(19, 6).                       |
|                                                                                         |
| 3. iDempiere Multi-Tenant & Multi-Branch Enterprise Isolation:                          |
|    - Dual-column multi-tenancy on every transactional table:                            |
|      * tenant_id (iDempiere AD_Client_ID - Complete SaaS / Group Client Isolation)       |
|      * company_id (iDempiere AD_Org_ID - Multi-Branch / Subsidiary Operating Isolation) |
|                                                                                         |
| 4. ERPNext Document Lifecycle (docstatus) & Business Subsystems:                        |
|    - Lightweight Document Immutability Engine:                                          |
|      * docstatus = 0 (Draft - fully editable, zero ledger impact)                       |
|      * docstatus = 1 (Submitted/Posted - immutable, triggers GL and Stock movements)     |
|      * docstatus = 2 (Cancelled - never deleted, appends explicit contra reversals)     |
|    - Clean declarative field design for HR, Leave Applications & Salary Slip Payroll    |
|    - Clean Manufacturing Bill of Materials (BOM) & Work Orders                          |
+-----------------------------------------------------------------------------------------+
```

---

## 3. Component Stack

### 2.1 Backend (Go API Server)
- **Language:** Go 1.22+
- **HTTP / RPC Framework:** Standard `net/http` with `chi` router or high-performance lightweight equivalent.
- **Database Driver:** `jackc/pgx/v5` utilizing native PostgreSQL binary protocol and connection pooling.
- **Math & Decimals:** `github.com/shopspring/decimal` for all monetary computations.
- **Real-Time Dispatch:** Server-Sent Events (SSE) broadcaster via native Go channels.
- **Job Processing:** Dedicated Go background worker pool utilizing PostgreSQL `FOR UPDATE SKIP LOCKED`.

### 2.2 Database Layer (PostgreSQL 16+)
- **Ledger Partitioning:** `gl_entry` and `stock_ledger_entry` partitioned by `RANGE (posting_date)` (Monthly or Quarterly).
- **Dual-Currency Balancing:** Columns `(debit, credit)` in transaction currency, and `(debit_base, credit_base)` in functional/base currency.
- **CQRS Read Models:** `MATERIALIZED VIEW` for C-Level Dashboards, Profit & Loss, and Balance Sheets refreshed concurrently in the background.

### 2.3 Frontend (React SPA)
- **Build Tool:** Vite + TypeScript
- **UI Library:** Ant Design (`ant-design`) with compact/dense theme tokens.
- **Data Grids:** AG Grid (`ag-grid-react`) for high-volume tabular ledgers and journals.
- **Global Search:** `cmdk` command palette for Mac Spotlight-style navigation.
- **Data Fetching:** `@tanstack/react-query` for optimistic updates and smart client caching.

### 2.4 AI Services
- **Gemini 1.5 Flash:** Multimodal Intelligent Document Processing (IDP) for auto-extracting line items and metadata from supplier invoices and receipts.
- **Claude 3.5 Sonnet:** Action Brain converting natural language business queries into validated JSON execution payloads.

---

## 4. Financial Engine Rules (Zero-Bug & Audit-Proof)

1. **Balance Invariant:** Every posted transaction must satisfy:
   $$\sum \text{Debit Base} == \sum \text{Credit Base} \quad (\text{Diff} = 0.000000)$$
2. **Pro-Tip Round Off Guard (`account_code: 5090`):**
   - Microscopic multi-currency rounding fractions ($\le 0.05$ Base Currency) automatically post to dedicated `5090 - Round Off Account` as an independent credit or debit line.
   - Taxes and revenue lines remain pristine and untouched for auditing.
3. **Automated Forex Realized Gain/Loss Engine:**
   - Settling multi-currency invoices generates automatic balancing entries:
     $$\Delta = (\text{Paid Amount} \times \text{Payment Date Rate}) - (\text{Allocated Amount} \times \text{Invoice Date Rate})$$
   - If $\Delta > 0$: Credit `4200 - Forex Realized Gain`
   - If $\Delta < 0$: Debit `5200 - Forex Realized Loss`
4. **Immutable Entries & Reversal Contra Postings:**
   - Database triggers strictly reject `UPDATE` or `DELETE` statements on `gl_entries`.
   - Cancellations produce opposite-signed contra reversal rows preserving full audit history.

---

## 5. Double-Entry Inventory & Perpetual COGS Engine

1. **Tryton Double-Entry Locations:**
   - No stock appears or vanishes without explicit origin and destination:
     - Inflow (Purchase Receipt): `Supplier (Virtual)` $\to$ `Internal Warehouse`
     - Outflow (Delivery Note): `Internal Warehouse` $\to$ `Customer (Virtual)`
     - Transfer: `Source Warehouse` $\to$ `Destination Warehouse`
2. **Moving Average Costing ($O(1)$ Time Complexity):**
   - On Inflow:
     $$\text{New Valuation Rate} = \frac{(\text{Current Qty} \times \text{Current Rate}) + (\text{Incoming Qty} \times \text{Incoming Rate})}{\text{Current Qty} + \text{Incoming Qty}}$$
   - On Outflow:
     $$\text{COGS Amount} = \text{Delivered Qty} \times \text{Valuation Rate}$$
3. **Perpetual Inventory Auto-Posting to `gl_entries`:**
   - Every stock dispatch atomically commits:
     - **Debit:** `5010 - Cost of Goods Sold (COGS)` (Amount: COGS)
     - **Credit:** `1030 - Stock In Hand (Inventory Asset)` (Amount: COGS)
4. **Row-Level Locking Concurrency Guard:**
   - Delivery validation locks balance with `SELECT ... FOR UPDATE` on `item_warehouse_balances`.
   - Strict `ErrInsufficientStock` returned if requested quantity exceeds available stock, completely eliminating negative stock anomalies.

---

## 6. Multimodal Intelligent Document Processing (IDP) & SSE

1. **Async Job Queue Ingestion (`FOR UPDATE SKIP LOCKED`):**
   - Invoices or receipts uploaded via `POST /api/v1/idp/upload` are enqueued into `background_jobs` table.
   - Go worker pool atomically locks tasks and prevents duplicate processing across instances.
2. **Real-time SSE Streaming (`/api/v1/idp/stream`):**
   - The Go server pushes live event steps to the React client:
     $$\text{analyzing\_document} \longrightarrow \text{extracting\_line\_items} \longrightarrow \text{validating\_math} \longrightarrow \text{completed}$$
3. **Zero-Float Gemini Structured Extraction & Math Guard:**
   - All amounts extracted strictly as **Strings** (`ResponseMIMEType: "application/json"`).
   - Go executes mathematical verification:
     $$\sum (\text{item.qty} \times \text{item.unit\_price}) + \text{tax\_total} == \text{grand\_total}$$
   - Any anomaly sets `math_mismatch_warning: true` with calculated variance.
4. **Human-in-the-Loop Verification:**
   - Documents are **never** posted directly to the general ledger. They pre-fill into the context Slide Drawer as **Draft (0)**, allowing accountant inspection before one-click atomic posting.

---

## 7. Claude 3.5 Sonnet Natural Language Action Brain (Command+K Engine)

1. **Natural Language to Action Translation:**
   - Receives business instructions typed into the global Command+K spotlight or spoken via voice.
   - Converts unstructured language into typed `ERPAction` JSON payloads.
2. **Action Taxonomies:**
   - `open_drawer`: Opens context Slide Drawer for `PaymentEntry`, `SalesInvoice`, or `DeliveryNote` with pre-filled entity parameters.
   - `navigate`: Switches active view in 0ms (`dashboard`, `ledger`, `invoices`).
   - `query_doctype`: Queries overdue invoices or unposted vouchers dynamically.
3. **Example Workflow:**
   - User Input: *"Show me overdue invoices for Acme and draft receipt"*
   - AI Output:
     ```json
     {
       "type": "open_drawer",
       "target": "PaymentEntry",
       "parameters": {
         "party_name": "Acme",
         "payment_type": "Receive",
         "filter_overdue_invoices": true,
         "auto_reconcile": true
       },
       "ui_message": "Identified intent: Customer Receipt for Acme. Opened Payment Drawer with auto-reconciliation enabled.",
       "confidence": 0.98
     }
     ```

---

## 8. Repository Layout

```
erp/
├── backend/                  # Go backend micro-core
│   ├── cmd/
│   │   ├── server/           # HTTP & SSE API server
│   │   └── seed/             # Database migrator & seed automation
│   ├── internal/
│   │   ├── accounting/       # Double-entry posting, Sub-ledger guard & Trial Balance
│   │   ├── doctype/          # DocType registry (Sales, Purchase, Payments, Salary Slip, BOM)
│   │   ├── hr/               # Human Resources, Leaves, Salary Structures & Payroll
│   │   ├── inventory/        # Moving average ($O(1)$) costing & stock ledgers
│   │   ├── module/           # Pluggable ERPNext-style module registry
│   │   ├── party/            # Tryton universal party (party.party, identifiers, branches)
│   │   ├── security/         # Core user management, bitmask RBAC, lockout & JWT
│   │   ├── queue/            # SKIP LOCKED background worker pool
│   │   ├── sse/              # Server-Sent Events (SSE) broadcaster
│   │   └── ai/               # Gemini 1.5 Flash IDP & Claude 3.5 Sonnet Action Brain
│   ├── migrations/           # PostgreSQL migrations (000001 through 000010)
│   └── go.mod
├── frontend/                 # Refine + React 18 + Ant Design 5
│   ├── src/
│   │   ├── components/       # SlideDrawer (IDP), CommandPalette (Claude 3.5)
│   │   ├── pages/            # Dashboard, GeneralLedger, TrialBalance, Parties, Users, HR Payroll
│   │   ├── providers/        # Refine dataProvider (Go REST adapter) & authProvider
│   │   └── App.tsx           # Refine routing, Active branch switcher, RBAC
│   └── package.json
├── docker-compose.yml        # PostgreSQL container setup
└── Makefile                  # Developer workflow automation
```

---

## 9. Tryton Universal Legal Entities & Multi-Branch Architecture

### 9.1 The Five Pillars of Party & Multi-Branch Tenancy
1. **`party.party` (`party_party`):** Universal legal entity representing Customers, Suppliers, Employees, or Internal Companies under a single unified table.
2. **`party.identifier` (`party_identifier`):** Preserves statutory tax identification codes (BIN, TIN, VAT, NID, Trade License) with `is_primary` validation.
3. **`party.address` & `party.contact_mechanism`:** Multi-address (Invoice vs Delivery) and multi-channel contacts (Email, Phone, WhatsApp) linked via UUID foreign keys.
4. **`company.company` (`companies`):** Self-referential tree structure (`parent_id`) allowing unlimited multi-branch and subsidiary hierarchies.
5. **Multi-Branch Hierarchy CTE (`vw_company_hierarchy`):** Recursive query computing branch lineage, root parentage, depth levels, and material path names.

### 9.2 Sub-Ledger Integrity & Inter-Branch Clearing
- **Sub-Ledger Invariant Trigger (`trg_validate_subledger_party`):** Strictly enforces that postings to Accounts Receivable (`1020`) and Accounts Payable (`2001`) must supply a non-null `party_id`.
- **Inter-Branch Clearing Accounts (`1990` & `2990`):** Transfers between branches credit/debit reciprocal clearing accounts:
  - Branch A transfers to Branch B: Branch A records `Due From Branch B (1990)`, Branch B records `Due To Branch A (2990)`.
  - In Consolidated Trial Balance reporting, clearing accounts net out to `0.000000`, guaranteeing balance sheet integrity.

---

## 10. Refine Headless Dual-Panel Framework

### 10.1 Refine Core Integration (`sources/refine`)
- **Dual Panel Architecture:** Single unified application providing both high-density **Admin/CFO** management tools and streamlined **Operational / User** workflows.
- **Dynamic Branch Switching:** Custom header dropdown switches active operating branch context in 0ms, purging query cache via `useInvalidate()` without page reloads.
- **1-Click Role Switcher:** Login screen enables instant role emulation (Admin/CFO, Sales Officer, Visitor/Auditor) with pre-configured JWT claims and bitmask permissions.
- **60 FPS AG Grid Virtualization:** Handles general ledger views with 100,000+ entries using virtual DOM row rendering.

---

## 11. Human Resources & Payroll Subsystem

### 11.1 ERPNext Fields + Tryton docstatus Financial Engine
- **Salary Slip Lifecycle:**
  - `docstatus = 0 (Draft)`: Unposted; allows adjusting base salary, house rent allowance (HRA), medical, conveyance, provident fund, and tax withholding.
  - `docstatus = 1 (Submitted / Posted)`: Locks the slip and auto-posts balanced general ledger entries:
    - **Dr 5020 (Operating Salary Expense):** Gross Pay
    - **Cr 2010 (Salary Payable to Employee Bank):** Net Take-Home Pay
    - **Cr 2020 (Tax & PF Withheld Liability):** Total Statutory Deductions
  - `docstatus = 2 (Cancelled)`: Reverses the payroll slip without database row deletion by appending explicit contra reversal vouchers.
- **Universal Employee Identity:**
  - All employees (`company_employee`) are strictly linked to `party_party` via `party_id`. This allows an employee to also be a vendor/claimant for expense reimbursements without data duplication.
- **Leave Management:**
  - Standardized leave types (Annual, Casual, Sick) with manager approval lifecycle (`docstatus: 0 -> 1`).
