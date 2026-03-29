import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:flutter_animate/flutter_animate.dart';
import 'package:shimmer/shimmer.dart';
import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_text_styles.dart';
import '../../../../core/auth/auth_provider.dart';
import '../../../../core/api/api_client.dart';
import '../../../../shared/widgets/app_widgets.dart';

// ─── Dashboard data provider ──────────────────────────────────────────────────
final dashboardProvider = FutureProvider<Map<String, dynamic>>((ref) async {
  final api = ref.watch(apiClientProvider);
  return api.getDashboard();
});

class HomeScreen extends ConsumerWidget {
  const HomeScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final user = ref.watch(currentUserProvider);
    final dashboardAsync = ref.watch(dashboardProvider);
    final isDark = Theme.of(context).brightness == Brightness.dark;

    return Scaffold(
      backgroundColor: isDark ? AppColors.darkBgPrimary : AppColors.bgSecondary,
      body: RefreshIndicator(
        onRefresh: () async {
          ref.invalidate(dashboardProvider);
          await ref.read(authProvider.notifier).refreshUser();
        },
        color: AppColors.brand500,
        child: CustomScrollView(
          slivers: [
            // ─── Header / Hero ─────────────────────────────────────────────
            SliverToBoxAdapter(
              child: _HomeHero(user: user, dashboardAsync: dashboardAsync),
            ),

            // ─── Quick Actions ─────────────────────────────────────────────
            SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.fromLTRB(16, 20, 16, 0),
                child: _QuickActions(),
              ),
            ),

            // ─── Active Draw Banner ────────────────────────────────────────
            SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.fromLTRB(16, 16, 16, 0),
                child: dashboardAsync.when(
                  data: (data) => _DrawBanner(data: data),
                  loading: () => _ShimmerCard(height: 120),
                  error: (_, __) => const SizedBox.shrink(),
                ),
              ),
            ),

