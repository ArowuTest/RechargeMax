import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:rechargemax/shared/widgets/app_widgets.dart';
import 'package:rechargemax/core/theme/app_theme.dart';

// Helper: wraps any widget in a minimal ProviderScope + MaterialApp
Widget testApp(Widget child) {
  return ProviderScope(
    child: MaterialApp(
      theme: AppTheme.dark,
      home: Scaffold(body: child),
    ),
  );
}

void main() {
  group('AppGradientButton', () {
    testWidgets('renders label text', (tester) async {
      await tester.pumpWidget(testApp(
        AppGradientButton(label: 'Recharge Now', onPressed: () {}),
      ));
      expect(find.text('Recharge Now'), findsOneWidget);
    });

    testWidgets('shows CircularProgressIndicator when isLoading=true', (tester) async {
      await tester.pumpWidget(testApp(
        AppGradientButton(label: 'Submit', onPressed: () {}, isLoading: true),
      ));
      expect(find.byType(CircularProgressIndicator), findsOneWidget);
      expect(find.text('Submit'), findsNothing);
    });

    testWidgets('calls onPressed when tapped', (tester) async {
      bool tapped = false;
      await tester.pumpWidget(testApp(
        AppGradientButton(label: 'Go', onPressed: () { tapped = true; }),
      ));
      await tester.tap(find.byType(AppGradientButton));
      await tester.pump();
      expect(tapped, true);
    });

    testWidgets('does not call onPressed when isLoading=true', (tester) async {
      bool tapped = false;
      await tester.pumpWidget(testApp(
        AppGradientButton(label: 'Go', onPressed: () { tapped = true; }, isLoading: true),
      ));
      await tester.tap(find.byType(AppGradientButton));
      await tester.pump();
      expect(tapped, false);
    });

    testWidgets('renders icon when provided', (tester) async {
      await tester.pumpWidget(testApp(
        AppGradientButton(
          label: 'With Icon',
          onPressed: () {},
          icon: const Icon(Icons.send_rounded),
        ),
      ));
      expect(find.byIcon(Icons.send_rounded), findsOneWidget);
    });
  });

  group('GlassCard', () {
    testWidgets('renders child widget', (tester) async {
      await tester.pumpWidget(testApp(
        GlassCard(child: Text('Card Content')),
      ));
      expect(find.text('Card Content'), findsOneWidget);
    });

    testWidgets('renders without overflow', (tester) async {
      await tester.pumpWidget(testApp(
        SizedBox(
          width: 300,
          child: GlassCard(
            child: Column(children: List.generate(3, (i) => Text('Row $i'))),
          ),
        ),
      ));
      expect(tester.takeException(), isNull);
    });
  });

  group('AppCard', () {
    testWidgets('renders child and calls onTap', (tester) async {
      bool tapped = false;
      await tester.pumpWidget(testApp(
        AppCard(
          child: Text('Tap Me'),
          onTap: () { tapped = true; },
        ),
      ));
      expect(find.text('Tap Me'), findsOneWidget);
      await tester.tap(find.text('Tap Me'));
      await tester.pump();
      expect(tapped, true);
    });

    testWidgets('accepts margin parameter without error', (tester) async {
      await tester.pumpWidget(testApp(
        AppCard(
          child: Text('Margined'),
          margin: const EdgeInsets.all(16),
        ),
      ));
      expect(find.text('Margined'), findsOneWidget);
      expect(tester.takeException(), isNull);
    });
  });

  group('TierBadge', () {
    testWidgets('shows tier label for BRONZE', (tester) async {
      await tester.pumpWidget(testApp(TierBadge(tier: 'BRONZE')));
      expect(find.text('BRONZE'), findsOneWidget);
    });

    testWidgets('shows tier label for GOLD', (tester) async {
      await tester.pumpWidget(testApp(TierBadge(tier: 'GOLD')));
      expect(find.text('GOLD'), findsOneWidget);
    });
  });

  group('PointsDisplay', () {
    testWidgets('shows formatted points', (tester) async {
      await tester.pumpWidget(testApp(PointsDisplay(points: 1500)));
      expect(find.textContaining('1,500'), findsOneWidget);
    });

    testWidgets('shows 0 points without crash', (tester) async {
      await tester.pumpWidget(testApp(PointsDisplay(points: 0)));
      expect(find.textContaining('0'), findsOneWidget);
      expect(tester.takeException(), isNull);
    });
  });

  group('ShimmerBox', () {
    testWidgets('renders with given dimensions', (tester) async {
      await tester.pumpWidget(testApp(ShimmerBox(width: 100, height: 20)));
      expect(tester.takeException(), isNull);
    });
  });

  group('SectionHeader', () {
    testWidgets('shows title text', (tester) async {
      await tester.pumpWidget(testApp(SectionHeader(title: 'Recent Recharges')));
      expect(find.text('Recent Recharges'), findsOneWidget);
    });

    testWidgets('shows action text and calls onAction', (tester) async {
      bool acted = false;
      await tester.pumpWidget(testApp(
        SectionHeader(title: 'History', actionLabel: 'See All', onAction: () { acted = true; }),
      ));
      expect(find.text('See All'), findsOneWidget);
      await tester.tap(find.text('See All'));
      await tester.pump();
      expect(acted, true);
    });
  });

  group('AppEmptyState', () {
    testWidgets('shows message', (tester) async {
      await tester.pumpWidget(testApp(
        const AppEmptyState(icon: Icons.inbox_rounded, title: 'No transactions yet'),
      ));
      expect(find.text('No transactions yet'), findsOneWidget);
    });
  });

  group('NetworkBadge', () {
    testWidgets('shows network name', (tester) async {
      await tester.pumpWidget(testApp(NetworkBadge(network: 'MTN')));
      expect(find.text('MTN'), findsOneWidget);
    });
  });
}
