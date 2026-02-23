#!/usr/bin/env python3
"""
Benchmark test suite across all models in ModelsByTier.

Sends a standardized coding/reasoning prompt to each model (OpenAI + Anthropic),
records latency, token usage, and output quality, then writes results into the
GlobalGradeManager JSON format so the Performance-by-Model tab shows real data.

Usage:
    python3 scripts/benchmark-models.py [--runs N] [--role ROLE] [--dry-run]
"""

import os
import sys
import json
import time
import math
import datetime
import argparse
import re
from pathlib import Path
from typing import Optional
from concurrent.futures import ThreadPoolExecutor, as_completed

# Suppress deprecation warnings from third-party libraries
import warnings
warnings.filterwarnings("ignore", category=DeprecationWarning)

# ---------------------------------------------------------------------------
# Configuration
# ---------------------------------------------------------------------------

PROJECT_ROOT = Path(__file__).resolve().parent.parent
# Default matches the agent server's DataDir (~/.claude on macOS/Linux).
# Override with --grades-dir if needed.
_DATA_DIR = Path(os.environ.get("AGENT_DATA_DIR", Path.home() / ".claude"))
GRADES_DIR = _DATA_DIR / "performance_grades"
PROJECT_ID = str(PROJECT_ROOT)       # matches what the Go agent records

# Roles to benchmark – now covers all roles in RoleDefaultTier
DEFAULT_ROLES = [
    "engineer",
    "tester",
    "reviewer",
    "orchestrator",
    "architect",
    "inspector",
    "spelunker",
    "cartographer",
    "archaeologist",
    "product-manager",
    "designer",
    "strategist",
]

# Number of benchmark runs per model per role
DEFAULT_RUNS = 5

# --------------------------------------------------------------------------
# Models – mirrors monitoring/model_selector.go ModelsByTier
# --------------------------------------------------------------------------

# Each entry: (id, provider, tier, api_model_name)
# api_model_name is the string sent to the API (may differ from the registry ID)
MODELS = [
    # TierMinimal
    ("gpt-4o-mini",           "openai",     "minimal", "gpt-4o-mini"),
    ("gpt-4.1-nano",          "openai",     "minimal", "gpt-4.1-nano"),
    ("claude-haiku-4-5",      "anthropic",  "minimal", "claude-haiku-4-5"),
    ("gemini-2.5-flash-lite", "gemini",     "minimal", "gemini-2.5-flash-lite"),

    # TierLow
    ("gpt-4.1-mini",          "openai",     "low",     "gpt-4.1-mini"),
    ("o4-mini",               "openai",     "low",     "o4-mini"),
    ("gemini-2.5-flash",      "gemini",     "low",     "gemini-2.5-flash"),

    # TierMedium
    ("gpt-5.1-codex-mini",    "openai",     "medium",  "gpt-5.1-codex-mini"),
    ("gpt-4.1",               "openai",     "medium",  "gpt-4.1"),
    ("claude-sonnet-4-6",     "anthropic",  "medium",  "claude-sonnet-4-6"),
    ("claude-sonnet-4-5",     "anthropic",  "medium",  "claude-sonnet-4-5-20250929"),
    ("gemini-2.5-pro",        "gemini",     "medium",  "gemini-2.5-pro"),

    # TierHigh
    ("gpt-5.1-codex",         "openai",     "high",    "gpt-5.1-codex"),
    ("gpt-5.2-codex",         "openai",     "high",    "gpt-5.2-codex"),
    ("claude-opus-4-5",       "anthropic",  "high",    "claude-opus-4-5-20251101"),
    ("claude-opus-4-6",       "anthropic",  "high",    "claude-opus-4-6"),
]

# ---------------------------------------------------------------------------
# Benchmark prompts per role – varied so repeated calls exercise real reasoning
# ---------------------------------------------------------------------------

# -- engineer ----------------------------------------------------------------
ENGINEER_PROMPTS = [
    {
        "system": "You are a senior software engineer. Respond with concise, correct code only.",
        "user": (
            "Implement a Python function `sieve_of_eratosthenes(n: int) -> list[int]` "
            "that returns all prime numbers up to n using the Sieve of Eratosthenes. "
            "Include type hints and a brief docstring. Do not include a main block."
        ),
    },
    {
        "system": "You are a senior software engineer. Respond with concise, correct code only.",
        "user": (
            "Implement a Python function `binary_search(arr: list, target) -> int` "
            "that returns the index of target in sorted arr, or -1 if not found. "
            "Include type hints and a brief docstring."
        ),
    },
    {
        "system": "You are a senior software engineer. Respond with concise, correct code only.",
        "user": (
            "Implement a Python class `LRUCache` with methods `get(key)` and `put(key, value)` "
            "using O(1) average time complexity. Include type hints and a brief docstring."
        ),
    },
    {
        "system": "You are a senior software engineer. Respond with concise, correct code only.",
        "user": (
            "Write a Python function `deep_merge(base: dict, override: dict) -> dict` "
            "that recursively merges two dictionaries, with override values taking precedence. "
            "Include type hints and a brief docstring."
        ),
    },
    {
        "system": "You are a senior software engineer. Respond with concise, correct code only.",
        "user": (
            "Implement a Python function `topological_sort(graph: dict[str, list[str]]) -> list[str]` "
            "that performs a topological sort on a DAG represented as an adjacency list. "
            "Raise ValueError on cycles. Include type hints and a brief docstring."
        ),
    },
]

# -- tester ------------------------------------------------------------------
TESTER_PROMPTS = [
    {
        "system": "You are a senior QA engineer. Respond with concise, correct test code only.",
        "user": (
            "Write pytest unit tests for a function `parse_date(s: str) -> datetime.date` "
            "that parses strings in YYYY-MM-DD format. Cover happy path, invalid format, "
            "and edge cases (leap year Feb 29, month boundaries). Use pytest.mark.parametrize."
        ),
    },
    {
        "system": "You are a senior QA engineer. Respond with concise, correct test code only.",
        "user": (
            "Write pytest unit tests for a function `divide(a: float, b: float) -> float` "
            "that raises ZeroDivisionError when b is 0. Cover normal division, negative numbers, "
            "zero numerator, and the error case."
        ),
    },
    {
        "system": "You are a senior QA engineer. Respond with concise, correct test code only.",
        "user": (
            "Write pytest unit tests for an `EmailValidator.is_valid(email: str) -> bool` method. "
            "Cover valid emails, missing @, missing domain, empty string, and None input."
        ),
    },
    {
        "system": "You are a senior QA engineer. Respond with concise, correct test code only.",
        "user": (
            "Write pytest tests for a `Stack` class with push, pop, peek, and is_empty methods. "
            "Cover LIFO ordering, pop from empty stack raising IndexError, and peek correctness."
        ),
    },
    {
        "system": "You are a senior QA engineer. Respond with concise, correct test code only.",
        "user": (
            "Write pytest integration tests for a REST endpoint POST /api/users that creates a user. "
            "Use a requests mock. Cover 201 Created, 400 Bad Request for missing fields, "
            "and 409 Conflict for duplicate email."
        ),
    },
]

# -- reviewer ----------------------------------------------------------------
REVIEWER_PROMPTS = [
    {
        "system": "You are a senior code reviewer. Provide concise, actionable review feedback.",
        "user": (
            "Review the following Python code and list the top issues with severity (Critical/Major/Minor):\n\n"
            "```python\n"
            "def get_user(id):\n"
            "    db = connect_db()\n"
            "    result = db.execute('SELECT * FROM users WHERE id=' + str(id))\n"
            "    user = result.fetchone()\n"
            "    db.close()\n"
            "    return user\n"
            "```"
        ),
    },
    {
        "system": "You are a senior code reviewer. Provide concise, actionable review feedback.",
        "user": (
            "Review this Python function and identify bugs, performance issues, and style violations:\n\n"
            "```python\n"
            "def find_duplicates(lst):\n"
            "    dups = []\n"
            "    for i in range(len(lst)):\n"
            "        for j in range(len(lst)):\n"
            "            if i != j and lst[i] == lst[j] and lst[i] not in dups:\n"
            "                dups.append(lst[i])\n"
            "    return dups\n"
            "```"
        ),
    },
    {
        "system": "You are a senior code reviewer. Provide concise, actionable review feedback.",
        "user": (
            "Review this Python async code and identify issues:\n\n"
            "```python\n"
            "async def fetch_all(urls):\n"
            "    results = []\n"
            "    for url in urls:\n"
            "        resp = await fetch(url)\n"
            "        results.append(resp.json())\n"
            "    return results\n"
            "```"
        ),
    },
    {
        "system": "You are a senior code reviewer. Provide concise, actionable review feedback.",
        "user": (
            "Review this Python config loader and identify issues with error handling and security:\n\n"
            "```python\n"
            "def load_config(path):\n"
            "    with open(path) as f:\n"
            "        config = eval(f.read())\n"
            "    os.environ.update(config.get('env', {}))\n"
            "    return config\n"
            "```"
        ),
    },
    {
        "system": "You are a senior code reviewer. Provide concise, actionable review feedback.",
        "user": (
            "Review this Python caching decorator and identify correctness and design issues:\n\n"
            "```python\n"
            "_cache = {}\n"
            "def cache(fn):\n"
            "    def wrapper(*args):\n"
            "        if args in _cache:\n"
            "            return _cache[args]\n"
            "        result = fn(*args)\n"
            "        _cache[args] = result\n"
            "        return result\n"
            "    return wrapper\n"
            "```"
        ),
    },
]

