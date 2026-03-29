import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:flutter_animate/flutter_animate.dart';
import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_text_styles.dart';
import '../../../../core/auth/auth_provider.dart';
import '../../../../shared/widgets/app_widgets.dart';

class ProfileScreen extends ConsumerWidget {
  const ProfileScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final user = ref.watch(currentUserProvider);
    final isDark = Theme.of(context).brightness == Brightness.dark;

    return Scaffold(
      backgroundColor: isDark ? AppColors.darkBgPrimary : AppColors.bgSecondary,
      body: CustomScrollView(
        slivers: [
          // Header
          SliverToBoxAdapter(
            child: Container(
              decoration: const BoxDecoration(gradient: AppColors.heroGradient),
              padding: EdgeInsets.only(
                top: MediaQuery.of(context).padding.top + 16,
                bottom: 28,
                left: 20,
                right: 20,
              ),
              child: Column(
                children: [
                  // Avatar
                  Container(
                    width: 80,
                    height: 80,
                    decoration: BoxDecoration(
                      gradient: AppColors.brandGradient,
                      shape: BoxShape.circle,
                      boxShadow: [
                        BoxShadow(
                          color: AppColors.brand500.withValues(alpha: 0.4),
                          blurRadius: 16,
                          spreadRadius: 2,
                        ),
                      ],
                    ),
                    child: Center(
                      child: Text(
                        user?.initials ?? '?',
                        style: AppTextStyles.headingXl.copyWith(
                          color: Colors.white,
                          fontWeight: FontWeight.w800,
                        ),
                      ),
                    ),
                  ).animate().scale(begin: const Offset(0.7, 0.7), duration: 400.ms, curve: Curves.elasticOut),

                  const SizedBox(height: 12),

                  Text(
                    user?.displayName ?? 'RechargeMax User',
                    style: AppTextStyles.headingXl.copyWith(
                      color: Colors.white,
                      fontWeight: FontWeight.w700,
                    ),
                  ).animate(delay: 100.ms).fadeIn(),

                  const SizedBox(height: 4),

                  Text(
                    user?.msisdn ?? '',
                    style: AppTextStyles.bodyMd.copyWith(color: AppColors.brand200),
                  ).animate(delay: 150.ms).fadeIn(),

                  const SizedBox(height: 12),

                  // Tier + points row
                  Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      TierBadge(tier: user?.tier ?? 'BRONZE', large: true),
                      const SizedBox(width: 12),
                      PointsDisplay(points: user?.points ?? 0),
                    ],
                  ).animate(delay: 200.ms).fadeIn(),
                ],
              ),
            ),
          ),

          // Menu items
          SliverPadding(
            padding: const EdgeInsets.all(16),
            sliver: SliverList(
              delegate: SliverChildListDelegate([
                _MenuSection(
                  title: 'Account',
                  items: [
                    _MenuItem(
                      icon: Icons.receipt_long_rounded,
                      label: 'Transaction History',
                      subtitle: 'View all your recharges',
                      onTap: () => context.push('/profile/history'),
                    ),
                    _MenuItem(
                      icon: Icons.calendar_today_rounded,
                      label: 'Daily Subscription',
                      subtitle: 'Manage ₦20/day subscription',
                      badge: 'NEW',
                      onTap: () => context.push('/profile/subscription'),
                    ),
                    _MenuItem(
                      icon: Icons.group_rounded,
                      label: 'Affiliate Program',
                      subtitle: 'Earn 5% on referral recharges',
                      onTap: () => context.push('/profile/affiliate'),
                    ),
                    _MenuItem(
                      icon: Icons.person_outline_rounded,
                      label: 'Edit Profile',
                      subtitle: 'Update your name and email',
                      onTap: () => context.push('/profile-setup'),
                    ),
                  ],
                ),

                const SizedBox(height: 8),

                _MenuSection(
                  title: 'Tier Progress',
                  items: [
                    _TierProgressItem(tier: user?.tier ?? 'BRONZE'),
                  ],
                ),

                const SizedBox(height: 8),

                _MenuSection(
                  title: 'More',
                  items: [
                    _MenuItem(
                      icon: Icons.help_outline_rounded,
                      label: 'Help & Support',
                      onTap: () {},
                    ),
                    _MenuItem(
                      icon: Icons.security_rounded,
                      label: 'Privacy Policy',
                      onTap: () {},
                    ),
                    _MenuItem(
                      icon: Icons.logout_rounded,
                      label: 'Sign Out',
                      color: AppColors.error500,
                      onTap: () => _confirmLogout(context, ref),
                    ),
                  ],
                ),

                const SizedBox(height: 24),

                Center(
                  child: Text(
                    'RechargeMax v1.0.0\nby BridgeTunes',
                    style: AppTextStyles.bodySm.copyWith(color: AppColors.textDisabled),
                    textAlign: TextAlign.center,
                  ),
                ),

                const SizedBox(height: 24),
              ]),
            ),
          ),
        ],
      ),
    );
  }

  Future<void> _confirmLogout(BuildContext context, WidgetRef ref) async {
    final confirm = await showDialog<bool>(
      context: context,
      builder: (_) => AlertDialog(
        title: const Text('Sign Out?'),
        content: const Text('You will need to verify your phone number again to log in.'),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context, false), child: const Text('Cancel')),
          TextButton(
            onPressed: () => Navigator.pop(context, true),
            child: Text('Sign Out', style: TextStyle(color: AppColors.error500)),
          ),
        ],
      ),
    );
    if (confirm == true) {
      await ref.read(authProvider.notifier).logout();
      if (context.mounted) context.go('/login');
    }
  }
}