            // ─── Recent Winners ────────────────────────────────────────────
            SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.fromLTRB(16, 20, 16, 0),
                child: SectionHeader(
                  title: 'Recent Winners',
                  actionLabel: 'See all',
                  onAction: () => context.go('/draws'),
                ),
              ),
            ),

            SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.fromLTRB(0, 12, 0, 0),
                child: dashboardAsync.when(
                  data: (data) => _WinnersCarousel(winners: data['recent_winners'] ?? []),
                  loading: () => _ShimmerCard(height: 120),
                  error: (_, __) => const SizedBox.shrink(),
                ),
              ),
            ),

            // ─── Recent Transactions ───────────────────────────────────────
            SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.fromLTRB(16, 20, 16, 0),
                child: SectionHeader(
                  title: 'Recent Recharges',
                  actionLabel: 'History',
                  onAction: () => context.go('/profile/history'),
                ),
              ),
            ),

            SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.fromLTRB(16, 12, 16, 20),
                child: dashboardAsync.when(
                  data: (data) {
                    final txns = (data['recent_transactions'] ?? []) as List;
                    if (txns.isEmpty) {
                      return AppEmptyState(
                        icon: Icons.receipt_long_rounded,
                        title: 'No recharges yet',
                        subtitle: 'Your transaction history will appear here',
                        actionLabel: 'Recharge Now',
                        onAction: () => context.go('/recharge'),
                      );
                    }
                    return _TransactionList(transactions: txns);
                  },
                  loading: () => _ShimmerCard(height: 240),
                  error: (_, __) => const SizedBox.shrink(),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

// ─── Hero Section ─────────────────────────────────────────────────────────────
class _HomeHero extends StatelessWidget {
  final UserProfile? user;
  final AsyncValue<Map<String, dynamic>> dashboardAsync;

  const _HomeHero({required this.user, required this.dashboardAsync});

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: const BoxDecoration(
        gradient: AppColors.heroGradient,
        borderRadius: BorderRadius.vertical(bottom: Radius.circular(28)),
      ),
      child: Stack(
        children: [
          // Radial glow
          Positioned.fill(
            child: Container(
              decoration: const BoxDecoration(
                gradient: AppColors.heroRadialGlow,
                borderRadius: BorderRadius.vertical(bottom: Radius.circular(28)),
              ),
            ),
          ),

          SafeArea(
            bottom: false,
            child: Padding(
              padding: const EdgeInsets.fromLTRB(20, 16, 20, 28),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Top bar: greeting + avatar + notification
                  Row(
                    children: [
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              _greeting(),
                              style: AppTextStyles.bodyMd.copyWith(
                                color: AppColors.brand200,
                              ),
                            ),
                            Text(
                              user?.displayName ?? 'Welcome back!',
                              style: AppTextStyles.headingXl.copyWith(
                                color: Colors.white,
                                fontWeight: FontWeight.w700,
                              ),
                              maxLines: 1,
                              overflow: TextOverflow.ellipsis,
                            ),
                          ],
                        ),
                      ),
                      GestureDetector(
                        onTap: () => context.go('/profile'),
                        child: CircleAvatar(
                          radius: 20,
                          backgroundColor: AppColors.brand500.withValues(alpha: 0.2),
                          child: Text(
                            user?.initials ?? '?',
                            style: AppTextStyles.labelMd.copyWith(
                              color: Colors.white,
                              fontWeight: FontWeight.w700,
                            ),
                          ),
                        ),
                      ),
                    ],
                  ).animate().fadeIn(duration: 400.ms),

                  const SizedBox(height: 24),

                  // Points + Tier card
                  _PointsCard(user: user, dashboardAsync: dashboardAsync),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  String _greeting() {
    final hour = DateTime.now().hour;
    if (hour < 12) return 'Good morning ☀️';
    if (hour < 17) return 'Good afternoon 👋';
    return 'Good evening 🌙';
  }
}

class _PointsCard extends StatelessWidget {
  final UserProfile? user;
  final AsyncValue<Map<String, dynamic>> dashboardAsync;

  const _PointsCard({required this.user, required this.dashboardAsync});

  @override
  Widget build(BuildContext context) {
    return GlassCard(
      borderRadius: 20,
      backgroundOpacity: 0.12,
      padding: const EdgeInsets.all(20),
      child: Row(
        children: [
          // Points
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'Your Points',
                  style: AppTextStyles.labelMd.copyWith(
                    color: Colors.white.withValues(alpha: 0.7),
                  ),
                ),
                const SizedBox(height: 6),
                PointsDisplay(points: user?.points ?? 0),
                const SizedBox(height: 4),
                Text(
                  'Every ₦200 = 1 entry',
                  style: AppTextStyles.bodySm.copyWith(
                    color: Colors.white.withValues(alpha: 0.5),
                  ),
                ),
              ],
            ),
          ),

          // Divider
          Container(
            width: 1,
            height: 60,
            color: Colors.white.withValues(alpha: 0.15),
          ),

          const SizedBox(width: 20),

          // Tier
          Column(
            crossAxisAlignment: CrossAxisAlignment.end,
            children: [
              TierBadge(tier: user?.tier ?? 'BRONZE', large: true),
              const SizedBox(height: 8),
              dashboardAsync.when(
                data: (data) {
                  final drawEntries = (data['draw_entries'] ?? data['total_entries'] ?? 0) as int;
                  return Text(
                    '$drawEntries entries',
                    style: AppTextStyles.labelLg.copyWith(
                      color: Colors.white,
                      fontWeight: FontWeight.w700,
                    ),
                  );
                },
                loading: () => Container(width: 60, height: 16, color: Colors.white24),
                error: (_, __) => const SizedBox.shrink(),
              ),
              Text(
                'draw entries',
                style: AppTextStyles.bodySm.copyWith(
                  color: Colors.white.withValues(alpha: 0.5),
                ),
              ),
            ],
          ),
        ],
      ),
    ).animate(delay: 150.ms).fadeIn().slideY(begin: 0.15, end: 0);
  }
}

