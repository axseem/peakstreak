# PeakStreak

PeakStreak is a full-stack, social habit-tracking application designed to help you build better habits with friends. It features a sleek Go backend, a reactive frontend, and a fully containerized, reproducible development environment using Nix and Docker.

![PeakStreak Screenshot](./docs/preview.png)

## Features

- **User Authentication**: Secure sign-up and login with JWT.
- **Habit Management**: Create, update, and delete habits. Supports both simple (yes/no) and numerical habits.
- **Dynamic Calendar View**: An intuitive and visually appealing blob-style calendar grid to track your progress over the years.
- **Social Features**: Follow/unfollow users to see their progress.
- **User Profiles**: Public user profiles with customizable avatars, habit displays, and follower/following counts.
- **Leaderboard**: See who is at the top of their game.
- **Explore Page**: Discover what habits other users are tracking.
- **User Search**: Find and connect with other users.
- **RESTful API**: A clean and well-defined API built with Go.

## Tech Stack

- **Backend**:
  - [Go](https://go.dev/)
- **Frontend**:
  - [TypeScript](https://www.typescriptlang.org/)
  - [Hyperapp](https://github.com/jorgebucaran/hyperapp) - A tiny, declarative, Elm inspired JavaScript framework
  - [Tailwind CSS](https://tailwindcss.com/) - Utility-first CSS framework
- **Database**:
  - [PostgreSQL](https://www.postgresql.org/)
- **DevOps & Tooling**:
  - [Nix](https://nixos.org/) with [Flakes](https://nixos.wiki/wiki/Flakes) - For a reproducible development environment
  - [Docker](https://www.docker.com/) & [Dockerfile](Dockerfile) - For containerization and deployment
  - [Makefile](Makefile) - For task automation

## Getting Started

There are three ways to get the project running: using Nix (recommended), Docker, or setting it up manually.

### Prerequisites

- For the recommended approach: [Nix](https://nixos.org/download.html) with Flakes enabled.
- For the containerized approach: [Docker](https://docs.docker.com/get-docker/) or [Podman](https://podman.io/).
- [Make](https://www.gnu.org/software/make/)

### 1. Using Nix (Recommended)

The `flake.nix` file defines a complete, reproducible development environment. This is the easiest and most reliable way to get started.

1.  **Enter the Development Shell:**
    Simply run the following command in the project root. It will automatically download all required dependencies and start a PostgreSQL instance for you.

    ```bash
    nix develop
    ```

    You will see a welcome message with the next steps.

2.  **Apply Database Migrations:**
    In the same shell, run the `migrate-up` command to set up the database schema.

    ```bash
    make migrate-up
    ```

3.  **Run the Development Servers:**
    This command will start both the Go backend and the Vite frontend server with hot-reloading.

    ```bash
    make dev
    ```

    - The backend API will be available at `http://localhost:8080`.
    - The frontend will be available at `http://localhost:5173`.

### 2. Using Docker

If you prefer using Docker, you can build and run the application in containers.

1.  **Start the Database:**
    The `Makefile` includes a command to start a PostgreSQL database using Podman or Docker.

    ```bash
    make run-db
    ```

2.  **Configure Environment Variables:**
    The application is configured using a `.env` file. Create one by copying the example:

    ```bash
    cp .env.example .env
    ```

    The default values should work with the `run-db` command.

3.  **Build the Docker Image:**
    This command builds the frontend and then creates the final production Docker image.

    ```bash
    make docker-build
    ```

4.  **Run the Application Container:**
    ```bash
    docker run -d -p 8080:8080 --name peakstreak-app -e DB_URL="postgres://user:password@host.docker.internal:5432/peakstreak?sslmode=disable" -e JWT_SECRET="your-super-secret-key" peakstreak:latest
    ```
    > **Note:** `host.docker.internal` is used to connect from the app container to the database running on the host. If your setup differs, you might need to adjust the `DB_URL`.


## Configuration

The application is configured via environment variables. Create a `.env` file in the project root.

```dotenv
# .env.example

# The connection string for your PostgreSQL database
DB_URL="postgres://user:password@localhost:5432/peakstreak?sslmode=disable"

# The port the Go server will run on
SERVER_PORT="8080"

# A secure secret for signing JWTs
JWT_SECRET="a-very-secure-secret-key-that-is-long-and-random"

# How long a JWT is valid for (e.g., 24h, 7d, 1h)
JWT_EXPIRES_IN="24h"
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
