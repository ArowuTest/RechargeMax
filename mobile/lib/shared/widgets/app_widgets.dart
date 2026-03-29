import 'package:flutter/material.dart';
import '../../core/theme/app_colors.dart';
import '../../core/theme/app_text_styles.dart';

// ─── Gradient Button ──────────────────────────────────────────────────────────
class AppGradientButton extends StatefulWidget {
  final String label;
  final VoidCallback? onPressed;
  final bool isLoading;
  final double? width;
  final double height;
  final Widget? icon;
  final Gradient? gradient;
  final double borderRadius;
  final TextStyle? textStyle;

  const AppGradientButton({
    super.key,
    required this.label,
    this.onPressed,
    this.isLoading = false,
    this.width,
    this.height = 52,
    this.icon,
    this.gradient,
    this.borderRadius = 14,
    this.textStyle,
  });

  @override
  State<AppGradientButton> createState() => _AppGradientButtonState();
}

class _AppGradientButtonState extends State<AppGradientButton>
    with SingleTickerProviderStateMixin {
  late AnimationController _controller;
  late Animation<double> _scale;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 100),
    );
    _scale = Tween<double>(begin: 1.0, end: 0.97).animate(
      CurvedAnimation(parent: _controller, curve: Curves.easeInOut),
    );
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final gradient = widget.gradient ?? AppColors.brandGradient;
    final isEnabled = widget.onPressed != null && !widget.isLoading;

    return GestureDetector(
      onTapDown: isEnabled ? (_) => _controller.forward() : null,
      onTapUp: isEnabled ? (_) => _controller.reverse() : null,
      onTapCancel: isEnabled ? () => _controller.reverse() : null,
      onTap: isEnabled ? widget.onPressed : null,
      child: AnimatedBuilder(
        animation: _scale,
        builder: (context, child) => Transform.scale(
          scale: _scale.value,
          child: child,
        ),
        child: AnimatedContainer(
          duration: const Duration(milliseconds: 200),
          width: widget.width ?? double.infinity,
          height: widget.height,
          decoration: BoxDecoration(
            gradient: isEnabled ? gradient : null,
            color: isEnabled ? null : AppColors.slate200,
            borderRadius: BorderRadius.circular(widget.borderRadius),
            boxShadow: isEnabled
                ? [
                    BoxShadow(
                      color: AppColors.brand600.withOpacity(0.35),
                      blurRadius: 16,
                      offset: const Offset(0, 6),
                    )
                  ]
                : [],
          ),
          child: Center(
            child: widget.isLoading
                ? const SizedBox(
                    width: 22,
                    height: 22,
                    child: CircularProgressIndicator(
                      strokeWidth: 2.5,
                      valueColor: AlwaysStoppedAnimation(Colors.white),
                    ),
                  )
                : Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      if (widget.icon != null) ...[
                        widget.icon!,
                        const SizedBox(width: 8),
                      ],
                      Text(
                        widget.label,
                        style: (widget.textStyle ?? AppTextStyles.labelXl).copyWith(
                          color: isEnabled ? Colors.white : AppColors.textDisabled,
                          fontWeight: FontWeight.w700,
                        ),
                      ),
                    ],
                  ),
          ),
        ),
      ),
    );
  }
}

// ─── Glass Card ───────────────────────────────────────────────────────────────
class GlassCard extends StatelessWidget {
  final Widget child;
  final EdgeInsets? padding;
  final double borderRadius;
  final Color? backgroundColor;
  final double backgroundOpacity;
  final bool hasBorder;
  final VoidCallback? onTap;

  const GlassCard({
    super.key,
    required this.child,
    this.padding,
    this.borderRadius = 16,
    this.backgroundColor,
    this.backgroundOpacity = 0.08,
    this.hasBorder = true,
    this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;

    return GestureDetector(
      onTap: onTap,
      child: Container(
        padding: padding ?? const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: (backgroundColor ?? (isDark ? Colors.white : AppColors.brand900))
              .withOpacity(backgroundOpacity),
          borderRadius: BorderRadius.circular(borderRadius),
          border: hasBorder
              ? Border.all(
                  color: isDark
                      ? Colors.white.withOpacity(0.1)
                      : AppColors.brand800.withOpacity(0.2),
                  width: 1,
                )
              : null,
        ),
        child: child,
      ),
    );
  }
}

