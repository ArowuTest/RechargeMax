import 'dart:math';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:flutter_animate/flutter_animate.dart';
import 'package:confetti/confetti.dart';
import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_text_styles.dart';
import '../../../../core/api/api_client.dart';
import '../../../../shared/widgets/app_widgets.dart';

// ─── Spin state ───────────────────────────────────────────────────────────────
enum SpinStatus { idle, spinning, result, ineligible }

class SpinState {
  final SpinStatus status;
  final bool eligible;
  final Map<String, dynamic>? result;
  final String? error;
  final int spinsRemaining;

  const SpinState({
    this.status = SpinStatus.idle,
    this.eligible = false,
    this.result,
    this.error,
    this.spinsRemaining = 0,
  });

  SpinState copyWith({
    SpinStatus? status,
    bool? eligible,
    Map<String, dynamic>? result,
    String? error,
    int? spinsRemaining,
  }) =>
      SpinState(
        status: status ?? this.status,
        eligible: eligible ?? this.eligible,
        result: result ?? this.result,
        error: error,
        spinsRemaining: spinsRemaining ?? this.spinsRemaining,
      );
}

class SpinNotifier extends StateNotifier<SpinState> {
  final ApiClient _api;
  SpinNotifier(this._api) : super(const SpinState());

  Future<void> checkEligibility() async {
    try {
      final data = await _api.checkSpinEligibility();
      state = state.copyWith(
        eligible: data['eligible'] == true,
        spinsRemaining: (data['spins_remaining'] ?? data['spin_count'] ?? 0) as int,
        status: (data['eligible'] == true) ? SpinStatus.idle : SpinStatus.ineligible,
      );
    } catch (e) {
      state = state.copyWith(status: SpinStatus.ineligible);
    }
  }

  Future<Map<String, dynamic>> playSpin() async {
    state = state.copyWith(status: SpinStatus.spinning, result: null, error: null);
    try {
      final result = await _api.playSpin();
      state = state.copyWith(
        status: SpinStatus.result,
        result: result,
        spinsRemaining: state.spinsRemaining - 1,
      );
      return result;
    } catch (e) {
      state = state.copyWith(
        status: SpinStatus.idle,
        error: e.toString().replaceAll('Exception: ', ''),
      );
      rethrow;
    }
  }

  void reset() {
    state = state.copyWith(
      status: state.eligible ? SpinStatus.idle : SpinStatus.ineligible,
      result: null,
      error: null,
    );
  }
}

final spinProvider = StateNotifierProvider.autoDispose<SpinNotifier, SpinState>((ref) {
  final notifier = SpinNotifier(ref.watch(apiClientProvider));
  notifier.checkEligibility();
  return notifier;
});

final spinHistoryProvider = FutureProvider.autoDispose<List<dynamic>>((ref) async {
  final api = ref.watch(apiClientProvider);
  final data = await api.getSpinHistory();
  return (data['spins'] ?? data['history'] ?? []) as List<dynamic>;
});

// ─── Spin Wheel segments ──────────────────────────────────────────────────────
const _kSegments = [
  _WheelSegment(label: '₦500', icon: '💰', color: AppColors.gold500, value: 500),
  _WheelSegment(label: 'Try Again', icon: '🔄', color: AppColors.slate400, value: 0),
  _WheelSegment(label: '1GB Data', icon: '📶', color: AppColors.success500, value: 0),
  _WheelSegment(label: '₦200', icon: '💵', color: AppColors.brand400, value: 200),
  _WheelSegment(label: 'Bonus', icon: '⭐', color: AppColors.warning500, value: 0),
  _WheelSegment(label: '₦1,000', icon: '🎉', color: AppColors.error500, value: 1000),
  _WheelSegment(label: '500MB', icon: '📡', color: AppColors.success400, value: 0),
  _WheelSegment(label: '₦100', icon: '🪙', color: AppColors.brand300, value: 100),
];

class _WheelSegment {
  final String label;
  final String icon;
  final Color color;
  final int value;
  const _WheelSegment({required this.label, required this.icon, required this.color, required this.value});
}

// ─── Screen ───────────────────────────────────────────────────────────────────
class SpinScreen extends ConsumerStatefulWidget {
  const SpinScreen({super.key});

  @override
  ConsumerState<SpinScreen> createState() => _SpinScreenState();
}

