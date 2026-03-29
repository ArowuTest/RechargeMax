import 'package:flutter_test/flutter_test.dart';
import 'package:rechargemax/features/recharge/presentation/screens/recharge_screen.dart';
import '../helpers/mock_api_client.dart';

void main() {
  late MockApiClient mockApi;
  setUp(() => mockApi = MockApiClient());

  group('RechargeFormState', () {
    test('default state: phone/network default to empty string', () {
      const s = RechargeFormState();
      expect(s.isLoading, false);
      expect(s.phone, '');     // defaults to empty string
      expect(s.network, '');   // defaults to empty string
      expect(s.rechargeType, 'airtime');
      expect(s.amount, isNull);
      expect(s.selectedBundle, isNull);
      expect(s.error, isNull);
    });

    test('copyWith only changes specified fields', () {
      const s = RechargeFormState(phone: '08012345678', network: 'MTN');
      final c = s.copyWith(isLoading: true);
      expect(c.isLoading, true);
      expect(c.phone, '08012345678');  // preserved
      expect(c.network, 'MTN');        // preserved
    });
  });

  group('RechargeNotifier form setters', () {
    test('setPhone updates state.phone', () {
      final n = RechargeNotifier(mockApi);
      n.setPhone('08012345678');
      expect(n.state.phone, '08012345678');
    });

    test('setNetwork updates state.network', () {
      final n = RechargeNotifier(mockApi);
      n.setNetwork('Airtel');
      expect(n.state.network, 'Airtel');
    });

    test('setType updates rechargeType', () {
      final n = RechargeNotifier(mockApi);
      n.setType('data');
      expect(n.state.rechargeType, 'data');
    });

    test('setAmount updates amount', () {
      final n = RechargeNotifier(mockApi);
      n.setAmount(500);
      expect(n.state.amount, 500);
    });

    test('setBundle sets both bundleId and amount together', () {
      final n = RechargeNotifier(mockApi);
      n.setBundle('bundle_5', 1500);
      expect(n.state.selectedBundle, 'bundle_5');
      expect(n.state.amount, 1500);
    });
  });

  group('RechargeNotifier initiateRecharge', () {
    test('success: isLoading=false after completion, returns result', () async {
      final n = RechargeNotifier(mockApi);
      n.setPhone('08012345678');
      n.setNetwork('MTN');
      n.setAmount(500);
      final result = await n.initiateRecharge();
      expect(n.state.isLoading, false);
      expect(result['success'], true);
      expect(result['reference'], 'REF123');
    });

    test('success result includes draw_entries and spin_unlocked', () async {
      final n = RechargeNotifier(mockApi);
      n.setPhone('08012345678');
      n.setNetwork('MTN');
      n.setAmount(500);
      final result = await n.initiateRecharge();
      expect(result['draw_entries'], 1);
      expect(result['spin_unlocked'], true);
    });

    test('API error: rethrows and isLoading resets to false', () async {
      mockApi.injectError(Exception('Payment gateway timeout'));
      final n = RechargeNotifier(mockApi);
      n.setPhone('08012345678');
      n.setNetwork('MTN');
      n.setAmount(500);
      bool threw = false;
      try {
        await n.initiateRecharge();
      } catch (_) {
        threw = true;
      }
      expect(threw, true);
      expect(n.state.isLoading, false);
    });
  });
}
