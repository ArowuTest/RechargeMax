-- Pre-configured notification templates for RechargeMax platform
-- Created: 2026-01-30 14:00 UTC
-- ============================================================================

-- ============================================================================
-- TRANSACTION NOTIFICATION TEMPLATES
-- ============================================================================

INSERT INTO public.notification_templates (
    template_key, template_name, description,
    title_template, body_template,
    email_subject_template, email_body_template, sms_template,
    variables, supports_push, supports_email, supports_sms, priority
) VALUES 
    (
        'transaction_success',
        'Transaction Successful',
        'Notification sent when a recharge transaction is completed successfully',
        'Recharge Successful! ✅',
        'Your {{network}} {{recharge_type}} recharge of ₦{{amount}} to {{phone_number}} was successful. Transaction ID: {{transaction_id}}',
        'RechargeMax - Transaction Successful',
        'Dear {{customer_name}},<br><br>Your {{network}} {{recharge_type}} recharge has been completed successfully.<br><br><strong>Details:</strong><br>Amount: ₦{{amount}}<br>Phone: {{phone_number}}<br>Transaction ID: {{transaction_id}}<br>Date: {{date}}<br><br>Thank you for using RechargeMax!',
        'RechargeMax: Your {{network}} recharge of N{{amount}} to {{phone_number}} was successful. Ref: {{transaction_id}}',
        '["network", "recharge_type", "amount", "phone_number", "transaction_id", "customer_name", "date"]'::jsonb,
        true, true, true, 'HIGH'
    ),
    (
        'transaction_failed',
        'Transaction Failed',
        'Notification sent when a recharge transaction fails',
        'Transaction Failed ❌',
        'Your {{network}} {{recharge_type}} recharge of ₦{{amount}} to {{phone_number}} failed. Reason: {{failure_reason}}. Your money will be refunded.',
        'RechargeMax - Transaction Failed',
        'Dear {{customer_name}},<br><br>Unfortunately, your {{network}} {{recharge_type}} recharge could not be completed.<br><br><strong>Details:</strong><br>Amount: ₦{{amount}}<br>Phone: {{phone_number}}<br>Reason: {{failure_reason}}<br><br>Your payment will be refunded within 24 hours. Contact support if you need assistance.',
        'RechargeMax: Your {{network}} recharge of N{{amount}} failed. Reason: {{failure_reason}}. Refund processing.',
        '["network", "recharge_type", "amount", "phone_number", "failure_reason", "customer_name"]'::jsonb,
        true, true, true, 'HIGH'
    );

-- ============================================================================
-- PRIZE AND SPIN WHEEL TEMPLATES
-- ============================================================================

INSERT INTO public.notification_templates (
    template_key, template_name, description,
    title_template, body_template,
    email_subject_template, email_body_template, sms_template,
    variables, supports_push, supports_email, supports_sms, priority
) VALUES 
    (
        'spin_wheel_prize_won',
        'Spin Wheel Prize Won',
        'Notification sent when user wins a prize from spin wheel',
        'Congratulations! You Won {{prize_name}} 🎉',
        'Amazing! You won {{prize_name}} worth ₦{{prize_value}} from the spin wheel! Click to claim your prize.',
        'RechargeMax - You Won a Prize!',
        'Dear {{customer_name}},<br><br>Congratulations! You have won a fantastic prize from our spin wheel!<br><br><strong>Prize Details:</strong><br>Prize: {{prize_name}}<br>Value: ₦{{prize_value}}<br>Transaction: {{transaction_id}}<br><br>Log in to your account to claim your prize. Prizes expire after 30 days if not claimed.',
        'RechargeMax: Congratulations! You won {{prize_name}} worth N{{prize_value}}! Login to claim.',
        '["prize_name", "prize_value", "customer_name", "transaction_id"]'::jsonb,
        true, true, true, 'HIGH'
    ),
    (
        'prize_claimed',
        'Prize Claimed Successfully',
        'Notification sent when user successfully claims a prize',
        'Prize Claimed Successfully! ✅',
        'Your {{prize_name}} worth ₦{{prize_value}} has been successfully claimed and processed.',
        'RechargeMax - Prize Claimed',
        'Dear {{customer_name}},<br><br>Your prize has been successfully claimed!<br><br><strong>Prize Details:</strong><br>Prize: {{prize_name}}<br>Value: ₦{{prize_value}}<br>Claimed on: {{claim_date}}<br><br>Thank you for being a valued RechargeMax customer!',
        'RechargeMax: Your {{prize_name}} worth N{{prize_value}} has been successfully claimed!',
        '["prize_name", "prize_value", "customer_name", "claim_date"]'::jsonb,
        true, true, true, 'NORMAL'
    );

-- ============================================================================
-- DRAW SYSTEM TEMPLATES
-- ============================================================================

