import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_animate/flutter_animate.dart';
import 'package:go_router/go_router.dart';
import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_text_styles.dart';
import '../../../../core/api/api_client.dart';
import '../../../../shared/widgets/app_widgets.dart';

final activeDrawsProvider = FutureProvider.autoDispose<Map<String, dynamic>>((ref) async {
  final api = ref.watch(apiClientProvider);
  return api.getActiveDraws();
});

final myEntriesProvider = FutureProvider.autoDispose<Map<String, dynamic>>((ref) async {
  final api = ref.watch(apiClientProvider);
  return api.getMyDrawEntries();
});

class DrawsScreen extends ConsumerStatefulWidget {
  const DrawsScreen({super.key});

  @override
  ConsumerState<DrawsScreen> createState() => _DrawsScreenState();
}

class _DrawsScreenState extends ConsumerState<DrawsScreen>
    with SingleTickerProviderStateMixin {
  late TabController _tabs;

  @override
  void initState() {
    super.initState();
    _tabs = TabController(length: 3, vsync: this);
  }

  @override
  void dispose() {
    _tabs.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;

    return Scaffold(
      backgroundColor: isDark ? AppColors.darkBgPrimary : AppColors.bgSecondary,
      body: NestedScrollView(
        headerSliverBuilder: (context, _) => [
          SliverAppBar(
            pinned: true,
            expandedHeight: 120,
            backgroundColor: isDark ? AppColors.darkBgPrimary : AppColors.bgPrimary,
            flexibleSpace: FlexibleSpaceBar(
              title: Text(
                'Prize Draws',
                style: AppTextStyles.headingLg.copyWith(
                  color: isDark ? AppColors.darkTextPrimary : AppColors.textPrimary,
                ),
              ),
              titlePadding: const EdgeInsets.only(left: 16, bottom: 52),
            ),
            bottom: TabBar(
              controller: _tabs,
              tabs: const [
                Tab(text: 'Active Draws'),
                Tab(text: 'My Entries'),
                Tab(text: 'Winners'),
              ],
            ),
          ),
        ],
        body: TabBarView(
          controller: _tabs,
          children: [
            _ActiveDrawsTab(),
            _MyEntriesTab(),
            _WinnersTab(),
          ],
        ),
      ),
    );
  }
}

class _ActiveDrawsTab extends ConsumerWidget {
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final drawsAsync = ref.watch(activeDrawsProvider);

    return drawsAsync.when(
      data: (data) {
        final draws = (data['draws'] ?? data['active_draws'] ?? []) as List;
        if (draws.isEmpty) {
          return AppEmptyState(
            icon: Icons.emoji_events_outlined,
            title: 'No active draws',
            subtitle: 'Check back soon for the next draw!',
            actionLabel: 'Recharge to earn entries',
            onAction: () => context.go('/recharge'),
          );
        }
        return RefreshIndicator(
          onRefresh: () async => ref.invalidate(activeDrawsProvider),
          child: ListView.builder(
            padding: const EdgeInsets.all(16),
            itemCount: draws.length,
            itemBuilder: (context, i) {
              return _DrawCard(draw: draws[i] as Map<String, dynamic>, index: i);
            },
          ),
        );
      },
      loading: () => const Center(child: CircularProgressIndicator()),
      error: (e, _) => AppEmptyState(
        icon: Icons.error_outline,
        title: 'Could not load draws',
        subtitle: e.toString(),
      ),
    );
  }
}

class _DrawCard extends StatefulWidget {
  final Map<String, dynamic> draw;
  final int index;
  const _DrawCard({required this.draw, required this.index});

  @override
  State<_DrawCard> createState() => _DrawCardState();
}

class _DrawCardState extends State<_DrawCard> {
  late Duration _remaining;
  late DateTime _endTime;

  @override
  void initState() {
    super.initState();
    try {
      _endTime = DateTime.parse(widget.draw['draw_date'] ?? widget.draw['end_time'] ?? '');
    } catch (_) {
      _endTime = DateTime.now().add(const Duration(hours: 20));
    }
    _updateRemaining();
    _startTimer();
  }