class _SpinScreenState extends ConsumerState<SpinScreen>
    with TickerProviderStateMixin {
  late AnimationController _wheelController;
  late Animation<double> _wheelAnimation;
  late ConfettiController _confetti;
  int _targetSegment = 0;
  double _currentAngle = 0;

  @override
  void initState() {
    super.initState();
    _wheelController = AnimationController(
      vsync: this,
      duration: const Duration(seconds: 4),
    );
    _wheelAnimation = Tween<double>(begin: 0, end: 0).animate(
      CurvedAnimation(parent: _wheelController, curve: Curves.easeInOutCubic),
    );
    _confetti = ConfettiController(duration: const Duration(seconds: 3));
  }

  @override
  void dispose() {
    _wheelController.dispose();
    _confetti.dispose();
    super.dispose();
  }

  Future<void> _spin() async {
    final notifier = ref.read(spinProvider.notifier);
    try {
      // Pick a random target segment for the animation
      // (the backend decides the real prize)
      _targetSegment = Random().nextInt(_kSegments.length);
      final segmentAngle = 2 * pi / _kSegments.length;
      final extraSpins = 5 * 2 * pi; // 5 full rotations
      final targetAngle = extraSpins + (_targetSegment * segmentAngle);

      _wheelAnimation = Tween<double>(
        begin: _currentAngle,
        end: _currentAngle + targetAngle,
      ).animate(CurvedAnimation(parent: _wheelController, curve: Curves.easeInOutCubic));

      _wheelController.forward(from: 0);

      // Call API in parallel
      final result = await notifier.playSpin();

      // Wait for animation to complete
      await Future.delayed(const Duration(seconds: 4));
      _currentAngle = (_currentAngle + targetAngle) % (2 * pi);

      // Show result
      final prizeAmount = result['prize_amount'] ?? result['amount'] ?? 0;
      if ((prizeAmount as num) > 0) {
        _confetti.play();
      }

      _showResultSheet(result);
    } catch (e) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(e.toString().replaceAll('Exception: ', '')),
          backgroundColor: AppColors.error500,
        ),
      );
    }
  }

  void _showResultSheet(Map<String, dynamic> result) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      builder: (_) => _SpinResultSheet(result: result, onClose: () {
        Navigator.pop(context);
        ref.read(spinProvider.notifier).reset();
      }),
    );
  }

  @override
  Widget build(BuildContext context) {
    final spinState = ref.watch(spinProvider);
    final isDark = Theme.of(context).brightness == Brightness.dark;

    return Scaffold(
      backgroundColor: isDark ? AppColors.darkBgPrimary : AppColors.bgSecondary,
      body: Stack(
        alignment: Alignment.topCenter,
        children: [
          // Dark purple gradient background
          Positioned.fill(
            child: Container(
              decoration: const BoxDecoration(gradient: AppColors.heroGradient),
            ),
          ),

          // Confetti
          ConfettiWidget(
            confettiController: _confetti,
            blastDirectionality: BlastDirectionality.explosive,
            numberOfParticles: 50,
            colors: const [AppColors.gold500, AppColors.brand400, Colors.white, AppColors.success400],
          ),

          SafeArea(
            child: Column(
              children: [
                // App bar
                Padding(
                  padding: const EdgeInsets.fromLTRB(16, 8, 16, 0),
                  child: Row(
                    children: [
                      Text(
                        'Spin & Win',
                        style: AppTextStyles.headingXl.copyWith(
                          color: Colors.white,
                          fontWeight: FontWeight.w800,
                        ),
                      ),
                      const Spacer(),
                      if (spinState.spinsRemaining > 0)
                        Container(
                          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                          decoration: BoxDecoration(
                            color: AppColors.gold500.withOpacity(0.2),
                            borderRadius: BorderRadius.circular(20),
                            border: Border.all(color: AppColors.gold500.withOpacity(0.4)),
                          ),
                          child: Row(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              const Icon(Icons.casino_rounded, color: AppColors.gold400, size: 16),
                              const SizedBox(width: 4),
                              Text(
                                '${spinState.spinsRemaining} left',
                                style: AppTextStyles.labelMd.copyWith(
                                  color: AppColors.gold400,
                                  fontWeight: FontWeight.w700,
                                ),
                              ),
                            ],
                          ),
                        ),
                    ],
                  ),
                ),

                const SizedBox(height: 20),

                // Wheel area
                Expanded(
                  child: _buildWheelContent(spinState),
                ),

                // Bottom: Spin button + history tab
                _SpinBottomBar(spinState: spinState, onSpin: _spin),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildWheelContent(SpinState spinState) {
    if (spinState.status == SpinStatus.ineligible) {
      return _IneligibleState();
    }

    return Column(
      children: [
        // Wheel
        Expanded(
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: 24),
            child: AnimatedBuilder(
              animation: _wheelAnimation,
              builder: (context, child) {
                return _SpinWheel(
                  angle: _wheelAnimation.value,
                  segments: _kSegments,
                  isSpinning: spinState.status == SpinStatus.spinning,
                );
              },
            ),
          ),
        ),

        const SizedBox(height: 16),

        // Hint text
        Text(
          spinState.status == SpinStatus.spinning
              ? '✨ Spinning...'
              : 'Tap SPIN to try your luck!',
          style: AppTextStyles.bodyLg.copyWith(
            color: Colors.white.withOpacity(0.7),
          ),
        ).animate(onPlay: (c) => c.repeat(reverse: true))
            .fadeIn(duration: 600.ms),

        const SizedBox(height: 8),
      ],
    );
  }
}

// ─── Spin Wheel Widget ────────────────────────────────────────────────────────
class _SpinWheel extends StatelessWidget {
  final double angle;
  final List<_WheelSegment> segments;
  final bool isSpinning;

  const _SpinWheel({
    required this.angle,
    required this.segments,
    required this.isSpinning,
  });

  @override
  Widget build(BuildContext context) {
    return AspectRatio(
      aspectRatio: 1,
      child: Stack(
        alignment: Alignment.center,
        children: [
          // Outer glow
          AnimatedContainer(
            duration: const Duration(milliseconds: 300),
            decoration: BoxDecoration(
              shape: BoxShape.circle,
              boxShadow: [
                BoxShadow(
                  color: AppColors.brand500.withOpacity(isSpinning ? 0.6 : 0.3),
                  blurRadius: isSpinning ? 40 : 20,
                  spreadRadius: isSpinning ? 10 : 2,
                ),
              ],
            ),
          ),

          // Wheel
          Transform.rotate(
            angle: angle,
            child: CustomPaint(
              painter: _WheelPainter(segments: segments),
              child: Container(),
            ),
          ),

          // Center hub
          Container(
            width: 50,
            height: 50,
            decoration: BoxDecoration(
              gradient: AppColors.brandGradient,
              shape: BoxShape.circle,
              boxShadow: [
                BoxShadow(
                  color: AppColors.brand700.withOpacity(0.5),
                  blurRadius: 12,
                  spreadRadius: 2,
                ),
              ],
            ),
            child: const Icon(Icons.bolt_rounded, color: Colors.white, size: 26),
          ),

          // Pointer (top center)
          Positioned(
            top: -2,
            child: _WheelPointer(),
          ),
        ],
      ),
    );
  }
}

class _WheelPointer extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Container(
      width: 24,
      height: 32,
      child: CustomPaint(painter: _PointerPainter()),
    );
  }
}

class _PointerPainter extends CustomPainter {
  @override
  void paint(Canvas canvas, Size size) {
    final paint = Paint()
      ..color = AppColors.gold500
      ..style = PaintingStyle.fill;

    final shadow = Paint()
      ..color = AppColors.gold600.withOpacity(0.4)
      ..maskFilter = const MaskFilter.blur(BlurStyle.normal, 4);

    final path = Path()
      ..moveTo(size.width / 2, size.height)
      ..lineTo(0, 0)
      ..lineTo(size.width, 0)
      ..close();

    canvas.drawPath(path, shadow);
    canvas.drawPath(path, paint);
  }

  @override
  bool shouldRepaint(_) => false;
}

// ─── Wheel Painter ────────────────────────────────────────────────────────────
class _WheelPainter extends CustomPainter {
  final List<_WheelSegment> segments;

  _WheelPainter({required this.segments});

  @override
  void paint(Canvas canvas, Size size) {
    final center = Offset(size.width / 2, size.height / 2);
    final radius = min(size.width, size.height) / 2;
    final segmentAngle = 2 * pi / segments.length;
    final textPainter = TextPainter(textDirection: TextDirection.ltr);

    for (int i = 0; i < segments.length; i++) {
      final seg = segments[i];
      final startAngle = i * segmentAngle - pi / 2;

      // Segment fill
      final paint = Paint()
        ..color = seg.color
        ..style = PaintingStyle.fill;

      canvas.drawArc(
        Rect.fromCircle(center: center, radius: radius),
        startAngle,
        segmentAngle,
        true,
        paint,
      );

      // Segment border
      final borderPaint = Paint()
        ..color = Colors.white.withOpacity(0.2)
        ..style = PaintingStyle.stroke
        ..strokeWidth = 1.5;

      canvas.drawArc(
        Rect.fromCircle(center: center, radius: radius),
        startAngle,
        segmentAngle,
        true,
        borderPaint,
      );

      // Label + icon
      final midAngle = startAngle + segmentAngle / 2;
      final labelRadius = radius * 0.65;
      final labelX = center.dx + labelRadius * cos(midAngle);
      final labelY = center.dy + labelRadius * sin(midAngle);

      canvas.save();
      canvas.translate(labelX, labelY);
      canvas.rotate(midAngle + pi / 2);

      // Icon / emoji
      textPainter.text = TextSpan(
        text: seg.icon,
        style: const TextStyle(fontSize: 18),
      );
      textPainter.layout();
      textPainter.paint(canvas, Offset(-textPainter.width / 2, -textPainter.height - 2));

      // Label text
      textPainter.text = TextSpan(
        text: seg.label,
        style: const TextStyle(
          color: Colors.white,
          fontSize: 11,
          fontWeight: FontWeight.w700,
          shadows: [Shadow(color: Colors.black38, blurRadius: 3)],
        ),
      );
      textPainter.layout();
      textPainter.paint(canvas, Offset(-textPainter.width / 2, 4));

      canvas.restore();
    }
  }

  @override
  bool shouldRepaint(_WheelPainter old) => false;
}

// ─── Bottom Bar ───────────────────────────────────────────────────────────────
class _SpinBottomBar extends ConsumerStatefulWidget {
  final SpinState spinState;
  final VoidCallback onSpin;

  const _SpinBottomBar({required this.spinState, required this.onSpin});

  @override
  ConsumerState<_SpinBottomBar> createState() => _SpinBottomBarState();
}

class _SpinBottomBarState extends ConsumerState<_SpinBottomBar>
    with SingleTickerProviderStateMixin {
  late TabController _tabController;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 2, vsync: this);
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final isSpinning = widget.spinState.status == SpinStatus.spinning;
    final canSpin = widget.spinState.eligible && !isSpinning;

    return Container(
      height: 200,
      padding: const EdgeInsets.fromLTRB(16, 0, 16, 0),
      child: Column(
        children: [
          // SPIN button
          Padding(
            padding: const EdgeInsets.symmetric(vertical: 12),
            child: AppGradientButton(
              label: isSpinning ? '🎰 Spinning...' : '🎰 SPIN!',
              onPressed: canSpin ? widget.onSpin : null,
              isLoading: isSpinning,
              height: 56,
              gradient: const LinearGradient(
                colors: [AppColors.gold500, AppColors.gold600],
              ),
              textStyle: AppTextStyles.headingMd.copyWith(
                color: Colors.white,
                fontWeight: FontWeight.w900,
              ),
            ),
          ),

          // Tab: How to unlock | History
          TabBar(
            controller: _tabController,
            tabs: const [
              Tab(text: 'How to Unlock'),
              Tab(text: 'My History'),
            ],
          ),

          Expanded(
            child: TabBarView(
              controller: _tabController,
              children: [
                _UnlockInfo(),
                _SpinHistoryList(),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _UnlockInfo extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.all(8),
      child: Text(
        'Recharge ₦1,000 or more to unlock 1 free spin. More recharges = more spins!',
        style: AppTextStyles.bodyMd.copyWith(
          color: Colors.white.withOpacity(0.7),
        ),
        textAlign: TextAlign.center,
      ),
    );
  }
}

class _SpinHistoryList extends ConsumerWidget {
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final historyAsync = ref.watch(spinHistoryProvider);
    return historyAsync.when(
      data: (history) {
        if (history.isEmpty) {
          return Center(
            child: Text(
              'No spins yet',
              style: AppTextStyles.bodyMd.copyWith(color: Colors.white54),
            ),
          );
        }
        return ListView.builder(
          itemCount: history.length.clamp(0, 5),
          itemBuilder: (context, i) {
            final item = history[i] as Map<String, dynamic>;
            final prize = item['prize'] ?? item['prize_name'] ?? '';
            final amount = (item['prize_amount'] ?? 0) as num;
            return ListTile(
              dense: true,
              leading: const Text('🎰', style: TextStyle(fontSize: 18)),
              title: Text(prize.toString(), style: AppTextStyles.labelMd.copyWith(color: Colors.white)),
              trailing: amount > 0
                  ? Text(amount.toInt().toNaira(), style: AppTextStyles.labelMd.copyWith(color: AppColors.gold400, fontWeight: FontWeight.w700))
                  : null,
            );
          },
        );
      },
      loading: () => const Center(child: CircularProgressIndicator(color: Colors.white54, strokeWidth: 2)),
      error: (_, __) => const SizedBox.shrink(),
    );
  }
}

// ─── Ineligible State ─────────────────────────────────────────────────────────
class _IneligibleState extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.all(32),
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Container(
            width: 100,
            height: 100,
            decoration: BoxDecoration(
              color: Colors.white.withOpacity(0.08),
              shape: BoxShape.circle,
            ),
            child: const Icon(Icons.lock_rounded, color: Colors.white54, size: 48),
          ).animate().scale(begin: const Offset(0.7, 0.7), duration: 400.ms, curve: Curves.elasticOut),

          const SizedBox(height: 24),

          Text(
            'Spin Wheel Locked',
            style: AppTextStyles.headingXl.copyWith(
              color: Colors.white,
              fontWeight: FontWeight.w700,
            ),
            textAlign: TextAlign.center,
          ),

          const SizedBox(height: 12),

          Text(
            'Recharge ₦1,000 or more to unlock the spin wheel and win amazing prizes!',
            style: AppTextStyles.bodyLg.copyWith(
              color: Colors.white.withOpacity(0.6),
            ),
            textAlign: TextAlign.center,
          ),

          const SizedBox(height: 32),

          AppGradientButton(
            label: 'Recharge Now',
            onPressed: () => context.go('/recharge'),
            icon: const Icon(Icons.bolt_rounded, color: Colors.white, size: 18),
          ).animate(delay: 300.ms).fadeIn().slideY(begin: 0.2, end: 0),
        ],
      ),
    );
  }
}

