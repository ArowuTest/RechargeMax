import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:flutter_animate/flutter_animate.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_text_styles.dart';
import '../../../../shared/constants/app_constants.dart';
import '../../../../shared/widgets/app_widgets.dart';

class OnboardingScreen extends ConsumerStatefulWidget {
  const OnboardingScreen({super.key});

  @override
  ConsumerState<OnboardingScreen> createState() => _OnboardingScreenState();
}

class _OnboardingScreenState extends ConsumerState<OnboardingScreen> {
  late PageController _pageController;
  int _currentPage = 0;

  static const _pages = [
    _OnboardingData(
      emoji: '⚡',
      title: 'Recharge in\nSeconds',
      description:
          'Buy airtime and data for any Nigerian network instantly. MTN, Glo, Airtel, 9Mobile — all covered.',
      gradientStart: AppColors.brand900,
      gradientEnd: AppColors.brand950,
      accentColor: AppColors.brand400,
    ),
    _OnboardingData(
      emoji: '🎰',
      title: 'Spin & Win\nInstantly',
      description:
          'Recharge ₦1,000 or more to unlock the spin wheel. Win airtime, data bundles, and bonus entries!',
      gradientStart: Color(0xFF1A0533),
      gradientEnd: Color(0xFF2D1B69),
      accentColor: AppColors.gold500,
    ),
    _OnboardingData(
      emoji: '🏆',
      title: 'Win Life-Changing\nPrizes Daily',
      description:
          'Every ₦200 earns a draw entry. Daily jackpots, weekly mega draws — your chance to win big.',
      gradientStart: Color(0xFF1A0A00),
      gradientEnd: Color(0xFF2D1B69),
      accentColor: AppColors.success500,
    ),
  ];

  @override
  void initState() {
    super.initState();
    _pageController = PageController();
  }

  @override
  void dispose() {
    _pageController.dispose();
    super.dispose();
  }

  Future<void> _markOnboardingDone() async {
    const storage = FlutterSecureStorage();
    await storage.write(key: AppConstants.onboardingDoneKey, value: 'true');
  }

  void _next() {
    if (_currentPage < _pages.length - 1) {
      _pageController.nextPage(
        duration: const Duration(milliseconds: 400),
        curve: Curves.easeInOut,
      );
    } else {
      _getStarted();
    }
  }

  Future<void> _getStarted() async {
    await _markOnboardingDone();
    if (mounted) context.go('/login');
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Stack(
        children: [
          // Page view
          PageView.builder(
            controller: _pageController,
            onPageChanged: (i) => setState(() => _currentPage = i),
            itemCount: _pages.length,
            itemBuilder: (context, index) => _OnboardingPage(data: _pages[index]),
          ),

          // Bottom controls
          Positioned(
            left: 0,
            right: 0,
            bottom: 0,
            child: _OnboardingControls(
              currentPage: _currentPage,
              pageCount: _pages.length,
              onNext: _next,
              onSkip: _getStarted,
            ),
          ),
        ],
      ),
    );
  }
}

class _OnboardingPage extends StatelessWidget {
  final _OnboardingData data;
  const _OnboardingPage({required this.data});

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: BoxDecoration(
        gradient: LinearGradient(
          begin: Alignment.topCenter,
          end: Alignment.bottomCenter,
          colors: [data.gradientStart, data.gradientEnd],
        ),
      ),
      child: SafeArea(
        child: Padding(
          padding: const EdgeInsets.fromLTRB(28, 40, 28, 120),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Illustration area
              Expanded(
                child: Center(
                  child: _IllustrationArea(data: data),
                ),
              ),

              const SizedBox(height: 40),

              // Title
              Text(
                data.title,
                style: AppTextStyles.displayMd.copyWith(
                  color: Colors.white,
                  fontWeight: FontWeight.w800,
                ),
              )
                  .animate()
                  .fadeIn(duration: 500.ms)
                  .slideX(begin: 0.2, end: 0),

              const SizedBox(height: 16),

              // Description
              Text(
                data.description,
                style: AppTextStyles.bodyXl.copyWith(
                  color: Colors.white.withValues(alpha: 0.72),
                  height: 1.6,
                ),
              )
                  .animate(delay: 150.ms)
                  .fadeIn(duration: 400.ms)
                  .slideX(begin: 0.15, end: 0),
            ],
          ),
        ),
      ),
    );
  }
}

class _IllustrationArea extends StatelessWidget {
  final _OnboardingData data;
  const _IllustrationArea({required this.data});

