import 'package:flutter_test/flutter_test.dart';
import 'package:rechargemax/core/auth/auth_provider.dart';

void main() {
  group('AuthState model', () {
    test('default state is unknown (pre-hydration)', () {
      const s = AuthState();
      expect(s.status, AuthStatus.unknown);
      expect(s.user, isNull);
      expect(s.error, isNull);
    });

    test('isAuthenticated is true only for authenticated', () {
      expect(const AuthState(status: AuthStatus.authenticated).isAuthenticated, true);
      expect(const AuthState(status: AuthStatus.unauthenticated).isAuthenticated, false);
      expect(const AuthState(status: AuthStatus.unknown).isAuthenticated, false);
    });

    test('isLoading is true only for unknown', () {
      expect(const AuthState(status: AuthStatus.unknown).isLoading, true);
      expect(const AuthState(status: AuthStatus.authenticated).isLoading, false);
      expect(const AuthState(status: AuthStatus.unauthenticated).isLoading, false);
    });

    test('copyWith preserves unmodified fields', () {
      const s = AuthState(status: AuthStatus.unauthenticated);
      final copy = s.copyWith(error: 'wrong OTP');
      expect(copy.status, AuthStatus.unauthenticated);  // preserved
      expect(copy.error, 'wrong OTP');
      expect(copy.user, isNull);                        // preserved
    });

    test('copyWith can set user while preserving other fields', () {
      const original = AuthState(status: AuthStatus.unauthenticated);
      final user = UserProfile(
        id: '1', msisdn: '0801', points: 0, tier: 'BRONZE', isAffiliate: false,
      );
      final copy = original.copyWith(status: AuthStatus.authenticated, user: user);
      expect(copy.status, AuthStatus.authenticated);
      expect(copy.user!.msisdn, '0801');
    });

    test('AuthStatus enum has all required values', () {
      expect(AuthStatus.values, containsAll([
        AuthStatus.unknown,
        AuthStatus.authenticated,
        AuthStatus.unauthenticated,
      ]));
    });
  });

  group('UserProfile constructor', () {
    test('required fields can be set directly', () {
      const p = UserProfile(
        id: '42', msisdn: '08012345678',
        points: 2500, tier: 'GOLD', isAffiliate: true,
      );
      expect(p.id, '42');
      expect(p.msisdn, '08012345678');
      expect(p.points, 2500);
      expect(p.tier, 'GOLD');
      expect(p.isAffiliate, true);
    });

    test('optional fields default to null', () {
      const p = UserProfile(
        id: '1', msisdn: '0801', points: 0, tier: 'BRONZE', isAffiliate: false,
      );
      expect(p.name, isNull);
      expect(p.email, isNull);
      expect(p.referralCode, isNull);
    });
  });
}