// ─── App Card (Light / content sections) ─────────────────────────────────────
class AppCard extends StatelessWidget {
  final Widget child;
  final EdgeInsets? padding;
  final double borderRadius;
  final Color? backgroundColor;
  final VoidCallback? onTap;
  final bool hasShadow;

  const AppCard({
    super.key,
    required this.child,
    this.padding,
    this.borderRadius = 16,
    this.backgroundColor,
    this.onTap,
    this.hasShadow = true,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final isDark = theme.brightness == Brightness.dark;

    return GestureDetector(
      onTap: onTap,
      child: Container(
        padding: padding ?? const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: backgroundColor ?? (isDark ? AppColors.darkBgCard : AppColors.bgPrimary),
          borderRadius: BorderRadius.circular(borderRadius),
          border: Border.all(
            color: isDark ? AppColors.darkBorderSecondary : AppColors.borderSecondary,
            width: 1,
          ),
          boxShadow: hasShadow
              ? [
                  BoxShadow(
                    color: Colors.black.withOpacity(isDark ? 0.2 : 0.04),
                    blurRadius: 12,
                    offset: const Offset(0, 2),
                  ),
                ]
              : [],
        ),
        child: child,
      ),
    );
  }
}

// ─── Tier Badge ───────────────────────────────────────────────────────────────
class TierBadge extends StatelessWidget {
  final String tier;
  final bool large;

  const TierBadge({super.key, required this.tier, this.large = false});

  static Color _tierColor(String tier) => switch (tier.toUpperCase()) {
    'SILVER' => AppColors.tierSilver,
    'GOLD' => AppColors.tierGold,
    'PLATINUM' => AppColors.tierPlatinum,
    _ => AppColors.tierBronze,
  };

  static IconData _tierIcon(String tier) => switch (tier.toUpperCase()) {
    'SILVER' => Icons.star_rounded,
    'GOLD' => Icons.workspace_premium_rounded,
    'PLATINUM' => Icons.diamond_rounded,
    _ => Icons.military_tech_rounded,
  };

  @override
  Widget build(BuildContext context) {
    final color = _tierColor(tier);
    final size = large ? 14.0 : 11.0;
    final padding = large
        ? const EdgeInsets.symmetric(horizontal: 10, vertical: 4)
        : const EdgeInsets.symmetric(horizontal: 7, vertical: 3);

    return Container(
      padding: padding,
      decoration: BoxDecoration(
        color: color.withOpacity(0.15),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: color.withOpacity(0.4), width: 1),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(_tierIcon(tier), size: size + 1, color: color),
          const SizedBox(width: 4),
          Text(
            tier.toUpperCase(),
            style: AppTextStyles.labelXs.copyWith(
              color: color,
              fontSize: size,
              fontWeight: FontWeight.w700,
              letterSpacing: 0.8,
            ),
          ),
        ],
      ),
    );
  }
}

// ─── Points Display ──────────────────────────────────────────────────────────
class PointsDisplay extends StatelessWidget {
  final int points;
  final bool compact;

  const PointsDisplay({super.key, required this.points, this.compact = false});

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.center,
      children: [
        Container(
          padding: const EdgeInsets.all(4),
          decoration: BoxDecoration(
            color: AppColors.gold500.withOpacity(0.2),
            shape: BoxShape.circle,
          ),
          child: Icon(
            Icons.star_rounded,
            color: AppColors.gold500,
            size: compact ? 14 : 18,
          ),
        ),
        const SizedBox(width: 6),
        Text(
          points.toLocaleString(),
          style: compact
              ? AppTextStyles.labelLg.copyWith(
                  color: AppColors.gold500,
                  fontWeight: FontWeight.w700,
                )
              : AppTextStyles.headingMd.copyWith(
                  color: AppColors.gold500,
                  fontWeight: FontWeight.w800,
                ),
        ),
        if (!compact) ...[
          const SizedBox(width: 4),
          Text(
            'pts',
            style: AppTextStyles.bodyMd.copyWith(
              color: AppColors.gold400.withOpacity(0.8),
            ),
          ),
        ],
      ],
    );
  }
}

// ─── Shimmer Loader ───────────────────────────────────────────────────────────
class ShimmerBox extends StatelessWidget {
  final double width;
  final double height;
  final double borderRadius;

