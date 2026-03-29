import 'package:flutter/material.dart';
import 'package:flutter_animate/flutter_animate.dart';
import 'package:go_router/go_router.dart';
import 'package:confetti/confetti.dart';
import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_text_styles.dart';
import '../../../../shared/widgets/app_widgets.dart';

class RechargeSuccessScreen extends StatefulWidget {
  final Map<String, dynamic> data;
  const RechargeSuccessScreen({super.key, required this.data});

  @override
  State<RechargeSuccessScreen> createState() => _RechargeSuccessScreenState();
}

class _RechargeSuccessScreenState extends State<RechargeSuccessScreen> {
  late ConfettiController _confetti;

  @override
  void initState() {
    super.initState();
    _confetti = ConfettiController(duration: const Duration(seconds: 4));
    WidgetsBinding.instance.addPostFrameCallback((_) => _confetti.play());
  }

  @override
  void dispose() {
    _confetti.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final amount = (widget.data['amount'] ?? 0) as num;
    final phone = widget.data['msisdn'] ?? widget.data['phone'] ?? '';
    final entries = widget.data['draw_entries'] ?? ((amount / 200).floor());
    final spinUnlocked = widget.data['spin_unlocked'] == true || amount >= 1000;

    return Scaffold(
      body: Stack(
        alignment: Alignment.topCenter,
        children: [
          // Background gradient
          Container(
            decoration: const BoxDecoration(
              gradient: LinearGradient(
                begin: Alignment.topCenter,
                end: Alignment.bottomCenter,
                colors: [AppColors.brand950, AppColors.bgSecondary],
                stops: [0.3, 0.6],
              ),
            ),
          ),

          // Confetti
          ConfettiWidget(
            confettiController: _confetti,
            blastDirectionality: BlastDirectionality.explosive,
            shouldLoop: false,
            numberOfParticles: 40,
            colors: const [
              AppColors.brand400,
              AppColors.gold500,
              AppColors.success500,
              Colors.white,
              AppColors.brand200,
            ],
          ),

          SafeArea(
            child: Padding(
              padding: const EdgeInsets.all(24),
              child: Column(
                children: [
                  const Spacer(),

                  // Success icon
                  Container(
                    width: 100,
                    height: 100,
                    decoration: BoxDecoration(
                      gradient: AppColors.successGradient,
                      shape: BoxShape.circle,
                      boxShadow: [
                        BoxShadow(
                          color: AppColors.success500.withValues(alpha: 0.4),
                          blurRadius: 24,
                          spreadRadius: 4,
                        ),
                      ],
                    ),
                    child: const Icon(Icons.check_rounded, color: Colors.white, size: 54),
                  )
                      .animate()
                      .scale(
                        begin: const Offset(0.3, 0.3),
                        duration: 500.ms,
                        curve: Curves.elasticOut,
                      ),

                  const SizedBox(height: 28),

                  Text(
                    'Recharge Successful!',
                    style: AppTextStyles.displaySm.copyWith(
                      color: Colors.white,
                      fontWeight: FontWeight.w800,
                    ),
                    textAlign: TextAlign.center,
                  ).animate(delay: 300.ms).fadeIn().slideY(begin: 0.3, end: 0),

                  const SizedBox(height: 8),

                  Text(
                    '${amount.toInt().toNaira()} sent to $phone',
                    style: AppTextStyles.bodyXl.copyWith(
                      color: AppColors.brand200,
                    ),
                    textAlign: TextAlign.center,
                  ).animate(delay: 400.ms).fadeIn(),

                  const SizedBox(height: 40),

                  // Rewards card
                  _RewardsCard(
                    entries: entries as int,
                    spinUnlocked: spinUnlocked,
                  ).animate(delay: 500.ms).fadeIn().slideY(begin: 0.2, end: 0),

                  const Spacer(),

                  // Actions
                  Column(
                    children: [
                      if (spinUnlocked)
                        AppGradientButton(
                          label: '🎰 Spin Now!',
                          onPressed: () => context.go('/spin'),
                          gradient: const LinearGradient(
                            colors: [AppColors.gold500, AppColors.gold600],
                          ),
                        ).animate(delay: 700.ms).fadeIn().slideY(begin: 0.2, end: 0),

                      if (spinUnlocked) const SizedBox(height: 12),

                      OutlinedButton(
                        onPressed: () => context.go('/home'),
                        style: OutlinedButton.styleFrom(
                          foregroundColor: Colors.white,
                          side: BorderSide(color: Colors.white.withValues(alpha: 0.3)),
                          minimumSize: const Size(double.infinity, 52),
                          shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(14),
                          ),
                        ),
                        child: Text(
                          'Back to Home',
                          style: AppTextStyles.labelXl.copyWith(color: Colors.white),
                        ),
                      ).animate(delay: 800.ms).fadeIn(),
                    ],
                  ),

                  const SizedBox(height: 16),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _RewardsCard extends StatelessWidget {
  final int entries;
  final bool spinUnlocked;

  const _RewardsCard({required this.entries, required this.spinUnlocked});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: Colors.white.withValues(alpha: 0.08),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: Colors.white.withValues(alpha: 0.15)),
      ),
      child: Column(
        children: [
          Text(
            'You Earned',
            style: AppTextStyles.labelLg.copyWith(
              color: Colors.white.withValues(alpha: 0.7),
            ),
          ),
          const SizedBox(height: 16),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceAround,
            children: [
              _RewardItem(
                icon: '🏆',
                value: '$entries',
                label: 'Draw ${entries == 1 ? 'Entry' : 'Entries'}',
              ),
              Container(width: 1, height: 50, color: Colors.white.withValues(alpha: 0.15)),
              _RewardItem(
                icon: spinUnlocked ? '🎰' : '🔒',
                value: spinUnlocked ? 'UNLOCKED' : 'LOCKED',
                label: 'Spin Wheel',
                highlight: spinUnlocked,
              ),
            ],
          ),
        ],
      ),
    );
  }
}

class _RewardItem extends StatelessWidget {
  final String icon;
  final String value;
  final String label;
  final bool highlight;

  const _RewardItem({
    required this.icon,
    required this.value,
    required this.label,
    this.highlight = false,
  });

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Text(icon, style: const TextStyle(fontSize: 28)),
        const SizedBox(height: 8),
        Text(
          value,
          style: AppTextStyles.headingMd.copyWith(
            color: highlight ? AppColors.gold400 : Colors.white,
            fontWeight: FontWeight.w800,
          ),
        ),
        Text(
          label,
          style: AppTextStyles.bodySm.copyWith(
            color: Colors.white.withValues(alpha: 0.6),
          ),
        ),
      ],
    );
  }
}
