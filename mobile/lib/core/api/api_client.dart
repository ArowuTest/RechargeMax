import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:pretty_dio_logger/pretty_dio_logger.dart';
import '../../shared/constants/app_constants.dart';

// ─── Secure Storage Provider ────────────────────────────────────────────────
final secureStorageProvider = Provider<FlutterSecureStorage>((ref) {
  return const FlutterSecureStorage(
    aOptions: AndroidOptions(encryptedSharedPreferences: true),
    iOptions: IOSOptions(accessibility: KeychainAccessibility.first_unlock),
  );
});

// ─── Dio Provider ────────────────────────────────────────────────────────────
final dioProvider = Provider<Dio>((ref) {
  final storage = ref.watch(secureStorageProvider);

  final dio = Dio(
    BaseOptions(
      baseUrl: AppConstants.baseUrl,
      connectTimeout: AppConstants.connectTimeout,
      receiveTimeout: AppConstants.receiveTimeout,
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
      },
    ),
  );

  // Auth interceptor — attaches JWT to every request
  dio.interceptors.add(
    InterceptorsWrapper(
      onRequest: (options, handler) async {
        final token = await storage.read(key: AppConstants.accessTokenKey);
        if (token != null) {
          options.headers['Authorization'] = 'Bearer $token';
        }
        handler.next(options);
      },
      onError: (error, handler) async {
        // 401 → token expired → clear and redirect to login
        if (error.response?.statusCode == 401) {
          await storage.delete(key: AppConstants.accessTokenKey);
          // Navigation will be handled by GoRouter redirect
        }
        handler.next(error);
      },
    ),
  );

  // Logger (debug only)
  assert(() {
    dio.interceptors.add(
      PrettyDioLogger(
        requestHeader: false,
        requestBody: true,
        responseBody: true,
        error: true,
        compact: true,
      ),
    );
    return true;
  }());

  return dio;
});

// ─── API Client ──────────────────────────────────────────────────────────────

/// Abstract interface — implemented by ApiClient (production) and MockApiClient (tests)
abstract class ApiBase {
  Future<Map<String, dynamic>> sendOtp(String msisdn);
  Future<Map<String, dynamic>> verifyOtp(String msisdn, String otp);
  Future<Map<String, dynamic>> getDashboard();
  Future<Map<String, dynamic>> getProfile();
  Future<Map<String, dynamic>> updateProfile(Map<String, dynamic> data);
  Future<Map<String, dynamic>> initiateRecharge(Map<String, dynamic> data);
  Future<Map<String, dynamic>> getRechargeHistory({int page = 1});
  Future<List<dynamic>> getDataBundles(String network);
  Future<Map<String, dynamic>> getActiveDraws();
  Future<Map<String, dynamic>> getDrawHistory({int page = 1, String? drawId});
  Future<Map<String, dynamic>> getMyDrawEntries({int page = 1});
  Future<Map<String, dynamic>> checkSpinEligibility();
  Future<Map<String, dynamic>> playSpin();
  Future<Map<String, dynamic>> getSpinHistory({int page = 1});
  Future<Map<String, dynamic>> getSpinTiers();
  Future<Map<String, dynamic>> getSubscriptionStatus();
  Future<Map<String, dynamic>> subscribe(Map<String, dynamic> paymentData);
  Future<Map<String, dynamic>> cancelSubscription();
  Future<Map<String, dynamic>> registerAffiliate();
  Future<Map<String, dynamic>> getAffiliateDashboard();
  Future<Map<String, dynamic>> requestPayout(Map<String, dynamic> data);
  Future<Map<String, dynamic>> getWinners({int page = 1, String? drawId});
}

class ApiClient implements ApiBase {
  final Dio _dio;
  ApiClient(this._dio);