  const ShimmerBox({
    super.key,
    required this.width,
    required this.height,
    this.borderRadius = 8,
  });

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    return Container(
      width: width,
      height: height,
      decoration: BoxDecoration(
        color: isDark ? AppColors.darkBgTertiary : AppColors.slate100,
        borderRadius: BorderRadius.circular(borderRadius),
      ),
    );
  }
}

// ─── Section Header ───────────────────────────────────────────────────────────
class SectionHeader extends StatelessWidget {
  final String title;
  final String? actionLabel;
  final VoidCallback? onAction;

  const SectionHeader({
    super.key,
    required this.title,
    this.actionLabel,
    this.onAction,
  });

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Text(
          title,
          style: AppTextStyles.headingMd.copyWith(
            color: isDark ? AppColors.darkTextPrimary : AppColors.textPrimary,
          ),
        ),
        if (actionLabel != null)
          GestureDetector(
            onTap: onAction,
            child: Text(
              actionLabel!,
              style: AppTextStyles.labelMd.copyWith(
                color: AppColors.brand500,
                fontWeight: FontWeight.w600,
              ),
            ),
          ),
      ],
    );
  }
}

// ─── Empty State ─────────────────────────────────────────────────────────────
class AppEmptyState extends StatelessWidget {
  final IconData icon;
  final String title;
  final String? subtitle;
  final String? actionLabel;
  final VoidCallback? onAction;

  const AppEmptyState({
    super.key,
    required this.icon,
    required this.title,
    this.subtitle,
    this.actionLabel,
    this.onAction,
  });

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(32),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              width: 80,
              height: 80,
              decoration: BoxDecoration(
                color: AppColors.brand500.withOpacity(0.1),
                shape: BoxShape.circle,
              ),
              child: Icon(icon, size: 36, color: AppColors.brand400),
            ),
            const SizedBox(height: 20),
            Text(
              title,
              style: AppTextStyles.headingMd.copyWith(
                color: isDark ? AppColors.darkTextPrimary : AppColors.textPrimary,
              ),
              textAlign: TextAlign.center,
            ),
            if (subtitle != null) ...[
              const SizedBox(height: 8),
              Text(
                subtitle!,
                style: AppTextStyles.bodyMd.copyWith(
                  color: isDark ? AppColors.darkTextTertiary : AppColors.textTertiary,
                ),
                textAlign: TextAlign.center,
              ),
            ],
            if (actionLabel != null) ...[
              const SizedBox(height: 24),
              AppGradientButton(
                label: actionLabel!,
                onPressed: onAction,
                width: 200,
                height: 44,
              ),
            ],
          ],
        ),
      ),
    );
  }
}

// ─── Network Badge ────────────────────────────────────────────────────────────
class NetworkBadge extends StatelessWidget {
  final String network;
  final bool selected;

  const NetworkBadge({super.key, required this.network, this.selected = false});

  static Color _networkColor(String network) => switch (network.toLowerCase()) {
    'mtn' => const Color(0xFFFFCB05),
    'glo' => const Color(0xFF009A44),
    'airtel' => const Color(0xFFE30613),
    '9mobile' => const Color(0xFF006B3F),
    _ => AppColors.brand500,
  };

  @override
  Widget build(BuildContext context) {
    final color = _networkColor(network);
    return AnimatedContainer(
      duration: const Duration(milliseconds: 200),
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
      decoration: BoxDecoration(
        color: selected ? color.withOpacity(0.15) : Colors.transparent,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(
          color: selected ? color : AppColors.borderSecondary,
          width: selected ? 2 : 1,
        ),
      ),
      child: Text(
        network.toUpperCase(),
        style: AppTextStyles.labelMd.copyWith(
          color: selected ? color : AppColors.textSecondary,
          fontWeight: selected ? FontWeight.w700 : FontWeight.w500,
        ),
      ),
    );
  }
}

// ─── Extension: Number formatting ─────────────────────────────────────────────
extension IntFormatting on int {
  String toLocaleString() {
    final str = toString();
    final buffer = StringBuffer();
    int count = 0;
    for (int i = str.length - 1; i >= 0; i--) {
      if (count > 0 && count % 3 == 0) buffer.write(',');
      buffer.write(str[i]);
      count++;
    }
    return buffer.toString().split('').reversed.join();
  }

  String toNaira() => '₦${toLocaleString()}';
}
