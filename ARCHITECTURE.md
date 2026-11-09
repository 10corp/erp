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

## 2. The Hybrid Architectural Synthesis: Hexya + ERPNext + Tryton

Our system is engineered by synthesizing the best elements of the top 3 open-source ERP systems, eliminating their individual flaws:

```
+-----------------------------------------------------------------------------------------+
|                                    OUR HYBRID ENGINE                                    |
+-----------------------------------------------------------------------------------------+
| 1. Hexya Engine Layer:                                                                  |
|    - High-throughput Go routing and sub-millisecond execution.                          |
|    - Bitmask RBAC Security (Read, Write, Unlink, Post).                                 |
|    - Goroutine Background Worker loops.                                                 |
|                                                                                         |
| 2. ERPNext Schema Layer (DocType):                                                      |
|    - Single-table DocType schema (replaces Odoo's convoluted product.template/product). |
|    - Declarative JSON-based field definitions mapped directly to typed Go structs.      |
|    - Clean foreign-key Link fields & dynamic JSONB custom attributes.                   |
|                                                                                         |
| 3. Tryton Financial Engine Layer:                                                       |
|    - Strict GAAP/IFRS State Machine: Draft (0) -> Submitted (1) -> Cancelled (2).       |
|    - Absolute Immutability: Posted GL entries can NEVER be updated or deleted.          |
|    - Reversal Contra Postings only for corrections.                                     |
|    - Zero-Float Policy: Pure shopspring/decimal & NUMERIC(19, 6).                       |
|    - Invariant: SUM(Debit Base) == SUM(Credit Base) enforced at DB & Go levels.         |
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

## 4. Repository Layout

```
erp/
├── backend/                  # Go backend micro-core
│   ├── cmd/
│   │   ├── server/           # HTTP & SSE API server
│   │   └── worker/           # Background job worker
│   ├── internal/
│   │   ├── accounting/       # Double-entry posting & ledger logic
│   │   ├── doctype/          # DocType schema registry & validation
│   │   ├── queue/            # SKIP LOCKED background queue
│   │   ├── sse/              # SSE event broadcaster
│   │   └── ai/               # AI integrations (Gemini & Claude)
│   ├── migrations/           # Database migration scripts
│   └── go.mod
├── frontend/                 # React SPA
│   ├── src/
│   │   ├── components/       # Drawer, CommandK, AG Grid wrappers
│   │   ├── modules/          # Accounting, Stock, Buying, Selling
│   │   └── hooks/            # SSE and API hooks
│   └── package.json
├── docker-compose.yml        # PostgreSQL container setup
└── Makefile                  # Developer workflow automation
```
