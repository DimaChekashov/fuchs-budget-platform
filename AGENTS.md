# Fuchs Budget Platform

## About the Project

- **Purpose.** Fuchs Budget Platform helps people manage their personal budgets and track their income and expenses.
- **Target users.** Anyone who wants to keep track of their money and understand where it goes.
- **Core features.** Users can create an account, set up wallets in different currencies, add currencies, and record money coming into or going out of each wallet. The platform will provide spending statistics and exchange rates.

## Backend

- Use Go, PostgreSQL, and Redis in a microservices architecture.
- Document each service's API with OpenAPI (Swagger). Update the specification whenever the API contract changes.
- Use tokens for authentication and authorization.
- Add or update automated tests whenever backend behavior changes.
- Define monetary amount handling here once the representation and rounding rules are agreed.

### Database migrations

- Use versioned Goose SQL migrations in the owning service's `migrations/` directory.
- Make schema changes in a new migration. Keep migrations that have already been applied immutable.
- Use transactional migrations where PostgreSQL supports them. Provide a safe rollback when possible and document irreversible changes.
- For breaking schema changes, add the new schema first, migrate the data and update the application, then remove the old schema in a later migration.

## Frontend

- Use Next.js and npm. Organize the frontend according to Feature-Sliced Design (FSD).
- Support Russian and English and let users switch the interface language.
- Provide a home dashboard with wallets, sign-in and registration screens, an exchange rates screen, and screens for creating and editing wallets.
- Use token-based authentication and authorization for protected screens and API requests.
- Support three responsive layouts for mobile, tablet, and desktop.
- Cover key user flows with automated tests. Add or update tests whenever frontend behavior changes.
- Document all frontend setup, startup, and test commands in `README.md`.

## Working Agreements

- Write all Git commit messages in English.
- Document commands and procedures for local setup, running services and the frontend, applying migrations, and running checks in `README.md`. Update it when those procedures change.
- Dependencies — [when new libraries can be added]
- Compatibility — [which contracts and formats must be preserved]
- Verification — [which checks are required for each type of change]