  void _updateRemaining() {
    _remaining = _endTime.difference(DateTime.now());
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
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final h = _remaining.inHours;
    final m = _remaining.inMinutes % 60;
    final s = _remaining.inSeconds % 60;

    final name = widget.draw['name'] ?? widget.draw['draw_name'] ?? 'Prize Draw';
    final prizePool = (widget.draw['prize_pool'] ?? widget.draw['total_prize'] ?? 0) as num;
    final entries = (widget.draw['total_entries'] ?? 0) as num;
    final myEntries = (widget.draw['my_entries'] ?? 0) as num;
    final drawType = (widget.draw['draw_type'] ?? 'Daily') as String;

    // Colors by type
    final typeColor = drawType.toLowerCase().contains('daily')
        ? AppColors.brand500
        : drawType.toLowerCase().contains('weekly')
            ? AppColors.gold500
            : AppColors.error500;

    return AppCard(
      margin: const EdgeInsets.only(bottom: 12),
      padding: EdgeInsets.zero,
      borderRadius: 16,
      child: Column(
        children: [
          // Header
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
            decoration: BoxDecoration(
              color: typeColor.withOpacity(0.1),
              borderRadius: const BorderRadius.vertical(top: Radius.circular(16)),
            ),
            child: Row(
              children: [
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                  decoration: BoxDecoration(
                    color: typeColor,
                    borderRadius: BorderRadius.circular(20),
                  ),
                  child: Text(
                    drawType.toUpperCase(),
                    style: AppTextStyles.labelXs.copyWith(color: Colors.white),
                  ),
                ),
                const Spacer(),
                // Countdown
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
                  decoration: BoxDecoration(
                    color: isDark ? AppColors.darkBgTertiary : AppColors.bgTertiary,
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Text(
                    '${_pad(h)}:${_pad(m)}:${_pad(s)}',
                    style: AppTextStyles.labelMd.copyWith(
                      color: isDark ? AppColors.darkTextPrimary : AppColors.textPrimary,
                      fontWeight: FontWeight.w700,
                      letterSpacing: 2,
                      fontFeatures: [const FontFeature.tabularFigures()],
                    ),
                  ),
                ),
              ],
            ),
          ),

          // Content
          Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  name,
                  style: AppTextStyles.headingMd.copyWith(
                    color: isDark ? AppColors.darkTextPrimary : AppColors.textPrimary,
                    fontWeight: FontWeight.w700,
                  ),
                ),
                const SizedBox(height: 12),
                Row(
                  children: [
                    Expanded(
                      child: _DrawStat(label: 'Prize Pool', value: prizePool.toInt().toNaira(), color: AppColors.gold500),
                    ),
                    Expanded(
                      child: _DrawStat(label: 'Total Entries', value: entries.toInt().toLocaleString()),
                    ),
                    Expanded(
                      child: _DrawStat(label: 'My Entries', value: myEntries.toString(), color: typeColor),
                    ),
                  ],
                ),
                const SizedBox(height: 12),
                SizedBox(
                  width: double.infinity,
                  child: OutlinedButton.icon(
                    onPressed: () => context.go('/recharge'),
                    icon: const Icon(Icons.bolt_rounded, size: 16),
                    label: const Text('Recharge to Enter'),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    ).animate(delay: Duration(milliseconds: 60 * widget.index)).fadeIn().slideY(begin: 0.1, end: 0);
  }
}

class _DrawStat extends StatelessWidget {
  final String label;
  final String value;
  final Color? color;
  const _DrawStat({required this.label, required this.value, this.color});

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    return Column(
      children: [
        Text(
          value,
          style: AppTextStyles.headingMd.copyWith(
            color: color ?? (isDark ? AppColors.darkTextPrimary : AppColors.textPrimary),
            fontWeight: FontWeight.w800,
          ),
        ),
        Text(
          label,
          style: AppTextStyles.bodySm.copyWith(
            color: isDark ? AppColors.darkTextTertiary : AppColors.textTertiary,
          ),
          textAlign: TextAlign.center,
        ),
      ],
    );
  }
}

class _MyEntriesTab extends ConsumerWidget {
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final entriesAsync = ref.watch(myEntriesProvider);

