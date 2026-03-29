import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:flutter_animate/flutter_animate.dart';
import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_text_styles.dart';
import '../../../../core/api/api_client.dart';
import '../../../../shared/widgets/app_widgets.dart';

// Recharge state
class RechargeFormState {
  final String phone;
  final String network;
  final String rechargeType; // 'airtime' | 'data'
  final double? amount;
  final String? selectedBundle;
  final bool isLoading;
  final String? error;

  const RechargeFormState({
    this.phone = '',
    this.network = '',
    this.rechargeType = 'airtime',
    this.amount,
    this.selectedBundle,
    this.isLoading = false,
    this.error,
  });

  RechargeFormState copyWith({
    String? phone,
    String? network,
    String? rechargeType,
    double? amount,
    String? selectedBundle,
    bool? isLoading,
    String? error,
  }) =>
      RechargeFormState(
        phone: phone ?? this.phone,
        network: network ?? this.network,
        rechargeType: rechargeType ?? this.rechargeType,
        amount: amount ?? this.amount,
        selectedBundle: selectedBundle ?? this.selectedBundle,
        isLoading: isLoading ?? this.isLoading,
        error: error,
      );
}

class RechargeNotifier extends StateNotifier<RechargeFormState> {
  final ApiClient _api;
  RechargeNotifier(this._api) : super(const RechargeFormState());

  void setPhone(String phone) => state = state.copyWith(phone: phone);
  void setNetwork(String network) => state = state.copyWith(network: network, selectedBundle: null);
  void setType(String type) => state = state.copyWith(rechargeType: type, selectedBundle: null, amount: null);
  void setAmount(double amount) => state = state.copyWith(amount: amount, selectedBundle: null);
  void setBundle(String bundleId, double amount) => state = state.copyWith(selectedBundle: bundleId, amount: amount);

  Future<Map<String, dynamic>> initiateRecharge() async {
    state = state.copyWith(isLoading: true, error: null);
    try {
      final result = await _api.initiateRecharge({
        'msisdn': state.phone,
        'network': state.network,
        'recharge_type': state.rechargeType,
        'amount': state.amount,
        if (state.selectedBundle != null) 'bundle_id': state.selectedBundle,
      });
      return result;
    } catch (e) {
      state = state.copyWith(isLoading: false, error: e.toString());
      rethrow;
    } finally {
      state = state.copyWith(isLoading: false);
    }
  }
}

final rechargeProvider = StateNotifierProvider.autoDispose<RechargeNotifier, RechargeFormState>((ref) {
  return RechargeNotifier(ref.watch(apiClientProvider));
});

final dataBundlesProvider = FutureProvider.autoDispose.family<List<dynamic>, String>((ref, network) async {
  if (network.isEmpty) return [];
  final api = ref.watch(apiClientProvider);
  return api.getDataBundles(network);
});

// ─── Screen ───────────────────────────────────────────────────────────────────
class RechargeScreen extends ConsumerWidget {
  const RechargeScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(rechargeProvider);
    final isDark = Theme.of(context).brightness == Brightness.dark;

    return Scaffold(
      backgroundColor: isDark ? AppColors.darkBgPrimary : AppColors.bgSecondary,
      appBar: AppBar(
        title: const Text('Buy Airtime & Data'),
        backgroundColor: isDark ? AppColors.darkBgPrimary : AppColors.bgPrimary,
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Type selector: Airtime | Data
            _TypeSelector(),

            const SizedBox(height: 16),

            // Phone + Network
            _PhoneNetworkSection(),

            const SizedBox(height: 16),

            // Amount / Bundles
            if (state.rechargeType == 'airtime') _AirtimeAmounts(),
            if (state.rechargeType == 'data') _DataBundles(),

            const SizedBox(height: 20),

            // Bonus info
            if (state.amount != null && state.amount! >= 1000)
              _SpinUnlockBanner(amount: state.amount!),

            if (state.amount != null && state.amount! >= 200)
              _DrawEntryBanner(amount: state.amount!),

            const SizedBox(height: 24),

            // Proceed button
            AppGradientButton(
              label: state.rechargeType == 'airtime'
                  ? 'Buy Airtime — ${(state.amount?.toInt() ?? 0).toNaira()}'
                  : 'Buy Data Bundle',
              onPressed: (state.phone.length >= 11 && state.network.isNotEmpty && state.amount != null)
                  ? () => _proceed(context, ref)
                  : null,
              isLoading: state.isLoading,
              icon: const Icon(Icons.bolt_rounded, color: Colors.white, size: 18),
            ),
          ],
        ),
      ),
    );
  }

  Future<void> _proceed(BuildContext context, WidgetRef ref) async {
    try {
      final result = await ref.read(rechargeProvider.notifier).initiateRecharge();
      // Result contains Paystack payment URL
      final paymentUrl = result['payment_url'] ?? result['authorization_url'];
      if (paymentUrl != null && context.mounted) {
        // Navigate to WebView or external browser for payment
        // For now, pass to success screen with the data
        context.push('/recharge/success', extra: result);
      }
    } catch (e) {
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(e.toString().replaceAll('Exception: ', '')),
            backgroundColor: AppColors.error500,
          ),
        );
      }
    }
  }
}