# -- orchestrator ------------------------------------------------------------
ORCHESTRATOR_PROMPTS = [
    {
        "system": (
            "You are an AI orchestrator that coordinates specialist agents to solve problems. "
            "Respond with a concrete delegation plan listing each agent, its task, inputs, "
            "and expected outputs. Be specific and structured."
        ),
        "user": (
            "A product team needs to launch a new user authentication feature in 2 weeks. "
            "The work includes: defining requirements, designing the system architecture, "
            "implementing the backend, writing tests, and creating user docs. "
            "Create a delegation plan for specialist agents (product-manager, architect, "
            "engineer, tester, technical-writer) with clear task assignments, "
            "dependencies, and expected deliverables."
        ),
    },
    {
        "system": (
            "You are an AI orchestrator that coordinates specialist agents to solve problems. "
            "Respond with a concrete delegation plan listing each agent, its task, inputs, "
            "and expected outputs. Be specific and structured."
        ),
        "user": (
            "The team has found a critical production bug: the payment service occasionally "
            "charges users twice. You have access to a spelunker (code navigation), "
            "inspector (bug analysis), engineer (code fixes), and tester (regression tests). "
            "Create a delegation plan to diagnose, fix, and prevent recurrence. "
            "Include which agent goes first, what they produce, and what the next agent needs."
        ),
    },
    {
        "system": (
            "You are an AI orchestrator that coordinates specialist agents to solve problems. "
            "Respond with a concrete delegation plan listing each agent, its task, inputs, "
            "and expected outputs. Be specific and structured."
        ),
        "user": (
            "A new engineer joins the team and needs to understand a 5-year-old legacy codebase "
            "before contributing. You have an archaeologist (history analysis), cartographer "
            "(code structure mapping), and a documentation writer. "
            "Design an onboarding task delegation plan that helps the engineer understand the "
            "codebase architecture, key decisions, and entry points within 3 days."
        ),
    },
    {
        "system": (
            "You are an AI orchestrator that coordinates specialist agents to solve problems. "
            "Respond with a concrete delegation plan listing each agent, its task, inputs, "
            "and expected outputs. Be specific and structured."
        ),
        "user": (
            "The business wants to migrate a monolith application to microservices. "
            "You have a strategist, architect, engineer (×3), tester, and reviewer available. "
            "Create a phased delegation plan for this migration, identifying which agents "
            "work on which phases, what decisions need escalation, and how quality is assured."
        ),
    },
    {
        "system": (
            "You are an AI orchestrator that coordinates specialist agents to solve problems. "
            "Respond with a concrete delegation plan listing each agent, its task, inputs, "
            "and expected outputs. Be specific and structured."
        ),
        "user": (
            "The security team flagged that the app has 3 known vulnerabilities: SQL injection risk, "
            "missing rate limiting, and expired TLS certificates. You have a spelunker, inspector, "
            "engineer, and reviewer available. "
            "Create a delegation plan to triage, fix, and verify all three issues in priority order. "
            "Specify handoffs and what each agent produces."
        ),
    },
]

# -- architect ---------------------------------------------------------------
ARCHITECT_PROMPTS = [
    {
        "system": (
            "You are a senior software architect. Respond with a structured system design "
            "that includes components, data flow, technology choices, and trade-offs."
        ),
        "user": (
            "Design a URL shortener service that must handle 10,000 writes/sec and 100,000 reads/sec. "
            "Include: core components (API layer, storage, cache), data model for short-to-long URL mapping, "
            "how you ensure uniqueness of short codes, caching strategy, and 2 trade-offs in your design."
        ),
    },
    {
        "system": (
            "You are a senior software architect. Respond with a structured system design "
            "that includes components, data flow, technology choices, and trade-offs."
        ),
        "user": (
            "Design a real-time notification system for a social media platform with 50M daily active users. "
            "Requirements: deliver notifications in < 2 seconds, support web and mobile, "
            "allow users to configure notification preferences. "
            "Cover the push delivery mechanism, storage, fanout strategy, and key trade-offs."
        ),
    },
    {
        "system": (
            "You are a senior software architect. Respond with a structured system design "
            "that includes components, data flow, technology choices, and trade-offs."
        ),
        "user": (
            "Design the backend architecture for a multi-tenant SaaS application. "
            "Requirements: strict tenant data isolation, each tenant can have custom configurations, "
            "auto-scaling per tenant load. "
            "Describe the tenancy model (shared DB vs per-tenant), API authentication, "
            "data isolation strategy, and the 2 biggest architectural risks."
        ),
    },
    {
        "system": (
            "You are a senior software architect. Respond with a structured system design "
            "that includes components, data flow, technology choices, and trade-offs."
        ),
        "user": (
            "Design an event-driven order processing system for an e-commerce platform. "
            "The pipeline is: order placed → inventory reserved → payment charged → fulfillment triggered. "
            "Each step can fail independently. Cover the event bus choice, saga pattern vs 2PC, "
            "idempotency handling, and failure compensation strategy."
        ),
    },
    {
        "system": (
            "You are a senior software architect. Respond with a structured system design "
            "that includes components, data flow, technology choices, and trade-offs."
        ),
        "user": (
            "Design a distributed caching layer for a read-heavy API that serves product catalog data. "
            "The catalog has 5M items, updates happen ~100 times/day, and read latency must be < 20ms p99. "
            "Describe the caching tier, invalidation strategy, consistency model, "
            "and how you handle cache stampede."
        ),
    },
]

# -- inspector ---------------------------------------------------------------
INSPECTOR_PROMPTS = [
    {
        "system": (
            "You are a bug inspector and root-cause analyst. "
            "Reason systematically through the bug report and identify likely root causes, "
            "ranked by probability, with your evidence for each."
        ),
        "user": (
            "Bug report: In production, users intermittently receive HTTP 500 errors on the "
            "checkout endpoint. The error occurs roughly 1% of the time, more often under load. "
            "Logs show: 'database connection timeout after 30s'. The DB CPU is normal. "
            "The app uses a connection pool of size 10 and has 50 app instances. "
            "Identify the top 3 likely root causes and what evidence you'd gather to confirm each."
        ),
    },
    {
        "system": (
            "You are a bug inspector and root-cause analyst. "
            "Reason systematically through the bug report and identify likely root causes, "
            "ranked by probability, with your evidence for each."
        ),
        "user": (
            "Bug report: After a deploy last Tuesday, memory usage on the auth service "
            "climbs 50 MB/hour until the pod OOMKills after ~8 hours. "
            "The deploy added JWT token validation. No memory issues before the deploy. "
            "Identify the top 3 likely root causes (focus on memory leaks) and "
            "what diagnostic steps would confirm each."
        ),
    },
    {
        "system": (
            "You are a bug inspector and root-cause analyst. "
            "Reason systematically through the bug report and identify likely root causes, "
            "ranked by probability, with your evidence for each."
        ),
        "user": (
            "Bug report: A nightly batch job that processes 100k records suddenly takes 8 hours "
            "instead of its normal 45 minutes after the database was upgraded from PostgreSQL 13 to 16. "
            "Query plans look unchanged in EXPLAIN output. No new indexes were dropped. "
            "Identify the top 3 likely root causes and what queries/checks would isolate each."
        ),
    },
    {
        "system": (
            "You are a bug inspector and root-cause analyst. "
            "Reason systematically through the bug report and identify likely root causes, "
            "ranked by probability, with your evidence for each."
        ),
        "user": (
            "Bug report: The mobile app shows incorrect balances for ~0.1% of users after a "
            "backend refactor that split the accounts service into two microservices. "
            "The issue is non-deterministic — refreshing sometimes shows the correct balance. "
            "Both services write to the same Postgres database. "
            "List the top 3 root cause hypotheses and what you'd check in the code and DB to confirm."
        ),
    },
    {
        "system": (
            "You are a bug inspector and root-cause analyst. "
            "Reason systematically through the bug report and identify likely root causes, "
            "ranked by probability, with your evidence for each."
        ),
        "user": (
            "Bug report: A Python Flask API returns correct results locally and in staging, "
            "but in production it occasionally returns stale data that's up to 5 minutes old. "
            "Production uses 4 app replicas behind a load balancer. No explicit caching layer. "
            "The database reads are SELECT queries with no ORM caching configured. "
            "List the top 3 root causes and the steps to reproduce or confirm each."
        ),
    },
]

