import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_animate/flutter_animate.dart';
import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_text_styles.dart';
import '../../../../core/api/api_client.dart';
import '../../../../shared/widgets/app_widgets.dart';

final transactionHistoryProvider = FutureProvider.autoDispose.family<Map<String, dynamic>, int>((ref, page) async {
  final api = ref.watch(apiClientProvider);
  return api.getRechargeHistory(page: page);
});

class TransactionHistoryScreen extends ConsumerStatefulWidget {
  const TransactionHistoryScreen({super.key});

  @override
  ConsumerState<TransactionHistoryScreen> createState() => _TransactionHistoryScreenState();
}

class _TransactionHistoryScreenState extends ConsumerState<TransactionHistoryScreen> {
  int _page = 1;
  final ScrollController _scrollController = ScrollController();
  final List<Map<String, dynamic>> _transactions = [];
  bool _isLoadingMore = false;
  bool _hasMore = true;

  @override
  void initState() {
    super.initState();
    _scrollController.addListener(_onScroll);
  }

  void _onScroll() {
    if (_scrollController.position.pixels >= _scrollController.position.maxScrollExtent - 200) {
      if (!_isLoadingMore && _hasMore) _loadMore();
    }
  }

  Future<void> _loadMore() async {
    setState(() => _isLoadingMore = true);
    try {
      final api = ref.read(apiClientProvider);
      final data = await api.getRechargeHistory(page: _page + 1);
      final items = (data['recharges'] ?? data['transactions'] ?? []) as List;
      if (items.isEmpty) {
        _hasMore = false;
      } else {
        _page++;
        _transactions.addAll(items.cast<Map<String, dynamic>>());
      }
    } catch (_) {}
    if (mounted) setState(() => _isLoadingMore = false);
  }

  @override
  void dispose() {
    _scrollController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final historyAsync = ref.watch(transactionHistoryProvider(1));

    return Scaffold(
      backgroundColor: isDark ? AppColors.darkBgPrimary : AppColors.bgSecondary,
      appBar: AppBar(
        title: const Text('Transaction History'),
        backgroundColor: isDark ? AppColors.darkBgPrimary : AppColors.bgPrimary,
      ),
      body: historyAsync.when(
        data: (data) {
          final items = (data['recharges'] ?? data['transactions'] ?? []) as List;
          final total = (data['total'] ?? items.length) as num;

          // Merge initial page with pagination
          final allItems = _transactions.isEmpty
              ? items.cast<Map<String, dynamic>>()
              : _transactions;

          if (allItems.isEmpty) {
            return AppEmptyState(
              icon: Icons.receipt_long_outlined,
              title: 'No transactions yet',
              subtitle: 'Your recharge history will appear here',
            );
          }

          return Column(
            children: [
              // Total
              Container(
                margin: const EdgeInsets.all(16),
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  gradient: AppColors.brandGradient,
                  borderRadius: BorderRadius.circular(14),
                ),
                child: Row(
                  children: [
                    const Icon(Icons.receipt_long_rounded, color: Colors.white, size: 28),
                    const SizedBox(width: 12),
                    Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text('Total Transactions', style: AppTextStyles.labelMd.copyWith(color: Colors.white70)),
                        Text('${total.toInt().toLocaleString()} recharges', style: AppTextStyles.headingMd.copyWith(color: Colors.white, fontWeight: FontWeight.w800)),
                      ],
                    ),
                  ],
                ),
              ),

              Expanded(
                child: ListView.builder(
                  controller: _scrollController,
                  padding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
                  itemCount: allItems.length + (_isLoadingMore ? 1 : 0),
                  itemBuilder: (context, i) {
                    if (i == allItems.length) {
                      return const Padding(
                        padding: EdgeInsets.symmetric(vertical: 20),
                        child: Center(child: CircularProgressIndicator()),
                      );
                    }

                    final txn = allItems[i];
                    return _TransactionCard(txn: txn, index: i);
                  },
                ),
              ),
            ],
          );
        },
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => AppEmptyState(
          icon: Icons.error_outline,
          title: 'Could not load transactions',
          subtitle: e.toString(),
        ),
      ),
    );
  }
}