class _TypeSelector extends ConsumerWidget {
  const _TypeSelector();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final type = ref.watch(rechargeProvider).rechargeType;
    final isDark = Theme.of(context).brightness == Brightness.dark;

    return AppCard(
      padding: const EdgeInsets.all(4),
      borderRadius: 12,
      child: Row(
        children: ['airtime', 'data'].map((t) {
          final isSelected = type == t;
          return Expanded(
            child: GestureDetector(
              onTap: () => ref.read(rechargeProvider.notifier).setType(t),
              child: AnimatedContainer(
                duration: const Duration(milliseconds: 200),
                padding: const EdgeInsets.symmetric(vertical: 12),
                decoration: BoxDecoration(
                  gradient: isSelected ? AppColors.brandGradient : null,
                  borderRadius: BorderRadius.circular(10),
                ),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Icon(
                      t == 'airtime' ? Icons.bolt_rounded : Icons.wifi_rounded,
                      size: 18,
                      color: isSelected ? Colors.white : (isDark ? AppColors.darkTextTertiary : AppColors.textTertiary),
                    ),
                    const SizedBox(width: 6),
                    Text(
                      t == 'airtime' ? 'Airtime' : 'Data',
                      style: AppTextStyles.labelLg.copyWith(
                        color: isSelected
                            ? Colors.white
                            : (isDark ? AppColors.darkTextTertiary : AppColors.textTertiary),
                        fontWeight: isSelected ? FontWeight.w700 : FontWeight.w500,
                      ),
                    ),
                  ],
                ),
              ),
            ),
          );
        }).toList(),
      ),
    );
  }
}

class _PhoneNetworkSection extends ConsumerStatefulWidget {
  const _PhoneNetworkSection();

  @override
  ConsumerState<_PhoneNetworkSection> createState() => _PhoneNetworkSectionState();
}

class _PhoneNetworkSectionState extends ConsumerState<_PhoneNetworkSection> {
  final _controller = TextEditingController();

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  String _detectNetwork(String phone) {
    final digits = phone.replaceAll(RegExp(r'\D'), '');
    final prefix = digits.length >= 4 ? digits.substring(0, 4) : digits;
    const networks = {
      'mtn': ['0803', '0806', '0703', '0706', '0813', '0816', '0810', '0814', '0903', '0906', '0913', '0916'],
      'glo': ['0805', '0807', '0705', '0815', '0811', '0905', '0915'],
      'airtel': ['0802', '0808', '0708', '0812', '0701', '0902', '0907', '0901'],
      '9mobile': ['0809', '0818', '0817', '0909', '0908'],
    };
    for (final entry in networks.entries) {
      if (entry.value.contains(prefix)) return entry.key;
    }
    return '';
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(rechargeProvider);
    final isDark = Theme.of(context).brightness == Brightness.dark;

    return AppCard(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text('Phone Number', style: AppTextStyles.labelLg.copyWith(
            color: isDark ? AppColors.darkTextSecondary : AppColors.textSecondary,
          )),
          const SizedBox(height: 10),
          TextField(
            controller: _controller,
            keyboardType: TextInputType.phone,
            style: AppTextStyles.headingMd.copyWith(
              color: isDark ? AppColors.darkTextPrimary : AppColors.textPrimary,
              letterSpacing: 1.5,
            ),
            decoration: InputDecoration(
              hintText: '0812 345 6789',
              prefixIcon: const Icon(Icons.phone_android_rounded),
              suffixIcon: state.network.isNotEmpty
                  ? Padding(
                      padding: const EdgeInsets.only(right: 8),
                      child: NetworkBadge(network: state.network, selected: true),
                    )
                  : null,
            ),
            onChanged: (v) {
              ref.read(rechargeProvider.notifier).setPhone(v.trim());
              final detected = _detectNetwork(v.trim());
              if (detected.isNotEmpty && detected != state.network) {
                ref.read(rechargeProvider.notifier).setNetwork(detected);
              }
            },
          ),
          const SizedBox(height: 12),
          Text('Network', style: AppTextStyles.labelLg.copyWith(
            color: isDark ? AppColors.darkTextSecondary : AppColors.textSecondary,
          )),
          const SizedBox(height: 8),
          Row(
            children: ['mtn', 'glo', 'airtel', '9mobile'].map((n) {
              return Padding(
                padding: const EdgeInsets.only(right: 8),
                child: GestureDetector(
                  onTap: () => ref.read(rechargeProvider.notifier).setNetwork(n),
                  child: NetworkBadge(network: n, selected: state.network == n),
                ),
              );
            }).toList(),
          ),
        ],
      ),
    );
  }
}

