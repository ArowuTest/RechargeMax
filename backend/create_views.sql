-- Create views to map clean table names to timestamped tables
-- This allows GORM entities to work with the migrated database

CREATE OR REPLACE VIEW admin_users AS SELECT * FROM admin_users_2026_01_30_14_00;
CREATE OR REPLACE VIEW admin_sessions AS SELECT * FROM admin_sessions_2026_01_30_14_00;
CREATE OR REPLACE VIEW admin_activity_logs AS SELECT * FROM admin_activity_logs_2026_01_30_14_00;
CREATE OR REPLACE VIEW affiliates AS SELECT * FROM affiliates_2026_01_30_14_00;
CREATE OR REPLACE VIEW affiliate_commissions AS SELECT * FROM affiliate_commissions_2026_01_30_14_00;
CREATE OR REPLACE VIEW affiliate_payouts AS SELECT * FROM affiliate_payouts_2026_01_30_14_00;
CREATE OR REPLACE VIEW affiliate_clicks AS SELECT * FROM affiliate_clicks_2026_01_30_14_00;
CREATE OR REPLACE VIEW affiliate_bank_accounts AS SELECT * FROM affiliate_bank_accounts_2026_01_30_14_00;
CREATE OR REPLACE VIEW affiliate_analytics AS SELECT * FROM affiliate_analytics_2026_01_30_14_00;
CREATE OR REPLACE VIEW transactions AS SELECT * FROM transactions_2026_01_30_14_00;
CREATE OR REPLACE VIEW draws AS SELECT * FROM draws_2026_01_30_14_00;
CREATE OR REPLACE VIEW draw_entries AS SELECT * FROM draw_entries_2026_01_30_14_00;
CREATE OR REPLACE VIEW draw_winners AS SELECT * FROM draw_winners_2026_01_30_14_00;
CREATE OR REPLACE VIEW daily_subscriptions AS SELECT * FROM daily_subscriptions_2026_01_30_14_00;
CREATE OR REPLACE VIEW daily_subscription_config AS SELECT * FROM daily_subscription_config_2026_01_30_14_00;
CREATE OR REPLACE VIEW data_plans AS SELECT * FROM data_plans_2026_01_30_14_00;
CREATE OR REPLACE VIEW network_configs AS SELECT * FROM network_configs_2026_01_30_14_00;
CREATE OR REPLACE VIEW spin_results AS SELECT * FROM spin_results_2026_01_30_14_00;
CREATE OR REPLACE VIEW payment_logs AS SELECT * FROM payment_logs_2026_01_30_14_00;
CREATE OR REPLACE VIEW platform_settings AS SELECT * FROM platform_settings_2026_01_30_14_00;
CREATE OR REPLACE VIEW notification_templates AS SELECT * FROM notification_templates_2026_01_30_14_00;
CREATE OR REPLACE VIEW notification_delivery_log AS SELECT * FROM notification_delivery_log_2026_01_30_14_00;
CREATE OR REPLACE VIEW user_notification_preferences AS SELECT * FROM user_notification_preferences_2026_01_30_14_00;
CREATE OR REPLACE VIEW otp_verifications AS SELECT * FROM otp_verifications_2026_01_30_14_00;
CREATE OR REPLACE VIEW application_logs AS SELECT * FROM application_logs_2026_01_30_14_00;
CREATE OR REPLACE VIEW application_metrics AS SELECT * FROM application_metrics_2026_01_30_14_00;