# -- spelunker ---------------------------------------------------------------
SPELUNKER_PROMPTS = [
    {
        "system": (
            "You are a code spelunker — an expert at navigating unfamiliar codebases. "
            "Given a question about where something lives in a codebase, reason through "
            "where to look, what files/patterns to search, and what you'd expect to find."
        ),
        "user": (
            "Codebase question: In a Django REST Framework project, a user reports that "
            "the API rate limiting isn't working. Where would you look to find and fix the "
            "rate limiting configuration? List the files/locations in order you'd check, "
            "what you'd look for in each, and what a misconfiguration might look like."
        ),
    },
    {
        "system": (
            "You are a code spelunker — an expert at navigating unfamiliar codebases. "
            "Given a question about where something lives in a codebase, reason through "
            "where to look, what files/patterns to search, and what you'd expect to find."
        ),
        "user": (
            "Codebase question: In a Node.js Express app, database queries are slow and "
            "you suspect there's no connection pooling. Where would you look to verify "
            "connection pooling is configured? What file patterns, config keys, and library "
            "calls would confirm or deny pooling is active?"
        ),
    },
    {
        "system": (
            "You are a code spelunker — an expert at navigating unfamiliar codebases. "
            "Given a question about where something lives in a codebase, reason through "
            "where to look, what files/patterns to search, and what you'd expect to find."
        ),
        "user": (
            "Codebase question: A Go microservice is emitting logs but they never appear "
            "in the centralized log aggregator. Where would you look to trace the log "
            "pipeline — from the code that writes logs, through any middleware or hooks, "
            "to the output destination? List search terms and file patterns."
        ),
    },
    {
        "system": (
            "You are a code spelunker — an expert at navigating unfamiliar codebases. "
            "Given a question about where something lives in a codebase, reason through "
            "where to look, what files/patterns to search, and what you'd expect to find."
        ),
        "user": (
            "Codebase question: In a React/TypeScript frontend, users report that changing "
            "their profile photo doesn't reflect immediately — they need to hard-refresh. "
            "Where would you look to find the caching or state management issue? "
            "Describe the file types, component patterns, and state management code to inspect."
        ),
    },
    {
        "system": (
            "You are a code spelunker — an expert at navigating unfamiliar codebases. "
            "Given a question about where something lives in a codebase, reason through "
            "where to look, what files/patterns to search, and what you'd expect to find."
        ),
        "user": (
            "Codebase question: A Java Spring Boot app has a feature flag system but you don't "
            "know how it works. Where would you look to understand how feature flags are defined, "
            "evaluated, and toggled? List the package structures, annotation names, and config "
            "patterns typical for Spring-based feature flag implementations."
        ),
    },
]

# -- cartographer ------------------------------------------------------------
CARTOGRAPHER_PROMPTS = [
    {
        "system": (
            "You are a code cartographer — an expert at mapping codebase structure. "
            "Given a file/directory listing, describe the architecture, component responsibilities, "
            "likely entry points, and how the pieces fit together."
        ),
        "user": (
            "Describe the architecture of a codebase with this structure:\n\n"
            "```\n"
            "myapp/\n"
            "├── cmd/\n"
            "│   └── server/main.go\n"
            "├── internal/\n"
            "│   ├── api/\n"
            "│   │   ├── handlers/\n"
            "│   │   └── middleware/\n"
            "│   ├── domain/\n"
            "│   │   ├── user/\n"
            "│   │   └── order/\n"
            "│   ├── repository/\n"
            "│   │   └── postgres/\n"
            "│   └── service/\n"
            "├── pkg/\n"
            "│   ├── config/\n"
            "│   └── logger/\n"
            "└── migrations/\n"
            "```\n\n"
            "What architectural pattern does this follow? What are the key components and their "
            "responsibilities? Where would you add a new REST endpoint?"
        ),
    },
    {
        "system": (
            "You are a code cartographer — an expert at mapping codebase structure. "
            "Given a file/directory listing, describe the architecture, component responsibilities, "
            "likely entry points, and how the pieces fit together."
        ),
        "user": (
            "Describe the architecture of a Python project with this structure:\n\n"
            "```\n"
            "ecommerce/\n"
            "├── apps/\n"
            "│   ├── catalog/\n"
            "│   │   ├── models.py\n"
            "│   │   ├── views.py\n"
            "│   │   ├── serializers.py\n"
            "│   │   └── urls.py\n"
            "│   ├── orders/\n"
            "│   └── payments/\n"
            "├── core/\n"
            "│   ├── settings/\n"
            "│   │   ├── base.py\n"
            "│   │   ├── dev.py\n"
            "│   │   └── prod.py\n"
            "│   └── celery.py\n"
            "├── config/\n"
            "└── tests/\n"
            "    ├── unit/\n"
            "    └── integration/\n"
            "```\n\n"
            "What framework is this? Describe the architectural pattern, how apps communicate, "
            "and how background tasks are handled."
        ),
    },
    {
        "system": (
            "You are a code cartographer — an expert at mapping codebase structure. "
            "Given a file/directory listing, describe the architecture, component responsibilities, "
            "likely entry points, and how the pieces fit together."
        ),
        "user": (
            "Describe the architecture of this microservices repository:\n\n"
            "```\n"
            "platform/\n"
            "├── services/\n"
            "│   ├── user-service/\n"
            "│   │   ├── Dockerfile\n"
            "│   │   └── src/\n"
            "│   ├── product-service/\n"
            "│   ├── order-service/\n"
            "│   └── notification-service/\n"
            "├── gateway/\n"
            "│   └── nginx.conf\n"
            "├── infra/\n"
            "│   ├── terraform/\n"
            "│   └── k8s/\n"
            "│       ├── deployments/\n"
            "│       └── services/\n"
            "├── shared/\n"
            "│   ├── proto/\n"
            "│   └── events/\n"
            "└── docker-compose.yml\n"
            "```\n\n"
            "Describe the overall architecture, how services communicate, the role of each top-level "
            "directory, and what shared/ suggests about inter-service contracts."
        ),
    },
    {
        "system": (
            "You are a code cartographer — an expert at mapping codebase structure. "
            "Given a file/directory listing, describe the architecture, component responsibilities, "
            "likely entry points, and how the pieces fit together."
        ),
        "user": (
            "Describe the architecture of this frontend project:\n\n"
            "```\n"
            "frontend/\n"
            "├── src/\n"
            "│   ├── app/\n"
            "│   │   ├── store/\n"
            "│   │   ├── router/\n"
            "│   │   └── App.tsx\n"
            "│   ├── features/\n"
            "│   │   ├── auth/\n"
            "│   │   │   ├── authSlice.ts\n"
            "│   │   │   ├── LoginPage.tsx\n"
            "│   │   │   └── authApi.ts\n"
            "│   │   └── dashboard/\n"
            "│   ├── shared/\n"
            "│   │   ├── components/\n"
            "│   │   ├── hooks/\n"
            "│   │   └── utils/\n"
            "│   └── index.tsx\n"
            "├── public/\n"
            "└── vite.config.ts\n"
            "```\n\n"
            "What frontend framework and state management library does this use? "
            "Describe the feature-based architecture and how a new feature should be added."
        ),
    },
    {
        "system": (
            "You are a code cartographer — an expert at mapping codebase structure. "
            "Given a file/directory listing, describe the architecture, component responsibilities, "
            "likely entry points, and how the pieces fit together."
        ),
        "user": (
            "Describe the architecture of this data pipeline project:\n\n"
            "```\n"
            "pipeline/\n"
            "├── dags/\n"
            "│   ├── ingest_raw.py\n"
            "│   ├── transform_events.py\n"
            "│   └── load_warehouse.py\n"
            "├── plugins/\n"
            "│   └── operators/\n"
            "│       ├── s3_operator.py\n"
            "│       └── snowflake_operator.py\n"
            "├── include/\n"
            "│   ├── sql/\n"
            "│   └── schemas/\n"
            "├── tests/\n"
            "│   └── dags/\n"
            "└── airflow.cfg\n"
            "```\n\n"
            "What orchestration tool is used? Describe the ELT pipeline stages, "
            "the role of operators, and how data flows from source to warehouse."
        ),
    },
]

