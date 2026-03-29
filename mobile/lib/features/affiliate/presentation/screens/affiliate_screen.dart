import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_animate/flutter_animate.dart';
import 'package:share_plus/share_plus.dart';
import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_text_styles.dart';
import '../../../../core/api/api_client.dart';
import '../../../../core/auth/auth_provider.dart';
import '../../../../shared/widgets/app_widgets.dart';

final affiliateDashboardProvider = FutureProvider.autoDispose<Map<String, dynamic>>((ref) async {
  final api = ref.watch(apiClientProvider);
  return api.getAffiliateDashboard();
});

class AffiliateScreen extends ConsumerStatefulWidget {
  const AffiliateScreen({super.key});

  @override
  ConsumerState<AffiliateScreen> createState() => _AffiliateScreenState();
}

class _AffiliateScreenState extends ConsumerState<AffiliateScreen> {
  bool _isRegistering = false;
  bool _isRequestingPayout = false;
  final _bankNameCtrl = TextEditingController();
  final _accountNoCtrl = TextEditingController();

  @override
  void dispose() {
    _bankNameCtrl.dispose();
    _accountNoCtrl.dispose();
    super.dispose();
  }

  Future<void> _register() async {
    setState(() => _isRegistering = true);
    try {
      await ref.read(apiClientProvider).registerAffiliate();
      ref.invalidate(affiliateDashboardProvider);
      await ref.read(authProvider.notifier).refreshUser();
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(e.toString()), backgroundColor: AppColors.error500),
        );
      }
    } finally {
      if (mounted) setState(() => _isRegistering = false);
    }
  }

  Future<void> _requestPayout(double balance) async {
    if (balance < 1000) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Minimum payout is ₦1,000')),
      );
      return;
    }

    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      builder: (_) => _PayoutSheet(
        balance: balance,
        onRequest: (bank, account) async {
          Navigator.pop(context);
          setState(() => _isRequestingPayout = true);
          try {
            await ref.read(apiClientProvider).requestPayout({
              'amount': balance,
              'bank_name': bank,
              'account_number': account,
            });
            ref.invalidate(affiliateDashboardProvider);
            if (mounted) {
              ScaffoldMessenger.of(context).showSnackBar(
                const SnackBar(content: Text('✅ Payout request submitted!')),
              );
            }
          } catch (e) {
            if (mounted) {
              ScaffoldMessenger.of(context).showSnackBar(
                SnackBar(content: Text(e.toString()), backgroundColor: AppColors.error500),
              );
            }
          } finally {
            if (mounted) setState(() => _isRequestingPayout = false);
          }
        },
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final user = ref.watch(currentUserProvider);
    final isAffiliate = user?.isAffiliate ?? false;
    final isDark = Theme.of(context).brightness == Brightness.dark;

    return Scaffold(
      backgroundColor: isDark ? AppColors.darkBgPrimary : AppColors.bgSecondary,
      appBar: AppBar(
        title: const Text('Affiliate Program'),
        backgroundColor: isDark ? AppColors.darkBgPrimary : AppColors.bgPrimary,
      ),
      body: isAffiliate
          ? _AffiliateDashboard(onRequestPayout: _requestPayout)
          : _AffiliateJoin(onJoin: _register, isLoading: _isRegistering),
    );
  }
}

// ─── Join screen ──────────────────────────────────────────────────────────────
class _AffiliateJoin extends StatelessWidget {
  final VoidCallback onJoin;
  final bool isLoading;

