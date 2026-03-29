import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../core/theme/app_colors.dart';
import '../core/theme/app_text_styles.dart';

// ─── Bottom Nav State ─────────────────────────────────────────────────────────
final bottomNavIndexProvider = StateProvider<int>((ref) => 0);

class MainScaffold extends ConsumerWidget {
  final Widget child;
  const MainScaffold({super.key, required this.child});

  static const _tabs = [
    _NavTab(icon: Icons.home_outlined, activeIcon: Icons.home_rounded, label: 'Home', path: '/home'),
    _NavTab(icon: Icons.bolt_outlined, activeIcon: Icons.bolt_rounded, label: 'Recharge', path: '/recharge'),
    _NavTab(icon: Icons.casino_outlined, activeIcon: Icons.casino_rounded, label: 'Spin', path: '/spin'),
    _NavTab(icon: Icons.emoji_events_outlined, activeIcon: Icons.emoji_events_rounded, label: 'Draws', path: '/draws'),
    _NavTab(icon: Icons.person_outline_rounded, activeIcon: Icons.person_rounded, label: 'Profile', path: '/profile'),
  ];

  int _locationToIndex(String location) {
    if (location.startsWith('/recharge')) return 1;
    if (location.startsWith('/spin')) return 2;
    if (location.startsWith('/draws')) return 3;
    if (location.startsWith('/profile')) return 4;
    return 0;
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final location = GoRouterState.of(context).matchedLocation;
    final currentIndex = _locationToIndex(location);

    return Scaffold(
      body: child,
      bottomNavigationBar: _AppBottomNav(
        currentIndex: currentIndex,
        tabs: _tabs,
        onTap: (index) => context.go(_tabs[index].path),
      ),
    );
  }
}

class _NavTab {
  final IconData icon;
  final IconData activeIcon;
  final String label;
  final String path;
  const _NavTab({required this.icon, required this.activeIcon, required this.label, required this.path});
}

class _AppBottomNav extends StatelessWidget {
  final int currentIndex;
  final List<_NavTab> tabs;
  final ValueChanged<int> onTap;

  const _AppBottomNav({
    required this.currentIndex,
    required this.tabs,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final isDark = theme.brightness == Brightness.dark;

    return Container(
      decoration: BoxDecoration(
        color: isDark ? AppColors.darkBgSecondary : AppColors.bgPrimary,
        border: Border(
          top: BorderSide(
            color: isDark ? AppColors.darkBorderSecondary : AppColors.borderSecondary,
            width: 1,
          ),
        ),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(isDark ? 0.3 : 0.06),
            blurRadius: 20,
            offset: const Offset(0, -4),
          ),
        ],
      ),
      child: SafeArea(
        child: SizedBox(
          height: 64,
          child: Row(
            children: List.generate(tabs.length, (index) {
              final tab = tabs[index];
              final isSelected = index == currentIndex;
              // Spin tab gets special treatment
              final isSpin = index == 2;

              if (isSpin) {
                return Expanded(
                  child: GestureDetector(
                    onTap: () => onTap(index),
                    behavior: HitTestBehavior.opaque,
                    child: _SpinNavItem(isSelected: isSelected),
                  ),
                );
              }

              return Expanded(
                child: GestureDetector(
                  onTap: () => onTap(index),
                  behavior: HitTestBehavior.opaque,
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      AnimatedSwitcher(
                        duration: const Duration(milliseconds: 200),
                        child: Icon(
                          isSelected ? tab.activeIcon : tab.icon,
                          key: ValueKey(isSelected),
                          size: 24,
                          color: isSelected
                              ? AppColors.brand500
                              : isDark
                                  ? AppColors.darkTextTertiary
                                  : AppColors.textTertiary,
                        ),
                      ),
                      const SizedBox(height: 4),
                      Text(
                        tab.label,
                        style: AppTextStyles.labelXs.copyWith(
                          color: isSelected
                              ? AppColors.brand500
                              : isDark
                                  ? AppColors.darkTextTertiary
                                  : AppColors.textTertiary,
                          fontWeight: isSelected ? FontWeight.w600 : FontWeight.w400,
                        ),
                      ),
                    ],
                  ),
                ),
              );
            }),
          ),
        ),
      ),
    );
  }
}

// Special spin button in bottom nav
class _SpinNavItem extends StatelessWidget {
  final bool isSelected;
  const _SpinNavItem({required this.isSelected});

  @override
  Widget build(BuildContext context) {
    return Column(
      mainAxisAlignment: MainAxisAlignment.center,
      children: [
        AnimatedContainer(
          duration: const Duration(milliseconds: 250),
          width: 48,
          height: 48,
          decoration: BoxDecoration(
            gradient: isSelected
                ? AppColors.brandGradient
                : const LinearGradient(colors: [AppColors.brand400, AppColors.brand600]),
            shape: BoxShape.circle,
            boxShadow: isSelected
                ? [
                    BoxShadow(
                      color: AppColors.brand500.withOpacity(0.4),
                      blurRadius: 12,
                      spreadRadius: 0,
                      offset: const Offset(0, 4),
                    )
                  ]
                : [],
          ),
          child: const Icon(Icons.casino_rounded, color: Colors.white, size: 22),
        ),
        const SizedBox(height: 4),
        Text(
          'Spin',
          style: AppTextStyles.labelXs.copyWith(
            color: isSelected ? AppColors.brand500 : AppColors.textTertiary,
            fontWeight: isSelected ? FontWeight.w600 : FontWeight.w400,
          ),
        ),
      ],
    );
  }
}
