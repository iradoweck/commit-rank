# Changelog

All notable changes and milestones in the **Most Active GitHub Users (Commit Rank)** project are documented here by date.

---

### 2026-11-10 — Version 3.4.0

- **UI & Developer Profile Experience**
  - **Dedicated Developer Profile View (`#user/:login`)**: High-fidelity personal profile page featuring:
    - Dynamic national ranking badge with country flag (`#X in Country`).
    - Continental ranking calculation comparing developers within the same regional group (`#X in Continent`).
    - Worldwide ranking position and global activity tier (`Worldwide Top 100` / `Global Scope Top`).
    - Comprehensive breakdown of Git commits, pull requests, issues, and combined impact scores.
    - Official GitHub profile link and company affiliation.
  - **Triple Embeddable Profile Badges**:
    - **Country Badge** (`[Country] - Top #X`).
    - **Continent Badge** (`[Continent] - Top #X`).
    - **Worldwide Global Badge** (`Worldwide - Top #X`).
    - Integrated 1-click Markdown copy buttons ready for GitHub Profile `README.md`.
  - **Side-by-Side Top Ranking Cut Filter**:
    - Inline filter selector on country leaderboards: `Top 5`, `Top 10`, `Top 50`, `Top 100`, and `All`.
    - Positioned side-by-side with metric selectors (`All`, `Commits`, `Contribs`) without layout breaking or wrapping.
    - Dynamically resizes tables and pagination while strictly preserving each developer's authentic placement.
  - **Live Global Search on Homepage**:
    - Real-time instant search bar supporting countries and developers simultaneously.
    - Pre-cached multi-country developer index with floating autocomplete dropdown displaying avatars, country flags, ranks, and total scores.
  - **Community Call to Action Banner**:
    - Sleek modern footer banner on the homepage inviting community contributions with direct links to *"Star on GitHub"* and *"Suggest Country"*.

---

### 2026-11-09 — Version 3.3.0

- **Special Countries & Territories Expansion**
  - Activated 5 special autonomous regions and territories in the directory: **Hong Kong** (`HK`), **Macau** (`MO`), **Taiwan** (`TW`), **Kosovo** (`XK`), and **Kurdistan** (`KRD`).
  - Added dedicated placeholder datasets in strict UTF-8 without BOM.
  - Sanitized query presets in `presets.go`, removing non-ASCII and CJK characters to ensure 100% Go runtime encoding stability.
- **Crawler & Workflow Batch 5 Scheduling**
  - Upgraded GitHub Actions daily workflow rotation from a 5-day to a 6-day cycle (`DAY_MOD % 6`).
  - Allocated **Batch 5** exclusively to the 5 special territories for fast, lightweight automated updates.
- **US Query Preset Optimization**
  - Trimmed redundant municipal search terms for the United States preset to reduce query payload size and eliminate HTTP 502/504 gateway timeouts on GitHub's Elasticsearch backend.

---

### 2026-11-06 — Version 3.2.0

- **Crawler Engine Resilience & Deep Rate-Limit Defense**
  - **Secondary Rate-Limit Handling**: Added automatic 65-second backoff with temporal jitter and `Retry-After` header parsing in `net/net.go`.
  - **Partial Edge Salvaging**: Engineered resilient parsing in `github/github.go` to salvage valid user edges when GraphQL returns partial field errors (`RESOURCE_LIMITS_EXCEEDED`).
  - **Toxic Cursor Bypass**: Implemented automatic cursor skipping (`followers:<N`) to bypass corrupted bot accounts and infinite query loops.
  - **Go 1.24 Runtime Modernization**: Upgraded Go compiler to 1.24 across `go.mod`, Dockerfiles, and GitHub Actions workflows.
- **Americas Regional Presets Expansion**
  - Added 10 countries across the Americas: **United States**, **Canada**, **Mexico**, **The Bahamas**, **Panama**, **Brazil**, **Argentina**, **Colombia**, **Venezuela**, and **Uruguay**.
  - Intelligent 5-day rotation batches implemented in `.github/workflows/update-rankings.yml`.

---

### 2026-11-04 — Version 3.1.0

- **UI & Mobile-First Experience**
  - **Animated Bidirectional Git & GitHub Tree**: Introduced an interactive graph below the main title showing continuous commit flow and branching pipelines between Git and GitHub.
  - **Live Repository Stars Badge**: Real-time GitHub stars counter integrated into the header alongside pull requests and issues telemetry.
  - **Mobile-First Layout**: Adaptive responsive paddings, touch-friendly scrollable action toolbars, and visual swipe guidance for rankings tables on mobile viewports.
  - **Brand Assets**: Added official `favicon.svg` and binary `favicon.ico` to eliminate console 404 errors.
  - **Clean Presentation**: Removed external floating Ko-fi script, keeping donations strictly modal-driven and unintrusive.

