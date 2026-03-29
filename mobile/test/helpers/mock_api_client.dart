// Mock API — implements ApiBase for use in tests (no Dio needed)
import 'package:rechargemax/core/api/api_client.dart';

class MockApiClient implements ApiBase {
  Map<String, dynamic> _eligibility = {'eligible': true, 'spins_remaining': 2};
  Map<String, dynamic> _spinResult = {'prize': '500 Cash', 'prize_amount': 500};
  Map<String, dynamic> _rechargeResult = {
    'success': true, 'reference': 'REF123', 'draw_entries': 1, 'spin_unlocked': true,
  };
  List<dynamic> _bundles = [
    {'id': '1', 'name': '1GB Daily', 'amount': 300, 'validity': '1 day'},
  ];

  Exception? _nextError;
  void injectError(Exception e) => _nextError = e;

  Future<T> _respond<T>(T data) async {
    await Future.delayed(const Duration(milliseconds: 2));
    if (_nextError != null) {
      final e = _nextError!; _nextError = null;
      throw e;
    }
    return data;
  }

  void setSpinEligibility(bool eligible, int remaining) =>
      _eligibility = {'eligible': eligible, 'spins_remaining': remaining};
  void setSpinResult(Map<String, dynamic> r) => _spinResult = r;
  void setRechargeResult(Map<String, dynamic> r) => _rechargeResult = r;

  @override Future<Map<String, dynamic>> sendOtp(String msisdn) => _respond({'message': 'OTP sent'});
  @override Future<Map<String, dynamic>> verifyOtp(String msisdn, String otp) => _respond({
    'token': 'jwt_test_token', 'id': '1', 'msisdn': msisdn,
    'name': 'Test User', 'email': 'test@test.com',
    'points': 1500, 'tier': 'BRONZE', 'is_affiliate': false, 'referral_code': 'REF001',
  });
  @override Future<Map<String, dynamic>> getProfile() => _respond({
    'id': '1', 'msisdn': '08012345678', 'name': 'Test User',
    'email': 'test@test.com', 'points': 1500, 'tier': 'BRONZE',
    'is_affiliate': false, 'referral_code': 'REF001',
  });
  @override Future<Map<String, dynamic>> getDashboard() => _respond({
    'points': 1500, 'tier': 'BRONZE', 'draw_entries': 3,
  });
  @override Future<Map<String, dynamic>> updateProfile(Map<String, dynamic> data) => _respond({'success': true});
  @override Future<Map<String, dynamic>> checkSpinEligibility() => _respond(_eligibility);
  @override Future<Map<String, dynamic>> playSpin() => _respond(_spinResult);
  @override Future<Map<String, dynamic>> initiateRecharge(Map<String, dynamic> data) => _respond(_rechargeResult);
  @override Future<List<dynamic>> getDataBundles(String network) => _respond(_bundles);
  @override Future<Map<String, dynamic>> getRechargeHistory({int page = 1}) => _respond({'recharges': [], 'total': 0});
  @override Future<Map<String, dynamic>> getActiveDraws() => _respond({'draws': []});
  @override Future<Map<String, dynamic>> getDrawHistory({int page = 1, String? drawId}) => _respond({'draws': []});
  @override Future<Map<String, dynamic>> getMyDrawEntries({int page = 1}) => _respond({'entries': [], 'total': 0});
  @override Future<Map<String, dynamic>> getSpinHistory({int page = 1}) => _respond({'spins': []});
  @override Future<Map<String, dynamic>> getSpinTiers() => _respond({'tiers': []});
  @override Future<Map<String, dynamic>> getSubscriptionStatus() => _respond({'active': false});
  @override Future<Map<String, dynamic>> subscribe(Map<String, dynamic> paymentData) => _respond({'success': true});
  @override Future<Map<String, dynamic>> cancelSubscription() => _respond({'success': true});
  @override Future<Map<String, dynamic>> registerAffiliate() => _respond({'success': true});
  @override Future<Map<String, dynamic>> getAffiliateDashboard() => _respond({'total_earnings': 0, 'balance': 0});
  @override Future<Map<String, dynamic>> requestPayout(Map<String, dynamic> data) => _respond({'success': true});
  @override Future<Map<String, dynamic>> getWinners({int page = 1, String? drawId}) => _respond({'winners': []});
}
