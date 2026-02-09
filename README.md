# 💰 SplitBill - Smart Group Bill Splitting App

A full-stack mobile application for splitting bills among groups, with optimized debt settlement using the Min-Cash Flow algorithm.

## 🏗️ Architecture

- **Backend**: Go (Gin) + MongoDB + Redis
- **Frontend**: React Native CLI + TypeScript
- **Auth**: Firebase Auth (Phone OTP) with dev mode fallback
- **State**: Zustand
- **Algorithm**: Greedy Min-Cash Flow for optimal settlements

## 📁 Project Structure

```
├── split-bill-backend/          # Go API Server
│   ├── cmd/server/main.go       # Entry point
│   ├── internal/
│   │   ├── config/              # App configuration
│   │   ├── database/            # MongoDB & Redis connections
│   │   ├── handlers/            # HTTP handlers
│   │   ├── middleware/          # Auth, CORS middleware
│   │   ├── models/              # Data models & DTOs
│   │   ├── repository/          # Database access layer
│   │   ├── services/            # Business logic
│   │   └── utils/               # Helpers
│   ├── config.yaml              # Configuration file
│   ├── Dockerfile               # Multi-stage Docker build
│   └── go.mod                   # Go dependencies
│
├── split-bill-mobile/           # React Native App
│   ├── App.tsx                  # App entry point
│   ├── src/
│   │   ├── api/                 # Axios client & API services
│   │   ├── components/          # Reusable UI components
│   │   ├── navigation/          # React Navigation setup
│   │   ├── screens/             # App screens
│   │   │   ├── auth/            # Login screen
│   │   │   ├── home/            # Dashboard
│   │   │   ├── group/           # Group management
│   │   │   ├── bill/            # Bill creation & detail
│   │   │   ├── settlement/      # Balances & settlements
│   │   │   └── profile/         # User profile
│   │   ├── store/               # Zustand state stores
│   │   ├── theme/               # Colors, spacing, typography
│   │   └── types/               # TypeScript interfaces
│   └── package.json
│
├── docker-compose.yml           # Full stack Docker setup
└── plans/                       # Architecture documentation
```

## 🚀 Quick Start

### Prerequisites

- **Go** 1.21+ 
- **Node.js** 18+
- **Docker & Docker Compose**
- **React Native CLI** setup (Android SDK / Xcode)

### 1. Start Backend Services

```bash
# Start MongoDB, Redis, and API server
docker-compose up -d

# Or run backend locally:
cd split-bill-backend
cp config.yaml config.local.yaml  # Edit your config
go mod tidy
go run cmd/server/main.go
```

The API server runs at `http://localhost:8080`

### 2. Setup Mobile App

```bash
cd split-bill-mobile

# Install dependencies
npm install

# iOS (macOS only)
cd ios && pod install && cd ..
npx react-native run-ios

# Android
npx react-native run-android
```

### 3. API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/verify` | Verify Firebase token |
| GET | `/api/v1/auth/me` | Get current user |
| PUT | `/api/v1/auth/profile` | Update profile |
| POST | `/api/v1/groups` | Create group |
| GET | `/api/v1/groups` | List user's groups |
| GET | `/api/v1/groups/:id` | Get group detail |
| POST | `/api/v1/groups/:id/members` | Add member |
| POST | `/api/v1/groups/join` | Join by invite code |
| POST | `/api/v1/groups/:id/bills` | Create bill |
| GET | `/api/v1/groups/:id/bills` | List group bills |
| GET | `/api/v1/groups/:id/balances` | Get group balances |
| GET | `/api/v1/groups/:id/settlements` | Get optimal settlements |
| POST | `/api/v1/transactions` | Create transaction |
| PUT | `/api/v1/transactions/:id/confirm` | Confirm transaction |

### 4. Dev Mode

The backend supports a **dev mode** where Firebase Auth is bypassed. Set in `config.yaml`:

```yaml
firebase:
  credentials_file: ""  # Leave empty to enable dev mode
```

In dev mode, send any string as the Authorization token - it will be used as the user ID.

## ✨ Key Features

### Phase 1 (MVP) ✅
- 📱 Phone OTP authentication
- 👥 Group creation with invite codes
- 💸 Equal & by-item bill splitting
- ⚡ Optimized debt settlement (Min-Cash Flow algorithm)
- 💳 Banking app deeplinks (Momo, ZaloPay, VNPay)
- 📊 Balance tracking dashboard

### Phase 2 (Planned)
- 📷 OCR receipt scanning (Google Vision API)
- 📝 Bill history & analytics
- 🔔 Push notifications
- 🖼️ Receipt image attachments

### Phase 3 (Planned)
- 🤖 AI smart suggestions
- 💱 Multi-currency support
- 📈 Spending analytics & charts
- 🔗 Deep linking for group invites

## 🧮 Min-Cash Flow Algorithm

The app uses a **Greedy Min-Cash Flow** algorithm to minimize the number of transactions needed to settle all debts:

1. Calculate net balance for each person (total paid - total share)
2. Find the person with max credit and max debit
3. Settle the minimum of the two amounts
4. Repeat until all balances are zero

This reduces N*(N-1)/2 potential transactions to at most N-1 transactions.

## 🛠️ Tech Stack Details

| Component | Technology |
|-----------|-----------|
| Mobile App | React Native CLI + TypeScript |
| Navigation | React Navigation 6 |
| State Management | Zustand |
| HTTP Client | Axios |
| Backend Framework | Go + Gin |
| Database | MongoDB 7 |
| Cache | Redis 7 |
| Authentication | Firebase Auth |
| OCR (Phase 2) | Google Cloud Vision API |
| Containerization | Docker + Docker Compose |

## 📄 License

MIT
