/**
 * PlatformSettingsPage
 * A focused, category-organised view for all platform_settings.
 * Replaces the flat list in ComprehensiveAdminPortal "settings" tab.
 */

import React, { useState, useEffect, useCallback } from 'react';
import { apiClient } from '@/lib/api-client';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { Switch } from '@/components/ui/switch';
import { useToast } from '@/hooks/useToast';
import { Settings, RefreshCw, Save, ChevronDown, ChevronRight } from 'lucide-react';



interface Setting {
  key: string;
  value: string;
  description?: string;
}

interface CategoryGroup {
  label: string;
  description: string;
  keys: string[];
}

const CATEGORY_GROUPS: CategoryGroup[] = [
  {
    label: 'Loyalty Tiers',
    description: 'Points thresholds and draw-entry multipliers for each tier',
    keys: [
      'loyalty.bronze_min_points',
      'loyalty.silver_min_points',
      'loyalty.silver_multiplier',
      'loyalty.gold_min_points',
      'loyalty.gold_multiplier',
      'loyalty.platinum_min_points',
      'loyalty.platinum_multiplier',
    ],
  },
  {
    label: 'Affiliate Programme',
    description: 'Commission rates and payout settings',
    keys: [
      'affiliate.commission_rate',
      'affiliate.auto_release_days',
      'affiliate.min_payout_amount',
      'affiliate_program_enabled',
      'commission_payout_threshold',
    ],
  },
  {
    label: 'Draw System',
    description: 'Draw entry allocation and claim window',
    keys: [
      'draw_system_enabled',
      'draw.claim_window_days',
      'draw.max_entries_per_msisdn',
      'draw_entries_per_200_points',
      'points.draw_entries_per_point',
      'points.naira_per_point',
      'naira_per_point',
    ],
  },
  {
    label: 'Spin Wheel',
    description: 'Spin wheel availability and limits',
    keys: [
      'spin_wheel_enabled',
      'spin_wheel_minimum',
      'spin.daily_spin_limit',
      'spin.min_recharge_kobo',
    ],
  },
  {
    label: 'Daily Subscription',
    description: 'Subscription pricing and draw entry rewards',
    keys: [
      'daily_subscription_enabled',
      'daily_subscription_amount',
      'daily_subscription_naira_per_point',
      'subscription.price_kobo',
      'subscription.points_per_sub',
      'subscription.draw_entries',
    ],
  },
  {
    label: 'Recharge Limits',
    description: 'Min/max recharge amounts and daily limits',
    keys: [
      'minimum_recharge_amount',
      'maximum_recharge_amount',
      'max_daily_amount_per_user',
      'max_daily_transactions_per_user',
      'guest_recharge_enabled',
      'points.min_recharge_kobo',
    ],
  },
  {
    label: 'USSD',
    description: 'USSD channel settings',
    keys: [
      'ussd.points_per_200_naira',
      'ussd.draw_entries_per_200_naira',
    ],
  },
  {
    label: 'Network / HLR',
    description: 'Network detection settings',
    keys: [
      'network.hlr_enabled',
      'network.hlr_timeout_seconds',
    ],
  },
  {
    label: 'Platform',
    description: 'General platform settings',
    keys: [
      'platform_name',
      'platform_tagline',
      'support_email',
      'support_phone',
      'registration_enabled',
      'maintenance_mode',
      'prize_claim_expiry_days',
    ],
  },
];

// Bool keys that should render as toggles
const BOOL_KEYS = new Set([
  'draw_system_enabled',
  'spin_wheel_enabled',
  'daily_subscription_enabled',
  'affiliate_program_enabled',
  'guest_recharge_enabled',
  'registration_enabled',
  'maintenance_mode',
  'network.hlr_enabled',
]);

function isBoolKey(key: string): boolean {
  return BOOL_KEYS.has(key);
}

