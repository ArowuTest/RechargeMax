import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:flutter_animate/flutter_animate.dart';
import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_text_styles.dart';
import '../../../../core/api/api_client.dart';
import '../../../../shared/widgets/app_widgets.dart';

// State for OTP sending
final _sendOtpLoadingProvider = StateProvider<bool>((ref) => false);

class PhoneEntryScreen extends ConsumerStatefulWidget {
  const PhoneEntryScreen({super.key});

  @override
  ConsumerState<PhoneEntryScreen> createState() => _PhoneEntryScreenState();
}

class _PhoneEntryScreenState extends ConsumerState<PhoneEntryScreen> {
  final _controller = TextEditingController();
  final _focusNode = FocusNode();
  String? _error;

  // Nigerian number formatter
  String _normalise(String input) {
    final digits = input.replaceAll(RegExp(r'\D'), '');
    if (digits.startsWith('234') && digits.length == 13) return '+$digits';
    if (digits.startsWith('0') && digits.length == 11) {
      return '+234${digits.substring(1)}';
    }
    if (digits.length == 10) return '+234$digits';
    return '+$digits';
  }

  bool _isValid(String input) {
    final norm = _normalise(input);
    return RegExp(r'^\+234[789]\d{9}$').hasMatch(norm);
  }

  Future<void> _sendOtp() async {
    final raw = _controller.text.trim();
    if (!_isValid(raw)) {
      setState(() => _error = 'Please enter a valid Nigerian phone number');
      return;
    }

    setState(() => _error = null);
    ref.read(_sendOtpLoadingProvider.notifier).state = true;

    try {
      final msisdn = _normalise(raw);
      final api = ref.read(apiClientProvider);
      await api.sendOtp(msisdn);
      if (mounted) {
        context.push('/login/otp', extra: msisdn);
      }
    } catch (e) {
      setState(() => _error = 'Failed to send OTP. Please try again.');
    } finally {
      ref.read(_sendOtpLoadingProvider.notifier).state = false;
    }
  }