  @override
  Widget build(BuildContext context) {
    return Stack(
      alignment: Alignment.center,
      children: [
        // Outer ring
        Container(
          width: 260,
          height: 260,
          decoration: BoxDecoration(
            shape: BoxShape.circle,
            border: Border.all(
              color: data.accentColor.withValues(alpha: 0.12),
              width: 1,
            ),
          ),
        ),
        // Middle ring
        Container(
          width: 200,
          height: 200,
          decoration: BoxDecoration(
            shape: BoxShape.circle,
            border: Border.all(
              color: data.accentColor.withValues(alpha: 0.18),
              width: 1,
            ),
          ),
        ),
        // Inner glow circle
        Container(
          width: 150,
          height: 150,
          decoration: BoxDecoration(
            shape: BoxShape.circle,
            gradient: RadialGradient(
              colors: [
                data.accentColor.withValues(alpha: 0.2),
                data.accentColor.withValues(alpha: 0.05),
                Colors.transparent,
              ],
            ),
          ),
        ),
        // Emoji
        Text(
          data.emoji,
          style: const TextStyle(fontSize: 72),
        ).animate(onPlay: (c) => c.repeat(reverse: true))
            .scale(
          begin: const Offset(0.9, 0.9),
          end: const Offset(1.05, 1.05),
          duration: 1800.ms,
          curve: Curves.easeInOut,
        ),
      ],
    ).animate().scale(
          begin: const Offset(0.8, 0.8),
          end: const Offset(1.0, 1.0),
          duration: 600.ms,
          curve: Curves.elasticOut,
        );
  }
}

class _OnboardingControls extends StatelessWidget {
  final int currentPage;
  final int pageCount;
  final VoidCallback onNext;
  final VoidCallback onSkip;

  const _OnboardingControls({
    required this.currentPage,
    required this.pageCount,
    required this.onNext,
    required this.onSkip,
  });

  @override
  Widget build(BuildContext context) {
    final isLast = currentPage == pageCount - 1;

    return Container(
      padding: const EdgeInsets.fromLTRB(28, 20, 28, 40),
      decoration: BoxDecoration(
        gradient: LinearGradient(
          begin: Alignment.topCenter,
          end: Alignment.bottomCenter,
          colors: [
            Colors.transparent,
            AppColors.brand950.withValues(alpha: 0.95),
          ],
        ),
      ),
      child: SafeArea(
        top: false,
        child: Column(
          children: [
            // Dots
            Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: List.generate(pageCount, (i) {
                return AnimatedContainer(
                  duration: const Duration(milliseconds: 250),
                  margin: const EdgeInsets.symmetric(horizontal: 4),
                  width: i == currentPage ? 24 : 8,
                  height: 8,
                  decoration: BoxDecoration(
                    color: i == currentPage
                        ? AppColors.brand400
                        : Colors.white.withValues(alpha: 0.25),
                    borderRadius: BorderRadius.circular(4),
                  ),
                );
              }),
            ),
            const SizedBox(height: 28),

            // Buttons
            Row(
              children: [
                if (!isLast)
                  Expanded(
                    child: OutlinedButton(
                      onPressed: onSkip,
                      style: OutlinedButton.styleFrom(
                        foregroundColor: Colors.white.withValues(alpha: 0.6),
                        side: BorderSide(color: Colors.white.withValues(alpha: 0.2)),
                        padding: const EdgeInsets.symmetric(vertical: 14),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(14),
                        ),
                      ),
                      child: Text(
                        'Skip',
                        style: AppTextStyles.labelXl.copyWith(
                          color: Colors.white.withValues(alpha: 0.6),
                        ),
                      ),
                    ),
                  ),
                if (!isLast) const SizedBox(width: 12),
                Expanded(
                  flex: isLast ? 1 : 2,
                  child: AppGradientButton(
                    label: isLast ? 'Get Started' : 'Next',
                    onPressed: onNext,
                    icon: Icon(
                      isLast ? Icons.rocket_launch_rounded : Icons.arrow_forward_rounded,
                      color: Colors.white,
                      size: 18,
                    ),
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _OnboardingData {
  final String emoji;
  final String title;
  final String description;
  final Color gradientStart;
  final Color gradientEnd;
  final Color accentColor;

  const _OnboardingData({
    required this.emoji,
    required this.title,
    required this.description,
    required this.gradientStart,
    required this.gradientEnd,
    required this.accentColor,
  });
}