# -- archaeologist -----------------------------------------------------------
ARCHAEOLOGIST_PROMPTS = [
    {
        "system": (
            "You are a code archaeologist — an expert at reconstructing the history and rationale "
            "behind legacy code and technical decisions. Given a code artifact or description, "
            "reason about why it was built this way, what problems it solved, and what technical "
            "debt it represents."
        ),
        "user": (
            "You discover this pattern in a 10-year-old banking codebase:\n\n"
            "```python\n"
            "# DO NOT REMOVE - legacy support\n"
            "ACCOUNT_TYPE_MAP = {\n"
            "    'CHK': 'checking',\n"
            "    'SAV': 'savings',\n"
            "    'CHK2': 'checking',  # added 2017\n"
            "    'CHK_LEGACY': 'checking',  # migration artifact\n"
            "    'SAV_OLD': 'savings',\n"
            "}\n"
            "```\n\n"
            "Reconstruct the likely history: why does this map exist, what events probably led to "
            "the duplicate mappings, what risks does this represent, and how would you safely "
            "modernize it?"
        ),
    },
    {
        "system": (
            "You are a code archaeologist — an expert at reconstructing the history and rationale "
            "behind legacy code and technical decisions. Given a code artifact or description, "
            "reason about why it was built this way, what problems it solved, and what technical "
            "debt it represents."
        ),
        "user": (
            "You find a 7-year-old microservice that has three separate database connections:\n"
            "1. A MySQL connection for user data\n"
            "2. A MongoDB connection for the same users' preferences\n"
            "3. A Redis connection also storing some user session data\n\n"
            "All three have slightly different schemas for the 'user' concept. "
            "Reconstruct the probable history of how this fragmentation happened, "
            "what organizational or technical pressures caused it, and what the "
            "consolidation strategy should consider."
        ),
    },
    {
        "system": (
            "You are a code archaeologist — an expert at reconstructing the history and rationale "
            "behind legacy code and technical decisions. Given a code artifact or description, "
            "reason about why it was built this way, what problems it solved, and what technical "
            "debt it represents."
        ),
        "user": (
            "A codebase has this file structure artifact:\n\n"
            "```\n"
            "utils/\n"
            "├── helpers.py          (3000 lines, imports from helpers2.py)\n"
            "├── helpers2.py         (1500 lines, comment: 'overflow from helpers.py')\n"
            "├── helpers_new.py      (800 lines, comment: 'refactor in progress - do not use yet')\n"
            "├── helpers_v2.py       (400 lines, comment: 'deprecated, use helpers_new')\n"
            "└── helpers_final.py    (200 lines, 5 years old)\n"
            "```\n\n"
            "Narrate the archaeological history of this utils directory. What does each file "
            "represent in the project's evolution? What risks exist in the current state? "
            "Propose a consolidation plan."
        ),
    },
    {
        "system": (
            "You are a code archaeologist — an expert at reconstructing the history and rationale "
            "behind legacy code and technical decisions. Given a code artifact or description, "
            "reason about why it was built this way, what problems it solved, and what technical "
            "debt it represents."
        ),
        "user": (
            "You find a critical business rule buried in a comment inside a 12-year-old stored procedure:\n\n"
            "```sql\n"
            "-- IMPORTANT: Never delete orders with status 'P'\n"
            "-- Ask Mike if you need to change this (2011)\n"
            "-- Mike left the company in 2015\n"
            "-- DO NOT TOUCH without asking Legal (2019)\n"
            "IF @status = 'P' RETURN  -- compliance requirement\n"
            "```\n\n"
            "Analyze this code artifact archaeologically. What does 'P' probably stand for? "
            "What organizational knowledge is at risk? How would you safely excavate the "
            "rationale and document it properly?"
        ),
    },
    {
        "system": (
            "You are a code archaeologist — an expert at reconstructing the history and rationale "
            "behind legacy code and technical decisions. Given a code artifact or description, "
            "reason about why it was built this way, what problems it solved, and what technical "
            "debt it represents."
        ),
        "user": (
            "A codebase has two parallel REST API implementations:\n"
            "- /api/v1/* — Django REST Framework, 5 years old, 60 endpoints\n"
            "- /api/v2/* — FastAPI, 2 years old, 40 endpoints\n\n"
            "Both are in production. Some v2 endpoints shadow v1 endpoints with different response shapes. "
            "Reconstruct the history of how this dual-API state emerged, what the original intent was, "
            "why the migration stalled, and what a realistic remediation plan looks like."
        ),
    },
]

# -- product-manager ---------------------------------------------------------
PRODUCT_MANAGER_PROMPTS = [
    {
        "system": (
            "You are a senior product manager. Given a user need or business problem, "
            "produce a concise requirements specification with user stories, acceptance criteria, "
            "and success metrics. Be specific and measurable."
        ),
        "user": (
            "User need: Our e-commerce customers frequently abandon checkout when they can't "
            "remember which card they used last time. Write a requirements spec for a "
            "'Saved Payment Methods' feature. Include: 2-3 user stories, acceptance criteria "
            "for each, non-functional requirements (security, UX), and 2 success metrics."
        ),
    },
    {
        "system": (
            "You are a senior product manager. Given a user need or business problem, "
            "produce a concise requirements specification with user stories, acceptance criteria, "
            "and success metrics. Be specific and measurable."
        ),
        "user": (
            "User need: B2B SaaS customers want to let their managers approve or reject "
            "employee expense reports within the app. Currently, approvals happen via email. "
            "Write a requirements spec for an in-app expense approval workflow. "
            "Cover the approval lifecycle, notification requirements, audit trail needs, "
            "and 3 measurable success criteria."
        ),
    },
    {
        "system": (
            "You are a senior product manager. Given a user need or business problem, "
            "produce a concise requirements specification with user stories, acceptance criteria, "
            "and success metrics. Be specific and measurable."
        ),
        "user": (
            "User need: Users of a project management tool say they lose track of what changed "
            "when they return from vacation. Write a requirements spec for a 'Catch Me Up' "
            "digest feature that summarizes project activity since a user last logged in. "
            "Define scope, user stories, what counts as a meaningful activity vs. noise, "
            "and how you'd measure success."
        ),
    },
    {
        "system": (
            "You are a senior product manager. Given a user need or business problem, "
            "produce a concise requirements specification with user stories, acceptance criteria, "
            "and success metrics. Be specific and measurable."
        ),
        "user": (
            "User need: A developer tools company sees 40% of new users churn within 7 days "
            "because they never complete the initial setup. Write a requirements spec for an "
            "onboarding wizard that guides developers through first-time configuration. "
            "Define the steps, skip/resume behavior, success state, and the metrics that "
            "indicate the onboarding is working."
        ),
    },
    {
        "system": (
            "You are a senior product manager. Given a user need or business problem, "
            "produce a concise requirements specification with user stories, acceptance criteria, "
            "and success metrics. Be specific and measurable."
        ),
        "user": (
            "User need: Mobile app users complain they miss important alerts because they "
            "get too many notifications and start ignoring them all. Write a requirements spec "
            "for a smart notification preference center. Cover notification categories, "
            "frequency controls, quiet hours, channel selection (push/email/SMS), "
            "and how you'd know the feature improved notification engagement."
        ),
    },
]

# -- designer ----------------------------------------------------------------
DESIGNER_PROMPTS = [
    {
        "system": (
            "You are a senior UX designer. Given a feature, describe the user experience flow "
            "in detail: screens, user actions, system responses, edge cases, and design principles. "
            "Be specific about what the user sees and does at each step."
        ),
        "user": (
            "Design the UX flow for a 'Forgot Password' feature in a mobile banking app. "
            "Cover: entry point (where user starts), identity verification steps, "
            "password reset form with validation feedback, success and error states. "
            "Include accessibility considerations and what happens if verification fails 3 times."
        ),
    },
    {
        "system": (
            "You are a senior UX designer. Given a feature, describe the user experience flow "
            "in detail: screens, user actions, system responses, edge cases, and design principles. "
            "Be specific about what the user sees and does at each step."
        ),
        "user": (
            "Design the UX flow for bulk-editing multiple items in a task management app. "
            "The user wants to reassign 20 tasks from one team member to another at once. "
            "Cover: how the user enters bulk-select mode, selection interaction, "
            "the bulk action menu, confirmation step, progress feedback, and undo capability."
        ),
    },
    {
        "system": (
            "You are a senior UX designer. Given a feature, describe the user experience flow "
            "in detail: screens, user actions, system responses, edge cases, and design principles. "
            "Be specific about what the user sees and does at each step."
        ),
        "user": (
            "Design the UX flow for an in-app subscription upgrade from free to paid tier. "
            "The user hits a paywall while using a premium feature. "
            "Cover: the paywall moment (what the user sees), the upgrade flow steps, "
            "payment entry, confirmation, and the moment they gain access. "
            "Address what happens if payment fails and how to minimize friction."
        ),
    },
    {
        "system": (
            "You are a senior UX designer. Given a feature, describe the user experience flow "
            "in detail: screens, user actions, system responses, edge cases, and design principles. "
            "Be specific about what the user sees and does at each step."
        ),
        "user": (
            "Design the UX flow for a first-time user onboarding in a personal finance app. "
            "The user has just created their account. "
            "Cover: the welcome screen, connecting their first bank account (OAuth flow), "
            "the moment transactions load for the first time, and how to surface the app's "
            "key value quickly. Address the empty state before transactions are available."
        ),
    },
    {
        "system": (
            "You are a senior UX designer. Given a feature, describe the user experience flow "
            "in detail: screens, user actions, system responses, edge cases, and design principles. "
            "Be specific about what the user sees and does at each step."
        ),
        "user": (
            "Design the UX flow for a real-time collaborative document editor (like Google Docs). "
            "Focus on the collaboration aspects: how the user knows others are present, "
            "how simultaneous edits are shown, cursor/selection visibility, "
            "conflict resolution feedback, and how commenting and @mentions work in context."
        ),
    },
]

