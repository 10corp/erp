# Ultra-Fast Enterprise Headless ERP

> **Next-Generation, Bullet-Speed, Real-Time Enterprise Resource Planning**  
> Built with **Go (Golang)**, **PostgreSQL**, and **React.js**.

---

## ⚡ Highlights

- **Rocket-Speed API:** Pure headless backend written in Native Go 1.22+ with zero template/rendering overhead and sub-millisecond route dispatching.
- **Zero-Bug Financial Engine:** Zero-float policy using `NUMERIC(19, 6)` in PostgreSQL and `shopspring/decimal` in Go. Immutable double-entry general ledger with Tryton-style state validation and multi-currency dual-column balancing.
- **Enterprise Scale Without Bloat:** PostgreSQL range partitioning for billions of rows; lock-free background job queue using `FOR UPDATE SKIP LOCKED` (no Redis/RabbitMQ operational overhead required).
- **5-Star Dense UI:** High-density Ant Design + AG Grid interface, context-preserving slide-over drawers, keyboard-first Command+K palette, and real-time Server-Sent Events (SSE).
- **Native Multimodal AI:** Automated invoice/receipt document extraction powered by Gemini 1.5 Flash, and conversational action execution powered by Claude 3.5 Sonnet.

---

## 🏗 System Architecture

For the complete architectural blueprint, see [ARCHITECTURE.md](./ARCHITECTURE.md).

```
[ React SPA (AntD + AG Grid + cmdk) ]
                 |
                 | REST / JSON-RPC + SSE
                 v
[ Go API Server (Goroutines + shopspring/decimal) ]
                 |
                 | Native pgx Binary Protocol
                 v
[ PostgreSQL (Partitioned Tables + Materialized Views + SKIP LOCKED Queue) ]
```

---

## 📁 Repository Structure

```
.
├── backend/            # Go API Server, Financial Engine, Job Queue & SSE
├── frontend/           # React Single-Page Application (AntD, AG Grid, Vite)
├── docs/               # Architecture documents and API specifications
├── docker-compose.yml  # Local PostgreSQL 16 development environment
└── Makefile            # Developer commands for build, test, and run
```

---

## 🚀 Getting Started

Instructions for running the backend, database migrations, and frontend will be documented as each module is initialized.
