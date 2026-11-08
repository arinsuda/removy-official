# 🎬 POPCORN (Movie Hub Official)

**POPCORN** is a full-stack web application designed for movie and TV series enthusiasts to browse, track, rate, review, and socially interact around entertainment content. Powered by **Vue 3** on the frontend and **Go (Fiber v3)** on the backend, integrated with **The Movie Database (TMDB) API**, **PostgreSQL**, **MinIO S3 Object Storage**, **Brevo Transactional Email Service**, and **Socket.io WebSockets**.

---

## 🚀 Live Demo & Hosting

- **Frontend Deployment**: [https://removy-official.pages.dev/](https://removy-official.pages.dev/) *(Hosted on Cloudflare Pages)*
- **Search Engine & PWA**: Optimized with `sitemap.xml`, `robots.txt`, Open Graph social metadata, and Web App Manifest (PWA Add to Home Screen).

---

## ✨ Features Breakdown

### 🔐 1. Authentication & Security
- **Email & Password Authentication**: Secure registration and login with bcrypt password hashing.
- **Email Verification**: Transactional verification emails sent via **Brevo API**.
- **Google OAuth 2.0 Integration**: Single Sign-On (SSO) using Google accounts (`user_identities` table).
- **JWT Token Management**: Access Tokens (15 min TTL) & Refresh Tokens (7 days TTL) stored in secure HTTP-Only cookies with token rotation.
- **Password Reset**: Secure time-limited password reset tokens delivered via email.

### 🎬 2. Movie & TV Series Discovery (TMDB API)
- **Trending & Popular Media**: Live data fetched from TMDB for Movies, TV Series, Top Rated, and Upcoming releases.
- **Detailed Media Views**: Synopsis, backdrop/poster images, release dates, runtime, genres, trailers, cast & crew.
- **Actor & Crew Profiles**: Dedicated actor pages with full filmography and bio.
- **Advanced Search & Filtering**: Multi-field search by title, genre, year, and media type.
- **Favorite Genre Onboarding**: Interactive genre preference selection upon initial registration.

### ✍️ 3. Ratings & Reviews System
- **Star Ratings**: 0.5 to 5.0 rating scale with strict database check constraints (`chk_reviews_rating_step`).
- **Rich Text Reviews**: Write detailed reviews with optional **Spoiler Warning** toggles.
- **Media Attachments**: Upload custom image attachments to reviews (stored on MinIO).
- **Social Engagement**: Like reviews, mark reviews as helpful, and post threaded comments.

### 📚 4. Personal Library & Watch Logs
- **Custom Media Lists**: Track items as `Watching`, `Completed`, `Plan to Watch`, `Dropped`, `Favorites`, or `Watchlist`.
- **Episode & Movie Watch Logs**: Log watched episodes and rewatch counts with dates and personal notes.
- **Data Integrity**: Partial unique indexes in PostgreSQL (`idx_library_items_user_media_list`) to prevent duplicate list entries.

### 🏆 5. Gamification & Achievements
- **Unlockable Badges**: System triggers automatic achievement unlocks based on user actions (e.g., number of reviews written, movies watched, social interactions).
- **Achievement Showcase**: Display unlocked achievements on user profiles.

### 👥 6. Social Feed & User Profiles
- **Follow System**: Follow/Unfollow users with follower and following counters.
- **Activity Feed**: Dynamic timeline displaying real-time updates from followed users (reviews, list additions, achievements).
- **Fine-Grained Privacy Settings**: Profile & activity visibility controls (`default`, `public`, `followers`, `private`).
- **Best Movie of Life (BMOL)**: Featured showcase on user profile pages for their all-time favorite media.
- **User Statistics & Charts**: Visual breakdown of user rating distributions, genre preferences, and total watch time using **Chart.js**.

### 🔔 7. Real-Time WebSockets & Notifications
- **Socket.io WebSocket Engine**: Real-time notifications for review likes, comments, new followers, and achievement unlocks.
- **In-App Notification Center**: Unread notification counter, mark as read, and direct navigation.

### 🛡️ 8. Admin Workspace
- **Admin Dashboard** (`/admin`): Restricted workspace for administrators to manage users, roles (`admin` / `user`), review moderation, audit logs, and view system metrics.

---

## 🛠️ Tech Stack

### Frontend (`/client`)
| Category | Technology |
|---|---|
| Framework | **Vue 3** (Composition API, `<script setup>`) |
| Language | **TypeScript** (Strict mode) |
| Build Tool | **Vite 8** |
| Styling | **Tailwind CSS 4** + primeicons (icon font) |
| State Management | **Pinia 3** |
| Routing | **Vue Router 5** |
| Data Fetching | **TanStack Vue Query 5** + **Axios** |
| Internationalization | **Vue I18n 10** (Thai & English) |
| Real-time Client | **Socket.io Client** |
| Data Visualization | **Chart.js** + **vue-chartjs** |
| Icons & Animation | **Lucide Vue Next** + **GSAP** |

### Backend (`/server`)
| Category | Technology |
|---|---|
| Language | **Go 1.26** |
| Web Framework | **Fiber v3** |
| ORM / Database | **GORM v1.31** + **PostgreSQL 16** |
| Object Storage | **MinIO S3** (Private Bucket for user avatars, banners, review media) |
| Email Service | **Brevo REST API** (Transactional Email Client) |
| External API | **TMDB API** (The Movie Database v3 API) |
| Real-time WebSockets | **Socket.io Go Server** (`github.com/zishang520/socket.io`) |
| Authentication | **JWT** (`golang-jwt/jwt/v5`) + **bcrypt** |

---

## 📁 Project Structure

```
movie-hub-official/
├── client/                     # Vue 3 + TypeScript Frontend Application
│   ├── public/                 # Static assets, sitemap.xml, robots.txt, manifest.json
│   ├── src/
│   │   ├── api/                # Axios & Query API handlers
│   │   ├── components/         # Reusable UI components
│   │   ├── composables/        # Vue composables
│   │   ├── i18n/               # Internationalization locale configs
│   │   ├── layouts/            # MainLayout, AdminLayout, AuthLayout
│   │   ├── stores/             # Pinia state management stores
│   │   ├── views/              # View pages (Home, Movies, TV, Admin, User, etc.)
│   │   ├── App.vue
│   │   └── main.ts
│   ├── index.html              # HTML entrypoint with SEO & OG Tags
│   ├── vite.config.ts
│   └── package.json
│
├── server/                     # Go (Fiber v3) Backend Application
│   ├── config/                 # Environment configuration loader
│   ├── database/               # DB Connection, GORM AutoMigrate & SQL Migrations
│   │   ├── migrate/            # Custom SQL migration scripts
│   │   └── seeder/             # Initial JSON seed files (e.g., achievements.json)
│   ├── internal/               # Core domain modules
│   │   ├── achievements_module/
│   │   ├── admin_module/
│   │   ├── auth_module/
│   │   ├── feed_module/
│   │   ├── follow_module/
│   │   ├── library_module/
│   │   ├── mailer/             # Brevo email sender service
│   │   ├── movie_module/
│   │   ├── notification_module/# Socket.io hub & notification logger
│   │   ├── review_module/
│   │   ├── tmdb_module/        # TMDB API client wrapper
│   │   ├── user_module/
│   │   └── watch_log_module/
│   ├── middleware/             # JWT auth & Admin RBAC middlewares
│   ├── router/                 # API route registrations & Fiber setup
│   ├── main.go                 # Entrypoint
│   └── go.mod
│
├── docker-compose.yml          # Infrastructure services (PostgreSQL 16 & MinIO)
└── README.md                   # Project documentation
```

---

## ⚙️ Environment Variables Setup

Create `server/.env` from `server/.env.example`. The server reads real environment variables first, then `server/.env`, then a repo-root `.env`.

### Server `.env` (`server/.env`)
```env
PORT=8080
APP_BASE_URL=http://localhost:8080
# Comma-separated exact origins, no wildcard
CORS_ALLOWED_ORIGIN=http://localhost:3000,http://localhost:5173

# PostgreSQL Database
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=your_postgres_password
POSTGRES_DB=moviehub
POSTGRES_SSLMODE=disable

# JWT Secrets
JWT_ACCESS_SECRET=your_jwt_access_secret_key
JWT_REFRESH_SECRET=your_jwt_refresh_secret_key

# TMDB API (Get your key from https://www.themoviedb.org/settings/api)
THE_MOVIE_BASE_API=https://api.themoviedb.org/3
THE_MOVIE_API_KEY=your_tmdb_api_key

# Brevo Email API (https://brevo.com - Free 300 emails/day)
BREVO_API_KEY=your_brevo_api_key
BREVO_SENDER_EMAIL=removy.official@gmail.com
BREVO_SENDER_NAME=POPCORN

# MinIO Object Storage
MINIO_ENDPOINT=localhost:9000
MINIO_ROOT_USER=minioadmin
MINIO_ROOT_PASSWORD=minioadmin
MINIO_BUCKET_NAME=removy-private
MINIO_USE_SSL=false

# Google OAuth 2.0 (Optional)
GOOGLE_OAUTH_ENABLED=false
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
GOOGLE_REDIRECT_URL=http://localhost:8080/api/auth/google/callback
GOOGLE_FRONTEND_SUCCESS_URL=http://localhost:5173/
GOOGLE_FRONTEND_ERROR_URL=http://localhost:5173/login
```

### Client `.env` (`client/.env`)
```env
VITE_API_BASE_URL=http://localhost:8080
VITE_SOCKET_URL=http://localhost:8080
```

---

## 📦 Getting Started & Local Development

### Prerequisites
- [Docker & Docker Compose](https://www.docker.com/)
- [Node.js](https://nodejs.org/) (v20.19.0+ or v22.12.0+)
- [Go](https://go.dev/) (v1.21+)

---

### Step-by-Step Run Instructions

#### 1. Clone the repository
```bash
git clone https://github.com/arinsuda/removy-official.git
cd movie-hub-official
```

#### 2. Start Infrastructure Services (PostgreSQL & MinIO)
```bash
docker compose --env-file server/.env up -d
```
Services started:
- **PostgreSQL**: `localhost:5432`
- **MinIO API**: `localhost:9000`
- **MinIO Web Console**: `localhost:9001` (login with `MINIO_ROOT_USER` & `MINIO_ROOT_PASSWORD`)

#### 3. Run Backend Server (Go)
```bash
cd server
go mod tidy
go run main.go
```
*The backend API will run on `http://localhost:8080` and auto-run database migrations and seeders.*

#### 4. Run Frontend App (Vue 3)
```bash
cd client
npm install
npm run dev
```
*Open `http://localhost:3000` in your browser.*

---

## 🧪 Testing & Code Quality

### Frontend
```bash
cd client
npm run lint         # Run Oxlint & ESLint checks
npm run type-check   # Run Vue TypeScript compiler checks
npm run test         # Run Vitest test suite
```

### Backend
```bash
cd server
go test ./...        # Run unit & integration tests
```

---

## 📜 License

This project is open-source under the [MIT License](LICENSE).
