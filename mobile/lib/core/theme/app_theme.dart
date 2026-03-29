import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'app_colors.dart';
import 'app_text_styles.dart';

class AppTheme {
  static ThemeData get light => _buildTheme(Brightness.light);
  static ThemeData get dark => _buildTheme(Brightness.dark);

  static ThemeData _buildTheme(Brightness brightness) {
    final isDark = brightness == Brightness.dark;

    final colorScheme = isDark
        ? const ColorScheme.dark(
            primary: AppColors.brand500,
            onPrimary: Colors.white,
            primaryContainer: AppColors.brand800,
            onPrimaryContainer: AppColors.brand100,
            secondary: AppColors.gold500,
            onSecondary: AppColors.brand950,
            secondaryContainer: AppColors.gold100,
            surface: AppColors.darkBgSecondary,
            onSurface: AppColors.darkTextPrimary,
            surfaceContainerHighest: AppColors.darkBgTertiary,
            outline: AppColors.darkBorderPrimary,
            error: AppColors.error500,
          )
        : const ColorScheme.light(
            primary: AppColors.brand500,
            onPrimary: Colors.white,
            primaryContainer: AppColors.brand50,
            onPrimaryContainer: AppColors.brand800,
            secondary: AppColors.gold500,
            onSecondary: AppColors.brand950,
            secondaryContainer: AppColors.gold50,
            surface: AppColors.bgPrimary,
            onSurface: AppColors.textPrimary,
            surfaceContainerHighest: AppColors.bgTertiary,
            outline: AppColors.borderPrimary,
            error: AppColors.error500,
          );

    return ThemeData(
      useMaterial3: true,
      brightness: brightness,
      colorScheme: colorScheme,
      fontFamily: AppTextStyles.fontFamily,

      // ─── AppBar ───────────────────────────────────────────────────────────
      appBarTheme: AppBarTheme(
        elevation: 0,
        scrolledUnderElevation: 0,
        centerTitle: false,
        backgroundColor: isDark ? AppColors.darkBgPrimary : AppColors.bgPrimary,
        foregroundColor: isDark ? AppColors.darkTextPrimary : AppColors.textPrimary,
        systemOverlayStyle: isDark
            ? SystemUiOverlayStyle.light.copyWith(
                statusBarColor: Colors.transparent,
                systemNavigationBarColor: AppColors.darkBgPrimary,
              )
            : SystemUiOverlayStyle.dark.copyWith(
                statusBarColor: Colors.transparent,
                systemNavigationBarColor: AppColors.bgPrimary,
              ),
        titleTextStyle: AppTextStyles.headingLg.copyWith(
          color: isDark ? AppColors.darkTextPrimary : AppColors.textPrimary,
        ),
      ),

      // ─── Bottom Navigation ────────────────────────────────────────────────
      bottomNavigationBarTheme: BottomNavigationBarThemeData(
        backgroundColor: isDark ? AppColors.darkBgSecondary : AppColors.bgPrimary,
        selectedItemColor: AppColors.brand500,
        unselectedItemColor: isDark ? AppColors.darkTextTertiary : AppColors.textTertiary,
        showSelectedLabels: true,
        showUnselectedLabels: true,
        type: BottomNavigationBarType.fixed,
        elevation: 0,
        selectedLabelStyle: AppTextStyles.labelSm.copyWith(fontWeight: FontWeight.w600),
        unselectedLabelStyle: AppTextStyles.labelSm,
      ),

      // ─── Navigation Bar (Material 3) ──────────────────────────────────────
      navigationBarTheme: NavigationBarThemeData(
        backgroundColor: isDark ? AppColors.darkBgSecondary : AppColors.bgPrimary,
        indicatorColor: AppColors.brand500.withOpacity(0.15),
        iconTheme: WidgetStateProperty.resolveWith((states) {
          if (states.contains(WidgetState.selected)) {
            return const IconThemeData(color: AppColors.brand500, size: 24);
          }
          return IconThemeData(
            color: isDark ? AppColors.darkTextTertiary : AppColors.textTertiary,
            size: 24,
          );
        }),
        labelTextStyle: WidgetStateProperty.resolveWith((states) {
          if (states.contains(WidgetState.selected)) {
            return AppTextStyles.labelSm.copyWith(
              color: AppColors.brand500,
              fontWeight: FontWeight.w600,
            );
          }
          return AppTextStyles.labelSm.copyWith(
            color: isDark ? AppColors.darkTextTertiary : AppColors.textTertiary,
          );
        }),
        elevation: 0,
        shadowColor: Colors.transparent,
      ),

      // ─── Cards ────────────────────────────────────────────────────────────
      cardTheme: CardThemeData(
        elevation: 0,
        color: isDark ? AppColors.darkBgCard : AppColors.bgPrimary,
        surfaceTintColor: Colors.transparent,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(16),
          side: BorderSide(
            color: isDark ? AppColors.darkBorderSecondary : AppColors.borderSecondary,
            width: 1,
          ),
        ),
        margin: EdgeInsets.zero,
        clipBehavior: Clip.antiAlias,
      ),

      // ─── Elevated Button ──────────────────────────────────────────────────
      elevatedButtonTheme: ElevatedButtonThemeData(
        style: ElevatedButton.styleFrom(
          backgroundColor: AppColors.brand500,
          foregroundColor: Colors.white,
          disabledBackgroundColor: isDark ? AppColors.darkBgTertiary : AppColors.slate200,
          disabledForegroundColor: isDark ? AppColors.darkTextTertiary : AppColors.textDisabled,
          elevation: 0,
          shadowColor: Colors.transparent,
          padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 14),
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
          textStyle: AppTextStyles.labelLg.copyWith(fontWeight: FontWeight.w600),
          minimumSize: const Size(0, 48),
        ),
      ),

