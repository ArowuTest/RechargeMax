import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../auth/auth_provider.dart';
import '../../features/auth/presentation/screens/splash_screen.dart';
import '../../features/auth/presentation/screens/onboarding_screen.dart';
import '../../features/auth/presentation/screens/phone_entry_screen.dart';
import '../../features/auth/presentation/screens/otp_screen.dart';
import '../../features/auth/presentation/screens/profile_setup_screen.dart';
import '../../features/home/presentation/screens/home_screen.dart';
import '../../features/recharge/presentation/screens/recharge_screen.dart';
import '../../features/recharge/presentation/screens/recharge_success_screen.dart';
import '../../features/spin/presentation/screens/spin_screen.dart';
import '../../features/draws/presentation/screens/draws_screen.dart';
import '../../features/subscription/presentation/screens/subscription_screen.dart';
import '../../features/affiliate/presentation/screens/affiliate_screen.dart';
import '../../features/profile/presentation/screens/profile_screen.dart';
import '../../features/profile/presentation/screens/transaction_history_screen.dart';
import '../../shared/widgets/main_scaffold.dart';

final routerProvider = Provider<GoRouter>((ref) {
  final authState = ref.watch(authProvider);

  return GoRouter(
    initialLocation: '/splash',
    debugLogDiagnostics: false,
    redirect: (context, state) {
      final auth = authState.valueOrNull;
      final isLoading = auth == null || auth.status == AuthStatus.unknown;
      final isAuthed = auth?.isAuthenticated ?? false;
      final location = state.matchedLocation;

      if (isLoading) {
        return location == '/splash' ? null : '/splash';
      }

      final publicRoutes = ['/splash', '/onboarding', '/login', '/login/otp'];
      final isPublic = publicRoutes.any((r) => location.startsWith(r));

      if (!isAuthed && !isPublic) return '/login';
      if (isAuthed && isPublic && location != '/splash') return '/home';

      return null;
    },
    routes: [
      // ─── Auth ─────────────────────────────────────────────────────────────
      GoRoute(
        path: '/splash',
        builder: (context, state) => const SplashScreen(),
      ),
      GoRoute(
        path: '/onboarding',
        builder: (context, state) => const OnboardingScreen(),
      ),
      GoRoute(
        path: '/login',
        builder: (context, state) => const PhoneEntryScreen(),
        routes: [
          GoRoute(
            path: 'otp',
            builder: (context, state) {
              final msisdn = state.extra as String? ?? '';
              return OtpScreen(msisdn: msisdn);
            },
          ),
        ],
      ),
      GoRoute(
        path: '/profile-setup',
        builder: (context, state) => const ProfileSetupScreen(),
      ),

      // ─── Main App Shell (with bottom nav) ────────────────────────────────
      ShellRoute(
        builder: (context, state, child) => MainScaffold(child: child),
        routes: [
          GoRoute(
            path: '/home',
            builder: (context, state) => const HomeScreen(),
          ),
          GoRoute(
            path: '/recharge',
            builder: (context, state) => const RechargeScreen(),
            routes: [
              GoRoute(
                path: 'success',
                builder: (context, state) {
                  final data = state.extra as Map<String, dynamic>? ?? {};
                  return RechargeSuccessScreen(data: data);
                },
              ),
            ],
          ),
          GoRoute(
            path: '/spin',
            builder: (context, state) => const SpinScreen(),
          ),
          GoRoute(
            path: '/draws',
            builder: (context, state) => const DrawsScreen(),
          ),
          GoRoute(
            path: '/profile',
            builder: (context, state) => const ProfileScreen(),
            routes: [
              GoRoute(
                path: 'history',
                builder: (context, state) => const TransactionHistoryScreen(),
              ),
              GoRoute(
                path: 'subscription',
                builder: (context, state) => const SubscriptionScreen(),
              ),
              GoRoute(
                path: 'affiliate',
                builder: (context, state) => const AffiliateScreen(),
              ),
            ],
          ),
        ],
      ),
    ],
    errorBuilder: (context, state) => Scaffold(
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.error_outline, size: 64, color: Colors.red),
            const SizedBox(height: 16),
            Text('Page not found: ${state.uri}'),
            TextButton(
              onPressed: () => context.go('/home'),
              child: const Text('Go Home'),
            ),
          ],
        ),
      ),
    ),
  );
});
