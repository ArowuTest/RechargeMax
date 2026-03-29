import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:flutter_animate/flutter_animate.dart';
import 'package:pin_code_fields/pin_code_fields.dart';
import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_text_styles.dart';
import '../../../../core/auth/auth_provider.dart';
import '../../../../core/api/api_client.dart';
import '../../../../shared/widgets/app_widgets.dart';

class OtpScreen extends ConsumerStatefulWidget {
  final String msisdn;
  const OtpScreen({super.key, required this.msisdn});

  @override
  ConsumerState<OtpScreen> createState() => _OtpScreenState();
}

class _OtpScreenState extends ConsumerState<OtpScreen> {
  String _otp = '';
  bool _isVerifying = false;
  bool _isResending = false;
  String? _error;
  int _countdown = 60;
  Timer? _timer;

  @override
  void initState() {
    super.initState();
    _startCountdown();
  }

  void _startCountdown() {
    _timer?.cancel();
    setState(() => _countdown = 60);
    _timer = Timer.periodic(const Duration(seconds: 1), (t) {
      if (_countdown <= 0) {
        t.cancel();
      } else {
        setState(() => _countdown--);
      }
    });
  }

  Future<void> _resendOtp() async {
    if (_countdown > 0 || _isResending) return;
    setState(() {
      _isResending = true;
      _error = null;
    });
    try {
      final api = ref.read(apiClientProvider);
      await api.sendOtp(widget.msisdn);
      _startCountdown();
    } catch (e) {
      setState(() => _error = 'Failed to resend OTP. Please try again.');
    } finally {
      setState(() => _isResending = false);
    }
  }

  Future<void> _verifyOtp() async {
    if (_otp.length != 6 || _isVerifying) return;

    setState(() {
      _isVerifying = true;
      _error = null;
    });

    try {
      await ref.read(authProvider.notifier).loginWithOtp(widget.msisdn, _otp);
      if (mounted) {
        // Check if profile needs setup
        final user = ref.read(currentUserProvider);
        if (user?.name == null || user!.name!.isEmpty) {
          context.go('/profile-setup');
        } else {
          context.go('/home');
        }
      }
    } catch (e) {
      setState(() {
        _error = 'Invalid OTP. Please check and try again.';
        _isVerifying = false;
      });
    }
  }

  @override
  void dispose() {
    _timer?.cancel();
    super.dispose();
  }

  // Format msisdn for display: +234 812 345 6789
  String get _maskedPhone {
    final m = widget.msisdn;
    if (m.startsWith('+234') && m.length >= 13) {
      return '${m.substring(0, 4)} ${m.substring(4, 7)}*** ${m.substring(m.length - 4)}';
    }
    return m;
  }

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;