# -- strategist --------------------------------------------------------------
STRATEGIST_PROMPTS = [
    {
        "system": (
            "You are a senior business strategist. Given a business problem, "
            "produce a structured strategy outline with situation analysis, strategic options, "
            "recommended approach, key risks, and success metrics. Be specific and actionable."
        ),
        "user": (
            "Business problem: A B2B SaaS company has 90% of revenue from 3 enterprise clients. "
            "The largest (50% of revenue) just announced they're evaluating competitors. "
            "Outline a strategy to reduce revenue concentration risk and retain the at-risk client. "
            "Cover: immediate retention actions, medium-term diversification approach, "
            "and how to measure progress."
        ),
    },
    {
        "system": (
            "You are a senior business strategist. Given a business problem, "
            "produce a structured strategy outline with situation analysis, strategic options, "
            "recommended approach, key risks, and success metrics. Be specific and actionable."
        ),
        "user": (
            "Business problem: A developer tools startup has 10k free users and 200 paid users "
            "(2% conversion). Industry benchmark is 5%. "
            "Outline a strategy to double paid conversion to 4% within 6 months. "
            "Cover: hypothesis for why conversion is low, 3 strategic levers to pull, "
            "prioritization, and how you'd measure what's working."
        ),
    },
    {
        "system": (
            "You are a senior business strategist. Given a business problem, "
            "produce a structured strategy outline with situation analysis, strategic options, "
            "recommended approach, key risks, and success metrics. Be specific and actionable."
        ),
        "user": (
            "Business problem: A marketplace platform that connects freelancers with clients "
            "is facing a 'chicken and egg' problem at launch: freelancers won't join without "
            "clients, and clients won't join without freelancers. "
            "Outline a cold-start strategy to bootstrap both sides of the marketplace. "
            "Cover supply-side and demand-side tactics, sequencing, and how to know when "
            "you've reached critical mass."
        ),
    },
    {
        "system": (
            "You are a senior business strategist. Given a business problem, "
            "produce a structured strategy outline with situation analysis, strategic options, "
            "recommended approach, key risks, and success metrics. Be specific and actionable."
        ),
        "user": (
            "Business problem: A mobile app in a competitive market is growing through paid "
            "acquisition (CAC = $18, LTV = $22). As the market matures, CAC is rising 20% QoQ. "
            "Outline a strategy to shift toward organic/product-led growth to improve unit economics. "
            "Cover: 3 organic growth levers, how to transition without losing current growth rate, "
            "and the metrics that signal the strategy is working."
        ),
    },
    {
        "system": (
            "You are a senior business strategist. Given a business problem, "
            "produce a structured strategy outline with situation analysis, strategic options, "
            "recommended approach, key risks, and success metrics. Be specific and actionable."
        ),
        "user": (
            "Business problem: An established SaaS company (Series B, $15M ARR) is seeing "
            "a new well-funded competitor offer their core product for free. "
            "Outline a competitive response strategy. Cover: whether to compete on price, "
            "where to differentiate, how to protect existing customers, and what signals would "
            "indicate whether to fight, partner, or find a new market segment."
        ),
    },
]

# Map role → prompts list
ROLE_PROMPTS = {
    "engineer":       ENGINEER_PROMPTS,
    "tester":         TESTER_PROMPTS,
    "reviewer":       REVIEWER_PROMPTS,
    "orchestrator":   ORCHESTRATOR_PROMPTS,
    "architect":      ARCHITECT_PROMPTS,
    "inspector":      INSPECTOR_PROMPTS,
    "spelunker":      SPELUNKER_PROMPTS,
    "cartographer":   CARTOGRAPHER_PROMPTS,
    "archaeologist":  ARCHAEOLOGIST_PROMPTS,
    "product-manager": PRODUCT_MANAGER_PROMPTS,
    "designer":       DESIGNER_PROMPTS,
    "strategist":     STRATEGIST_PROMPTS,
}

# For backward compatibility keep the original name too
BENCHMARK_PROMPTS = ENGINEER_PROMPTS

# ---------------------------------------------------------------------------
# Quality evaluators — one per role
# ---------------------------------------------------------------------------

def evaluate_engineer(output: str, prompt_index: int) -> bool:
    """Must contain a Python function/class definition with minimal length."""
    if not output or len(output) < 80:
        return False
    if not re.search(r'\b(def |class )\w+', output):
        return False
    if output.count(':') < 2:
        return False
    lower = output.lower()
    refusals = [r"i cannot", r"i'm unable", r"i am unable", r"as an ai",
                r"i don't have", r"i won't"]
    if any(re.search(p, lower) for p in refusals):
        return False
    return True


def evaluate_tester(output: str, prompt_index: int) -> bool:
    """Must contain pytest test functions."""
    if not output or len(output) < 80:
        return False
    if not re.search(r'def test_\w+', output):
        return False
    if not re.search(r'\b(assert|pytest\.raises)\b', output):
        return False
    return True


def evaluate_reviewer(output: str, prompt_index: int) -> bool:
    """Must identify at least one issue or concern in the code."""
    if not output or len(output) < 60:
        return False
    # Accept any of: severity labels, or common code-review feedback terms
    review_pattern = (
        r'\b(critical|major|minor|bug|issue|problem|vulnerability|risk|'
        r'error|flaw|concern|incorrect|unsafe|inefficient|performance|'
        r'fix|improve|suggest|recommend|avoid|should|instead|better|'
        r'correctness|logic|security|injection|style|violation)\b'
    )
    if not re.search(review_pattern, output.lower()):
        return False
    return True


def evaluate_orchestrator(output: str, prompt_index: int) -> bool:
    """Must contain a structured delegation plan."""
    if not output or len(output) < 100:
        return False
    # Must mention roles/agents OR task delegation concepts
    agent_pattern = (
        r'\b(agent|engineer|architect|tester|reviewer|spelunker|inspector|'
        r'orchestrator|product.manager|designer|strategist|team|task|step|'
        r'phase|assign|delegate|handoff|parallel|sequen|depend|owner|'
        r'responsible|coordinate|plan|workflow|pipeline)\b'
    )
    if not re.search(agent_pattern, output.lower()):
        return False
    # Must have some structure (numbered list, bullet, or heading)
    if not re.search(r'(\d+\.|[-*•]|##|\n[-*])', output):
        return False
    return True


def evaluate_architect(output: str, prompt_index: int) -> bool:
    """Must contain system design elements: components, trade-offs."""
    if not output or len(output) < 100:
        return False
    design_pattern = r'\b(component|service|database|cache|api|layer|trade.?off|scalab|'
    design_pattern += r'latency|throughput|consistency|availability)\b'
    if not re.search(design_pattern, output.lower()):
        return False
    return True


def evaluate_inspector(output: str, prompt_index: int) -> bool:
    """Must identify root causes with reasoning."""
    if not output or len(output) < 100:
        return False
    cause_pattern = (
        r'\b(root cause|likely cause|hypothesis|probable|evidence|diagnos|'
        r'investigate|check|look for|cause|reason|suspect|because|due to|'
        r'explain|indicate|symptom|contribut|trigger|factor|analyz)\b'
    )
    if not re.search(cause_pattern, output.lower()):
        return False
    return True


def evaluate_spelunker(output: str, prompt_index: int) -> bool:
    """Must describe where to look in code with file/pattern references."""
    if not output or len(output) < 80:
        return False
    nav_pattern = r'\b(file|directory|module|class|function|search|grep|look|check|find|'
    nav_pattern += r'config|settings|middleware|handler)\b'
    if not re.search(nav_pattern, output.lower()):
        return False
    return True


def evaluate_cartographer(output: str, prompt_index: int) -> bool:
    """Must describe architectural patterns and component roles."""
    if not output or len(output) < 80:
        return False
    arch_pattern = r'\b(architecture|pattern|component|layer|service|module|responsibility|'
    arch_pattern += r'entry.?point|structure|mvc|clean|hexagonal|domain)\b'
    if not re.search(arch_pattern, output.lower()):
        return False
    return True


def evaluate_archaeologist(output: str, prompt_index: int) -> bool:
    """Must reason about history, legacy, and technical debt."""
    if not output or len(output) < 80:
        return False
    history_pattern = (
        r'\b(history|legacy|technical debt|evolved|migration|over time|'
        r'original|deprecated|refactor|consolidat|risk|old|grew|accumul|'
        r'started|began|added|grew|pattern|debt|duplication|scattered|'
        r'duplicate|redundan|obsolete|replaced|supersed|version|old code)\b'
    )
    if not re.search(history_pattern, output.lower()):
        return False
    return True


def evaluate_product_manager(output: str, prompt_index: int) -> bool:
    """Must contain user stories or requirements-related content."""
    if not output or len(output) < 80:
        return False
    pm_pattern = (
        r'\b(user story|as a user|acceptance criteria|given|when|then|'
        r'success metric|kpi|measure|requirement|feature|functionality|'
        r'onboarding|notification|workflow|specification|spec|criteria|'
        r'objective|goal|outcome|user need|use case|priority)\b'
    )
    if not re.search(pm_pattern, output.lower()):
        return False
    return True


def evaluate_designer(output: str, prompt_index: int) -> bool:
    """Must describe UX flow with screens or steps."""
    if not output or len(output) < 80:
        return False
    ux_pattern = r'\b(screen|step|user|click|tap|button|form|flow|state|error|'
    ux_pattern += r'success|feedback|navigation|modal|page)\b'
    if not re.search(ux_pattern, output.lower()):
        return False
    return True