// ─── Quick Actions ────────────────────────────────────────────────────────────
class _QuickActions extends StatelessWidget {
  const _QuickActions();

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          'Quick Actions',
          style: AppTextStyles.headingMd.copyWith(
            color: Theme.of(context).brightness == Brightness.dark
                ? AppColors.darkTextPrimary
                : AppColors.textPrimary,
          ),
        ),
        const SizedBox(height: 14),
        Row(
          children: [
            Expanded(
              child: _QuickActionCard(
                icon: Icons.bolt_rounded,
                label: 'Airtime',
                sublabel: 'Instant top-up',
                color: AppColors.brand500,
                onTap: () => context.go('/recharge'),
              ),
            ),
            const SizedBox(width: 10),
            Expanded(
              child: _QuickActionCard(
                icon: Icons.wifi_rounded,
                label: 'Data',
                sublabel: 'Buy bundles',
                color: AppColors.success500,
                onTap: () => context.go('/recharge'),
              ),
            ),
            const SizedBox(width: 10),
            Expanded(
              child: _QuickActionCard(
                icon: Icons.casino_rounded,
                label: 'Spin',
                sublabel: 'Win prizes',
                color: AppColors.gold500,
                onTap: () => context.go('/spin'),
              ),
            ),
            const SizedBox(width: 10),
            Expanded(
              child: _QuickActionCard(
                icon: Icons.emoji_events_rounded,
                label: 'Draws',
                sublabel: 'Enter now',
                color: AppColors.warning500,
                onTap: () => context.go('/draws'),
              ),
            ),
          ],
        ),
      ],
    ).animate(delay: 200.ms).fadeIn().slideY(begin: 0.1, end: 0);
  }
}

class _QuickActionCard extends StatelessWidget {
  final IconData icon;
  final String label;
  final String sublabel;
  final Color color;
  final VoidCallback onTap;

  const _QuickActionCard({
    required this.icon,
    required this.label,
    required this.sublabel,
    required this.color,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;

    return GestureDetector(
      onTap: onTap,
      child: AppCard(
        padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 14),
        borderRadius: 14,
        child: Column(
          children: [
            Container(
              width: 44,
              height: 44,
              decoration: BoxDecoration(
                color: color.withValues(alpha: 0.12),
                borderRadius: BorderRadius.circular(12),
              ),
              child: Icon(icon, color: color, size: 22),
            ),
            const SizedBox(height: 8),
            Text(
              label,
              style: AppTextStyles.labelMd.copyWith(
                color: isDark ? AppColors.darkTextPrimary : AppColors.textPrimary,
                fontWeight: FontWeight.w600,
              ),
              textAlign: TextAlign.center,
            ),
            Text(
              sublabel,
              style: AppTextStyles.labelXs.copyWith(
                color: isDark ? AppColors.darkTextTertiary : AppColors.textTertiary,
              ),
              textAlign: TextAlign.center,
            ),
          ],
        ),
      ),
    );
  }
}

// ─── Active Draw Banner ───────────────────────────────────────────────────────
class _DrawBanner extends StatefulWidget {
  final Map<String, dynamic> data;
  const _DrawBanner({required this.data});

  @override
  State<_DrawBanner> createState() => _DrawBannerState();
}

class _DrawBannerState extends State<_DrawBanner> {
  late Duration _remaining;
  late DateTime _end;

  @override
  void initState() {
    super.initState();
    // Parse end time from data, default to midnight tonight
    final draws = (widget.data['active_draws'] ?? []) as List;
    if (draws.isNotEmpty) {
      try {
        _end = DateTime.parse(draws[0]['draw_date'] ?? draws[0]['end_date'] ?? '');
      } catch (_) {
        _end = DateTime.now().add(const Duration(hours: 20));
      }
    } else {
      _end = DateTime.now().add(const Duration(hours: 20));
    }
    _updateRemaining();
    _startTimer();
  }

  void _updateRemaining() {
    _remaining = _end.difference(DateTime.now());
    if (_remaining.isNegative) _remaining = Duration.zero;
  }

  void _startTimer() {
    Future.delayed(const Duration(seconds: 1), () {
      if (mounted) {
        setState(_updateRemaining);
        _startTimer();
      }
    });
  }

  String _pad(int n) => n.toString().padLeft(2, '0');

