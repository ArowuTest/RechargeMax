import 'package:flutter_test/flutter_test.dart';
import 'package:rechargemax/core/auth/auth_provider.dart';

void main() {
  group('UserProfile.fromJson', () {
    test('parses full profile from flat JSON', () {
      final p = UserProfile.fromJson({
        'id': '42',
        'msisdn': '08012345678',
        'name': 'Ada Obi',
        'email': 'ada@test.com',
        'points': 2500,
        'tier': 'SILVER',
        'is_affiliate': true,
        'referral_code': 'REFABC',
      });
      expect(p.id, '42');
      expect(p.msisdn, '08012345678');
      expect(p.name, 'Ada Obi');
      expect(p.email, 'ada@test.com');
      expect(p.points, 2500);
      expect(p.tier, 'SILVER');
      expect(p.isAffiliate, true);
      expect(p.referralCode, 'REFABC');
    });

    test('parses profile nested under "user" key (login response)', () {
      final p = UserProfile.fromJson({
        'token': 'jwt_abc',
        'user': {
          'id': '10',
          'msisdn': '08011111111',
          'name': 'Nested User',
          'points': 100,
          'tier': 'BRONZE',
          'is_affiliate': false,
        },
      });
      expect(p.msisdn, '08011111111');
      expect(p.name, 'Nested User');
      expect(p.points, 100);
    });

    test('falls back to full_name when name is absent', () {
      final p = UserProfile.fromJson({'id': '1', 'msisdn': '0801', 'full_name': 'Fallback Name'});
      expect(p.name, 'Fallback Name');
    });

    test('handles missing optional fields with safe defaults', () {
      final p = UserProfile.fromJson({'msisdn': '0801'});
      expect(p.id, '');
      expect(p.email, isNull);
      expect(p.points, 0);
      expect(p.tier, 'BRONZE');
      expect(p.isAffiliate, false);
      expect(p.referralCode, isNull);
    });

    test('normalises tier to uppercase', () {
      final p = UserProfile.fromJson({'msisdn': '0801', 'tier': 'gold'});
      expect(p.tier, 'GOLD');
    });

    test('treats affiliate=true when affiliate object is non-null', () {
      final p = UserProfile.fromJson({'msisdn': '0801', 'affiliate': {'code': 'X1'}});
      expect(p.isAffiliate, true);
    });
  });

  group('UserProfile.initials', () {
    test('two-word name -> first letters of first and second word', () {
      final p = UserProfile.fromJson({'msisdn': '0801', 'name': 'John Doe'});
      expect(p.initials, 'JD');
    });

    test('three-word name -> first letters of first and second word', () {
      final p = UserProfile.fromJson({'msisdn': '0801', 'name': 'Chidi Obi Nwosu'});
      expect(p.initials, 'CO');
    });

    test('single-word name -> first letter', () {
      final p = UserProfile.fromJson({'msisdn': '0801', 'name': 'Madonna'});
      expect(p.initials, 'M');
    });

    test('no name -> last 2 digits of msisdn', () {
      final p = UserProfile.fromJson({'msisdn': '08098765432'});
      expect(p.initials, '32');
    });
  });

  group('UserProfile.displayName', () {
    test('shows name when available', () {
      final p = UserProfile.fromJson({'msisdn': '08012345678', 'name': 'Tolu Adeyemi'});
      expect(p.displayName, 'Tolu Adeyemi');
    });

    test('falls back to msisdn when name is null', () {
      final p = UserProfile.fromJson({'msisdn': '08012345678'});
      expect(p.displayName, '08012345678');
    });
  });

  group('UserProfile.toJson', () {
    test('round-trip: fromJson -> toJson preserves all fields', () {
      final original = {
        'id': '99',
        'msisdn': '08099999999',
        'name': 'Round Trip',
        'email': 'rt@test.com',
        'points': 750,
        'tier': 'GOLD',
        'is_affiliate': false,
        'referral_code': 'RTCODE',
      };
      final json = UserProfile.fromJson(original).toJson();
      expect(json['id'], '99');
      expect(json['msisdn'], '08099999999');
      expect(json['name'], 'Round Trip');
      expect(json['email'], 'rt@test.com');
      expect(json['points'], 750);
      expect(json['tier'], 'GOLD');
      expect(json['referral_code'], 'RTCODE');
    });
  });

  group('AuthState', () {
    test('default state is unknown (pre-hydration)', () {
      const s = AuthState();
      expect(s.status, AuthStatus.unknown);
      expect(s.user, isNull);
      expect(s.error, isNull);
    });

    test('isAuthenticated true only when status==authenticated', () {
      expect(const AuthState(status: AuthStatus.authenticated).isAuthenticated, true);
      expect(const AuthState(status: AuthStatus.unauthenticated).isAuthenticated, false);
      expect(const AuthState(status: AuthStatus.unknown).isAuthenticated, false);
    });

    test('isLoading true only when status==unknown (hydrating)', () {
      expect(const AuthState(status: AuthStatus.unknown).isLoading, true);
      expect(const AuthState(status: AuthStatus.authenticated).isLoading, false);
      expect(const AuthState(status: AuthStatus.unauthenticated).isLoading, false);
    });

    test('copyWith preserves unset fields', () {
      const original = AuthState(status: AuthStatus.unauthenticated);
      final copy = original.copyWith(error: 'wrong OTP');
      expect(copy.status, AuthStatus.unauthenticated);  // preserved
      expect(copy.error, 'wrong OTP');
      expect(copy.user, isNull);                        // preserved
    });

    test('copyWith can set user', () {
      const original = AuthState(status: AuthStatus.unauthenticated);
      final user = UserProfile.fromJson({'id': '1', 'msisdn': '0801', 'name': 'X', 'points': 0, 'tier': 'BRONZE', 'is_affiliate': false});
      final copy = original.copyWith(status: AuthStatus.authenticated, user: user);
      expect(copy.status, AuthStatus.authenticated);
      expect(copy.user!.msisdn, '0801');
    });
  });
}
