# CV Checker

**Stop sending generic resumes into the ATS black hole. Tailor every application with precision, speed, and full control over your data.**

## Overview

Landing interviews starts with speaking the exact language of the job description—without fabricating experience or handing your career history over to third-party SaaS platforms.

**CV Checker** is a fast, cross-platform desktop application built in **Go** and **Fyne v2** that turns your Markdown resume into an interview-ready, ATS-aligned version for every role you target:

- **See why you match—or don't:** Get an explainable ATS compatibility score that breaks down exact matches, partial matches, and missing requirements before you hit apply.
- **Tailor without hallucination:** Let AI optimize your CV for each job offer while built-in anti-invention validation keeps every edit grounded in your real experience.
- **Inspect every change like code:** Review Git-style line diffs across CV and prompt versions so you always know what changed and why.
- **Your AI, your rules:** Switch effortlessly between **Gemini**, **OpenAI**, **Groq**, or 100% offline **Ollama** models—and test safely anytime with the built-in **Mock** provider.
- **Private by design:** Your CVs, job applications, prompts, and audit logs stay on your machine in local Markdown files and SQLite, with API keys encrypted at rest via **AES-256-GCM**.

## Key Features

- **Markdown CV & Version Management:** Create, import, edit, duplicate, search, favorite, and version CVs stored locally as Markdown files.
- **Job Application Tracking:** Store job offers and track applications alongside their extracted requirements.
- **Explainable ATS Scoring:** Compare a CV against a job offer's requirements to identify matches, partial matches, and gaps with a transparent ATS score.
- **AI-Powered Agentic CV Optimization:** Run a Go-native agentic optimization loop (`ReflectionOptimizerAgent` + `CriticAgent`, up to 2 reflection iterations) with deterministic anti-invention validation, semantic fact-checking, and traceability to source CV evidence.
- **Gap Interviewer Agent:** Automatically detect missing or partial requirements and answer targeted clarifying questions (`GapInterviewerAgent`) to incorporate verified, unrecorded experience before tailoring.
- **Git-Style Diffs:** Inspect line-level differences between original and optimized CVs, as well as prompt revisions.
- **Multi-Provider AI Support:** Switch at runtime between **Gemini**, **OpenAI**, **Groq**, **Ollama** (local LLMs), or a built-in **Mock** provider for offline testing.
- **Prompt Versioning:** Manage, edit, and roll back system prompts used across the pipeline and agents (`JOB_ANALYSIS`, `CV_MATCHING`, `ATS_SCORING`, `CV_OPTIMIZATION`, `TRANSLATION`, `CRITIC_AUDIT`, and `GAP_INTERVIEW`).
- **Local Security & Audit Trail:** Encrypt stored AI provider API keys with AES-256-GCM and review a local audit log of actions.

## Architecture

The project follows a Clean / Hexagonal Architecture layout:

```text
cv-checker/
├── cmd/cvchecker/          # Application entrypoint and dependency wiring
├── internal/
│   ├── domain/             # Core entities, ATS scoring, anti-invention validator, ports
│   ├── app/                # Application services, pipeline, and native AI agents (agents/)
│   ├── infrastructure/     # SQLite persistence, Markdown store, AI providers, AES-GCM, Git diff
│   └── ui/                 # Fyne v2 desktop GUI, custom AppTheme, and views
└── docs/                   # Architecture, requirements, and test case documentation (Spanish)
```

## Tech Stack

- **Language:** Go (`1.27+`)
- **Desktop UI:** [Fyne v2](https://fyne.io/) (`fyne.io/fyne/v2`)
- **Database:** SQLite via `modernc.org/sqlite`
- **Security:** AES-256-GCM local key encryption

## Getting Started

### Prerequisites

- Go `1.27` or later
- C compiler and X11 / graphics development headers required by Fyne (on Linux, e.g. `libgl1-mesa-dev`, `xorg-dev`)

### Build & Run

```bash
# Build the binary to bin/cvchecker
make build

# Build and run the desktop application
make run
```

On startup, the application initializes local storage under `data/`:
- `data/cvchecker.db` — SQLite database
- `data/cvs/` — Markdown CV files
- `data/security/` — Local encryption key material

### Running Tests

```bash
make test
```

### Cleaning Build Artifacts

```bash
make clean
```

## AI Provider Configuration

Open the **Configuración IA** tab in the app to configure your preferred provider:

| Provider | Notes |
| --- | --- |
| **Gemini** | Requires a Gemini API key |
| **OpenAI** | Requires an OpenAI API key (custom base URL supported) |
| **Groq** | Requires a Groq API key (OpenAI-compatible endpoint) |
| **Ollama** | Connects to a local Ollama instance (defaults to `http://localhost:11434`) |
| **Mock** | Default fallback when no provider is configured |

## Documentation

Additional project documentation is available in `docs/` (written in Spanish):

- [Architecture](docs/architecture.md)
- [Requirements](docs/requirements.md)
- [Test Cases](docs/test_cases.md)