class _TransactionCard extends StatelessWidget {
  final Map<String, dynamic> txn;
  final int index;

  const _TransactionCard({required this.txn, required this.index});

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final amount = (txn['amount'] ?? 0) as num;
    final network = (txn['network'] ?? '') as String;
    final type = (txn['recharge_type'] ?? txn['type'] ?? 'Airtime') as String;
    final status = (txn['status'] ?? 'completed') as String;
    final phone = (txn['msisdn'] ?? txn['recipient'] ?? '') as String;
    final entries = (txn['draw_entries'] ?? 0) as num;
    final spinUnlocked = txn['spin_unlocked'] == true;
    final date = txn['created_at'] as String?;

    final isSuccess = status.toLowerCase() == 'completed' || status.toLowerCase() == 'success';
    final statusColor = isSuccess ? AppColors.success500 : AppColors.error500;

    return AppCard(
      margin: const EdgeInsets.only(bottom: 10),
      padding: EdgeInsets.zero,
      child: Column(
        children: [
          Padding(
            padding: const EdgeInsets.all(14),
            child: Row(
              children: [
                Container(
                  width: 44,
                  height: 44,
                  decoration: BoxDecoration(
                    color: AppColors.brand500.withValues(alpha: 0.1),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Icon(
                    type.toLowerCase() == 'data' ? Icons.wifi_rounded : Icons.bolt_rounded,
                    color: AppColors.brand500,
                    size: 22,
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        '${network.toUpperCase()} $type',
                        style: AppTextStyles.labelLg.copyWith(
                          color: isDark ? AppColors.darkTextPrimary : AppColors.textPrimary,
                          fontWeight: FontWeight.w700,
                        ),
                      ),
                      Text(
                        phone,
                        style: AppTextStyles.bodyMd.copyWith(
                          color: isDark ? AppColors.darkTextSecondary : AppColors.textSecondary,
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
                      style: AppTextStyles.headingMd.copyWith(
                        color: isDark ? AppColors.darkTextPrimary : AppColors.textPrimary,
                        fontWeight: FontWeight.w800,
                      ),
                    ),
                    Container(
                      padding: const EdgeInsets.symmetric(horizontal: 7, vertical: 2),
                      decoration: BoxDecoration(
                        color: statusColor.withValues(alpha: 0.1),
                        borderRadius: BorderRadius.circular(20),
                      ),
                      child: Text(
                        status.toUpperCase(),
                        style: AppTextStyles.labelXs.copyWith(
                          color: statusColor,
                          fontSize: 9,
                        ),
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),

          // Rewards row
          if (entries > 0 || spinUnlocked)
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
              decoration: BoxDecoration(
                color: isDark ? AppColors.darkBgTertiary : AppColors.bgTertiary,
                borderRadius: const BorderRadius.vertical(bottom: Radius.circular(16)),
              ),
              child: Row(
                children: [
                  if (entries > 0) ...[
                    const Icon(Icons.confirmation_number_rounded, size: 14, color: AppColors.brand400),
                    const SizedBox(width: 4),
                    Text('${entries.toInt()} ${entries == 1 ? 'entry' : 'entries'}',
                        style: AppTextStyles.labelSm.copyWith(color: AppColors.brand400)),
                  ],
                  if (entries > 0 && spinUnlocked) const SizedBox(width: 12),
                  if (spinUnlocked) ...[
                    const Text('🎰', style: TextStyle(fontSize: 14)),
                    const SizedBox(width: 4),
                    Text('Spin unlocked', style: AppTextStyles.labelSm.copyWith(color: AppColors.gold500)),
                  ],
                  const Spacer(),
                  if (date != null)
                    Text(_fmt(date), style: AppTextStyles.bodySm.copyWith(color: isDark ? AppColors.darkTextTertiary : AppColors.textTertiary)),
                ],
              ),
            ),
        ],
      ),
    ).animate(delay: Duration(milliseconds: 30 * index.clamp(0, 10))).fadeIn().slideY(begin: 0.05, end: 0);
  }

  String _fmt(String iso) {
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
