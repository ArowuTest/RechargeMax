import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_animate/flutter_animate.dart';
import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_text_styles.dart';
import '../../../../core/api/api_client.dart';
import '../../../../shared/widgets/app_widgets.dart';

final subscriptionStatusProvider = FutureProvider.autoDispose<Map<String, dynamic>>((ref) async {
  final api = ref.watch(apiClientProvider);
  return api.getSubscriptionStatus();
});

class SubscriptionScreen extends ConsumerStatefulWidget {
  const SubscriptionScreen({super.key});

  @override
  ConsumerState<SubscriptionScreen> createState() => _SubscriptionScreenState();
}

class _SubscriptionScreenState extends ConsumerState<SubscriptionScreen> {
  bool _isSubscribing = false;
  bool _isCancelling = false;

  Future<void> _subscribe() async {
    setState(() => _isSubscribing = true);
    try {
      final api = ref.read(apiClientProvider);
      await api.subscribe({'payment_method': 'wallet'});
      ref.invalidate(subscriptionStatusProvider);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('✅ Subscribed successfully!')),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(e.toString()), backgroundColor: AppColors.error500),
        );
      }
    } finally {
      if (mounted) setState(() => _isSubscribing = false);
    }
  }

  Future<void> _cancel() async {
    final confirm = await showDialog<bool>(
      context: context,
      builder: (_) => AlertDialog(
        title: const Text('Cancel Subscription?'),
        content: const Text('You will lose your daily draw entries. This cannot be undone.'),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context, false), child: const Text('Keep')),
          TextButton(
            onPressed: () => Navigator.pop(context, true),
            child: Text('Cancel', style: TextStyle(color: AppColors.error500)),
          ),
        ],
      ),
    );
    if (confirm != true) return;

    setState(() => _isCancelling = true);
    try {
      final api = ref.read(apiClientProvider);
      await api.cancelSubscription();
      ref.invalidate(subscriptionStatusProvider);
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(e.toString()), backgroundColor: AppColors.error500),
        );
      }
    } finally {
      if (mounted) setState(() => _isCancelling = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final statusAsync = ref.watch(subscriptionStatusProvider);
    final isDark = Theme.of(context).brightness == Brightness.dark;

    return Scaffold(
      backgroundColor: isDark ? AppColors.darkBgPrimary : AppColors.bgSecondary,
      appBar: AppBar(
        title: const Text('Daily Subscription'),
        backgroundColor: isDark ? AppColors.darkBgPrimary : AppColors.bgPrimary,
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          children: [
            // Hero card
            Container(
              width: double.infinity,
              padding: const EdgeInsets.all(24),
              decoration: BoxDecoration(
                gradient: AppColors.brandGradient,
                borderRadius: BorderRadius.circular(20),
                boxShadow: [
                  BoxShadow(
                    color: AppColors.brand600.withValues(alpha: 0.3),
                    blurRadius: 20,
                    offset: const Offset(0, 8),
                  ),
                ],
              ),
              child: Column(
                children: [
                  const Text('📅', style: TextStyle(fontSize: 44)),
                  const SizedBox(height: 12),
                  Text(
                    'Daily Subscription',
                    style: AppTextStyles.headingXl.copyWith(color: Colors.white, fontWeight: FontWeight.w800),
                  ),
                  const SizedBox(height: 6),
                  RichText(
                    textAlign: TextAlign.center,
                    text: TextSpan(
                      style: AppTextStyles.displaySm.copyWith(color: Colors.white),
                      children: [
                        const TextSpan(text: '₦20', style: TextStyle(fontWeight: FontWeight.w900, fontSize: 40)),
                        TextSpan(text: '/day', style: AppTextStyles.bodyXl.copyWith(color: Colors.white70)),
                      ],
                    ),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    'Guaranteed daily draw entries',
                    style: AppTextStyles.bodyMd.copyWith(color: Colors.white.withValues(alpha: 0.8)),
                  ),
                ],
              ),
            ).animate().fadeIn(duration: 400.ms).slideY(begin: 0.1, end: 0),

            const SizedBox(height: 20),

            // Benefits
            AppCard(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text('What you get', style: AppTextStyles.headingMd),
                  const SizedBox(height: 12),
                  ..._benefits.map((b) => _BenefitRow(icon: b.$1, text: b.$2)),
                ],
              ),
            ).animate(delay: 100.ms).fadeIn().slideY(begin: 0.1, end: 0),

            const SizedBox(height: 16),

            // Status
            statusAsync.when(
              data: (status) {
                final isActive = status['active'] == true || status['status'] == 'active';
                final expiresAt = status['expires_at'] ?? status['next_billing'] as String?;

                return Column(
                  children: [
                    // Status indicator
                    AppCard(
                      child: Row(
                        children: [
                          Container(
                            width: 48,
                            height: 48,
                            decoration: BoxDecoration(
                              color: isActive
                                  ? AppColors.success500.withValues(alpha: 0.1)
                                  : AppColors.error500.withValues(alpha: 0.1),
                              borderRadius: BorderRadius.circular(12),
                            ),
                            child: Icon(
                              isActive ? Icons.check_circle_rounded : Icons.cancel_rounded,
                              color: isActive ? AppColors.success500 : AppColors.error500,
                              size: 26,
                            ),
                          ),
                          const SizedBox(width: 14),
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(
                                  isActive ? 'Subscription Active' : 'Not Subscribed',
                                  style: AppTextStyles.labelLg.copyWith(fontWeight: FontWeight.w700),
                                ),
                                if (expiresAt != null)
                                  Text(
                                    'Renews $expiresAt',
                                    style: AppTextStyles.bodySm.copyWith(color: AppColors.textTertiary),
                                  ),
                              ],
                            ),
                          ),
                          Container(
                            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
                            decoration: BoxDecoration(
                              color: isActive
                                  ? AppColors.success500.withValues(alpha: 0.1)
                                  : AppColors.error500.withValues(alpha: 0.1),
                              borderRadius: BorderRadius.circular(20),
                            ),
                            child: Text(
                              isActive ? 'ACTIVE' : 'INACTIVE',
                              style: AppTextStyles.labelXs.copyWith(
                                color: isActive ? AppColors.success500 : AppColors.error500,
                                fontWeight: FontWeight.w700,
                              ),
                            ),
                          ),
                        ],
                      ),
                    ),

                    const SizedBox(height: 16),

                    if (!isActive)
                      AppGradientButton(
                        label: 'Subscribe for ₦20/day',
                        onPressed: _subscribe,
                        isLoading: _isSubscribing,
                        icon: const Icon(Icons.calendar_today_rounded, color: Colors.white, size: 18),
                      )
                    else
                      OutlinedButton.icon(
                        onPressed: _isCancelling ? null : _cancel,
                        style: OutlinedButton.styleFrom(
                          foregroundColor: AppColors.error500,
                          side: const BorderSide(color: AppColors.error500),
                          minimumSize: const Size(double.infinity, 48),
                        ),
                        icon: _isCancelling
                            ? const SizedBox(
                                width: 16, height: 16,
                                child: CircularProgressIndicator(strokeWidth: 2, color: AppColors.error500),
                              )
                            : const Icon(Icons.cancel_outlined, size: 18),
                        label: const Text('Cancel Subscription'),
                      ),
                  ],
                );
              },
              loading: () => const Center(child: CircularProgressIndicator()),
              error: (_, __) => AppEmptyState(icon: Icons.error_outline, title: 'Could not load status'),
            ).animate(delay: 200.ms).fadeIn(),
          ],
        ),
      ),
    );
  }

  static const _benefits = [
    ('🏆', '1 guaranteed draw entry every day'),
    ('⚡', 'Boost your winning odds significantly'),
    ('💫', 'Access to subscriber-only prize draws'),
    ('📅', 'Auto-renews daily — cancel anytime'),
    ('💰', 'Cost: just ₦20 per day'),
  ];
}

class _BenefitRow extends StatelessWidget {
  final String icon;
  final String text;
  const _BenefitRow({required this.icon, required this.text});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 10),
      child: Row(
        children: [
          Text(icon, style: const TextStyle(fontSize: 20)),
          const SizedBox(width: 12),
          Expanded(
            child: Text(text, style: AppTextStyles.bodyMd),
          ),
        ],
      ),
    );
  }
}