class _MenuSection extends StatelessWidget {
  final String title;
  final List<Widget> items;

  const _MenuSection({required this.title, required this.items});

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Padding(
          padding: const EdgeInsets.only(left: 4, bottom: 8, top: 8),
          child: Text(
            title.toUpperCase(),
            style: AppTextStyles.labelXs.copyWith(
              color: isDark ? AppColors.darkTextTertiary : AppColors.textTertiary,
            ),
          ),
        ),
        AppCard(
          padding: EdgeInsets.zero,
          child: Column(children: items),
        ),
      ],
    );
  }
}

class _MenuItem extends StatelessWidget {
  final IconData icon;
  final String label;
  final String? subtitle;
  final String? badge;
  final Color? color;
  final VoidCallback onTap;

  const _MenuItem({
    required this.icon,
    required this.label,
    this.subtitle,
    this.badge,
    this.color,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final itemColor = color ?? (isDark ? AppColors.darkTextPrimary : AppColors.textPrimary);

    return InkWell(
      onTap: onTap,
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
        child: Row(
          children: [
            Container(
              width: 40,
              height: 40,
              decoration: BoxDecoration(
                color: itemColor.withValues(alpha: 0.08),
                borderRadius: BorderRadius.circular(10),
              ),
              child: Icon(icon, color: itemColor, size: 20),
            ),
            const SizedBox(width: 14),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Text(
                        label,
                        style: AppTextStyles.labelLg.copyWith(
                          color: itemColor,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                      if (badge != null) ...[
                        const SizedBox(width: 8),
                        Container(
                          padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                          decoration: BoxDecoration(
                            color: AppColors.brand500,
                            borderRadius: BorderRadius.circular(4),
                          ),
                          child: Text(
                            badge!,
                            style: AppTextStyles.labelXs.copyWith(
                              color: Colors.white,
                              fontSize: 9,
                            ),
                          ),
                        ),
                      ],
                    ],
                  ),
                  if (subtitle != null)
                    Text(
                      subtitle!,
                      style: AppTextStyles.bodySm.copyWith(
                        color: isDark ? AppColors.darkTextTertiary : AppColors.textTertiary,
                      ),
                    ),
                ],
              ),
            ),
            if (color == null)
              Icon(
                Icons.chevron_right_rounded,
                color: isDark ? AppColors.darkTextTertiary : AppColors.textTertiary,
                size: 20,
              ),
          ],
        ),
      ),
    );
  }
}

class _TierProgressItem extends StatelessWidget {
  final String tier;
  const _TierProgressItem({required this.tier});

  @override
  Widget build(BuildContext context) {
    const tiers = ['BRONZE', 'SILVER', 'GOLD', 'PLATINUM'];
    final currentIdx = tiers.indexOf(tier.toUpperCase());

    return Padding(
      padding: const EdgeInsets.all(16),
      child: Column(
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: tiers.asMap().entries.map((e) {
              final isActive = e.key <= currentIdx;
              final color = _tierColor(e.value);
              return Column(
                children: [
                  Container(
                    width: 36,
                    height: 36,
                    decoration: BoxDecoration(
                      color: isActive ? color.withValues(alpha: 0.15) : AppColors.bgTertiary,
                      shape: BoxShape.circle,
                      border: Border.all(
                        color: isActive ? color : AppColors.borderSecondary,
                        width: isActive ? 2 : 1,
                      ),
                    ),
                    child: Icon(
                      _tierIcon(e.value),
                      size: 18,
                      color: isActive ? color : AppColors.textDisabled,
                    ),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    e.value.substring(0, 2),
                    style: AppTextStyles.labelXs.copyWith(
                      color: isActive ? color : AppColors.textDisabled,
                      fontWeight: isActive ? FontWeight.w700 : FontWeight.w400,
                    ),
                  ),
                ],
              );
            }).toList(),
          ),
          const SizedBox(height: 8),
          Text(
            'You are at ${tier.toUpperCase()} tier',
            style: AppTextStyles.bodySm.copyWith(color: AppColors.textTertiary),
            textAlign: TextAlign.center,
          ),
        ],
      ),
    );
  }

  Color _tierColor(String tier) => switch (tier) {
    'SILVER' => AppColors.tierSilver,
    'GOLD' => AppColors.tierGold,
    'PLATINUM' => AppColors.tierPlatinum,
    _ => AppColors.tierBronze,
  };

  IconData _tierIcon(String tier) => switch (tier) {
    'SILVER' => Icons.star_rounded,
    'GOLD' => Icons.workspace_premium_rounded,
    'PLATINUM' => Icons.diamond_rounded,
    _ => Icons.military_tech_rounded,
  };
}
