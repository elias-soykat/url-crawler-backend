# Test task

## Introduction

This test task is for the Full-Stack Developer position with a front-end focus.
While your primary responsibility will be front-end development, you will also integrate with back-end services.

This exercise has two goals:

- **For us:** To evaluate your production-quality code, API design reasoning, and overall development approach.
- **For you:** To experience the day-to-day work at Sykell through a realistic project.

## Code and project ownership

This is purely an exercise; we have no intention of using your submission in production.
You retain full ownership of your work and may license it as you see fit.
If you choose not to move forward with us, you’re free to open-source it or include it in your portfolio.

## Areas of focus

We will evaluate your submission on:

- **Front-end**

  - UI must be a React Web application
  - Responsive design (desktop and mobile)
  - TypeScript

- **Back-end**

  - Go (Golang)
  - Framework of your choice (e.g., Gin, Echo, etc.)
  - MySQL for data storage

- **General**

  - Reproducible builds (use your preferred vendoring/packaging system)
  - Consistent error handling and reporting (no crashes under non-critical conditions)
  - API request to backend must use authorization mechanism
  - Show a chain of commits representing project progression, including initial scaffolding, incremental feature additions, and any refactoring steps
  - Automated front-end tests covering few happy-path scenarios
  - Total implementation time should not exceed 8 hours (feel free to spread this across multiple days)

> **Note:** As this role is front-end focused, the quality and completeness of your React/TypeScript code will carry greater weight.

# Task Description

Build a web application that accepts a website URL as input, crawls it and displays key information about that page.

**Back-end data collection (per URL):**

- HTML version
- Page title
- Count of heading tags by level (H1, H2, etc.)
- Number of internal vs. external links
- Number of inaccessible links (those returning 4xx or 5xx status codes)
- Presence of a login form

**Front-end requirements (high → low priority):**

1. **URL Management**

   - Add URLs for analysis
   - Start/stop processing on selected URLs

2. **Results Dashboard**

   - Paginated, sortable table (columns: Title, HTML Version, #Internal Links, etc.)
   - Column filters and a global search box (fuzzy or prefix matching)

3. **Details View**

   - Click a row to open a detail page showing:

     - Bar or donut chart of internal vs. external links
     - List of broken links with status codes

4. **Bulk Actions**

   - Re-run analysis or delete selected URLs via checkboxes

5. **Real-Time Progress**

   - Display crawl status for each URL (queued → running → done / error)

# How to submit test task

1. Create a **public GitHub repository** and commit your code there.
2. Include a **README.md** with clear instructions how to build and run your application.
3. Email the repository link to **[careers@sykell.com](mailto:careers@sykell.com)**.

# Evaluation criteria (scorecard)

| **Criterion**                      | **Description**                                                                    | **Weight** | **Score (1–5)** | **Weighted Score** | **Comments** |
| ---------------------------------- | ---------------------------------------------------------------------------------- | ---------: | --------------: | -----------------: | ------------ |
| **1. Front-End Implementation**    |                                                                                    |        25% |                 |                    |              |
| • React & TypeScript Quality       | Clarity, modularity, use of hooks, TS typing discipline                            |        10% |                 |                    |              |
| • Responsiveness & UX              | Works on mobile & desktop; polished, accessible UI                                 |         8% |                 |                    |              |
| • Testing (happy paths)            | Coverage of core flows; meaningful assertions                                      |         7% |                 |                    |              |
| **2. Back-End Implementation**     |                                                                                    |        20% |                 |                    |              |
| • Go Code Quality                  | Readability, error handling, idiomatic use of Go                                   |         8% |                 |                    |              |
| • MySQL Schema & Data Access       | Logical schema design; safe queries; migrations or seed scripts                    |         6% |                 |                    |              |
| • Reproducible Builds              | Vendoring/packaging consistency (e.g. modules, Dockerfile)                         |         6% |                 |                    |              |
| **3. Core Features**               |                                                                                    |        25% |                 |                    |              |
| • URL Ingestion & Crawl Control    | Add URLs, start/stop processing                                                    |         5% |                 |                    |              |
| • Data Collection Accuracy         | Correct detection of HTML version, headings, link counts, login form, broken links |        10% |                 |                    |              |
| • Authorized API Calls             | Proper auth mechanism in requests                                                  |         5% |                 |                    |              |
| • Real-Time Status Updates         | Working status indicators (polling or push)                                        |         5% |                 |                    |              |
| **4. Dashboard & Details View**    |                                                                                    |        20% |                 |                    |              |
| • Table: Sort, Pagination, Filters | Stable, fast, intuitive grid with sorting, paging, column filters, global search   |         8% |                 |                    |              |
| • Detail Page Charts & Reports     | Clear, accurate bar/donut chart and broken-links list                              |         7% |                 |                    |              |
| • Bulk Actions                     | Re-run and delete selections via checkboxes                                        |         5% |                 |                    |              |
| **5. Documentation & Setup**       |                                                                                    |        10% |                 |                    |              |
| • README & Run Instructions        | Completeness and clarity of setup steps                                            |         5% |                 |                    |              |
| • Code Organization & Comments     | Logical project structure, helpful inline comments, small isolated commits         |         5% |                 |                    |              |
| **TOTAL**                          |                                                                                    |       100% |                 |                    |              |

based on above task description develop the backend project, here is the backend define prompt and requirements

You are a senior Go backend developer. Build a RESTful API using Go (Golang) and the Gin framework that powers a website crawler.

The application will:

- Accept a website URL for analysis
- Crawl the page and collect:
  - HTML version
  - Page title
  - Count of heading tags (H1–H6)
  - Number of internal and external links
  - Number of broken links (4xx/5xx)
  - Presence of login form (input type="password")
- Store results in MySQL
- Return data to frontend in paginated, sortable API
- Support these endpoints:
  - `POST /urls` — submit new URL for analysis
  - `GET /urls` — list analyzed URLs (paginated, filterable)
  - `GET /urls/:id` — get detailed result of single URL
  - `POST /urls/bulk` — bulk re-run or delete URLs
  - `POST /auth/signup` — sign up user
  - `POST /auth/login` — return a JWT token

- Use middleware to require JWT auth on all endpoints except `/auth/login`
- Use GORM or SQLx to interact with MySQL
- Add a crawler that runs in background (async) using Go routines or a job queue
- Track crawling status: `queued`, `running`, `done`, `error`

Requirements:

- Go module-based project
- Gin framework
- MySQL
- Proper error handling
- Dockerfile + docker-compose.yml for local setup
- Seed script for MySQL
