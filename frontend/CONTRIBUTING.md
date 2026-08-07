# Contributing to GAVIS

Thank you for your interest in improving **GAVIS** (GSI AI Vision for Intelligent Surveillance). This guide will help you set up your environment and understand how to contribute effectively to our internal initiative.

---

## Overview

GAVIS is built with **Next.js**, **TypeScript**, **Tailwind CSS v4**, and **Shadcn UI**. The goal is to keep the codebase modular, scalable, and easy to extend as we add more AI models and dashboard features.

---

## Project Layout

We use a **colocation-based file system**. Each feature keeps its own pages, components, and logic.

```
src
├── app               # Next.js routes (App Router)
│   ├── (main)        # Main routing group
│   │   ├── auth      # Authentication layouts & screens
│   │   └── dashboard # Dashboard screens (CCTV, Analytics, About)
│   └── layout.tsx    # Root layout
├── components        # Shared UI components
├── hooks             # Reusable hooks
├── lib               # Config & utilities
├── server            # Server actions & utilities
└── styles            # Tailwind / theme setup
```

---

## Contribution Flow

- Always create a new branch before working on changes:
  ```bash
  git checkout -b feature/my-update
  ```

- Use clear commit messages:
  ```bash
  git commit -m "feat: integrate new person_detect_batch event"
  ```

- If your change adds a new UI screen or component, include a screenshot in your Pull Request description.

---

## Guidelines

- Prefer **TypeScript types** over `any`.
- **Husky pre-commit hooks** are enabled - linting (via Biome) runs automatically when you commit. Ensure your code passes `npx @biomejs/biome check --write` before committing.
- Follow the established **Shadcn UI** style & **Tailwind v4** conventions.
- Make sure to test WebSocket configurations and CCTV layers properly when adjusting the `LiveCctvCard` or any `canvas` overlays.
- Keep accessibility in mind (e.g., always provide `type="button"` for interactive buttons).
- Maintain responsive grid proportions using standard Tailwind breakpoints (`md:`, `lg:`, `xl:`).

---

## Submitting PRs

- Open a Pull Request on the internal Git repository once your changes are ready.
- Ensure your branch is up to date with `main` before submitting.
- Reference any related internal issue tracker or ticketing system ID in your PR for context.

---

Your contributions help strengthen our independent surveillance capabilities! 🚀
