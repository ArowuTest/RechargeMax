import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:flutter_animate/flutter_animate.dart';
import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_text_styles.dart';
import '../../../../core/auth/auth_provider.dart';
import '../../../../core/api/api_client.dart';
import '../../../../shared/widgets/app_widgets.dart';

class ProfileSetupScreen extends ConsumerStatefulWidget {
  const ProfileSetupScreen({super.key});

  @override
  ConsumerState<ProfileSetupScreen> createState() => _ProfileSetupScreenState();
}

class _ProfileSetupScreenState extends ConsumerState<ProfileSetupScreen> {
  final _nameController = TextEditingController();
  final _emailController = TextEditingController();
  final _referralController = TextEditingController();
  final _formKey = GlobalKey<FormState>();
  bool _isLoading = false;

  Future<void> _saveProfile() async {
    if (!(_formKey.currentState?.validate() ?? false)) return;
    setState(() => _isLoading = true);

    try {
      final api = ref.read(apiClientProvider);
      final data = <String, dynamic>{
        'name': _nameController.text.trim(),
      };
      if (_emailController.text.isNotEmpty) {
        data['email'] = _emailController.text.trim();
      }
      if (_referralController.text.isNotEmpty) {
        data['referral_code'] = _referralController.text.trim().toUpperCase();
      }

      await api.updateProfile(data);
      await ref.read(authProvider.notifier).refreshUser();

      if (mounted) context.go('/home');
    } catch (e) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Failed to save profile. Please try again.')),
      );
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  @override
  void dispose() {
    _nameController.dispose();
    _emailController.dispose();
    _referralController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;

    return Scaffold(
      backgroundColor: isDark ? AppColors.darkBgPrimary : AppColors.bgPrimary,
      body: SafeArea(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(24),
          child: Form(
            key: _formKey,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const SizedBox(height: 16),

                // Icon
                Container(
                  width: 72,
                  height: 72,
                  decoration: BoxDecoration(
                    gradient: AppColors.brandGradient,
                    borderRadius: BorderRadius.circular(20),
                  ),
                  child: const Icon(Icons.person_rounded, color: Colors.white, size: 36),
                ).animate().scale(
                  begin: const Offset(0.7, 0.7),
                  duration: 400.ms,
                  curve: Curves.elasticOut,
                ),

                const SizedBox(height: 24),

                Text(
                  'Set up your\nprofile',
                  style: AppTextStyles.displaySm.copyWith(
                    color: isDark ? AppColors.darkTextPrimary : AppColors.textPrimary,
                    fontWeight: FontWeight.w800,
                  ),
                ).animate(delay: 100.ms).fadeIn().slideY(begin: 0.2, end: 0),

                const SizedBox(height: 8),

                Text(
                  'Personalize your account. You can skip and do this later.',
                  style: AppTextStyles.bodyMd.copyWith(
                    color: isDark ? AppColors.darkTextTertiary : AppColors.textTertiary,
                  ),
                ).animate(delay: 200.ms).fadeIn(),

                const SizedBox(height: 36),

                // Name field
                TextFormField(
                  controller: _nameController,
                  textCapitalization: TextCapitalization.words,
                  keyboardType: TextInputType.name,
                  decoration: const InputDecoration(
                    labelText: 'Full Name',
                    hintText: 'e.g. Chidi Okafor',
                    prefixIcon: Icon(Icons.person_outline_rounded),
                  ),
                  validator: (v) {
                    if (v != null && v.isNotEmpty && v.trim().length < 2) {
                      return 'Name must be at least 2 characters';
                    }
                    return null;
                  },
                ).animate(delay: 300.ms).fadeIn().slideY(begin: 0.1, end: 0),

                const SizedBox(height: 16),

                // Email field
                TextFormField(
                  controller: _emailController,
                  keyboardType: TextInputType.emailAddress,
                  decoration: const InputDecoration(
                    labelText: 'Email Address (optional)',
                    hintText: 'you@example.com',
                    prefixIcon: Icon(Icons.email_outlined),
                  ),
                  validator: (v) {
                    if (v != null && v.isNotEmpty) {
                      if (!RegExp(r'^[^@]+@[^@]+\.[^@]+$').hasMatch(v.trim())) {
                        return 'Enter a valid email address';
                      }
                    }
                    return null;
                  },
                ).animate(delay: 350.ms).fadeIn().slideY(begin: 0.1, end: 0),

                const SizedBox(height: 16),

                // Referral code
                TextFormField(
                  controller: _referralController,
                  keyboardType: TextInputType.text,
                  textCapitalization: TextCapitalization.characters,
                  decoration: const InputDecoration(
                    labelText: 'Referral Code (optional)',
                    hintText: 'Enter referral code if you have one',
                    prefixIcon: Icon(Icons.card_giftcard_rounded),
                  ),
                ).animate(delay: 400.ms).fadeIn().slideY(begin: 0.1, end: 0),

                const SizedBox(height: 36),

                AppGradientButton(
                  label: 'Complete Setup',
                  onPressed: _saveProfile,
                  isLoading: _isLoading,
                  icon: const Icon(Icons.check_rounded, color: Colors.white, size: 18),
                ).animate(delay: 500.ms).fadeIn().slideY(begin: 0.2, end: 0),

                const SizedBox(height: 16),

                Center(
                  child: TextButton(
                    onPressed: () => context.go('/home'),
                    child: Text(
                      'Skip for now',
                      style: AppTextStyles.labelLg.copyWith(
                        color: isDark ? AppColors.darkTextTertiary : AppColors.textTertiary,
                      ),
                    ),
                  ),
                ).animate(delay: 600.ms).fadeIn(),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
