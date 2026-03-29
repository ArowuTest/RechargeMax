import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:flutter_animate/flutter_animate.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_text_styles.dart';
import '../../../../core/auth/auth_provider.dart';
import '../../../../shared/constants/app_constants.dart';

class SplashScreen extends ConsumerStatefulWidget {
  const SplashScreen({super.key});

  @override
  ConsumerState<SplashScreen> createState() => _SplashScreenState();
}

class _SplashScreenState extends ConsumerState<SplashScreen>
    with SingleTickerProviderStateMixin {
  late AnimationController _pulseController;

  @override
  void initState() {
    super.initState();
    _pulseController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 1500),
    )..repeat(reverse: true);

    _navigate();
  }

  Future<void> _navigate() async {
    // Wait for auth state to resolve + show splash
    await Future.delayed(const Duration(milliseconds: 2200));
    if (!mounted) return;

    final auth = ref.read(authProvider).valueOrNull;
    final storage = const FlutterSecureStorage();
    final onboardingDone = await storage.read(key: AppConstants.onboardingDoneKey);

    if (!mounted) return;

    if (auth?.isAuthenticated == true) {
      context.go('/home');
    } else if (onboardingDone == 'true') {
      context.go('/login');
    } else {
      context.go('/onboarding');
    }
  }

  @override
  void dispose() {
    _pulseController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Container(
        decoration: const BoxDecoration(gradient: AppColors.heroGradient),
        child: Stack(
          children: [
            // Background radial glow
            Positioned.fill(
              child: Container(
                decoration: const BoxDecoration(
                  gradient: AppColors.heroRadialGlow,
                ),
              ),
            ),

            // Floating orbs
            Positioned(
              top: -60,
              right: -60,
              child: _GlowOrb(size: 200, color: AppColors.brand500, opacity: 0.15),
            ),
            Positioned(
              bottom: 100,
              left: -80,
              child: _GlowOrb(size: 250, color: AppColors.brand400, opacity: 0.10),
            ),

            // Content
            Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  // Logo
                  _LogoMark(pulseController: _pulseController)
                      .animate()
                      .scale(
                        begin: const Offset(0.6, 0.6),
                        end: const Offset(1.0, 1.0),
                        duration: 600.ms,
                        curve: Curves.elasticOut,
                      ),

                  const SizedBox(height: 24),

                  // Brand name
                  Text(
                    'RechargeMax',
                    style: AppTextStyles.displayMd.copyWith(
                      color: Colors.white,
                      fontWeight: FontWeight.w800,
                    ),
                  )
                      .animate(delay: 300.ms)
                      .fadeIn(duration: 500.ms)
                      .slideY(begin: 0.3, end: 0),

                  const SizedBox(height: 8),

                  Text(
                    'Recharge & Win',
                    style: AppTextStyles.bodyXl.copyWith(
                      color: AppColors.brand200,
                      fontWeight: FontWeight.w500,
                    ),
                  )
                      .animate(delay: 500.ms)
                      .fadeIn(duration: 400.ms),

                  const SizedBox(height: 64),

                  // Loading dots
                  _LoadingDots()
                      .animate(delay: 800.ms)
                      .fadeIn(duration: 400.ms),
                ],
              ),
            ),

            // Bottom tagline
            Positioned(
              bottom: 48,
              left: 0,
              right: 0,
              child: Text(
                'by BridgeTunes',
                style: AppTextStyles.bodySm.copyWith(
                  color: Colors.white.withOpacity(0.4),
                ),
                textAlign: TextAlign.center,
              ).animate(delay: 1000.ms).fadeIn(duration: 400.ms),
            ),
          ],
        ),
      ),
    );
  }
}

class _LogoMark extends StatelessWidget {
  final AnimationController pulseController;
  const _LogoMark({required this.pulseController});

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: pulseController,
      builder: (context, child) {
        return Container(
          width: 100,
          height: 100,
          decoration: BoxDecoration(
            gradient: AppColors.brandGradient,
            shape: BoxShape.circle,
            boxShadow: [
              BoxShadow(
                color: AppColors.brand500.withOpacity(
                    0.3 + 0.2 * pulseController.value),
                blurRadius: 30 + 10 * pulseController.value,
                spreadRadius: 5 * pulseController.value,
              ),
            ],
          ),
          child: const Icon(Icons.bolt_rounded, color: Colors.white, size: 52),
        );
      },
    );
  }
}

class _GlowOrb extends StatelessWidget {
  final double size;
  final Color color;
  final double opacity;
  const _GlowOrb({required this.size, required this.color, required this.opacity});

  @override
  Widget build(BuildContext context) {
    return Container(
      width: size,
      height: size,
      decoration: BoxDecoration(
        shape: BoxShape.circle,
        gradient: RadialGradient(
          colors: [color.withOpacity(opacity), Colors.transparent],
        ),
      ),
    );
  }
}

class _LoadingDots extends StatefulWidget {
  @override
  State<_LoadingDots> createState() => _LoadingDotsState();
}

class _LoadingDotsState extends State<_LoadingDots>
    with SingleTickerProviderStateMixin {
  late AnimationController _controller;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 1200),
    )..repeat();
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: List.generate(3, (i) {
        return AnimatedBuilder(
          animation: _controller,
          builder: (context, child) {
            final offset = ((_controller.value * 3) - i).clamp(0.0, 1.0);
            final bounce = offset < 0.5 ? offset * 2 : (1 - offset) * 2;
            return Container(
              margin: const EdgeInsets.symmetric(horizontal: 4),
              width: 8,
              height: 8,
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                color: Colors.white.withOpacity(0.3 + 0.7 * bounce),
              ),
            );
          },
        );
      }),
    );
  }
}
