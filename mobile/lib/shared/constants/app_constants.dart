/// App-wide constants
abstract class AppConstants {
  // ─── API ──────────────────────────────────────────────────────────────────
  static const String baseUrl = 'https://rechargemax-api.onrender.com/api/v1';
  static const Duration connectTimeout = Duration(seconds: 30);
  static const Duration receiveTimeout = Duration(seconds: 30);

  // ─── Storage Keys ─────────────────────────────────────────────────────────
  static const String accessTokenKey = 'access_token';
  static const String refreshTokenKey = 'refresh_token';
  static const String userKey = 'user_data';
  static const String onboardingDoneKey = 'onboarding_done';
  static const String themeModeKey = 'theme_mode';

  // ─── Hive Boxes ───────────────────────────────────────────────────────────
  static const String userBox = 'user_box';
  static const String cacheBox = 'cache_box';

  // ─── Business Rules ───────────────────────────────────────────────────────
  static const int pointsPerNaira = 200;       // ₦200 = 1 point
  static const int spinThresholdNaira = 1000;  // ₦1000+ unlocks spin
  static const int subscriptionCostNaira = 20; // ₦20/day

  // ─── Tier Thresholds (₦ in 90 days) ─────────────────────────────────────
  static const int silverThreshold = 5000;
  static const int goldThreshold = 20000;

  // ─── Payout Minimum ──────────────────────────────────────────────────────
  static const int minPayoutNaira = 1000;

  // ─── Networks ─────────────────────────────────────────────────────────────
  static const Map<String, String> networkNames = {
    'mtn': 'MTN',
    'glo': 'Glo',
    'airtel': 'Airtel',
    '9mobile': '9Mobile',
  };

  static const Map<String, String> networkColors = {
    'mtn': '#FFCB05',
    'glo': '#009A44',
    'airtel': '#E30613',
    '9mobile': '#006B3F',
  };

  // ─── Animation Durations ─────────────────────────────────────────────────
  static const Duration microDuration = Duration(milliseconds: 150);
  static const Duration shortDuration = Duration(milliseconds: 250);
  static const Duration mediumDuration = Duration(milliseconds: 400);
  static const Duration longDuration = Duration(milliseconds: 600);
  static const Duration extraLongDuration = Duration(milliseconds: 1000);
  static const Duration spinDuration = Duration(seconds: 4);
}