  @override
  void dispose() {
    _controller.dispose();
    _focusNode.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final isLoading = ref.watch(_sendOtpLoadingProvider);
    final isDark = Theme.of(context).brightness == Brightness.dark;

    return Scaffold(
      body: Stack(
        children: [
          // Purple gradient top section
          Positioned(
            top: 0,
            left: 0,
            right: 0,
            height: MediaQuery.of(context).size.height * 0.42,
            child: Container(
              decoration: const BoxDecoration(
                gradient: AppColors.heroGradient,
              ),
              child: SafeArea(
                child: Padding(
                  padding: const EdgeInsets.all(24),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      // Logo row
                      Row(
                        children: [
                          Container(
                            width: 40,
                            height: 40,
                            decoration: BoxDecoration(
                              gradient: AppColors.brandGradient,
                              shape: BoxShape.circle,
                            ),
                            child: const Icon(Icons.bolt_rounded, color: Colors.white, size: 22),
                          ),
                          const SizedBox(width: 10),
                          Text(
                            'RechargeMax',
                            style: AppTextStyles.headingMd.copyWith(
                              color: Colors.white,
                              fontWeight: FontWeight.w700,
                            ),
                          ),
                        ],
                      ).animate().fadeIn(duration: 400.ms),

                      const Spacer(),

                      Text(
                        'Welcome back 👋',
                        style: AppTextStyles.bodyLg.copyWith(
                          color: AppColors.brand200,
                        ),
                      ).animate(delay: 100.ms).fadeIn(),

                      const SizedBox(height: 8),

                      Text(
                        'Enter your phone\nnumber to continue',
                        style: AppTextStyles.displaySm.copyWith(
                          color: Colors.white,
                          fontWeight: FontWeight.w800,
                        ),
                      ).animate(delay: 200.ms).fadeIn().slideY(begin: 0.2, end: 0),

                      const SizedBox(height: 20),
                    ],
                  ),
                ),
              ),
            ),
          ),

          // White card bottom
          Positioned(
            top: MediaQuery.of(context).size.height * 0.36,
            left: 0,
            right: 0,
            bottom: 0,
            child: Container(
              decoration: BoxDecoration(
                color: isDark ? AppColors.darkBgPrimary : AppColors.bgPrimary,
                borderRadius: const BorderRadius.vertical(top: Radius.circular(28)),
                boxShadow: [
                  BoxShadow(
                    color: Colors.black.withOpacity(0.15),
                    blurRadius: 20,
                    offset: const Offset(0, -4),
                  ),
                ],
              ),
              child: SingleChildScrollView(
                padding: const EdgeInsets.fromLTRB(24, 32, 24, 24),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'Phone Number',
                      style: AppTextStyles.labelLg.copyWith(
                        color: isDark ? AppColors.darkTextSecondary : AppColors.textSecondary,
                      ),
                    ),
                    const SizedBox(height: 10),

                    // Phone field
                    TextField(
                      controller: _controller,
                      focusNode: _focusNode,
                      keyboardType: TextInputType.phone,
                      inputFormatters: [FilteringTextInputFormatter.digitsOnly],
                      style: AppTextStyles.headingMd.copyWith(
                        color: isDark ? AppColors.darkTextPrimary : AppColors.textPrimary,
                        letterSpacing: 2,
                      ),
                      decoration: InputDecoration(
                        hintText: '0812 345 6789',
                        hintStyle: AppTextStyles.headingMd.copyWith(
                          color: AppColors.textDisabled,
                          letterSpacing: 2,
                        ),
                        prefixIcon: Padding(
                          padding: const EdgeInsets.symmetric(horizontal: 14),
                          child: Row(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              Text('🇳🇬', style: const TextStyle(fontSize: 20)),
                              const SizedBox(width: 8),
                              Text(
                                '+234',
                                style: AppTextStyles.labelXl.copyWith(
                                  color: isDark ? AppColors.darkTextSecondary : AppColors.textSecondary,
                                  fontWeight: FontWeight.w600,
                                ),
                              ),
                              const SizedBox(width: 8),
                              Container(
                                width: 1,
                                height: 20,
                                color: isDark ? AppColors.darkBorderPrimary : AppColors.borderPrimary,
                              ),
                            ],
                          ),
                        ),
                        prefixIconConstraints: const BoxConstraints(),
                        errorText: _error,
                        contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 18),
                      ),
                      onChanged: (_) => setState(() => _error = null),
                      onSubmitted: (_) => _sendOtp(),
                    ).animate(delay: 200.ms).fadeIn().slideY(begin: 0.1, end: 0),

                    const SizedBox(height: 12),

                    Text(
                      'We\'ll send a 6-digit OTP to verify your number.',
                      style: AppTextStyles.bodySm.copyWith(
                        color: isDark ? AppColors.darkTextTertiary : AppColors.textTertiary,
                      ),
                    ).animate(delay: 300.ms).fadeIn(),

                    const SizedBox(height: 32),

                    AppGradientButton(
                      label: 'Send OTP',
                      onPressed: _sendOtp,
                      isLoading: isLoading,
                      icon: const Icon(Icons.sms_rounded, color: Colors.white, size: 18),
                    ).animate(delay: 400.ms).fadeIn().slideY(begin: 0.2, end: 0),

                    const SizedBox(height: 40),

                    // Trust signals
                    Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        _TrustItem(icon: Icons.lock_rounded, label: 'Secure'),
                        const SizedBox(width: 28),
                        _TrustItem(icon: Icons.flash_on_rounded, label: 'Instant'),
                        const SizedBox(width: 28),
                        _TrustItem(icon: Icons.verified_rounded, label: 'Verified'),
                      ],
                    ).animate(delay: 500.ms).fadeIn(),
                  ],
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _TrustItem extends StatelessWidget {
  final IconData icon;
  final String label;
  const _TrustItem({required this.icon, required this.label});

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Icon(icon, size: 22, color: AppColors.brand400),
        const SizedBox(height: 4),
        Text(
          label,
          style: AppTextStyles.labelSm.copyWith(
            color: AppColors.textTertiary,
          ),
        ),
      ],
    );
  }
}
