import 'package:flutter_test/flutter_test.dart';
import 'package:rechargemax/shared/constants/app_constants.dart';

void main() {
  group('AppConstants — API config', () {
    test('baseUrl is HTTPS and points to render.com', () {
      expect(AppConstants.baseUrl, startsWith('https://'));
      expect(AppConstants.baseUrl, contains('rechargemax-backend'));
      expect(AppConstants.baseUrl, isNot(contains('localhost')));
    });

    test('timeouts are between 5–60 seconds', () {
      expect(AppConstants.connectTimeout.inSeconds, greaterThanOrEqualTo(5));
      expect(AppConstants.connectTimeout.inSeconds, lessThanOrEqualTo(60));
      expect(AppConstants.receiveTimeout.inSeconds, greaterThanOrEqualTo(5));
      expect(AppConstants.receiveTimeout.inSeconds, lessThanOrEqualTo(60));
    });
  });

  group('AppConstants — storage keys', () {
    test('all storage keys are non-empty strings', () {
      for (final key in [
        AppConstants.accessTokenKey,
        AppConstants.userKey,
        AppConstants.onboardingDoneKey,
        AppConstants.themeModeKey,
      ]) {
        expect(key, isNotEmpty, reason: 'Storage key must not be empty');
      }
    });

    test('storage keys are unique (no collisions)', () {
      final keys = [
        AppConstants.accessTokenKey,
        AppConstants.refreshTokenKey,
        AppConstants.userKey,
        AppConstants.onboardingDoneKey,
        AppConstants.themeModeKey,
      ];
      expect(keys.toSet().length, keys.length, reason: 'Storage keys must be unique');
    });
  });

  group('AppConstants — business rules', () {
    test('spin threshold is positive and <= 2000 naira', () {
      expect(AppConstants.spinThresholdNaira, greaterThan(0));
      expect(AppConstants.spinThresholdNaira, lessThanOrEqualTo(2000));
    });

    test('subscription cost is positive and affordable (< 100 naira)', () {
      expect(AppConstants.subscriptionCostNaira, greaterThan(0));
      expect(AppConstants.subscriptionCostNaira, lessThan(100));
    });

    test('payout minimum is positive', () {
      expect(AppConstants.minPayoutNaira, greaterThan(0));
    });

    test('silver threshold > 0', () {
      expect(AppConstants.silverThreshold, greaterThan(0));
    });

    test('gold threshold > silver threshold', () {
      expect(AppConstants.goldThreshold, greaterThan(AppConstants.silverThreshold));
    });
  });

  group('AppConstants — networks', () {
    test('all 4 Nigerian networks are present', () {
      final keys = AppConstants.networkNames.keys.toList();
      expect(keys, containsAll(['mtn', 'glo', 'airtel', '9mobile']));
    });

    test('every network has a display name', () {
      for (final entry in AppConstants.networkNames.entries) {
        expect(entry.value, isNotEmpty, reason: '${entry.key} display name is empty');
      }
    });

    test('every network has a hex color', () {
      for (final network in AppConstants.networkNames.keys) {
        expect(AppConstants.networkColors.containsKey(network), true,
            reason: 'Missing color for $network');
        expect(AppConstants.networkColors[network], startsWith('#'),
            reason: 'Color for $network must be hex (#RRGGBB)');
        expect(AppConstants.networkColors[network]!.length, 7,
            reason: 'Color for $network must be 7 chars #RRGGBB');
      }
    });
  });

  group('AppConstants — animation durations', () {
    test('durations are ordered: micro < short < medium < long < extraLong', () {
      expect(AppConstants.microDuration, lessThan(AppConstants.shortDuration));
      expect(AppConstants.shortDuration, lessThan(AppConstants.mediumDuration));
      expect(AppConstants.mediumDuration, lessThan(AppConstants.longDuration));
      expect(AppConstants.longDuration, lessThan(AppConstants.extraLongDuration));
    });

    test('spin duration is between 1 and 10 seconds', () {
      expect(AppConstants.spinDuration.inSeconds, greaterThanOrEqualTo(1));
      expect(AppConstants.spinDuration.inSeconds, lessThanOrEqualTo(10));
    });
  });
}