def evaluate_strategist(output: str, prompt_index: int) -> bool:
    """Must contain strategic analysis with recommendations and metrics."""
    if not output or len(output) < 100:
        return False
    strategy_pattern = r'\b(strategy|strategic|approach|option|recommend|metric|'
    strategy_pattern += r'risk|opportunit|objective|measure|kpi|growth|revenue)\b'
    if not re.search(strategy_pattern, output.lower()):
        return False
    return True


# Default fallback evaluator
def evaluate_quality(output: str, prompt_index: int) -> bool:
    """Generic evaluator: checks for reasonable length and structured response."""
    if not output or len(output) < 80:
        return False
    lower = output.lower()
    refusals = [r"i cannot", r"i'm unable", r"i am unable", r"as an ai",
                r"i don't have", r"i won't"]
    if any(re.search(p, lower) for p in refusals):
        return False
    return True


ROLE_EVALUATORS = {
    "engineer":        evaluate_engineer,
    "tester":          evaluate_tester,
    "reviewer":        evaluate_reviewer,
    "orchestrator":    evaluate_orchestrator,
    "architect":       evaluate_architect,
    "inspector":       evaluate_inspector,
    "spelunker":       evaluate_spelunker,
    "cartographer":    evaluate_cartographer,
    "archaeologist":   evaluate_archaeologist,
    "product-manager": evaluate_product_manager,
    "designer":        evaluate_designer,
    "strategist":      evaluate_strategist,
}


# ---------------------------------------------------------------------------
# LLM-as-judge rubrics — stricter quality gate applied after basic evaluator
# ---------------------------------------------------------------------------

JUDGE_RUBRICS = {
    "engineer": (
        "Evaluate this code response. Reply YES if: (1) the code is syntactically valid Python, "
        "(2) it correctly implements the described algorithm with no obvious logic errors, "
        "(3) it handles the core requirements (e.g. edge cases mentioned). "
        "Reply NO if the code is incomplete, pseudocode only, has wrong logic, or is missing key requirements. "
        "Reply with only YES or NO, nothing else."
    ),
    "tester": (
        "Evaluate this pytest test code. Reply YES if: (1) it contains properly named test functions (def test_...), "
        "(2) assertions test the specific behaviors described in the task, "
        "(3) it covers at least 2 distinct scenarios or cases. "
        "Reply NO if tests are trivial stubs, missing meaningful assertions, or don't test what was asked. "
        "Reply with only YES or NO, nothing else."
    ),
    "reviewer": (
        "Evaluate this code review response. Reply YES if: (1) it identifies at least 2 specific issues, "
        "(2) it explains WHY each is a problem (not just names it), "
        "(3) it provides actionable fixes or improvements. "
        "Reply NO if it only gives vague feedback, misses critical issues, or is a single paragraph. "
        "Reply with only YES or NO, nothing else."
    ),
    "orchestrator": (
        "Evaluate this task delegation plan. Reply YES if: (1) it assigns specific tasks to named agent roles, "
        "(2) it specifies what each agent produces as output, "
        "(3) it addresses task ordering or dependencies between agents. "
        "Reply NO if agent assignments are vague, outputs are unspecified, or it's just a generic plan. "
        "Reply with only YES or NO, nothing else."
    ),
    "architect": (
        "Evaluate this system design response. Reply YES if: (1) it names and describes specific components, "
        "(2) it addresses the stated scale/performance numbers from the task, "
        "(3) it explicitly discusses at least one trade-off. "
        "Reply NO if the design is generic, ignores the stated requirements, or lacks any trade-off analysis. "
        "Reply with only YES or NO, nothing else."
    ),
    "inspector": (
        "Evaluate this root cause analysis. Reply YES if: (1) it proposes at least 2 specific hypotheses, "
        "(2) each hypothesis has a causal explanation tied to the described symptoms, "
        "(3) it suggests concrete diagnostic steps to confirm each. "
        "Reply NO if hypotheses are generic, lack causal reasoning, or miss the key symptoms. "
        "Reply with only YES or NO, nothing else."
    ),
    "spelunker": (
        "Evaluate this code navigation guidance. Reply YES if: (1) it names specific files, directories, "
        "or modules to check, (2) it explains what to look for in each location, "
        "(3) the search strategy is logical and actionable for a developer. "
        "Reply NO if the advice is too vague to act on or doesn't reference concrete code locations. "
        "Reply with only YES or NO, nothing else."
    ),
    "cartographer": (
        "Evaluate this codebase architecture description. Reply YES if: (1) it correctly identifies the "
        "architectural pattern, (2) it describes the specific responsibility of key directories, "
        "(3) it explains how components interact. "
        "Reply NO if it misidentifies the architecture or gives only surface-level descriptions. "
        "Reply with only YES or NO, nothing else."
    ),
    "archaeologist": (
        "Evaluate this legacy code analysis. Reply YES if: (1) it constructs a plausible historical narrative "
        "explaining how the artifact evolved, (2) it identifies at least 2 specific risks or debt items, "
        "(3) it proposes a concrete remediation approach. "
        "Reply NO if it only describes what's visible without explaining history or risks. "
        "Reply with only YES or NO, nothing else."
    ),
    "product-manager": (
        "Evaluate this product requirements response. Reply YES if: (1) it includes user stories with "
        "clear user/action/benefit structure, (2) acceptance criteria are specific and testable, "
        "(3) at least one success metric is quantifiable (e.g. percentage, time, count). "
        "Reply NO if stories are vague, criteria are untestable, or metrics are unmeasurable. "
        "Reply with only YES or NO, nothing else."
    ),
    "designer": (
        "Evaluate this UX flow description. Reply YES if: (1) it describes specific screens or steps "
        "the user goes through, (2) it addresses at least one error state or edge case, "
        "(3) it specifies user feedback at key interaction points. "
        "Reply NO if the flow is incomplete, skips error handling, or is too abstract to implement. "
        "Reply with only YES or NO, nothing else."
    ),
    "strategist": (
        "Evaluate this business strategy response. Reply YES if: (1) it analyzes the specific problem "
        "with evidence-based reasoning, (2) it proposes at least 2 concrete strategic options, "
        "(3) it defines measurable success criteria or KPIs. "
        "Reply NO if the strategy is generic, avoids the specific constraints, or lacks measurable outcomes. "
        "Reply with only YES or NO, nothing else."
    ),
}

_DEFAULT_JUDGE_RUBRIC = (
    "Evaluate whether this response fully and specifically addresses the task. "
    "Reply YES if the response is detailed, accurate, and directly answers what was asked. "
    "Reply NO if it's vague, incomplete, or off-topic. "
    "Reply with only YES or NO, nothing else."
)


def judge_response(role: str, prompt: dict, response_text: str,
                   openai_key: str, anthropic_key: str) -> bool:
    """Call a cheap judge model to evaluate response quality. Returns True = pass."""
    rubric = JUDGE_RUBRICS.get(role, _DEFAULT_JUDGE_RUBRIC)
    judge_prompt = {
        "user": (
            f"{rubric}\n\n"
            f"TASK:\n{prompt['user'][:600]}\n\n"
            f"RESPONSE:\n{response_text[:4000]}"
        )
    }
    if openai_key:
        result = call_with_retry("openai", "gpt-4o-mini", judge_prompt, temperature=0.0)
    elif anthropic_key:
        result = call_with_retry("anthropic", "claude-haiku-4-5", judge_prompt, temperature=0.0)
    else:
        return True  # No judge available — fall back to lenient pass
    if not result["success"]:
        return True  # Judge call failed — be lenient
    return result["text"].strip().upper().startswith("YES")


# ---------------------------------------------------------------------------
# API wrappers
# ---------------------------------------------------------------------------

def call_openai(api_model: str, prompt: dict, timeout: int = 60, temperature: float = 0.3) -> dict:
    """Call OpenAI and return timing + token data.

    Routes to the appropriate API endpoint based on model family:
    - codex models → v1/responses (Responses API)
    - o-series (o1/o3/o4) → chat completions with max_completion_tokens, no system
    - all others → chat completions standard
    """
    import openai

    client = openai.OpenAI(api_key=os.environ["OPENAI_API_KEY"])

    # Codex models only work via the Responses API
    is_codex = "codex" in api_model.lower()
    is_reasoning_model = api_model.startswith(("o1", "o3", "o4"))

    start = time.monotonic()
    try:
        if is_codex:
            user_input = prompt["user"]
            if prompt.get("system"):
                user_input = f"{prompt['system']}\n\n{prompt['user']}"
            resp = client.responses.create(
                model=api_model,
                input=user_input,
                max_output_tokens=1024,
            )
            elapsed_ms = (time.monotonic() - start) * 1000
            text = resp.output_text if hasattr(resp, "output_text") else ""
            tokens_in = resp.usage.input_tokens if resp.usage else 0
            tokens_out = resp.usage.output_tokens if resp.usage else 0
        else:
            messages = []
            if prompt.get("system"):
                messages.append({"role": "system", "content": prompt["system"]})
            messages.append({"role": "user", "content": prompt["user"]})

            if is_reasoning_model:
                messages = [m for m in messages if m["role"] != "system"]
                # Reasoning models (o1/o3/o4) use reasoning tokens internally;
                # grant a larger budget so visible output tokens are non-empty.
                resp = client.chat.completions.create(
                    model=api_model,
                    messages=messages,
                    max_completion_tokens=8192,
                )
            else:
                resp = client.chat.completions.create(
                    model=api_model,
                    messages=messages,
                    max_tokens=1024,
                    temperature=temperature,
                )
            elapsed_ms = (time.monotonic() - start) * 1000
            text = resp.choices[0].message.content or "" if resp.choices else ""
            tokens_in = resp.usage.prompt_tokens if resp.usage else 0
            tokens_out = resp.usage.completion_tokens if resp.usage else 0

        return {
            "success": True,
            "text": text,
            "tokens_in": tokens_in,
            "tokens_out": tokens_out,
            "total_tokens": tokens_in + tokens_out,
            "elapsed_ms": elapsed_ms,
            "error": None,
        }
    except Exception as exc:
        elapsed_ms = (time.monotonic() - start) * 1000
        return {
            "success": False,
            "text": "",
            "tokens_in": 0,
            "tokens_out": 0,
            "total_tokens": 0,
            "elapsed_ms": elapsed_ms,
            "error": str(exc),
        }