  // Auth
  Future<Map<String, dynamic>> sendOtp(String msisdn) async {
    final res = await _dio.post('/auth/send-otp', data: {'msisdn': msisdn});
    return res.data as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> verifyOtp(String msisdn, String otp) async {
    final res = await _dio.post('/auth/verify-otp', data: {
      'msisdn': msisdn,
      'otp': otp,
    });
    return res.data as Map<String, dynamic>;
  }

  // User
  Future<Map<String, dynamic>> getDashboard() async {
    final res = await _dio.get('/user/dashboard');
    return res.data as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> getProfile() async {
    final res = await _dio.get('/user/profile');
    return res.data as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> updateProfile(Map<String, dynamic> data) async {
    // Backend uses POST /user/profile for updates (not PUT — router.go:user.POST("/profile"))
    final res = await _dio.post('/user/profile', data: data);
    return res.data as Map<String, dynamic>;
  }

  // Recharge
  Future<Map<String, dynamic>> initiateRecharge(Map<String, dynamic> data) async {
    final res = await _dio.post('/recharge/initiate', data: data);
    return res.data as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> getRechargeHistory({int page = 1}) async {
    final res = await _dio.get('/recharge/history', queryParameters: {'page': page});
    return res.data as Map<String, dynamic>;
  }

  Future<List<dynamic>> getDataBundles(String network) async {
    // Correct backend route: GET /networks/:networkId/bundles
    final res = await _dio.get('/networks/${network.toLowerCase()}/bundles');
    return (res.data['data'] ?? res.data['bundles'] ?? []) as List<dynamic>;
  }

  // Draws
  Future<Map<String, dynamic>> getActiveDraws() async {
    final res = await _dio.get('/draws/active');
    return res.data as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> getDrawHistory({int page = 1, String? drawId}) async {
    // Backend route: GET /draws (list of all draws with history) — /draws/history needs draw_id
    final res = await _dio.get('/draws', queryParameters: {'page': page});
    return res.data as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> getMyDrawEntries({int page = 1}) async {
    final res = await _dio.get('/draws/my-entries', queryParameters: {'page': page});
    return res.data as Map<String, dynamic>;
  }

  // Spin
  Future<Map<String, dynamic>> checkSpinEligibility() async {
    final res = await _dio.get('/spin/eligibility');
    return res.data as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> playSpin() async {
    final res = await _dio.post('/spin/play');
    return res.data as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> getSpinHistory({int page = 1}) async {
    final res = await _dio.get('/spin/history', queryParameters: {'page': page});
    return res.data as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> getSpinTiers() async {
    final res = await _dio.get('/spins/tiers');
    return res.data as Map<String, dynamic>;
  }

  // Subscription
  Future<Map<String, dynamic>> getSubscriptionStatus() async {
    final res = await _dio.get('/subscription/status');
    return res.data as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> subscribe(Map<String, dynamic> paymentData) async {
    // Correct backend route: POST /subscription/create
    final res = await _dio.post('/subscription/create', data: paymentData);
    return res.data as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> cancelSubscription() async {
    final res = await _dio.post('/subscription/cancel');
    return res.data as Map<String, dynamic>;
  }

  // Affiliate
  Future<Map<String, dynamic>> registerAffiliate() async {
    final res = await _dio.post('/affiliate/register');
    return res.data as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> getAffiliateDashboard() async {
    final res = await _dio.get('/affiliate/dashboard');
    return res.data as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> requestPayout(Map<String, dynamic> data) async {
    // Correct backend route: POST /affiliate/payout
    final res = await _dio.post('/affiliate/payout', data: data);
    return res.data as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> getWinners({int page = 1, String? drawId}) async {
    // Backend route: GET /draws/:id/winners OR GET /winners (public recent winners)
    if (drawId != null) {
      final res = await _dio.get('/draws/$drawId/winners');
      return res.data as Map<String, dynamic>;
    }
    // Fallback: public recent winners list
    final res = await _dio.get('/winners', queryParameters: {'page': page});
    return res.data as Map<String, dynamic>;
  }
}

final apiClientProvider = Provider<ApiClient>((ref) {
  return ApiClient(ref.watch(dioProvider));
});