  const _AffiliateJoin({required this.onJoin, required this.isLoading});

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(24),
      child: Column(
        children: [
          // Hero
          Container(
            width: double.infinity,
            padding: const EdgeInsets.all(28),
            decoration: BoxDecoration(
              gradient: const LinearGradient(
                colors: [AppColors.brand800, AppColors.brand600],
                begin: Alignment.topLeft,
                end: Alignment.bottomRight,
              ),
              borderRadius: BorderRadius.circular(20),
            ),
            child: Column(
              children: [
                const Text('🤝', style: TextStyle(fontSize: 52)),
                const SizedBox(height: 12),
                Text(
                  'Earn While You Share',
                  style: AppTextStyles.headingXl.copyWith(
                    color: Colors.white,
                    fontWeight: FontWeight.w800,
                  ),
                  textAlign: TextAlign.center,
                ),
                const SizedBox(height: 8),
                Text(
                  'Get 5% commission on every recharge made by your referrals. Unlimited earnings!',
                  style: AppTextStyles.bodyLg.copyWith(
                    color: Colors.white.withOpacity(0.8),
                  ),
                  textAlign: TextAlign.center,
                ),
              ],
            ),
          ).animate().fadeIn().slideY(begin: 0.1, end: 0),

          const SizedBox(height: 24),

          // How it works
          AppCard(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text('How It Works', style: AppTextStyles.headingMd),
                const SizedBox(height: 16),
                ...[
                  ('1', 'Join the program below', AppColors.brand500),
                  ('2', 'Share your referral link', AppColors.gold500),
                  ('3', 'Friends recharge using your link', AppColors.success500),
                  ('4', 'Earn 5% of every recharge they make', AppColors.warning500),
                ].map((s) => Padding(
                  padding: const EdgeInsets.only(bottom: 12),
                  child: Row(
                    children: [
                      Container(
                        width: 28,
                        height: 28,
                        decoration: BoxDecoration(
                          color: s.$3.withOpacity(0.15),
                          shape: BoxShape.circle,
                        ),
                        child: Center(
                          child: Text(s.$1, style: AppTextStyles.labelMd.copyWith(color: s.$3, fontWeight: FontWeight.w700)),
                        ),
                      ),
                      const SizedBox(width: 12),
                      Expanded(child: Text(s.$2, style: AppTextStyles.bodyMd)),
                    ],
                  ),
                )),
              ],
            ),
          ).animate(delay: 150.ms).fadeIn(),

          const SizedBox(height: 24),

          AppGradientButton(
            label: 'Join Affiliate Program',
            onPressed: onJoin,
            isLoading: isLoading,
            icon: const Icon(Icons.group_add_rounded, color: Colors.white, size: 20),
          ).animate(delay: 300.ms).fadeIn(),
        ],
      ),
    );
  }
}

// ─── Dashboard ────────────────────────────────────────────────────────────────
class _AffiliateDashboard extends ConsumerWidget {
  final Future<void> Function(double) onRequestPayout;

  const _AffiliateDashboard({required this.onRequestPayout});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final user = ref.watch(currentUserProvider);
    final dashAsync = ref.watch(affiliateDashboardProvider);
    final referralCode = user?.referralCode ?? '';
    final referralLink = 'https://rechargemax.com/r/$referralCode';

    return dashAsync.when(
      data: (data) {
        final totalEarnings = (data['total_earnings'] ?? data['total_commission'] ?? 0) as num;
        final pendingEarnings = (data['pending_earnings'] ?? data['pending_commission'] ?? 0) as num;
        final totalReferrals = (data['total_referrals'] ?? data['referral_count'] ?? 0) as num;
        final balance = (data['balance'] ?? data['wallet_balance'] ?? 0) as num;

        return RefreshIndicator(
          onRefresh: () async => ref.invalidate(affiliateDashboardProvider),
          child: SingleChildScrollView(
            physics: const AlwaysScrollableScrollPhysics(),
            padding: const EdgeInsets.all(16),
            child: Column(
              children: [
                // Stats grid
                Row(
                  children: [
                    Expanded(child: _StatCard(
                      icon: '💰',
                      label: 'Total Earned',
                      value: totalEarnings.toInt().toNaira(),
                      color: AppColors.gold500,
                    )),
                    const SizedBox(width: 12),
                    Expanded(child: _StatCard(
                      icon: '👥',
                      label: 'Referrals',
                      value: totalReferrals.toString(),
                      color: AppColors.brand500,
                    )),
                  ],
                ),
                const SizedBox(height: 12),
                Row(
                  children: [
                    Expanded(child: _StatCard(
                      icon: '⏳',
                      label: 'Pending',
                      value: pendingEarnings.toInt().toNaira(),
                      color: AppColors.warning500,
                    )),
                    const SizedBox(width: 12),
                    Expanded(child: _StatCard(
                      icon: '💳',
                      label: 'Available',
                      value: balance.toInt().toNaira(),
                      color: AppColors.success500,
                    )),
                  ],
                ),

                const SizedBox(height: 20),

                // Referral link card
                AppCard(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text('Your Referral Link', style: AppTextStyles.headingMd),
                      const SizedBox(height: 12),
                      Container(
                        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
                        decoration: BoxDecoration(
                          color: AppColors.brand50,
                          borderRadius: BorderRadius.circular(10),
                          border: Border.all(color: AppColors.brand200),
                        ),
                        child: Row(
                          children: [
                            Expanded(
                              child: Text(
                                referralLink,
                                style: AppTextStyles.bodyMd.copyWith(
                                  color: AppColors.brand700,
                                  fontWeight: FontWeight.w500,
                                ),
                                overflow: TextOverflow.ellipsis,
                              ),
                            ),
                            GestureDetector(
                              onTap: () {
                                Clipboard.copy(referralLink);
                                ScaffoldMessenger.of(context).showSnackBar(
                                  const SnackBar(content: Text('Copied!')),
                                );
                              },
                              child: const Icon(Icons.copy_rounded, color: AppColors.brand500, size: 20),
                            ),
                          ],
                        ),
                      ),
                      const SizedBox(height: 12),
                      Row(
                        children: [
                          Expanded(
                            child: OutlinedButton.icon(
                              onPressed: () => Share.share(
                                'Recharge with RechargeMax and win amazing prizes!\n$referralLink',
                              ),
                              icon: const Icon(Icons.share_rounded, size: 18),
                              label: const Text('Share Link'),
                            ),
                          ),
                          const SizedBox(width: 10),
                          Expanded(
                            child: ElevatedButton.icon(
                              onPressed: balance >= 1000 ? () => onRequestPayout(balance.toDouble()) : null,
                              icon: const Icon(Icons.account_balance_rounded, size: 18),
                              label: const Text('Withdraw'),
                            ),
                          ),
                        ],
                      ),
                    ],
                  ),
                ).animate(delay: 150.ms).fadeIn(),
              ],
            ),
          ),
        );
      },
      loading: () => const Center(child: CircularProgressIndicator()),
      error: (_, __) => AppEmptyState(icon: Icons.error_outline, title: 'Could not load affiliate data'),
    );
  }
}