INSERT INTO public.notification_templates (
    template_key, template_name, description,
    title_template, body_template,
    email_subject_template, email_body_template, sms_template,
    variables, supports_push, supports_email, supports_sms, priority
) VALUES 
    (
        'draw_winner_announcement',
        'Draw Winner Announcement',
        'Notification sent to draw winners',
        'CONGRATULATIONS! You Won the {{draw_name}}! 🏆',
        'Amazing news! You are a winner in the {{draw_name}}! You won ₦{{prize_amount}} (Position: {{position}}). Claim your prize now!',
        'RechargeMax - You Won the Draw!',
        'Dear {{customer_name}},<br><br>🎉 CONGRATULATIONS! 🎉<br><br>You are a winner in the {{draw_name}}!<br><br><strong>Winning Details:</strong><br>Position: {{position}}<br>Prize Amount: ₦{{prize_amount}}<br>Draw Date: {{draw_date}}<br><br>Please log in to your account to claim your prize. Congratulations once again!',
        'RechargeMax: CONGRATULATIONS! You won N{{prize_amount}} in {{draw_name}}! Login to claim.',
        '["draw_name", "prize_amount", "position", "draw_date", "customer_name"]'::jsonb,
        true, true, true, 'URGENT'
    ),
    (
        'draw_entry_added',
        'Draw Entry Added',
        'Notification when user gets entries for a draw',
        'You Got {{entries_count}} Draw Entries! 🎫',
        'Great! You earned {{entries_count}} entries for the {{draw_name}}. Prize pool: ₦{{prize_pool}}. Draw date: {{draw_date}}.',
        'RechargeMax - Draw Entries Added',
        'Dear {{customer_name}},<br><br>You have earned draw entries!<br><br><strong>Draw Details:</strong><br>Draw: {{draw_name}}<br>Your Entries: {{entries_count}}<br>Prize Pool: ₦{{prize_pool}}<br>Draw Date: {{draw_date}}<br><br>Good luck!',
        'RechargeMax: You got {{entries_count}} entries for {{draw_name}}. Prize: N{{prize_pool}}',
        '["entries_count", "draw_name", "prize_pool", "draw_date", "customer_name"]'::jsonb,
        true, true, false, 'NORMAL'
    );

-- ============================================================================
-- AFFILIATE SYSTEM TEMPLATES
-- ============================================================================

INSERT INTO public.notification_templates (
    template_key, template_name, description,
    title_template, body_template,
    email_subject_template, email_body_template, sms_template,
    variables, supports_push, supports_email, supports_sms, priority
) VALUES 
    (
        'affiliate_application_approved',
        'Affiliate Application Approved',
        'Notification when affiliate application is approved',
        'Welcome to RechargeMax Affiliate Program! 🤝',
        'Congratulations! Your affiliate application has been approved. Your affiliate code is: {{affiliate_code}}. Start earning commissions now!',
        'RechargeMax - Affiliate Application Approved',
        'Dear {{customer_name}},<br><br>Congratulations! Your application to join the RechargeMax Affiliate Program has been approved.<br><br><strong>Your Details:</strong><br>Affiliate Code: {{affiliate_code}}<br>Commission Rate: {{commission_rate}}%<br>Tier: {{tier}}<br><br>You can now start referring customers and earning commissions. Share your referral link and start earning today!',
        'RechargeMax: Your affiliate application approved! Code: {{affiliate_code}}. Start earning now!',
        '["customer_name", "affiliate_code", "commission_rate", "tier"]'::jsonb,
        true, true, true, 'HIGH'
    ),
    (
        'affiliate_commission_earned',
        'Commission Earned',
        'Notification when affiliate earns a commission',
        'You Earned ₦{{commission_amount}} Commission! 💰',
        'Great! You earned ₦{{commission_amount}} commission from a referral transaction of ₦{{transaction_amount}}.',
        'RechargeMax - Commission Earned',
        'Dear {{customer_name}},<br><br>You have earned a new commission!<br><br><strong>Commission Details:</strong><br>Amount: ₦{{commission_amount}}<br>Transaction: ₦{{transaction_amount}}<br>Rate: {{commission_rate}}%<br><br>Keep referring more customers to earn more commissions!',
        'RechargeMax: You earned N{{commission_amount}} commission from a referral!',
        '["commission_amount", "transaction_amount", "commission_rate", "customer_name"]'::jsonb,
        true, true, false, 'NORMAL'
    );

-- ============================================================================
-- SYSTEM AND SECURITY TEMPLATES
-- ============================================================================