      // ─── Outlined Button ──────────────────────────────────────────────────
      outlinedButtonTheme: OutlinedButtonThemeData(
        style: OutlinedButton.styleFrom(
          foregroundColor: AppColors.brand500,
          side: const BorderSide(color: AppColors.brand500, width: 1.5),
          elevation: 0,
          padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 14),
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
          textStyle: AppTextStyles.labelLg.copyWith(fontWeight: FontWeight.w600),
          minimumSize: const Size(0, 48),
        ),
      ),

      // ─── Text Button ──────────────────────────────────────────────────────
      textButtonTheme: TextButtonThemeData(
        style: TextButton.styleFrom(
          foregroundColor: AppColors.brand500,
          textStyle: AppTextStyles.labelLg.copyWith(fontWeight: FontWeight.w600),
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
        ),
      ),

      // ─── Input Decoration ─────────────────────────────────────────────────
      inputDecorationTheme: InputDecorationTheme(
        filled: true,
        fillColor: isDark ? AppColors.darkBgTertiary : AppColors.bgSecondary,
        contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: BorderSide(
            color: isDark ? AppColors.darkBorderPrimary : AppColors.borderPrimary,
          ),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: BorderSide(
            color: isDark ? AppColors.darkBorderSecondary : AppColors.borderSecondary,
          ),
        ),
        focusedBorder: const OutlineInputBorder(
          borderRadius: BorderRadius.all(Radius.circular(12)),
          borderSide: BorderSide(color: AppColors.brand500, width: 2),
        ),
        errorBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: AppColors.error500),
        ),
        focusedErrorBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: AppColors.error500, width: 2),
        ),
        labelStyle: AppTextStyles.labelLg.copyWith(
          color: isDark ? AppColors.darkTextSecondary : AppColors.textSecondary,
        ),
        hintStyle: AppTextStyles.bodyLg.copyWith(color: AppColors.textDisabled),
        prefixIconColor: isDark ? AppColors.darkTextTertiary : AppColors.textTertiary,
        suffixIconColor: isDark ? AppColors.darkTextTertiary : AppColors.textTertiary,
      ),

      // ─── Chip ─────────────────────────────────────────────────────────────
      chipTheme: ChipThemeData(
        backgroundColor: isDark ? AppColors.darkBgTertiary : AppColors.bgTertiary,
        selectedColor: AppColors.brand500.withOpacity(0.15),
        labelStyle: AppTextStyles.labelMd,
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
        side: BorderSide.none,
      ),

      // ─── Divider ──────────────────────────────────────────────────────────
      dividerTheme: DividerThemeData(
        color: isDark ? AppColors.darkBorderSecondary : AppColors.borderSecondary,
        thickness: 1,
        space: 1,
      ),

      // ─── SnackBar ─────────────────────────────────────────────────────────
      snackBarTheme: SnackBarThemeData(
        backgroundColor: isDark ? AppColors.darkBgCard : AppColors.slate800,
        contentTextStyle: AppTextStyles.bodyMd.copyWith(color: Colors.white),
        behavior: SnackBarBehavior.floating,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
        elevation: 4,
      ),

      // ─── Dialog ───────────────────────────────────────────────────────────
      dialogTheme: DialogThemeData(
        backgroundColor: isDark ? AppColors.darkBgSecondary : AppColors.bgPrimary,
        surfaceTintColor: Colors.transparent,
        elevation: 0,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
      ),

      // ─── Bottom Sheet ─────────────────────────────────────────────────────
      bottomSheetTheme: BottomSheetThemeData(
        backgroundColor: isDark ? AppColors.darkBgSecondary : AppColors.bgPrimary,
        surfaceTintColor: Colors.transparent,
        shape: const RoundedRectangleBorder(
          borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
        ),
        elevation: 0,
        showDragHandle: true,
        dragHandleColor: isDark ? AppColors.darkBorderPrimary : AppColors.borderPrimary,
      ),

      // ─── Switch ───────────────────────────────────────────────────────────
      switchTheme: SwitchThemeData(
        thumbColor: WidgetStateProperty.resolveWith((states) {
          if (states.contains(WidgetState.selected)) return Colors.white;
          return isDark ? AppColors.darkTextTertiary : AppColors.textDisabled;
        }),
        trackColor: WidgetStateProperty.resolveWith((states) {
          if (states.contains(WidgetState.selected)) return AppColors.brand500;
          return isDark ? AppColors.darkBgTertiary : AppColors.slate200;
        }),
      ),

      // ─── Progress Indicator ───────────────────────────────────────────────
      progressIndicatorTheme: const ProgressIndicatorThemeData(
        color: AppColors.brand500,
        linearTrackColor: AppColors.brand100,
        circularTrackColor: AppColors.brand100,
      ),

      // ─── Tab Bar ──────────────────────────────────────────────────────────
      tabBarTheme: TabBarThemeData(
        labelColor: AppColors.brand500,
        unselectedLabelColor: isDark ? AppColors.darkTextTertiary : AppColors.textTertiary,
        labelStyle: AppTextStyles.labelLg.copyWith(fontWeight: FontWeight.w600),
        unselectedLabelStyle: AppTextStyles.labelLg,
        indicatorColor: AppColors.brand500,
        indicatorSize: TabBarIndicatorSize.label,
        dividerColor: isDark ? AppColors.darkBorderSecondary : AppColors.borderSecondary,
      ),
    );
  }
}
