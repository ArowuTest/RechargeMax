# RechargeMax Mobile App

**Nigeria's #1 Gamified Recharge Platform — Mobile App**

Built with Flutter 3.x | Riverpod 2 | GoRouter 13 | Dio 5 | Plus Jakarta Sans

---

## 🎨 Design System

The app faithfully mirrors the live website at [rechargemax-frontend.onrender.com](https://rechargemax-frontend.onrender.com):

| Token | Value | Usage |
|-------|-------|-------|
| `brand500` | `#7C3AED` | Primary CTAs, icons |
| `brand950` | `#1A0533` | Hero backgrounds |
| `gold500` | `#F59E0B` | Prize pool, wins |
| `success500` | `#10B981` | Success states |
| Font | Plus Jakarta Sans | All text |
| Border radius | 12–20px | Cards, buttons |

Dark mode is the **primary mode** (matches the website hero). Light mode is clean white for content sections.

---

## 🏗️ Architecture

```
lib/
├── main.dart                        # App entry point
├── core/
│   ├── api/api_client.dart          # Dio + all API calls
│   ├── auth/auth_provider.dart      # JWT auth state (Riverpod)
│   ├── router/app_router.dart       # GoRouter + auth guards
│   └── theme/
│       ├── app_colors.dart          # All design tokens
│       ├── app_text_styles.dart     # Typography scale
│       └── app_theme.dart           # Material 3 theme
├── features/
│   ├── auth/                        # Splash, Onboarding, Phone, OTP, Profile Setup
│   ├── home/                        # Dashboard, quick actions, draw countdown
│   ├── recharge/                    # Airtime/data purchase, Paystack payment
│   ├── spin/                        # Custom wheel painter, prize reveal
│   ├── draws/                       # Active draws, my entries, winners
│   ├── subscription/                # ₦20/day subscription management
│   ├── affiliate/                   # Referral program, earnings, payout
│   └── profile/                    # Account, tier progress, history, logout
└── shared/
    ├── widgets/
    │   ├── main_scaffold.dart       # Bottom nav shell
    │   └── app_widgets.dart        # Reusable components
    └── constants/app_constants.dart # Business rules & API config
```

---

## 🚀 Getting Started

### Prerequisites

```bash
flutter --version   # Flutter 3.19+ required
dart --version      # Dart 3.3+ required
```

### 1. Clone and Install

```bash
cd rechargemax_app
flutter pub get
```

### 2. Download Fonts

Download **Plus Jakarta Sans** from [Google Fonts](https://fonts.google.com/specimen/Plus+Jakarta+Sans) and place the `.ttf` files in `assets/fonts/`:

```
assets/fonts/
├── PlusJakartaSans-Regular.ttf
├── PlusJakartaSans-Medium.ttf
├── PlusJakartaSans-SemiBold.ttf
├── PlusJakartaSans-Bold.ttf
└── PlusJakartaSans-ExtraBold.ttf
```

### 3. Configure API URL

In `lib/shared/constants/app_constants.dart`:

```dart
static const String baseUrl = 'https://rechargemax-api.onrender.com/api/v1';
```

Change to your actual backend URL if different.

### 4. Run

```bash
# Debug
flutter run

# Release APK
flutter build apk --release

# App Bundle for Play Store
flutter build appbundle --release
```

---

## 📱 Screen Map

| Screen | Route | Description |
|--------|-------|-------------|
| Splash | `/splash` | Logo animation, auth check |
| Onboarding | `/onboarding` | 3-slide intro (first launch only) |
| Phone Entry | `/login` | Nigerian phone number input |
| OTP Verify | `/login/otp` | 6-digit PIN input with countdown |
| Profile Setup | `/profile-setup` | Name, email, referral code |
| **Home** | `/home` | Dashboard, points, draw banner, winners |
| **Recharge** | `/recharge` | Airtime/data purchase |
| Recharge Success | `/recharge/success` | Confetti + rewards summary |
| **Spin** | `/spin` | Animated spin wheel |
| **Draws** | `/draws` | Active draws, entries, winners |
| Subscription | `/profile/subscription` | ₦20/day subscribe |
| Affiliate | `/profile/affiliate` | Referral link, earnings, payout |
| **Profile** | `/profile` | Account info, tier, logout |
| History | `/profile/history` | Paginated recharge history |

---

## 🔌 Backend API Contract

All endpoints hit `https://rechargemax-api.onrender.com/api/v1`:

| Method | Path | Auth |
|--------|------|------|
| POST | `/auth/send-otp` | No |
| POST | `/auth/verify-otp` | No |
| GET | `/user/dashboard` | JWT |
| GET | `/user/profile` | JWT |
| PUT | `/user/profile` | JWT |
| POST | `/recharge/initiate` | JWT |
| GET | `/recharge/history` | JWT |
| GET | `/recharge/data-bundles?network=mtn` | JWT |
| GET | `/draws/active` | No |
| GET | `/draws/my-entries` | JWT |
| GET | `/draws/winners` | No |
| GET | `/spin/eligibility` | JWT |
| POST | `/spin/play` | JWT |
| GET | `/spin/history` | JWT |
| GET | `/spins/tiers` | No |
| GET | `/subscription/status` | JWT |
| POST | `/subscription/subscribe` | JWT |
| POST | `/subscription/cancel` | JWT |
| POST | `/affiliate/register` | JWT |
| GET | `/affiliate/dashboard` | JWT |
| POST | `/affiliate/payment` | JWT |

---

## 🎰 Key Features

### Spin Wheel
- Custom `CustomPainter`-based wheel with 8 segments
- 5 full rotation animation + easing
- Confetti on wins
- `SpinStatus` state machine: idle → spinning → result → idle

### Recharge Flow
- Auto-detects Nigerian network from phone prefix (MTN/Glo/Airtel/9Mobile)
- Airtime: preset amounts + custom input
- Data: live bundles from API
- Unlocks spin + shows draw entry count preview
- Paystack payment URL redirect

### Auth
- Phone OTP with 60-second resend timer
- JWT stored in `FlutterSecureStorage` (encrypted)
- GoRouter auto-redirects unauthenticated users

---

## 🔒 Security

- **JWT tokens** in `FlutterSecureStorage` (Android Keystore / iOS Secure Enclave)
- **No tokens in SharedPreferences** or plain storage
- HTTPS-only API calls (no cleartext traffic in Android manifest)
- OTP expiry respected by countdown timer

---

## 🎨 UI Components

| Component | File |
|-----------|------|
| `AppGradientButton` | Press animation, loading state, disabled |
| `GlassCard` | Semi-transparent dark card for hero sections |
| `AppCard` | Light card with subtle border + shadow |
| `TierBadge` | Bronze/Silver/Gold/Platinum with icons |
| `PointsDisplay` | Animated gold star + formatted number |
| `NetworkBadge` | MTN/Glo/Airtel/9Mobile with brand colors |
| `AppEmptyState` | Icon + title + subtitle + optional action |
| `ShimmerBox` | Loading skeleton |
| `SectionHeader` | Title + "See all" link |
| `MainScaffold` | Bottom nav with special spin button |

---

## 📦 Dependencies

```yaml
flutter_riverpod: ^2.5.1     # State management
go_router: ^13.2.0           # Navigation
dio: ^5.4.3                  # HTTP client
flutter_secure_storage: ^9.0.0 # JWT storage
flutter_animate: ^4.5.0      # Animations
confetti: ^0.7.0             # Win celebrations
pin_code_fields: ^8.0.1      # OTP input
shimmer: ^3.0.0              # Loading skeletons
share_plus: ^9.0.0           # Referral link sharing
fl_chart: ^0.68.0            # Charts (earnings)
```

---

## 📝 Notes

1. **Fonts**: You must download Plus Jakarta Sans separately (Google Fonts — free, OFL license)
2. **Paystack**: In production, implement the full Paystack Flutter SDK or WebView for payment
3. **Push notifications**: Wire FCM via `firebase_messaging` for draw results and win alerts
4. **App icon**: Replace `assets/images/` with your actual icon set; use `flutter_launcher_icons`

---

*Built with ❤️ for Nigeria's #1 gamified recharge platform*