class _AirtimeAmounts extends ConsumerWidget {
  const _AirtimeAmounts();

  static const _amounts = [100.0, 200.0, 500.0, 1000.0, 2000.0, 5000.0];

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(rechargeProvider);
    final isDark = Theme.of(context).brightness == Brightness.dark;

    return AppCard(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text('Select Amount', style: AppTextStyles.labelLg.copyWith(
            color: isDark ? AppColors.darkTextSecondary : AppColors.textSecondary,
          )),
          const SizedBox(height: 12),
          Wrap(
            spacing: 10,
            runSpacing: 10,
            children: _amounts.map((amt) {
              final isSelected = state.amount == amt;
              return GestureDetector(
                onTap: () => ref.read(rechargeProvider.notifier).setAmount(amt),
                child: AnimatedContainer(
                  duration: const Duration(milliseconds: 200),
                  padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
                  decoration: BoxDecoration(
                    gradient: isSelected ? AppColors.brandGradient : null,
                    color: isSelected ? null : (isDark ? AppColors.darkBgTertiary : AppColors.bgTertiary),
                    borderRadius: BorderRadius.circular(10),
                    border: Border.all(
                      color: isSelected ? Colors.transparent : (isDark ? AppColors.darkBorderSecondary : AppColors.borderSecondary),
                    ),
                  ),
                  child: Text(
                    amt.toInt().toNaira(),
                    style: AppTextStyles.labelLg.copyWith(
                      color: isSelected ? Colors.white : (isDark ? AppColors.darkTextPrimary : AppColors.textPrimary),
                      fontWeight: isSelected ? FontWeight.w700 : FontWeight.w500,
                    ),
                  ),
                ),
              );
            }).toList(),
          ),
          const SizedBox(height: 12),
          // Custom amount
          TextField(
            keyboardType: const TextInputType.numberWithOptions(decimal: false),
            decoration: const InputDecoration(
              labelText: 'Or enter custom amount',
              prefixText: '₦ ',
              prefixStyle: TextStyle(fontWeight: FontWeight.w600),
            ),
            onChanged: (v) {
              final parsed = double.tryParse(v);
              if (parsed != null && parsed >= 50) {
                ref.read(rechargeProvider.notifier).setAmount(parsed);
              }
            },
          ),
        ],
      ),
    );
  }
}

class _DataBundles extends ConsumerWidget {
  const _DataBundles();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final network = ref.watch(rechargeProvider).network;
    final isDark = Theme.of(context).brightness == Brightness.dark;

    if (network.isEmpty) {
      return AppCard(
        padding: const EdgeInsets.all(20),
        child: Center(
          child: Text(
            'Select a network to see data bundles',
            style: AppTextStyles.bodyMd.copyWith(
              color: isDark ? AppColors.darkTextTertiary : AppColors.textTertiary,
            ),
          ),
        ),
      );
    }

    final bundlesAsync = ref.watch(dataBundlesProvider(network));