class _StatCard extends StatelessWidget {
  final String icon;
  final String label;
  final String value;
  final Color color;

  const _StatCard({required this.icon, required this.label, required this.value, required this.color});

  @override
  Widget build(BuildContext context) {
    return AppCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(icon, style: const TextStyle(fontSize: 24)),
          const SizedBox(height: 8),
          Text(value, style: AppTextStyles.headingMd.copyWith(color: color, fontWeight: FontWeight.w800)),
          Text(label, style: AppTextStyles.bodySm.copyWith(color: AppColors.textTertiary)),
        ],
      ),
    );
  }
}

class _PayoutSheet extends StatefulWidget {
  final double balance;
  final void Function(String bank, String account) onRequest;

  const _PayoutSheet({required this.balance, required this.onRequest});

  @override
  State<_PayoutSheet> createState() => _PayoutSheetState();
}

class _PayoutSheetState extends State<_PayoutSheet> {
  final _bankCtrl = TextEditingController();
  final _accountCtrl = TextEditingController();
  final _formKey = GlobalKey<FormState>();

  @override
  void dispose() {
    _bankCtrl.dispose();
    _accountCtrl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: EdgeInsets.only(bottom: MediaQuery.of(context).viewInsets.bottom),
      child: Container(
        padding: const EdgeInsets.all(24),
        child: Form(
          key: _formKey,
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text('Request Payout', style: AppTextStyles.headingXl),
              const SizedBox(height: 4),
              Text('Amount: ${widget.balance.toInt().toNaira()}', style: AppTextStyles.bodyLg.copyWith(color: AppColors.textSecondary)),
              const SizedBox(height: 20),
              TextFormField(
                controller: _bankCtrl,
                decoration: const InputDecoration(labelText: 'Bank Name', prefixIcon: Icon(Icons.account_balance_outlined)),
                validator: (v) => v?.isEmpty == true ? 'Required' : null,
              ),
              const SizedBox(height: 12),
              TextFormField(
                controller: _accountCtrl,
                keyboardType: TextInputType.number,
                decoration: const InputDecoration(labelText: 'Account Number', prefixIcon: Icon(Icons.credit_card_rounded)),
                validator: (v) => v?.length != 10 ? 'Enter 10-digit account number' : null,
              ),
              const SizedBox(height: 20),
              AppGradientButton(
                label: 'Submit Payout Request',
                onPressed: () {
                  if (_formKey.currentState?.validate() == true) {
                    widget.onRequest(_bankCtrl.text, _accountCtrl.text);
                  }
                },
              ),
            ],
          ),
        ),
      ),
    );
  }
}
