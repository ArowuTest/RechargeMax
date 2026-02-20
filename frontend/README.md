# RechargeMax Frontend

Modern, responsive React frontend for the RechargeMax rewards platform, built with Vite, TypeScript, and TailwindCSS.

## 🚀 Features

- **OTP Authentication** - Secure phone-based login
- **Airtime & Data Recharge** - Support for MTN, Airtel, Glo, 9mobile
- **Spin Wheel System** - Tier-based daily spins with prizes
- **Prize Draws** - Subscription-based lottery system
- **Affiliate Program** - Referral tracking and commission management
- **Admin Panel** - Comprehensive management dashboard
- **Responsive Design** - Mobile-first UI with shadcn/ui components

## 🛠️ Tech Stack

- **Framework:** React 18 + Vite
- **Language:** TypeScript
- **Styling:** TailwindCSS + shadcn/ui
- **State Management:** TanStack Query (React Query)
- **Forms:** React Hook Form + Zod validation
- **Routing:** React Router v6
- **API Client:** Axios
- **Animations:** Framer Motion

## 📋 Prerequisites

- Node.js 18+ and npm
- Go backend API running (see backend README)

## 🚀 Quick Start

### Development

1. **Install dependencies:**
   ```bash
   npm install
   ```

2. **Configure environment:**
   ```bash
   cp .env.example .env
   ```
   
   Edit `.env` and set:
   ```
   VITE_API_BASE_URL=http://localhost:8080/api/v1
   VITE_API_TIMEOUT=30000
   ```

3. **Start development server:**
   ```bash
   npm run dev
   ```
   
   Frontend will be available at `http://localhost:5173`

### Production Build

```bash
npm run build
npm run preview
```

## 🐳 Docker Deployment

### Using Docker Compose (Recommended)

```bash
# Build and start
docker-compose up -d

# View logs
docker-compose logs -f

# Stop
docker-compose down
```

Frontend will be available at `http://localhost:3000`

### Using Docker directly

```bash
# Build image
docker build -t rechargemax-frontend .

# Run container
docker run -d \
  -p 3000:80 \
  -e VITE_API_BASE_URL=http://your-backend-api:8080/api/v1 \
  --name rechargemax-frontend \
  rechargemax-frontend
```

## 📁 Project Structure

```
src/
├── components/          # React components
│   ├── admin/          # Admin panel components
│   ├── affiliate/      # Affiliate dashboard
│   ├── auth/           # Authentication components
│   ├── dashboard/      # User dashboard
│   ├── draws/          # Prize draw components
│   ├── recharge/       # Recharge forms
│   ├── spin/           # Spin wheel
│   ├── subscription/   # Subscription management
│   └── ui/             # shadcn/ui components (49 components)
├── lib/                # Utilities and API client
│   ├── api-client.ts   # Go backend API client
│   ├── api.ts          # Legacy API functions
│   └── utils.ts        # Helper functions
├── hooks/              # Custom React hooks
├── contexts/           # React contexts
├── application/        # Business logic services
├── domain/             # Domain types and interfaces
├── infrastructure/     # Infrastructure layer
└── pages/              # Page components

Total: 105 files
```

## 🔌 API Integration

The frontend connects to the Go backend API. All API calls are handled through `src/lib/api-client.ts`:

### Authentication
```typescript
import { authApi } from '@/lib/api-client';

// Send OTP
await authApi.sendOTP(phoneNumber);

// Verify OTP and login
await authApi.verifyOTP(phoneNumber, otp);
```

### Recharge
```typescript
import { rechargeApi, paymentApi } from '@/lib/api-client';

// Get networks
const networks = await rechargeApi.getNetworks();

// Initialize payment
const payment = await paymentApi.initializePayment({
  amount: 1000,
  email: 'user@example.com',
  phone_number: '08012345678',
  transaction_type: 'airtime',
});
```

### Admin Operations
```typescript
import { adminApi } from '@/lib/api-client';

// Get dashboard stats
const stats = await adminApi.getStats();

// Manage users
const users = await adminApi.users.getAll();
await adminApi.users.update(userId, updates);

// Manage affiliates
await adminApi.affiliates.approve(affiliateId);
```

## 🎨 UI Components

Built with **shadcn/ui** - 49 pre-built, accessible components:

- Forms: Input, Select, Checkbox, Radio, Switch, Textarea
- Feedback: Alert, Toast, Dialog, Drawer, Progress
- Navigation: Tabs, Breadcrumb, Menu, Sidebar
- Data Display: Table, Card, Badge, Avatar, Chart
- And more...

## 🔐 Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `VITE_API_BASE_URL` | Go backend API URL | `http://localhost:8080/api/v1` |
| `VITE_API_TIMEOUT` | API request timeout (ms) | `30000` |

## 📦 Scripts

| Command | Description |
|---------|-------------|
| `npm run dev` | Start development server |
| `npm run build` | Build for production |
| `npm run preview` | Preview production build |
| `npm run lint` | Run ESLint |

## 🚀 Deployment

### Production Checklist

1. ✅ Set production `VITE_API_BASE_URL`
2. ✅ Build with `npm run build`
3. ✅ Test with `npm run preview`
4. ✅ Deploy `dist/` folder or use Docker
5. ✅ Configure nginx for SPA routing
6. ✅ Enable HTTPS
7. ✅ Set up monitoring

### Nginx Configuration

The included `nginx.conf` provides:
- SPA routing (all routes → index.html)
- Static asset caching
- Gzip compression
- Security headers
- Health check endpoint

### Docker Production

```bash
# Build for production
docker build -t rechargemax-frontend:latest .

# Run with environment variables
docker run -d \
  -p 80:80 \
  -e VITE_API_BASE_URL=https://api.rechargemax.com/api/v1 \
  --restart unless-stopped \
  rechargemax-frontend:latest
```

## 🔧 Development

### Adding New Components

```bash
# Add shadcn/ui component
npx shadcn-ui@latest add [component-name]
```

### Code Style

- TypeScript strict mode enabled
- ESLint for code quality
- Prettier for formatting (recommended)
- Follow existing component patterns

### API Client Updates

When backend adds new endpoints, update `src/lib/api-client.ts`:

```typescript
export const newFeatureApi = {
  getData: async () => {
    const response = await apiClient.get('/new-feature');
    return response.data;
  },
};
```

## 📊 Performance

- **Vite** for lightning-fast HMR
- **Code splitting** via React.lazy()
- **Image optimization** recommended
- **Bundle size** optimized with tree-shaking

## 🐛 Troubleshooting

### API Connection Issues

```bash
# Check backend is running
curl http://localhost:8080/health

# Check CORS configuration in backend
# Ensure frontend origin is allowed
```

### Build Errors

```bash
# Clear cache and reinstall
rm -rf node_modules package-lock.json
npm install

# Clear Vite cache
rm -rf .vite
```

### Docker Issues

```bash
# View container logs
docker logs rechargemax-frontend

# Rebuild without cache
docker build --no-cache -t rechargemax-frontend .
```

## 📝 License

Proprietary - RechargeMax Platform

## 🤝 Support

For issues or questions, contact the development team.

---

**Built with ❤️ by the RechargeMax Team**