def call_anthropic(api_model: str, prompt: dict, timeout: int = 60, temperature: float = 0.3) -> dict:
    """Call Anthropic messages API and return timing + token data."""
    import anthropic

    client = anthropic.Anthropic(api_key=os.environ["ANTHROPIC_API_KEY"])

    messages = [{"role": "user", "content": prompt["user"]}]
    kwargs = dict(
        model=api_model,
        messages=messages,
        max_tokens=4096,
        temperature=temperature,
    )
    if prompt.get("system"):
        kwargs["system"] = prompt["system"]

    start = time.monotonic()
    try:
        resp = client.messages.create(**kwargs)  # type: ignore[arg-type]
        elapsed_ms = (time.monotonic() - start) * 1000
        text = resp.content[0].text if resp.content else ""
        tokens_in = resp.usage.input_tokens if resp.usage else 0
        tokens_out = resp.usage.output_tokens if resp.usage else 0
        return {
            "success": True,
            "text": text,
            "tokens_in": tokens_in,
            "tokens_out": tokens_out,
            "total_tokens": tokens_in + tokens_out,
            "elapsed_ms": elapsed_ms,
            "error": None,
        }
    except Exception as exc:
        elapsed_ms = (time.monotonic() - start) * 1000
        return {
            "success": False,
            "text": "",
            "tokens_in": 0,
            "tokens_out": 0,
            "total_tokens": 0,
            "elapsed_ms": elapsed_ms,
            "error": str(exc),
        }


def call_gemini(api_model: str, prompt: dict, timeout: int = 60) -> dict:
    """Call Google Gemini API (google-genai SDK) and return timing + token data."""
    try:
        from google import genai
        from google.genai import types as genai_types
    except ImportError:
        return {
            "success": False,
            "text": "",
            "tokens_in": 0,
            "tokens_out": 0,
            "total_tokens": 0,
            "elapsed_ms": 0.0,
            "error": "google-genai not installed; run: pip install google-genai",
        }

    client = genai.Client(api_key=os.environ["GEMINI_API_KEY"])

    config = genai_types.GenerateContentConfig(
        max_output_tokens=8192,  # thinking models (2.5-flash, 2.5-pro) use tokens for reasoning
        temperature=0.3,
    )
    if prompt.get("system"):
        config.system_instruction = prompt["system"]

    start = time.monotonic()
    try:
        resp = client.models.generate_content(
            model=api_model,
            contents=prompt["user"],
            config=config,
        )
        elapsed_ms = (time.monotonic() - start) * 1000
        text = resp.text if resp.text else ""
        usage = resp.usage_metadata
        tokens_in = (usage.prompt_token_count or 0) if usage else 0
        tokens_out = (usage.candidates_token_count or 0) if usage else 0
        return {
            "success": True,
            "text": text,
            "tokens_in": tokens_in,
            "tokens_out": tokens_out,
            "total_tokens": tokens_in + tokens_out,
            "elapsed_ms": elapsed_ms,
            "error": None,
        }
    except Exception as exc:
        elapsed_ms = (time.monotonic() - start) * 1000
        return {
            "success": False,
            "text": "",
            "tokens_in": 0,
            "tokens_out": 0,
            "total_tokens": 0,
            "elapsed_ms": elapsed_ms,
            "error": str(exc),
        }


# ---------------------------------------------------------------------------
# Grade file helpers (mirrors Go performance_grading.go logic)
# ---------------------------------------------------------------------------

def sanitize_filename(s: str) -> str:
    for ch in r'/\:*?"<>|':
        s = s.replace(ch, "_")
    return s


def grade_filename(model_id: str, role_id: str, project_id: str) -> Path:
    name = (
        f"{sanitize_filename(model_id)}_"
        f"{sanitize_filename(role_id)}_"
        f"{sanitize_filename(project_id)}.json"
    )
    return GRADES_DIR / name


def calculate_grade(success_rate: float, retry_rate: float, error_rate: float) -> str:
    """
    Mirrors PerformanceGradeManager.recalculateGrade() thresholds from config.go:
      A: success >= 0.90, retry <= 0.05
      B: success >= 0.80, retry <= 0.10
      C: success >= 0.70, retry <= 0.20
      D: success >= 0.60, retry <= 0.30
      F: otherwise
    """
    if success_rate >= 0.90 and retry_rate <= 0.05:
        return "A"
    if success_rate >= 0.80 and retry_rate <= 0.10:
        return "B"
    if success_rate >= 0.70 and retry_rate <= 0.20:
        return "C"
    if success_rate >= 0.60 and retry_rate <= 0.30:
        return "D"
    return "F"


def calculate_confidence(total_attempts: int) -> float:
    """Simple confidence score: tapers toward 1.0 as attempts → 50."""
    if total_attempts == 0:
        return 0.0
    return min(1.0, total_attempts / 50.0)


def load_or_create_grade(model_id: str, role_id: str, project_id: str) -> dict:
    path = grade_filename(model_id, role_id, project_id)
    if path.exists():
        with open(path) as f:
            return json.load(f)

    now = datetime.datetime.now(datetime.timezone.utc).isoformat().replace("+00:00", "Z")
    return {
        "model_id": model_id,
        "role_id": role_id,
        "project_id": project_id,
        "total_attempts": 0,
        "successes": 0,
        "failures": 0,
        "retries": 0,
        "total_tokens_used": 0,
        "total_execution_time_ms": 0,
        "average_tokens": 0,
        "average_execution_time": 0.0,
        "error_rate": 0.0,
        "retry_rate": 0.0,
        "success_rate": 0.0,
        "escalation_count": 0,
        "downgrade_count": 0,
        "grade": "F",
        "confidence_score": 0.0,
        "last_updated": now,
        "sample_size": 0,
        "source": "benchmark",
    }


def update_grade(grade: dict, elapsed_ms: float, tokens: int, success: bool) -> dict:
    """Add one benchmark result into an existing grade dict and recalculate."""
    # Ensure benchmark-produced grades are always tagged so the GUI can filter them
    grade.setdefault("source", "benchmark")
    grade["total_attempts"] += 1
    grade["total_execution_time_ms"] += elapsed_ms
    grade["total_tokens_used"] += tokens

    if success:
        grade["successes"] += 1
    else:
        grade["failures"] += 1

    n = grade["total_attempts"]
    grade["sample_size"] = n
    grade["success_rate"] = grade["successes"] / n
    grade["error_rate"] = grade["failures"] / n
    grade["retry_rate"] = grade["retries"] / n
    grade["average_tokens"] = int(grade["total_tokens_used"] / n)
    grade["average_execution_time"] = grade["total_execution_time_ms"] / n
    grade["grade"] = calculate_grade(
        grade["success_rate"], grade["retry_rate"], grade["error_rate"]
    )
    grade["confidence_score"] = calculate_confidence(n)
    grade["last_updated"] = datetime.datetime.now(datetime.timezone.utc).isoformat().replace("+00:00", "Z")
    return grade


def save_grade(grade: dict) -> None:
    path = grade_filename(grade["model_id"], grade["role_id"], grade["project_id"])
    GRADES_DIR.mkdir(parents=True, exist_ok=True)
    with open(path, "w") as f:
        json.dump(grade, f, indent=2)
        f.write("\n")


# ---------------------------------------------------------------------------
# Main benchmark runner
# ---------------------------------------------------------------------------

MAX_RETRIES = 3
RETRY_BASE_DELAY = 5.0  # seconds; doubled on each retry