  @override
  Widget build(BuildContext context) {
    final h = _remaining.inHours;
    final m = _remaining.inMinutes % 60;
    final s = _remaining.inSeconds % 60;

    return Container(
      decoration: BoxDecoration(
        gradient: const LinearGradient(
          colors: [AppColors.gold500, AppColors.gold600],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        borderRadius: BorderRadius.circular(16),
        boxShadow: [
          BoxShadow(
            color: AppColors.gold500.withValues(alpha: 0.3),
            blurRadius: 16,
            offset: const Offset(0, 6),
          ),
        ],
      ),
      padding: const EdgeInsets.all(16),
      child: Row(
        children: [
          const Text('🏆', style: TextStyle(fontSize: 36)),
          const SizedBox(width: 14),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  "Today's Prize Pool",
                  style: AppTextStyles.labelMd.copyWith(
                    color: AppColors.brand950.withValues(alpha: 0.7),
                  ),
                ),
                Text(
                  '₦500,000',
                  style: AppTextStyles.headingXl.copyWith(
                    color: AppColors.brand950,
                    fontWeight: FontWeight.w800,
                  ),
                ),
              ],
            ),
          ),
          Column(
            crossAxisAlignment: CrossAxisAlignment.end,
            children: [
              Text(
                'Draws in',
                style: AppTextStyles.labelSm.copyWith(
                  color: AppColors.brand950.withValues(alpha: 0.6),
                ),
              ),
              const SizedBox(height: 4),
              Text(
                '${_pad(h)}:${_pad(m)}:${_pad(s)}',
                style: AppTextStyles.headingMd.copyWith(
                  color: AppColors.brand950,
                  fontWeight: FontWeight.w800,
                  letterSpacing: 2,
                  fontFeatures: [const FontFeature.tabularFigures()],
                ),
              ),
            ],
          ),
        ],
      ),
    ).animate(delay: 250.ms).fadeIn().slideY(begin: 0.1, end: 0);
  }
}

// ─── Winners Carousel ─────────────────────────────────────────────────────────
class _WinnersCarousel extends StatelessWidget {
  final List winners;
  const _WinnersCarousel({required this.winners});

  @override
  Widget build(BuildContext context) {
    if (winners.isEmpty) {
      return Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16),
        child: AppEmptyState(
          icon: Icons.emoji_events_outlined,
          title: 'No winners yet',
          subtitle: 'Be the first to win today!',
        ),
      );
    }

    return SizedBox(
      height: 110,
      child: ListView.builder(
        scrollDirection: Axis.horizontal,
        padding: const EdgeInsets.symmetric(horizontal: 16),
        itemCount: winners.length,
        itemBuilder: (context, i) {
          final w = winners[i] as Map<String, dynamic>;
          return _WinnerCard(winner: w);
        },
      ),
    );
  }
}

class _WinnerCard extends StatelessWidget {
  final Map<String, dynamic> winner;
  const _WinnerCard({required this.winner});

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final name = (winner['name'] ?? winner['msisdn'] ?? 'Winner') as String;
    final prize = (winner['prize_amount'] ?? winner['prize'] ?? 0) as num;
    final network = (winner['network'] ?? '') as String;

    return Container(
      width: 140,
      margin: const EdgeInsets.only(right: 10),
      child: AppCard(
        padding: const EdgeInsets.all(12),
        borderRadius: 14,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                const Text('🏆', style: TextStyle(fontSize: 18)),
                const SizedBox(width: 6),
                Expanded(
                  child: Text(
                    name.length > 12 ? '${name.substring(0, 12)}...' : name,
                    style: AppTextStyles.labelMd.copyWith(
                      color: isDark ? AppColors.darkTextPrimary : AppColors.textPrimary,
                      fontWeight: FontWeight.w600,
                    ),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 8),
            Text(
              prize.toInt().toNaira(),
              style: AppTextStyles.headingMd.copyWith(
                color: AppColors.gold500,
                fontWeight: FontWeight.w800,
              ),
            ),
            if (network.isNotEmpty) ...[
              const SizedBox(height: 4),
              Text(
                network.toUpperCase(),
                style: AppTextStyles.labelXs.copyWith(
                  color: isDark ? AppColors.darkTextTertiary : AppColors.textTertiary,
                ),
              ),
            ],
          ],
        ),
      ),
    ).animate(delay: Duration(milliseconds: 100 * 0)).fadeIn().slideX(begin: 0.2, end: 0);
  }
}