    return Scaffold(
      backgroundColor: isDark ? AppColors.darkBgPrimary : AppColors.bgPrimary,
      appBar: AppBar(
        backgroundColor: Colors.transparent,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back_rounded),
          onPressed: () => context.pop(),
        ),
      ),
      body: SafeArea(
        child: SingleChildScrollView(
          padding: const EdgeInsets.fromLTRB(24, 8, 24, 24),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Header
              Container(
                width: 64,
                height: 64,
                decoration: BoxDecoration(
                  gradient: AppColors.brandGradient,
                  borderRadius: BorderRadius.circular(16),
                ),
                child: const Icon(Icons.sms_rounded, color: Colors.white, size: 30),
              ).animate().scale(
                begin: const Offset(0.7, 0.7),
                duration: 400.ms,
                curve: Curves.elasticOut,
              ),

              const SizedBox(height: 24),

              Text(
                'Verify your\nphone number',
                style: AppTextStyles.displaySm.copyWith(
                  color: isDark ? AppColors.darkTextPrimary : AppColors.textPrimary,
                  fontWeight: FontWeight.w800,
                ),
              ).animate(delay: 100.ms).fadeIn().slideY(begin: 0.2, end: 0),

              const SizedBox(height: 12),

              RichText(
                text: TextSpan(
                  style: AppTextStyles.bodyLg.copyWith(
                    color: isDark ? AppColors.darkTextSecondary : AppColors.textSecondary,
                  ),
                  children: [
                    const TextSpan(text: 'We sent a 6-digit code to\n'),
                    TextSpan(
                      text: _maskedPhone,
                      style: AppTextStyles.bodyLg.copyWith(
                        color: AppColors.brand500,
                        fontWeight: FontWeight.w700,
                      ),
                    ),
                  ],
                ),
              ).animate(delay: 200.ms).fadeIn(),

              const SizedBox(height: 40),

              // OTP Input
              PinCodeTextField(
                appContext: context,
                length: 6,
                onChanged: (v) => setState(() {
                  _otp = v;
                  _error = null;
                }),
                onCompleted: (_) => _verifyOtp(),
                keyboardType: TextInputType.number,
                animationType: AnimationType.scale,
                animationDuration: const Duration(milliseconds: 150),
                enableActiveFill: true,
                pinTheme: PinTheme(
                  shape: PinCodeFieldShape.box,
                  borderRadius: BorderRadius.circular(12),
                  fieldHeight: 60,
                  fieldWidth: 50,
                  activeColor: AppColors.brand500,
                  selectedColor: AppColors.brand400,
                  inactiveColor: isDark ? AppColors.darkBorderPrimary : AppColors.borderPrimary,
                  activeFillColor: isDark ? AppColors.darkBgTertiary : AppColors.brand50,
                  selectedFillColor: isDark ? AppColors.darkBgTertiary : AppColors.brand50,
                  inactiveFillColor: isDark ? AppColors.darkBgCard : AppColors.bgSecondary,
                  errorBorderColor: AppColors.error500,
                ),
                textStyle: AppTextStyles.headingXl.copyWith(
                  color: isDark ? AppColors.darkTextPrimary : AppColors.textPrimary,
                  fontWeight: FontWeight.w700,
                ),
                errorAnimationController: null,
              ).animate(delay: 300.ms).fadeIn().slideY(begin: 0.2, end: 0),

              // Error message
              if (_error != null) ...[
                const SizedBox(height: 8),
                Text(
                  _error!,
                  style: AppTextStyles.bodySm.copyWith(color: AppColors.error500),
                ).animate().fadeIn().shakeX(),
              ],

              const SizedBox(height: 32),

              // Verify button
              AppGradientButton(
                label: 'Verify & Continue',
                onPressed: _otp.length == 6 ? _verifyOtp : null,
                isLoading: _isVerifying,
                icon: const Icon(Icons.check_circle_rounded, color: Colors.white, size: 18),
              ).animate(delay: 400.ms).fadeIn().slideY(begin: 0.2, end: 0),

              const SizedBox(height: 28),

              // Resend OTP
              Center(
                child: _countdown > 0
                    ? Text(
                        'Resend code in ${_countdown}s',
                        style: AppTextStyles.bodyMd.copyWith(
                          color: isDark ? AppColors.darkTextTertiary : AppColors.textTertiary,
                        ),
                      )
                    : GestureDetector(
                        onTap: _resendOtp,
                        child: _isResending
                            ? const SizedBox(
                                width: 20,
                                height: 20,
                                child: CircularProgressIndicator(strokeWidth: 2),
                              )
                            : RichText(
                                text: TextSpan(
                                  text: "Didn't receive it? ",
                                  style: AppTextStyles.bodyMd.copyWith(
                                    color: isDark
                                        ? AppColors.darkTextTertiary
                                        : AppColors.textTertiary,
                                  ),
                                  children: [
                                    TextSpan(
                                      text: 'Resend',
                                      style: AppTextStyles.labelLg.copyWith(
                                        color: AppColors.brand500,
                                        fontWeight: FontWeight.w700,
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                      ),
              ).animate(delay: 500.ms).fadeIn(),
            ],
          ),
        ),
      ),
    );
  }
}