- **Security & Code Sanitization**
  - **XSS Defense**: Implemented strict HTML entity escaping (`escapeHtml()`) and URL encoding (`encodeURIComponent()`) across all dynamic fields (names, logins, company, and organization handles).
  - **Audit Alignment**: Full compliance with `/security-audit`, `/code-reviewer`, and `/backend-security-coder` guidelines.

- **Crawler Engine & Geographical Presets**
  - **Mozambique Complete Coverage**: Expanded preset to cover all 11 provinces, districts, major capitals, and colloquial aliases (`moz`, `mz`, `moç`).
  - **Africa Regional Presets**: Added comprehensive provincial and tech-hub mappings for **Angola**, **Nigeria**, **South Africa**, and **Kenya**.
  - **Worldwide & Africa Automated Daily Pipeline**: Configured `.github/workflows/update-rankings.yml` to automatically extract and publish daily datasets for `worldwide.json`, `mozambique.json`, `angola.json`, `nigeria.json`, `south-africa.json`, and `kenya.json`.

---

### 2026-11-03 — Version 3.0.1

- **Interactive Metric Sorting & Dynamic Re-Ranking**
  - Added metric filter selector: `All (Total)`, `Commits`, and `Contributions`.
  - Column headers in the developers table are now interactive and clickable with indicator carets (`▼`).
  - Implemented dynamic rank recalculation: ranks (`#1, #2, #3...`) dynamically re-order based on the selected metric.
  - Multi-page pagination (100 users per page up to Top 300) fully supports and preserves active sort order.

- **Organization Enhancements**
  - Expanded organization extraction capacity to Top 50 organizations.
  - Organization ranking table now fetches and displays real organization logos directly from GitHub with direct profile links.

- **Crawler & Query Throughput Optimization**
  - Increased GraphQL query page size to 15 users and reduced sub-node organizations to 10, eliminating `RESOURCE_LIMITS_EXCEEDED` errors and cutting crawler execution time down by over 60%.
  - Fixed natural pagination termination in Go CLI to avoid fake retries on completed searches.

- **Repository Hygiene & Legacy Code Cleanup**
  - Completely removed archived `museum-code/` folder to eliminate duplicate Go manifests, Dockerfiles, and stale workflows.
  - Historical attributions and memorial links for [@lauripiispanen](https://github.com/lauripiispanen) and [@ashkulz](https://github.com/ashkulz) are permanently preserved in the UI and README.
  - Cleaned `.gitignore` and `.geminiignore` to optimize build and token footprint.

---

### 2026-11-01 — Version 3.0.0

- **Dedicated Documentation Page (`#documentation`)**
  - Built a comprehensive integration guide covering profile and repository setup.
  - Added live dynamic Shields.io badge specifications with automatic unranked fallback.
  - Published the auditable GitHub GraphQL indexing algorithm and filter criteria.

- **Open Source Technologies Showcase**
  - Redesigned the technologies modal into a clean, responsive 2-column layout.
  - Highlighted full underlying stack: Go, GitHub GraphQL API, Tailwind CSS, Python, Inter, JetBrains Mono, Shields.io, GitHub Actions, CodeQL, and Ko-fi SDK.

- **Community Support & Infrastructure Goal**
  - Integrated dedicated support modal outlining the custom domain and telemetry server infrastructure goal.
  - Configured seamless Ko-fi checkout flow allowing direct contributions without leaving the page.
  - Added maintainer profile badge with avatar and GitHub Sponsors alternative.

- **Header Live Metrics & Navigation**
  - Added live GitHub PR and Issue counters fetched dynamically from the official GitHub API.
  - Introduced clean top navigation with direct access to About, Documentation, and Repository source.

---

### 2026-10-31

- **Country Directory & Drill-Down Views**
  - Built high-speed client-side search across worldwide countries with an A-Z quick jump bar.
  - Implemented dual ranking views: **Developers Ranking** (contributions, commits, PRs, company) and **Organizations Ranking** (associated developers, impact ratio).
  - Added regional collapsible accordion with localized dynamic badges, pseudocode extraction logic, and single-click JSON datasets.

- **DevSecOps, CI/CD & Repository Security**
  - Configured GitHub Actions CI pipeline running Go build suites on `main` and `beta`.
  - Implemented GitHub CodeQL static analysis workflows for continuous vulnerability detection.
  - Configured Dependabot automated versioning and security updates.
  - Published comprehensive `SECURITY.md` vulnerability reporting guidelines.
  - Configured `.geminiignore` adhering to FinOps principles to optimize AI token context and prevent payload waste.