def call_with_retry(provider: str, api_model: str, prompt: dict, temperature: float = 0.3) -> dict:
    """Call the appropriate API with exponential-backoff retry on failure."""
    last_result = None
    for attempt in range(MAX_RETRIES):
        if provider == "openai":
            result = call_openai(api_model, prompt, temperature=temperature)
        elif provider == "gemini":
            result = call_gemini(api_model, prompt)
        else:
            result = call_anthropic(api_model, prompt, temperature=temperature)

        if result["success"]:
            return result

        last_result = result
        if attempt < MAX_RETRIES - 1:
            delay = RETRY_BASE_DELAY * (2 ** attempt)
            print(f"FAIL ({result['error']}) — retry {attempt + 1}/{MAX_RETRIES - 1} in {delay:.0f}s...",
                  end=" ", flush=True)
            time.sleep(delay)

    return last_result


def benchmark_combo(
    model_idx: int,
    total_models: int,
    model_id: str,
    provider: str,
    tier: str,
    api_model: str,
    role: str,
    runs: int,
    skip_completed: bool,
    reset: bool,
    openai_key: str,
    anthropic_key: str,
) -> tuple:
    """Run benchmark for one (model, role) combo. Thread-safe — each combo writes
    to a unique grade file so there is no cross-thread file contention."""
    lines = [f"\n[{model_idx}/{total_models}] {model_id} ({tier}) → role={role}"]

    prompts = ROLE_PROMPTS.get(role, ENGINEER_PROMPTS)
    evaluator = ROLE_EVALUATORS.get(role, evaluate_quality)

    grade = load_or_create_grade(model_id, role, PROJECT_ID)
    if reset:
        grade["total_attempts"] = 0
        grade["successes"] = 0
        grade["failures"] = 0

    if skip_completed and not reset and grade["total_attempts"] >= runs:
        lines.append(f"  → Already complete ({grade['total_attempts']} runs, "
                     f"grade={grade['grade']}), skipping")
        return (model_id, role, grade["grade"], grade["success_rate"],
                grade["average_execution_time"], "\n".join(lines))

    for run_num in range(1, runs + 1):
        prompt_idx = (run_num - 1) % len(prompts)
        prompt = prompts[prompt_idx]

        result = call_with_retry(provider, api_model, prompt)

        if not result["success"]:
            lines.append(f"  run {run_num}/{runs} prompt[{prompt_idx}]... ERROR: {result['error']}")
            success = False
            tokens = 0
        else:
            tokens = result["total_tokens"]
            # Fast pre-filter first; if it passes, apply the LLM judge for real quality
            if evaluator(result["text"], prompt_idx):
                success = judge_response(role, prompt, result["text"], openai_key, anthropic_key)
            else:
                success = False
            status = "✓" if success else "✗"
            lines.append(f"  run {run_num}/{runs} prompt[{prompt_idx}]... {status}"
                         f"  ({result['elapsed_ms']:.0f}ms, {tokens} tok)")

        grade = update_grade(grade, result["elapsed_ms"], tokens, success)

    save_grade(grade)
    lines.append(f"\n  → Grade: {grade['grade']}  "
                 f"success={grade['success_rate']:.0%}  "
                 f"avg_latency={grade['average_execution_time']:.0f}ms  "
                 f"avg_tokens={grade['average_tokens']}")
    return (model_id, role, grade["grade"], grade["success_rate"],
            grade["average_execution_time"], "\n".join(lines))


def run_benchmark(
    runs: int = DEFAULT_RUNS,
    roles: list = None,
    dry_run: bool = False,
    model_filter: str = None,
    reset: bool = False,
    skip_completed: bool = False,
    workers: int = 6,
) -> None:
    if roles is None:
        roles = DEFAULT_ROLES

    openai_key = os.environ.get("OPENAI_API_KEY", "")
    anthropic_key = os.environ.get("ANTHROPIC_API_KEY", "")
    gemini_key = os.environ.get("GEMINI_API_KEY", "")

    if not openai_key:
        print("⚠️  OPENAI_API_KEY not set – OpenAI models will be skipped")
    if not anthropic_key:
        print("⚠️  ANTHROPIC_API_KEY not set – Anthropic models will be skipped")
    if not gemini_key:
        print("⚠️  GEMINI_API_KEY not set – Gemini models will be skipped")

    total_models = len(MODELS)
    results_summary = []

    # Build list of (model, role) combos to run
    combos = []
    for model_idx, (model_id, provider, tier, api_model) in enumerate(MODELS, 1):
        if model_filter and model_filter.lower() not in model_id.lower():
            continue
        if provider == "openai" and not openai_key:
            continue
        if provider == "anthropic" and not anthropic_key:
            continue
        if provider == "gemini" and not gemini_key:
            continue
        for role in roles:
            if dry_run:
                prompts = ROLE_PROMPTS.get(role, ENGINEER_PROMPTS)
                print(f"[{model_idx}/{total_models}] {model_id} ({tier}) → role={role}"
                      f"  [DRY RUN] {runs} runs × {len(prompts)} prompts")
                continue
            combos.append((model_idx, total_models, model_id, provider, tier, api_model, role))

    if dry_run:
        return

    print(f"\nRunning {len(combos)} combos with {workers} parallel workers "
          f"(LLM judge enabled)...\n")

    with ThreadPoolExecutor(max_workers=workers) as executor:
        future_to_combo = {
            executor.submit(
                benchmark_combo,
                *combo,
                runs, skip_completed, reset, openai_key, anthropic_key,
            ): combo
            for combo in combos
        }
        for future in as_completed(future_to_combo):
            try:
                model_id, role, grade_letter, sr, lat, output = future.result()
                print(output, flush=True)
                results_summary.append((model_id, role, grade_letter, sr, lat))
            except Exception as exc:
                combo = future_to_combo[future]
                print(f"\nERROR in {combo[2]}/{combo[6]}: {exc}", flush=True)

    # Final summary table
    print("\n" + "=" * 70)
    print("BENCHMARK SUMMARY")
    print("=" * 70)
    if results_summary:
        col_w = [20, 16, 6, 10, 12]
        header = (
            f"{'Model':<{col_w[0]}} {'Role':<{col_w[1]}} {'Grade':>{col_w[2]}} "
            f"{'Success':>{col_w[3]}} {'Latency(ms)':>{col_w[4]}}"
        )
        print(header)
        print("-" * 70)
        for model_id, role, grade_letter, sr, lat in sorted(results_summary):
            print(
                f"{model_id:<{col_w[0]}} {role:<{col_w[1]}} {grade_letter:>{col_w[2]}} "
                f"{sr:>{col_w[3]}.0%} {lat:>{col_w[4]}.0f}"
            )
    else:
        print("No results (dry run or all models filtered).")


# ---------------------------------------------------------------------------
# CLI
# ---------------------------------------------------------------------------

def main() -> None:
    parser = argparse.ArgumentParser(
        description="Benchmark AI models across all agent roles."
    )
    parser.add_argument(
        "--runs", type=int, default=DEFAULT_RUNS,
        help=f"Number of benchmark runs per model per role (default: {DEFAULT_RUNS})",
    )
    parser.add_argument(
        "--role", type=str, default=None,
        help="Benchmark only this role (e.g. engineer, orchestrator, architect)",
    )
    parser.add_argument(
        "--roles", type=str, default=None,
        help="Comma-separated list of roles to benchmark",
    )
    parser.add_argument(
        "--model", type=str, default=None,
        help="Filter to a specific model (substring match)",
    )
    parser.add_argument(
        "--dry-run", action="store_true",
        help="Print what would be run without making API calls",
    )
    parser.add_argument(
        "--reset", action="store_true",
        help="Reset existing grade data before benchmarking",
    )
    parser.add_argument(
        "--skip-completed", action="store_true",
        help="Skip model/role combos that already have >= --runs samples",
    )
    parser.add_argument(
        "--workers", type=int, default=6,
        help="Number of parallel workers (default: 6)",
    )
    parser.add_argument(
        "--list-roles", action="store_true",
        help="List all available roles and exit",
    )
    parser.add_argument(
        "--grades-dir", type=str, default=None,
        help="Override the directory where grade JSON files are written "
             "(default: <project-root>/.claude/performance_grades)",
    )
    args = parser.parse_args()

    # Redirect grade storage if --grades-dir was provided by the caller
    # (e.g. the agent server passes its own .claude/performance_grades path so
    # benchmark results land in the same store the server reads from).
    if args.grades_dir:
        global GRADES_DIR
        GRADES_DIR = Path(args.grades_dir)

    if args.list_roles:
        print("Available roles:")
        for r in sorted(ROLE_PROMPTS.keys()):
            print(f"  {r}  ({len(ROLE_PROMPTS[r])} prompts)")
        return

    # Determine roles
    if args.roles:
        roles = [r.strip() for r in args.roles.split(",") if r.strip()]
    elif args.role:
        roles = [args.role]
    else:
        roles = DEFAULT_ROLES

    # Validate roles
    unknown = [r for r in roles if r not in ROLE_PROMPTS]
    if unknown:
        print(f"ERROR: Unknown role(s): {', '.join(unknown)}")
        print(f"Valid roles: {', '.join(sorted(ROLE_PROMPTS.keys()))}")
        sys.exit(1)

    run_benchmark(
        runs=args.runs,
        roles=roles,
        dry_run=args.dry_run,
        model_filter=args.model,
        reset=args.reset,
        skip_completed=args.skip_completed,
        workers=args.workers,
    )


if __name__ == "__main__":
    main()
