import 'package:flutter_test/flutter_test.dart';
import 'package:rechargemax/features/spin/presentation/screens/spin_screen.dart';
import '../helpers/mock_api_client.dart';

void main() {
  late MockApiClient mockApi;
  setUp(() => mockApi = MockApiClient());

  group('SpinState', () {
    test('default values', () {
      const s = SpinState();
      expect(s.status, SpinStatus.idle);
      expect(s.eligible, false);
      expect(s.spinsRemaining, 0);
      expect(s.result, isNull);
      expect(s.error, isNull);
    });

    test('copyWith only changes specified fields', () {
      const s = SpinState(eligible: true, spinsRemaining: 5);
      final c = s.copyWith(status: SpinStatus.spinning);
      expect(c.status, SpinStatus.spinning);
      expect(c.eligible, true);
      expect(c.spinsRemaining, 5);
    });
  });

  group('SpinNotifier', () {
    test('starts at idle, not eligible, 0 spins', () {
      final n = SpinNotifier(mockApi);
      expect(n.state.status, SpinStatus.idle);
      expect(n.state.eligible, false);
      expect(n.state.spinsRemaining, 0);
    });

    test('checkEligibility eligible=true', () async {
      mockApi.setSpinEligibility(true, 3);
      final n = SpinNotifier(mockApi);
      await n.checkEligibility();
      expect(n.state.eligible, true);
      expect(n.state.spinsRemaining, 3);
      expect(n.state.status, SpinStatus.idle);
    });

    test('checkEligibility eligible=false -> ineligible status', () async {
      mockApi.setSpinEligibility(false, 0);
      final n = SpinNotifier(mockApi);
      await n.checkEligibility();
      expect(n.state.eligible, false);
      expect(n.state.status, SpinStatus.ineligible);
    });

    test('checkEligibility error -> ineligible gracefully', () async {
      mockApi.injectError(Exception('network timeout'));
      final n = SpinNotifier(mockApi);
      await n.checkEligibility();
      expect(n.state.status, SpinStatus.ineligible);
    });

    test('playSpin transitions: idle -> spinning (sync), then result (async)', () async {
      mockApi.setSpinEligibility(true, 2);
      mockApi.setSpinResult({'prize': '500 Cash', 'prize_amount': 500});
      final n = SpinNotifier(mockApi);
      await n.checkEligibility();

      final future = n.playSpin();
      expect(n.state.status, SpinStatus.spinning);  // sync check

      final result = await future;
      expect(n.state.status, SpinStatus.result);
      expect(result['prize_amount'], 500);
      expect(n.state.result!['prize_amount'], 500);
    });

    test('playSpin decrements spinsRemaining', () async {
      mockApi.setSpinEligibility(true, 3);
      mockApi.setSpinResult({'prize': '100', 'prize_amount': 100});
      final n = SpinNotifier(mockApi);
      await n.checkEligibility();
      await n.playSpin();
      expect(n.state.spinsRemaining, 2);
    });

    test('playSpin error -> back to idle with error message', () async {
      mockApi.setSpinEligibility(true, 1);
      final n = SpinNotifier(mockApi);
      await n.checkEligibility();
      mockApi.injectError(Exception('no spins left'));
      try {
        await n.playSpin();
      } catch (_) {}
      expect(n.state.status, SpinStatus.idle);
      expect(n.state.error, contains('no spins left'));
    });

    test('reset after result: clears result, back to idle', () async {
      mockApi.setSpinEligibility(true, 1);
      mockApi.setSpinResult({'prize': '200', 'prize_amount': 200});
      final n = SpinNotifier(mockApi);
      await n.checkEligibility();
      await n.playSpin();
      expect(n.state.status, SpinStatus.result);
      n.reset();
      expect(n.state.status, SpinStatus.idle);
      expect(n.state.result, isNull);
      expect(n.state.error, isNull);
    });

    test('reset when ineligible -> stays ineligible', () async {
      mockApi.setSpinEligibility(false, 0);
      final n = SpinNotifier(mockApi);
      await n.checkEligibility();
      n.reset();
      expect(n.state.status, SpinStatus.ineligible);
    });
  });
}