// ─── Transaction List ─────────────────────────────────────────────────────────
class _TransactionList extends StatelessWidget {
  final List transactions;
  const _TransactionList({required this.transactions});

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final items = transactions.take(5).toList();

    return Column(
      children: items.asMap().entries.map((e) {
        final txn = e.value as Map<String, dynamic>;
        final amount = (txn['amount'] ?? 0) as num;
        final network = (txn['network'] ?? '') as String;
        final type = (txn['recharge_type'] ?? txn['type'] ?? 'Airtime') as String;
        final status = (txn['status'] ?? 'completed') as String;
        final createdAt = txn['created_at'] as String?;

        return AppCard(
          margin: const EdgeInsets.only(bottom: 10),
          child: Row(
            children: [
              Container(
                width: 44,
                height: 44,
                decoration: BoxDecoration(
                  color: AppColors.brand500.withValues(alpha: 0.1),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: const Icon(Icons.bolt_rounded, color: AppColors.brand500, size: 22),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      '${network.toUpperCase()} $type Recharge',
                      style: AppTextStyles.labelLg.copyWith(
                        color: isDark ? AppColors.darkTextPrimary : AppColors.textPrimary,
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                    if (createdAt != null)
                      Text(
                        _formatDate(createdAt),
                        style: AppTextStyles.bodySm.copyWith(
                          color: isDark ? AppColors.darkTextTertiary : AppColors.textTertiary,
                        ),
                      ),
                  ],
                ),
              ),
              Column(
                crossAxisAlignment: CrossAxisAlignment.end,
                children: [
                  Text(
                    amount.toInt().toNaira(),
                    style: AppTextStyles.labelLg.copyWith(
                      color: isDark ? AppColors.darkTextPrimary : AppColors.textPrimary,
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                    decoration: BoxDecoration(
                      color: status == 'completed'
                          ? AppColors.success500.withValues(alpha: 0.1)
                          : AppColors.warning400.withValues(alpha: 0.1),
                      borderRadius: BorderRadius.circular(20),
                    ),
                    child: Text(
                      status.toUpperCase(),
                      style: AppTextStyles.labelXs.copyWith(
                        color: status == 'completed' ? AppColors.success500 : AppColors.warning400,
                      ),
                    ),
                  ),
                ],
              ),
            ],
          ),
        ).animate(delay: Duration(milliseconds: 50 * e.key)).fadeIn().slideY(begin: 0.1, end: 0);
      }).toList(),
    );
  }

  String _formatDate(String iso) {
    try {
      final dt = DateTime.parse(iso).toLocal();
      final now = DateTime.now();
      final diff = now.difference(dt);
      if (diff.inDays == 0) return 'Today ${dt.hour}:${dt.minute.toString().padLeft(2, '0')}';
      if (diff.inDays == 1) return 'Yesterday';
      return '${dt.day}/${dt.month}/${dt.year}';
    } catch (_) {
      return '';
    }
  }
}

// ─── Shimmer Placeholder ──────────────────────────────────────────────────────
class _ShimmerCard extends StatelessWidget {
  final double height;
  const _ShimmerCard({required this.height});

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    return Shimmer.fromColors(
      baseColor: isDark ? AppColors.darkBgTertiary : AppColors.slate100,
      highlightColor: isDark ? AppColors.darkBgCard : AppColors.bgPrimary,
      child: Container(
        height: height,
        decoration: BoxDecoration(
          color: isDark ? AppColors.darkBgTertiary : AppColors.slate100,
          borderRadius: BorderRadius.circular(16),
        ),
      ),
    );
  }
}

extension _AppCardMargin on AppCard {
  // Dummy — margin is handled by wrapping
}