    return bundlesAsync.when(
      data: (bundles) {
        if (bundles.isEmpty) {
          return AppEmptyState(
            icon: Icons.wifi_off_rounded,
            title: 'No bundles available',
            subtitle: 'Try a different network',
          );
        }

        return AppCard(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text('Data Bundles', style: AppTextStyles.labelLg.copyWith(
                color: isDark ? AppColors.darkTextSecondary : AppColors.textSecondary,
              )),
              const SizedBox(height: 12),
              ...bundles.map((b) {
                final bundle = b as Map<String, dynamic>;
                final id = bundle['id']?.toString() ?? '';
                final price = (bundle['price'] ?? bundle['amount'] ?? 0) as num;
                final size = bundle['size'] ?? bundle['data_size'] ?? '';
                final validity = bundle['validity'] ?? '';
                final selected = ref.watch(rechargeProvider).selectedBundle == id;

                return GestureDetector(
                  onTap: () => ref.read(rechargeProvider.notifier).setBundle(id, price.toDouble()),
                  child: AnimatedContainer(
                    duration: const Duration(milliseconds: 200),
                    margin: const EdgeInsets.only(bottom: 8),
                    padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
                    decoration: BoxDecoration(
                      gradient: selected ? AppColors.brandGradient : null,
                      color: selected ? null : (isDark ? AppColors.darkBgTertiary : AppColors.bgTertiary),
                      borderRadius: BorderRadius.circular(10),
                      border: Border.all(
                        color: selected ? Colors.transparent : (isDark ? AppColors.darkBorderSecondary : AppColors.borderSecondary),
                      ),
                    ),
                    child: Row(
                      children: [
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(size, style: AppTextStyles.labelLg.copyWith(
                                color: selected ? Colors.white : (isDark ? AppColors.darkTextPrimary : AppColors.textPrimary),
                                fontWeight: FontWeight.w700,
                              )),
                              if (validity.isNotEmpty)
                                Text(validity, style: AppTextStyles.bodySm.copyWith(
                                  color: selected ? Colors.white70 : (isDark ? AppColors.darkTextTertiary : AppColors.textTertiary),
                                )),
                            ],
                          ),
                        ),
                        Text(price.toInt().toNaira(), style: AppTextStyles.headingMd.copyWith(
                          color: selected ? Colors.white : AppColors.brand500,
                          fontWeight: FontWeight.w800,
                        )),
                      ],
                    ),
                  ),
                );
              }).toList(),
            ],
          ),
        );
      },
      loading: () => const Center(child: CircularProgressIndicator()),
      error: (_, __) => AppEmptyState(
        icon: Icons.error_outline,
        title: 'Could not load bundles',
        subtitle: 'Check your connection and try again',
      ),
    );
  }
}

class _SpinUnlockBanner extends StatelessWidget {
  final double amount;
  const _SpinUnlockBanner({required this.amount});

  @override
  Widget build(BuildContext context) {
    return Container(
      margin: const EdgeInsets.only(bottom: 10),
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
      decoration: BoxDecoration(
        gradient: const LinearGradient(
          colors: [AppColors.brand800, AppColors.brand600],
        ),
        borderRadius: BorderRadius.circular(12),
      ),
      child: Row(
        children: [
          const Text('🎰', style: TextStyle(fontSize: 22)),
          const SizedBox(width: 10),
          Expanded(
            child: Text(
              'This recharge unlocks the Spin Wheel!',
              style: AppTextStyles.labelMd.copyWith(
                color: Colors.white,
                fontWeight: FontWeight.w600,
              ),
            ),
          ),
        ],
      ),
    ).animate().fadeIn().slideY(begin: 0.1, end: 0);
  }
}

class _DrawEntryBanner extends StatelessWidget {
  final double amount;
  const _DrawEntryBanner({required this.amount});

  @override
  Widget build(BuildContext context) {
    final entries = (amount / 200).floor();
    return Container(
      margin: const EdgeInsets.only(bottom: 10),
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
      decoration: BoxDecoration(
        color: AppColors.gold500.withOpacity(0.15),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: AppColors.gold500.withOpacity(0.3)),
      ),
      child: Row(
        children: [
          const Text('🏆', style: TextStyle(fontSize: 22)),
          const SizedBox(width: 10),
          Expanded(
            child: Text(
              'You\'ll earn $entries draw ${entries == 1 ? 'entry' : 'entries'}!',
              style: AppTextStyles.labelMd.copyWith(
                color: AppColors.gold600,
                fontWeight: FontWeight.w600,
              ),
            ),
          ),
        ],
      ),
    ).animate().fadeIn().slideY(begin: 0.1, end: 0);
  }
}
