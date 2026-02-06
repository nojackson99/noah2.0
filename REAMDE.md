# Personal Assistant (Learning Project)

## Goal
Build a cloud-backed personal planning assistant that:
- Ingests my personal data (calendar, notes, etc.)
- Helps me plan my week/day
- Starts simple (read-only, CLI-driven)
- Grows into a more agentic system over time

## Philosophy
- Learning-first, not speed-first
- Prefer explicit implementations over magic
- Use managed infra only where it doesn’t reduce learning

## Tech (current decisions)
- Go backend (Gin)
- Postgres (Railway)
- CLI client (Go)
- React UI later (Vercel)
- RAG-style retrieval with pgvector
- Text-only v1

## Current milestone
Week 1: API skeleton + /health endpoint
Week 2: Google Calendar ingestion
Week 3: Markdown notes ingestion + embeddings
Week 4: Weekly planning endpoint