// ─── Result Bottom Sheet ──────────────────────────────────────────────────────
class _SpinResultSheet extends StatelessWidget {
  final Map<String, dynamic> result;
  final VoidCallback onClose;

  const _SpinResultSheet({required this.result, required this.onClose});

  @override
  Widget build(BuildContext context) {
    final prize = result['prize'] ?? result['prize_name'] ?? 'Better luck next time!';
    final amount = (result['prize_amount'] ?? result['amount'] ?? 0) as num;
    final isWin = amount > 0;

    return Container(
      padding: const EdgeInsets.all(28),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(
            isWin ? '🎉' : '😊',
            style: const TextStyle(fontSize: 56),
          ).animate().scale(begin: const Offset(0.5, 0.5), duration: 400.ms, curve: Curves.elasticOut),

          const SizedBox(height: 16),

          Text(
            isWin ? 'You Won!' : 'Try Again',
            style: AppTextStyles.displaySm.copyWith(fontWeight: FontWeight.w800),
          ),

          const SizedBox(height: 8),

          Text(
            prize.toString(),
            style: AppTextStyles.headingXl.copyWith(
              color: isWin ? AppColors.gold500 : AppColors.textSecondary,
              fontWeight: FontWeight.w700,
            ),
            textAlign: TextAlign.center,
          ),

          if (amount > 0) ...[
            const SizedBox(height: 4),
            Text(
              amount.toInt().toNaira(),
              style: AppTextStyles.displayMd.copyWith(
                color: AppColors.gold500,
                fontWeight: FontWeight.w800,
              ),
            ),
          ],

          const SizedBox(height: 28),

          AppGradientButton(
            label: isWin ? 'Claim & Continue' : 'Try Again',
            onPressed: onClose,
          ),

          const SizedBox(height: 12),
        ],
      ),
    );
  }
}