    return entriesAsync.when(
      data: (data) {
        final entries = (data['entries'] ?? []) as List;
        final total = (data['total'] ?? data['total_entries'] ?? entries.length) as num;

        return Column(
          children: [
            // Summary banner
            Container(
              margin: const EdgeInsets.all(16),
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                gradient: AppColors.brandGradient,
                borderRadius: BorderRadius.circular(16),
              ),
              child: Row(
                children: [
                  const Text('🏆', style: TextStyle(fontSize: 28)),
                  const SizedBox(width: 12),
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text('Total Draw Entries', style: AppTextStyles.labelMd.copyWith(color: Colors.white70)),
                      Text(total.toInt().toLocaleString(), style: AppTextStyles.headingXl.copyWith(color: Colors.white, fontWeight: FontWeight.w800)),
                    ],
                  ),
                ],
              ),
            ),

            if (entries.isEmpty)
              Expanded(
                child: AppEmptyState(
                  icon: Icons.confirmation_number_outlined,
                  title: 'No entries yet',
                  subtitle: 'Recharge to earn draw entries!',
                  actionLabel: 'Recharge Now',
                  onAction: () => context.go('/recharge'),
                ),
              )
            else
              Expanded(
                child: ListView.builder(
                  padding: const EdgeInsets.symmetric(horizontal: 16),
                  itemCount: entries.length,
                  itemBuilder: (context, i) {
                    final entry = entries[i] as Map<String, dynamic>;
                    final count = (entry['entry_count'] ?? entry['entries'] ?? 1) as num;
                    final draw = entry['draw_name'] ?? entry['draw_type'] ?? 'Draw';
                    final date = entry['created_at'] as String?;

                    return AppCard(
                      margin: const EdgeInsets.only(bottom: 8),
                      child: Row(
                        children: [
                          Container(
                            width: 40,
                            height: 40,
                            decoration: BoxDecoration(
                              color: AppColors.brand500.withOpacity(0.1),
                              borderRadius: BorderRadius.circular(10),
                            ),
                            child: const Icon(Icons.confirmation_number_rounded, color: AppColors.brand500, size: 20),
                          ),
                          const SizedBox(width: 12),
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(draw, style: AppTextStyles.labelLg.copyWith(fontWeight: FontWeight.w600)),
                                if (date != null)
                                  Text(_formatDate(date), style: AppTextStyles.bodySm.copyWith(color: AppColors.textTertiary)),
                              ],
                            ),
                          ),
                          Container(
                            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
                            decoration: BoxDecoration(
                              color: AppColors.brand500.withOpacity(0.1),
                              borderRadius: BorderRadius.circular(20),
                            ),
                            child: Text(
                              '×${count.toInt()}',
                              style: AppTextStyles.labelMd.copyWith(
                                color: AppColors.brand500,
                                fontWeight: FontWeight.w700,
                              ),
                            ),
                          ),
                        ],
                      ),
                    );
                  },
                ),
              ),
          ],
        );
      },
      loading: () => const Center(child: CircularProgressIndicator()),
      error: (_, __) => AppEmptyState(
        icon: Icons.error_outline,
        title: 'Could not load entries',
      ),
    );
  }

  String _formatDate(String iso) {
    try {
      final dt = DateTime.parse(iso).toLocal();
      return '${dt.day}/${dt.month}/${dt.year}';
    } catch (_) {
      return '';
    }
  }
}

class _WinnersTab extends ConsumerWidget {
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final winnersAsync = ref.watch(FutureProvider.autoDispose((ref) async {
      final api = ref.watch(apiClientProvider);
      return api.getWinners();
    }));

    return winnersAsync.when(
      data: (data) {
        final winners = (data['winners'] ?? []) as List;
        if (winners.isEmpty) {
          return AppEmptyState(
            icon: Icons.emoji_events_outlined,
            title: 'No winners yet',
            subtitle: 'Be the first to win!',
          );
        }
        return ListView.builder(
          padding: const EdgeInsets.all(16),
          itemCount: winners.length,
          itemBuilder: (context, i) {
            final w = winners[i] as Map<String, dynamic>;
            final name = w['name'] ?? w['msisdn'] ?? 'Winner';
            final prize = (w['prize_amount'] ?? w['amount'] ?? 0) as num;
            final draw = w['draw_name'] ?? w['draw_type'] ?? '';
            final date = w['created_at'] as String?;

            return AppCard(
              margin: const EdgeInsets.only(bottom: 10),
              child: Row(
                children: [
                  CircleAvatar(
                    backgroundColor: AppColors.gold500.withOpacity(0.15),
                    child: const Text('🏆'),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(name, style: AppTextStyles.labelLg.copyWith(fontWeight: FontWeight.w600)),
                        if (draw.isNotEmpty)
                          Text(draw, style: AppTextStyles.bodySm.copyWith(color: AppColors.textTertiary)),
                      ],
                    ),
                  ),
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.end,
                    children: [
                      Text(prize.toInt().toNaira(), style: AppTextStyles.labelLg.copyWith(color: AppColors.gold500, fontWeight: FontWeight.w800)),
                      if (date != null)
                        Text(_fmt(date), style: AppTextStyles.bodySm.copyWith(color: AppColors.textTertiary)),
                    ],
                  ),
                ],
              ),
            );
          },
        );
      },
      loading: () => const Center(child: CircularProgressIndicator()),
      error: (_, __) => AppEmptyState(icon: Icons.error_outline, title: 'Could not load winners'),
    );
  }

  String _fmt(String iso) {
    try {
      final dt = DateTime.parse(iso).toLocal();
      return '${dt.day}/${dt.month}/${dt.year}';
    } catch (_) {
      return '';
    }
  }
}