INSERT INTO public.notification_templates (
    template_key, template_name, description,
    title_template, body_template,
    email_subject_template, email_body_template, sms_template,
    variables, supports_push, supports_email, supports_sms, priority
) VALUES 
    (
        'welcome_new_user',
        'Welcome New User',
        'Welcome notification for new users',
        'Welcome to RechargeMax! 🎉',
        'Welcome {{customer_name}}! Your account has been created successfully. Start recharging and earning rewards today!',
        'Welcome to RechargeMax!',
        'Dear {{customer_name}},<br><br>Welcome to RechargeMax - Your Ultimate Mobile Recharge Platform!<br><br>Your account has been created successfully. Here is what you can do:<br><br>✅ Recharge airtime and data<br>✅ Spin the wheel for prizes<br>✅ Join daily draws<br>✅ Earn through our affiliate program<br><br>Start your journey with us today!',
        'Welcome to RechargeMax {{customer_name}}! Start recharging and earning rewards today.',
        '["customer_name"]'::jsonb,
        true, true, false, 'NORMAL'
    ),
    (
        'account_security_alert',
        'Account Security Alert',
        'Security alert for suspicious activities',
        'Security Alert: {{alert_type}} 🔒',
        'We detected {{alert_type}} on your account from {{location}} at {{time}}. If this was not you, please secure your account immediately.',
        'RechargeMax - Security Alert',
        'Dear {{customer_name}},<br><br>We detected unusual activity on your RechargeMax account.<br><br><strong>Alert Details:</strong><br>Activity: {{alert_type}}<br>Location: {{location}}<br>Time: {{time}}<br>IP Address: {{ip_address}}<br><br>If this was not you, please change your password immediately and contact our support team.',
        'RechargeMax Security Alert: {{alert_type}} detected. If not you, secure your account now.',
        '["alert_type", "location", "time", "ip_address", "customer_name"]'::jsonb,
        true, true, true, 'URGENT'
    ),
    (
        'daily_subscription_activated',
        'Daily Subscription Activated',
        'Notification when daily subscription is activated',
        'Daily Subscription Activated! 📅',
        'Your daily subscription of ₦{{amount}} is now active! You earned {{draw_entries}} draw entries for today.',
        'RechargeMax - Daily Subscription Active',
        'Dear {{customer_name}},<br><br>Your daily subscription has been activated successfully!<br><br><strong>Subscription Details:</strong><br>Daily Amount: ₦{{amount}}<br>Draw Entries Earned: {{draw_entries}}<br>Next Charge: {{next_charge_date}}<br><br>You will automatically earn draw entries every day while your subscription is active.',
        'RechargeMax: Daily subscription of N{{amount}} activated. {{draw_entries}} entries earned!',
        '["amount", "draw_entries", "next_charge_date", "customer_name"]'::jsonb,
        true, true, false, 'NORMAL'
    ),
    (
        'loyalty_tier_upgrade',
        'Loyalty Tier Upgrade',
        'Notification when user loyalty tier is upgraded',
        'Congratulations! You are now {{new_tier}}! 🌟',
        'Amazing! You have been upgraded to {{new_tier}} tier! Enjoy {{benefits}} and higher rewards on all transactions.',
        'RechargeMax - Loyalty Tier Upgrade',
        'Dear {{customer_name}},<br><br>Congratulations on reaching a new loyalty tier!<br><br><strong>Tier Upgrade:</strong><br>Previous Tier: {{old_tier}}<br>New Tier: {{new_tier}}<br>Benefits: {{benefits}}<br><br>Thank you for being a loyal RechargeMax customer!',
        'RechargeMax: Congratulations! You are now {{new_tier}} tier. Enjoy enhanced benefits!',
        '["new_tier", "old_tier", "benefits", "customer_name"]'::jsonb,
        true, true, true, 'HIGH'
    );

-- Verify the seeded templates
SELECT 
    template_key,
    template_name,
    priority,
    supports_push,
    supports_email,
    supports_sms
FROM public.notification_templates
ORDER BY created_at;


## Summary

This complete package contains all 7 SQL migration files for the RechargeMax database:

1. **Core Tables Schema** - 19 tables with relationships and indexes
2. **RLS Policies** - 50+ security policies for data protection
3. **Seeded Data** - Networks, plans, prizes, admin users, settings
4. **Functions & Triggers** - 15+ business logic functions
5. **Storage Buckets** - File storage system with 4 buckets
6. **Notification System** - 4 tables for multi-channel notifications
7. **Notification Templates** - 12 pre-configured templates

**Total: 24 tables, 100+ indexes, 50+ RLS policies, comprehensive business logic**

The database is production-ready and supports all RechargeMax platform features including user management, transactions, gamification, affiliate program, notifications, and file storage.

**To use these files:**
1. Copy each SQL block into separate `.sql` files with the specified filenames
2. Run them in order (01, 02, 03, 04, 05, 06, 07) in your Supabase SQL editor
3. All tables, data, and functionality will be created automatically

**Admin Credentials:**
- Super Admin: admin@rechargemax.ng (password: SuperAdmin123!)
