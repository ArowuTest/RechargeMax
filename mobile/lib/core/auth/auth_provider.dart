import 'dart:convert';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../api/api_client.dart';
import '../../shared/constants/app_constants.dart';

// ─── Auth State ──────────────────────────────────────────────────────────────
enum AuthStatus { unknown, authenticated, unauthenticated }

class AuthState {
  final AuthStatus status;
  final UserProfile? user;
  final String? error;

  const AuthState({
    this.status = AuthStatus.unknown,
    this.user,
    this.error,
  });

  AuthState copyWith({AuthStatus? status, UserProfile? user, String? error}) {
    return AuthState(
      status: status ?? this.status,
      user: user ?? this.user,
      error: error,
    );
  }

  bool get isAuthenticated => status == AuthStatus.authenticated;
  bool get isLoading => status == AuthStatus.unknown;
}

// ─── User Profile Model ───────────────────────────────────────────────────────
class UserProfile {
  final String id;
  final String msisdn;
  final String? name;
  final String? email;
  final int points;
  final String tier; // BRONZE, SILVER, GOLD, PLATINUM
  final bool isAffiliate;
  final String? referralCode;

  const UserProfile({
    required this.id,
    required this.msisdn,
    this.name,
    this.email,
    required this.points,
    required this.tier,
    required this.isAffiliate,
    this.referralCode,
  });

  factory UserProfile.fromJson(Map<String, dynamic> json) {
    final user = json['user'] ?? json;
    return UserProfile(
      id: (user['id'] ?? user['user_id'] ?? '').toString(),
      msisdn: user['msisdn'] ?? user['phone'] ?? '',
      name: user['name'] ?? user['full_name'],
      email: user['email'],
      points: (user['points'] ?? user['points_balance'] ?? 0) as int,
      tier: (user['tier'] ?? user['loyalty_tier'] ?? 'BRONZE').toString().toUpperCase(),
      isAffiliate: user['is_affiliate'] == true || user['affiliate'] != null,
      referralCode: user['referral_code'],
    );
  }

  Map<String, dynamic> toJson() => {
    'id': id,
    'msisdn': msisdn,
    'name': name,
    'email': email,
    'points': points,
    'tier': tier,
    'is_affiliate': isAffiliate,
    'referral_code': referralCode,
  };

  String get displayName => name?.isNotEmpty == true ? name! : msisdn;
  String get initials {
    if (name?.isNotEmpty == true) {
      final parts = name!.trim().split(' ');
      if (parts.length >= 2) return '${parts[0][0]}${parts[1][0]}'.toUpperCase();
      return name![0].toUpperCase();
    }
    return msisdn.length >= 2 ? msisdn.substring(msisdn.length - 2) : msisdn;
  }
}

// ─── Auth Notifier ────────────────────────────────────────────────────────────
class AuthNotifier extends AsyncNotifier<AuthState> {
  @override
  Future<AuthState> build() async {
    return _checkAuth();
  }

  Future<AuthState> _checkAuth() async {
    final storage = ref.read(secureStorageProvider);
    final token = await storage.read(key: AppConstants.accessTokenKey);

    if (token == null) {
      return const AuthState(status: AuthStatus.unauthenticated);
    }

    // Restore user from cache
    final cached = await storage.read(key: AppConstants.userKey);
    if (cached != null) {
      try {
        final user = UserProfile.fromJson(jsonDecode(cached));
        return AuthState(status: AuthStatus.authenticated, user: user);
      } catch (_) {}
    }

    // Fetch fresh profile
    try {
      final api = ref.read(apiClientProvider);
      final data = await api.getProfile();
      final user = UserProfile.fromJson(data);
      await _cacheUser(user);
      return AuthState(status: AuthStatus.authenticated, user: user);
    } catch (_) {
      await storage.delete(key: AppConstants.accessTokenKey);
      return const AuthState(status: AuthStatus.unauthenticated);
    }
  }

  Future<void> loginWithOtp(String msisdn, String otp) async {
    state = const AsyncLoading();
    try {
      final api = ref.read(apiClientProvider);
      final storage = ref.read(secureStorageProvider);

      final data = await api.verifyOtp(msisdn, otp);

      // Save token
      final token = data['token'] ?? data['access_token'];
      if (token != null) {
        await storage.write(key: AppConstants.accessTokenKey, value: token);
      }

      // Parse user
      final user = UserProfile.fromJson(data);
      await _cacheUser(user);

      state = AsyncData(AuthState(status: AuthStatus.authenticated, user: user));
    } catch (e) {
      state = AsyncData(AuthState(
        status: AuthStatus.unauthenticated,
        error: _parseError(e),
      ));
      rethrow;
    }
  }

  Future<void> logout() async {
    final storage = ref.read(secureStorageProvider);
    await storage.deleteAll();
    state = const AsyncData(AuthState(status: AuthStatus.unauthenticated));
  }

  Future<void> refreshUser() async {
    try {
      final api = ref.read(apiClientProvider);
      final data = await api.getProfile();
      final user = UserProfile.fromJson(data);
      await _cacheUser(user);
      state = AsyncData(AuthState(status: AuthStatus.authenticated, user: user));
    } catch (_) {}
  }

  Future<void> _cacheUser(UserProfile user) async {
    final storage = ref.read(secureStorageProvider);
    await storage.write(
      key: AppConstants.userKey,
      value: jsonEncode(user.toJson()),
    );
  }

  String _parseError(Object e) {
    if (e is Exception) return e.toString().replaceAll('Exception: ', '');
    return 'Something went wrong. Please try again.';
  }
}

final authProvider = AsyncNotifierProvider<AuthNotifier, AuthState>(
  AuthNotifier.new,
);

// Convenience provider for just the user
final currentUserProvider = Provider<UserProfile?>((ref) {
  return ref.watch(authProvider).valueOrNull?.user;
});

final isAuthenticatedProvider = Provider<bool>((ref) {
  return ref.watch(authProvider).valueOrNull?.isAuthenticated ?? false;
});