export const PlatformSettingsPage: React.FC = () => {
  const { toast } = useToast();
  const [settings, setSettings] = useState<Record<string, Setting>>({});
  const [editedValues, setEditedValues] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState<string | null>(null);
  const [collapsed, setCollapsed] = useState<Record<string, boolean>>({});

  const fetchSettings = useCallback(async () => {
    setLoading(true);
    try {
      const res = await apiClient.get('/admin/settings');
      const data = res.data;
      // data.data may be a flat array or nested map — normalise to key->Setting
      const map: Record<string, Setting> = {};
      if (Array.isArray(data.data)) {
        for (const s of data.data) {
          const k = s.setting_key || s.key;
          map[k] = { key: k, value: String(s.setting_value ?? s.value ?? ''), description: s.description };
        }
      } else if (data.data && typeof data.data === 'object') {
        // Nested by category
        for (const [cat, items] of Object.entries(data.data as Record<string, any>)) {
          if (Array.isArray(items)) {
            for (const s of items) {
              const k = s.setting_key || s.key;
              map[k] = { key: k, value: String(s.setting_value ?? s.value ?? ''), description: s.description };
            }
          }
        }
      }
      setSettings(map);
      setEditedValues({});
    } catch (e) {
      toast({ title: 'Error', description: 'Failed to load settings', variant: 'destructive' });
    } finally {
      setLoading(false);
    }
  }, [toast]);

  useEffect(() => { fetchSettings(); }, [fetchSettings]);

  const getValue = (key: string): string => {
    if (key in editedValues) return editedValues[key] ?? '';
    return settings[key]?.value ?? '';
  };

  const handleChange = (key: string, value: string) => {
    setEditedValues(prev => ({ ...prev, [key]: value }));
  };

  const handleToggle = async (key: string, checked: boolean) => {
    const value = checked ? 'true' : 'false';
    setEditedValues(prev => ({ ...prev, [key]: value }));
    await saveSetting(key, value);
  };

  const saveSetting = async (key: string, value?: string) => {
    const val = value ?? editedValues[key] ?? settings[key]?.value ?? '';  // already typed as string
    setSaving(key);
    try {
      const res = await apiClient.put('/admin/settings', { [key]: val });
      const data = res.data;
      if (data.success) {
        setSettings(prev => ({
          ...prev,
          [key]: { ...(prev[key] ?? { key, description: '' }), key, value: val },
        }));
        setEditedValues(prev => { const n = { ...prev }; delete n[key]; return n; });
        toast({ title: 'Saved', description: `${key} updated` });
      } else {
        toast({ title: 'Error', description: data.message || 'Save failed', variant: 'destructive' });
      }
    } catch {
      toast({ title: 'Error', description: 'Network error saving setting', variant: 'destructive' });
    } finally {
      setSaving(null);
    }
  };

  const hasPendingChanges = Object.keys(editedValues).length > 0;

  const toggleCategory = (label: string) => {
    setCollapsed(prev => ({ ...prev, [label]: !prev[label] }));
  };

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-xl font-bold flex items-center gap-2">
            <Settings className="w-5 h-5" /> Platform Settings
          </h2>
          <p className="text-sm text-gray-500 mt-1">All configurable rates and feature flags — changes are live immediately</p>
        </div>
        <div className="flex gap-2">
          {hasPendingChanges && (
            <Badge variant="secondary" className="animate-pulse">
              {Object.keys(editedValues).length} unsaved change{Object.keys(editedValues).length !== 1 ? 's' : ''}
            </Badge>
          )}
          <Button variant="outline" size="sm" onClick={fetchSettings} disabled={loading}>
            <RefreshCw className={`w-4 h-4 mr-1 ${loading ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
        </div>
      </div>

      {/* Category groups */}
      {CATEGORY_GROUPS.map(group => (
        <Card key={group.label}>
          <CardHeader
            className="cursor-pointer select-none"
            onClick={() => toggleCategory(group.label)}
          >
            <CardTitle className="flex items-center justify-between text-base">
              <span>{group.label}</span>
              {collapsed[group.label]
                ? <ChevronRight className="w-4 h-4 text-gray-400" />
                : <ChevronDown className="w-4 h-4 text-gray-400" />
              }
            </CardTitle>
            <CardDescription>{group.description}</CardDescription>
          </CardHeader>

          {!collapsed[group.label] && (
            <CardContent>
              <div className="divide-y">
                {group.keys.map(key => {
                  const current = getValue(key);
                  const isDirty = key in editedValues;
                  const boolVal = current === 'true';

                  return (
                    <div key={key} className="flex items-center justify-between py-3 gap-4">
                      <div className="flex-1 min-w-0">
                        <code className="text-sm font-mono text-gray-700">{key}</code>
                        {settings[key]?.description && (
                          <p className="text-xs text-gray-400 mt-0.5">{settings[key].description}</p>
                        )}
                        {!(key in settings) && (
                          <p className="text-xs text-amber-500 mt-0.5">Not yet in database</p>
                        )}
                      </div>

                      <div className="flex items-center gap-2 shrink-0">
                        {isBoolKey(key) ? (
                          <Switch
                            checked={boolVal}
                            onCheckedChange={checked => handleToggle(key, checked)}
                            disabled={saving === key}
                          />
                        ) : (
                          <>
                            <Input
                              value={current}
                              onChange={e => handleChange(key, e.target.value)}
                              className={`w-36 text-sm ${isDirty ? 'border-amber-400 bg-amber-50' : ''}`}
                              disabled={saving === key}
                            />
                            {isDirty && (
                              <Button
                                size="sm"
                                onClick={() => saveSetting(key)}
                                disabled={saving === key}
                              >
                                <Save className="w-3 h-3 mr-1" />
                                {saving === key ? 'Saving…' : 'Save'}
                              </Button>
                            )}
                          </>
                        )}
                      </div>
                    </div>
                  );
                })}
              </div>
            </CardContent>
          )}
        </Card>
      ))}
    </div>
  );
};

export default PlatformSettingsPage;
