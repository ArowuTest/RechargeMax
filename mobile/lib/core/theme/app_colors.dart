import 'package:flutter/material.dart';

/// RechargeMax Design Tokens
/// Extracted directly from the live website: rechargemax-frontend.onrender.com
///
/// Primary palette: Deep Purple + Violet + Amber Gold
/// Background: Dark purple-black for hero sections, clean white for content
abstract class AppColors {
  // ─── Brand Purple Scale ─────────────────────────────────────────────────────
  static const Color brand25 = Color(0xFFF5F3FF);
  static const Color brand50 = Color(0xFFEDE9FE);
  static const Color brand100 = Color(0xFFDDD6FE);
  static const Color brand200 = Color(0xFFC4B5FD);
  static const Color brand300 = Color(0xFFA78BFA);
  static const Color brand400 = Color(0xFF8B5CF6);
  static const Color brand500 = Color(0xFF7C3AED); // Primary brand
  static const Color brand600 = Color(0xFF6D28D9); // Interactive / CTA
  static const Color brand700 = Color(0xFF5B21B6); // Hover
  static const Color brand800 = Color(0xFF4C1D95);
  static const Color brand900 = Color(0xFF2D1B69); // Deep dark
  static const Color brand950 = Color(0xFF1A0533); // Hero background

  // ─── Amber / Gold (Prize / Win moments) ───────────────────────────────────
  static const Color gold50 = Color(0xFFFFFBEB);
  static const Color gold100 = Color(0xFFFEF3C7);
  static const Color gold300 = Color(0xFFFCD34D);
  static const Color gold400 = Color(0xFFFBBF24);
  static const Color gold500 = Color(0xFFF59E0B); // Prize pool banner
  static const Color gold600 = Color(0xFFD97706);

  // ─── Success / Win Green ──────────────────────────────────────────────────
  static const Color success50 = Color(0xFFECFDF5);
  static const Color success100 = Color(0xFFD1FAE5);
  static const Color success400 = Color(0xFF34D399);
  static const Color success500 = Color(0xFF10B981);
  static const Color success600 = Color(0xFF059669);

  // ─── Error / Alert Red ────────────────────────────────────────────────────
  static const Color error50 = Color(0xFFFEF2F2);
  static const Color error100 = Color(0xFFFEE2E2);
  static const Color error400 = Color(0xFFF87171);
  static const Color error500 = Color(0xFFEF4444);
  static const Color error600 = Color(0xFFDC2626);

  // ─── Warning Orange ──────────────────────────────────────────────────────
  static const Color warning400 = Color(0xFFFB923C);
  static const Color warning500 = Color(0xFFF97316);

  // ─── Neutral / Slate ─────────────────────────────────────────────────────
  static const Color slate50 = Color(0xFFF8FAFC);
  static const Color slate100 = Color(0xFFF1F5F9);
  static const Color slate200 = Color(0xFFE2E8F0);
  static const Color slate300 = Color(0xFFCBD5E1);
  static const Color slate400 = Color(0xFF94A3B8);
  static const Color slate500 = Color(0xFF64748B);
  static const Color slate600 = Color(0xFF475569);
  static const Color slate700 = Color(0xFF334155);
  static const Color slate800 = Color(0xFF1E293B);
  static const Color slate900 = Color(0xFF0F172A);

  // ─── Semantic Light Mode ──────────────────────────────────────────────────
  static const Color bgPrimary = Color(0xFFFFFFFF);
  static const Color bgSecondary = Color(0xFFF8FAFC);
  static const Color bgTertiary = Color(0xFFF1F5F9);
  static const Color bgBrandSection = brand950;
  static const Color bgBrandHero = brand900;

  static const Color textPrimary = Color(0xFF0F172A);
  static const Color textSecondary = Color(0xFF334155);
  static const Color textTertiary = Color(0xFF64748B);
  static const Color textDisabled = Color(0xFF94A3B8);
  static const Color textOnBrand = Color(0xFFFFFFFF);

  static const Color borderPrimary = Color(0xFFCBD5E1);
  static const Color borderSecondary = Color(0xFFE2E8F0);

  // ─── Semantic Dark Mode ───────────────────────────────────────────────────
  static const Color darkBgPrimary = Color(0xFF1A0533);    // deepest bg
  static const Color darkBgSecondary = Color(0xFF2D1B69);  // card bg
  static const Color darkBgTertiary = Color(0xFF3D2080);   // elevated
  static const Color darkBgCard = Color(0xFF231645);        // glass card
  static const Color darkTextPrimary = Color(0xFFF8FAFC);
  static const Color darkTextSecondary = Color(0xFFCBD5E1);
  static const Color darkTextTertiary = Color(0xFF94A3B8);
  static const Color darkBorderPrimary = Color(0xFF4C1D95);
  static const Color darkBorderSecondary = Color(0xFF3D2080);

  // ─── Gradients ────────────────────────────────────────────────────────────
  static const LinearGradient heroGradient = LinearGradient(
    begin: Alignment.topLeft,
    end: Alignment.bottomRight,
    colors: [brand950, brand900, Color(0xFF3D1A7A)],
  );

  static const LinearGradient brandGradient = LinearGradient(
    begin: Alignment.topLeft,
    end: Alignment.bottomRight,
    colors: [brand500, brand700],
  );

  static const LinearGradient goldGradient = LinearGradient(
    begin: Alignment.topLeft,
    end: Alignment.bottomRight,
    colors: [gold400, gold600],
  );

  static const LinearGradient successGradient = LinearGradient(
    begin: Alignment.topLeft,
    end: Alignment.bottomRight,
    colors: [success400, success600],
  );

  static const RadialGradient heroRadialGlow = RadialGradient(
    center: Alignment(0.6, -0.3),
    radius: 0.8,
    colors: [Color(0x55A855F7), Color(0x00000000)],
  );

  // ─── Tier Colors ─────────────────────────────────────────────────────────
  static const Color tierBronze = Color(0xFFCD7F32);
  static const Color tierSilver = Color(0xFF9CA3AF);
  static const Color tierGold = Color(0xFFF59E0B);
  static const Color tierPlatinum = Color(0xFF7C3AED);
}
