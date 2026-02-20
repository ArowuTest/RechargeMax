import React, { useState, useEffect } from 'react';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Switch } from '@/components/ui/switch';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Loader2, Trophy, AlertTriangle, CheckCircle } from 'lucide-react';

interface SpinTier {
  id?: string;
  tier_name: string;
  tier_display_name: string;
  min_daily_amount: number;
  max_daily_amount: number | null;
  spins_per_day: number;
  tier_color: string;
  tier_icon: string;
  tier_badge: string;
  tier_description: string;
  sort_order: number;
  is_active: boolean;
}

interface SpinTierDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  tier?: SpinTier | null;
  existingTiers: SpinTier[];
  onSave: (tierData: Omit<SpinTier, 'id' | 'created_at' | 'updated_at'>) => Promise<void>;
  loading?: boolean;
}

const TIER_COLORS = [
  { value: '#CD7F32', label: 'Bronze', name: 'bronze' },
  { value: '#C0C0C0', label: 'Silver', name: 'silver' },
  { value: '#FFD700', label: 'Gold', name: 'gold' },
  { value: '#E5E4E2', label: 'Platinum', name: 'platinum' },
  { value: '#B9F2FF', label: 'Diamond', name: 'diamond' },
  { value: '#50C878', label: 'Emerald', name: 'emerald' },
  { value: '#E0115F', label: 'Ruby', name: 'ruby' },
  { value: '#9966CC', label: 'Amethyst', name: 'amethyst' }
];

const TIER_ICONS = ['🥉', '🥈', '🥇', '💎', '👑', '⭐', '🏆', '🎖️'];

export const SpinTierDialog: React.FC<SpinTierDialogProps> = ({
  open,
  onOpenChange,
  tier,
  existingTiers,
  onSave,
  loading = false
}) => {
  const [formData, setFormData] = useState<Omit<SpinTier, 'id' | 'created_at' | 'updated_at'>>({
    tier_name: '',
    tier_display_name: '',
    min_daily_amount: 100000, // ₦1,000 in kobo
    max_daily_amount: 499900, // ₦4,999 in kobo
    spins_per_day: 1,
    tier_color: '#CD7F32',
    tier_icon: '🥉',
    tier_badge: 'BRONZE',
    tier_description: '',
    sort_order: 1,
    is_active: true
  });

  const [errors, setErrors] = useState<Record<string, string>>({});
  const [rangeWarning, setRangeWarning] = useState<string>('');
  const [hasMaxAmount, setHasMaxAmount] = useState(true);

  useEffect(() => {
    if (tier) {
      setFormData({
        tier_name: tier.tier_name,
        tier_display_name: tier.tier_display_name,
        min_daily_amount: tier.min_daily_amount,
        max_daily_amount: tier.max_daily_amount,
        spins_per_day: tier.spins_per_day,
        tier_color: tier.tier_color,
        tier_icon: tier.tier_icon,
        tier_badge: tier.tier_badge,
        tier_description: tier.tier_description,
        sort_order: tier.sort_order,
        is_active: tier.is_active
      });
      setHasMaxAmount(tier.max_daily_amount !== null);
    } else {
      // Set defaults for new tier
      const nextSortOrder = Math.max(...existingTiers.map(t => t.sort_order), 0) + 1;
      setFormData({
        tier_name: '',
        tier_display_name: '',
        min_daily_amount: 100000,
        max_daily_amount: 499900,
        spins_per_day: 1,
        tier_color: '#CD7F32',
        tier_icon: '🥉',
        tier_badge: 'BRONZE',
        tier_description: '',
        sort_order: nextSortOrder,
        is_active: true
      });
      setHasMaxAmount(true);
    }
    setErrors({});
  }, [tier, open, existingTiers]);

  useEffect(() => {
    // Check for range overlaps
    if (formData.min_daily_amount > 0) {
      const otherTiers = existingTiers.filter(t => t.id !== tier?.id && t.is_active);
      const maxAmount = hasMaxAmount ? formData.max_daily_amount : null;
      
      const overlapping = otherTiers.find(t => {
        const tMin = t.min_daily_amount;
        const tMax = t.max_daily_amount;
        
        // Check if ranges overlap
        if (maxAmount === null) {
          // Current tier is unlimited, check if it overlaps with others
          return tMax === null || tMax >= formData.min_daily_amount;
        } else {
          if (tMax === null) {
            // Other tier is unlimited
            return tMin <= maxAmount;
          } else {
            // Both have limits
            return (formData.min_daily_amount <= tMax && maxAmount >= tMin);
          }
        }
      });

      if (overlapping) {
        setRangeWarning(`Range overlaps with ${overlapping.tier_display_name} tier`);
      } else {
        setRangeWarning('');
      }
    }
  }, [formData.min_daily_amount, formData.max_daily_amount, hasMaxAmount, existingTiers, tier]);

  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.tier_name.trim()) {
      newErrors.tier_name = 'Tier name is required';
    }

    if (!formData.tier_display_name.trim()) {
      newErrors.tier_display_name = 'Display name is required';
    }

    if (formData.min_daily_amount < 0) {
      newErrors.min_daily_amount = 'Minimum amount cannot be negative';
    }

    if (hasMaxAmount && formData.max_daily_amount) {
      if (formData.max_daily_amount <= formData.min_daily_amount) {
        newErrors.max_daily_amount = 'Maximum must be greater than minimum';
      }
    }

    if (formData.spins_per_day < 1) {
      newErrors.spins_per_day = 'Must have at least 1 spin per day';
    }

    if (formData.spins_per_day > 100) {
      newErrors.spins_per_day = 'Cannot exceed 100 spins per day';
    }

    if (!formData.tier_badge.trim()) {
      newErrors.tier_badge = 'Badge text is required';
    }

    if (formData.sort_order < 1) {
      newErrors.sort_order = 'Sort order must be at least 1';
    }

    if (rangeWarning) {
      newErrors.range = rangeWarning;
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!validateForm()) {
      return;
    }

    try {
      const tierData = {
        ...formData,
        max_daily_amount: hasMaxAmount ? formData.max_daily_amount : null
      };
      await onSave(tierData);
      onOpenChange(false);
    } catch (error) {
      console.error('Failed to save spin tier:', error);
    }
  };

  const handleInputChange = (field: keyof typeof formData, value: any) => {
    setFormData(prev => ({ ...prev, [field]: value }));
    
    // Clear error for this field
    if (errors[field]) {
      setErrors(prev => ({ ...prev, [field]: '' }));
    }
  };

  const formatAmountDisplay = (kobo: number): string => {
    return `₦${(kobo / 100).toLocaleString()}`;
  };

  const handleAmountChange = (field: 'min_daily_amount' | 'max_daily_amount', nairaValue: string) => {
    const numValue = parseFloat(nairaValue) || 0;
    const koboValue = Math.round(numValue * 100);
    handleInputChange(field, koboValue);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Trophy className="w-5 h-5" />
            {tier ? 'Edit Spin Tier' : 'Add New Spin Tier'}
          </DialogTitle>
          <DialogDescription>
            Configure daily recharge tier thresholds and spin rewards
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          {/* Tier Names */}
          <div className="grid grid-cols-2 gap-4">
            <div>
              <Label htmlFor="tier_name">Tier Name (Internal)</Label>
              <Input
                id="tier_name"
                value={formData.tier_name}
                onChange={(e) => handleInputChange('tier_name', e.target.value.toUpperCase())}
                placeholder="e.g., BRONZE"
                className={errors.tier_name ? 'border-red-500' : ''}
              />
              {errors.tier_name && (
                <p className="text-red-500 text-sm mt-1">{errors.tier_name}</p>
              )}
            </div>

            <div>
              <Label htmlFor="tier_display_name">Display Name</Label>
              <Input
                id="tier_display_name"
                value={formData.tier_display_name}
                onChange={(e) => handleInputChange('tier_display_name', e.target.value)}
                placeholder="e.g., Bronze"
                className={errors.tier_display_name ? 'border-red-500' : ''}
              />
              {errors.tier_display_name && (
                <p className="text-red-500 text-sm mt-1">{errors.tier_display_name}</p>
              )}
            </div>
          </div>

          {/* Amount Range */}
          <div className="space-y-3">
            <Label>Daily Recharge Amount Range</Label>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <Label htmlFor="min_daily_amount" className="text-sm text-gray-600">
                  Minimum (₦)
                </Label>
                <Input
                  id="min_daily_amount"
                  type="number"
                  min="0"
                  step="0.01"
                  value={(formData.min_daily_amount / 100).toFixed(2)}
                  onChange={(e) => handleAmountChange('min_daily_amount', e.target.value)}
                  className={errors.min_daily_amount ? 'border-red-500' : ''}
                />
                {errors.min_daily_amount && (
                  <p className="text-red-500 text-sm mt-1">{errors.min_daily_amount}</p>
                )}
              </div>

              <div>
                <Label htmlFor="max_daily_amount" className="text-sm text-gray-600">
                  Maximum (₦)
                </Label>
                <Input
                  id="max_daily_amount"
                  type="number"
                  min="0"
                  step="0.01"
                  value={hasMaxAmount && formData.max_daily_amount ? (formData.max_daily_amount / 100).toFixed(2) : ''}
                  onChange={(e) => handleAmountChange('max_daily_amount', e.target.value)}
                  disabled={!hasMaxAmount}
                  placeholder={hasMaxAmount ? '' : 'Unlimited'}
                  className={errors.max_daily_amount ? 'border-red-500' : ''}
                />
                {errors.max_daily_amount && (
                  <p className="text-red-500 text-sm mt-1">{errors.max_daily_amount}</p>
                )}
              </div>
            </div>

            <div className="flex items-center gap-2">
              <Switch
                checked={!hasMaxAmount}
                onCheckedChange={(checked) => setHasMaxAmount(!checked)}
              />
              <Label className="text-sm">Unlimited maximum (highest tier)</Label>
            </div>

            {formData.min_daily_amount > 0 && (
              <div className="text-sm text-gray-600 bg-gray-50 p-3 rounded">
                <strong>Range:</strong> {formatAmountDisplay(formData.min_daily_amount)} - {
                  hasMaxAmount && formData.max_daily_amount 
                    ? formatAmountDisplay(formData.max_daily_amount)
                    : 'Unlimited'
                }
              </div>
            )}

            {rangeWarning && (
              <Alert className="bg-yellow-50 border-yellow-200">
                <AlertTriangle className="w-4 h-4 text-yellow-600" />
                <AlertDescription className="text-yellow-800">
                  {rangeWarning}
                </AlertDescription>
              </Alert>
            )}
          </div>

          {/* Spins Per Day */}
          <div>
            <Label htmlFor="spins_per_day">Spins Per Day</Label>
            <Input
              id="spins_per_day"
              type="number"
              min="1"
              max="100"
              value={formData.spins_per_day}
              onChange={(e) => handleInputChange('spins_per_day', parseInt(e.target.value) || 1)}
              className={errors.spins_per_day ? 'border-red-500' : ''}
            />
            {errors.spins_per_day && (
              <p className="text-red-500 text-sm mt-1">{errors.spins_per_day}</p>
            )}
          </div>

          {/* Visual Customization */}
          <div className="grid grid-cols-2 gap-4">
            <div>
              <Label htmlFor="tier_color">Tier Color</Label>
              <div className="flex gap-2">
                <Input
                  id="tier_color"
                  type="color"
                  value={formData.tier_color}
                  onChange={(e) => handleInputChange('tier_color', e.target.value)}
                  className="w-20 h-10 p-1 cursor-pointer"
                />
                <select
                  value={formData.tier_color}
                  onChange={(e) => handleInputChange('tier_color', e.target.value)}
                  className="flex-1 border rounded px-3 py-2"
                >
                  {TIER_COLORS.map((color) => (
                    <option key={color.value} value={color.value}>
                      {color.label}
                    </option>
                  ))}
                </select>
              </div>
            </div>

            <div>
              <Label htmlFor="tier_icon">Tier Icon</Label>
              <div className="flex gap-2">
                <div className="w-10 h-10 border rounded flex items-center justify-center text-2xl">
                  {formData.tier_icon}
                </div>
                <select
                  value={formData.tier_icon}
                  onChange={(e) => handleInputChange('tier_icon', e.target.value)}
                  className="flex-1 border rounded px-3 py-2"
                >
                  {TIER_ICONS.map((icon) => (
                    <option key={icon} value={icon}>
                      {icon}
                    </option>
                  ))}
                </select>
              </div>
            </div>
          </div>

          {/* Badge and Description */}
          <div>
            <Label htmlFor="tier_badge">Badge Text</Label>
            <Input
              id="tier_badge"
              value={formData.tier_badge}
              onChange={(e) => handleInputChange('tier_badge', e.target.value.toUpperCase())}
              placeholder="e.g., BRONZE MEMBER"
              className={errors.tier_badge ? 'border-red-500' : ''}
            />
            {errors.tier_badge && (
              <p className="text-red-500 text-sm mt-1">{errors.tier_badge}</p>
            )}
          </div>

          <div>
            <Label htmlFor="tier_description">Description</Label>
            <Textarea
              id="tier_description"
              value={formData.tier_description}
              onChange={(e) => handleInputChange('tier_description', e.target.value)}
              placeholder="Describe this tier's benefits..."
              rows={3}
            />
          </div>

          {/* Sort Order and Active Status */}
          <div className="grid grid-cols-2 gap-4">
            <div>
              <Label htmlFor="sort_order">Sort Order</Label>
              <Input
                id="sort_order"
                type="number"
                min="1"
                value={formData.sort_order}
                onChange={(e) => handleInputChange('sort_order', parseInt(e.target.value) || 1)}
                className={errors.sort_order ? 'border-red-500' : ''}
              />
              {errors.sort_order && (
                <p className="text-red-500 text-sm mt-1">{errors.sort_order}</p>
              )}
              <p className="text-sm text-gray-600 mt-1">
                Lower numbers appear first
              </p>
            </div>

            <div>
              <Label htmlFor="is_active">Status</Label>
              <div className="flex items-center gap-3 mt-2">
                <Switch
                  id="is_active"
                  checked={formData.is_active}
                  onCheckedChange={(checked) => handleInputChange('is_active', checked)}
                />
                <Label htmlFor="is_active" className="text-sm">
                  {formData.is_active ? (
                    <span className="text-green-600 flex items-center gap-1">
                      <CheckCircle className="w-4 h-4" />
                      Active
                    </span>
                  ) : (
                    <span className="text-gray-600">Inactive</span>
                  )}
                </Label>
              </div>
            </div>
          </div>

          {/* Preview */}
          <div className="border rounded-lg p-4 bg-gray-50">
            <Label className="text-sm text-gray-600 mb-2 block">Preview</Label>
            <div className="flex items-center gap-3">
              <div 
                className="w-12 h-12 rounded-full flex items-center justify-center text-white text-xl font-bold"
                style={{ backgroundColor: formData.tier_color }}
              >
                {formData.tier_icon}
              </div>
              <div>
                <div className="font-semibold text-lg">{formData.tier_display_name || 'Tier Name'}</div>
                <div className="text-sm text-gray-600">{formData.tier_badge || 'BADGE'}</div>
                <div className="text-xs text-gray-500 mt-1">
                  {formData.spins_per_day} {formData.spins_per_day === 1 ? 'spin' : 'spins'} per day
                </div>
              </div>
            </div>
          </div>

          {/* Form Errors */}
          {Object.keys(errors).length > 0 && errors.range && (
            <Alert className="bg-red-50 border-red-200">
              <AlertTriangle className="w-4 h-4 text-red-600" />
              <AlertDescription className="text-red-800">
                Please fix the errors above before saving.
              </AlertDescription>
            </Alert>
          )}

          {/* Actions */}
          <div className="flex justify-end gap-3 pt-4 border-t">
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={loading}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={loading || Object.keys(errors).length > 0}>
              {loading ? (
                <>
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                  Saving...
                </>
              ) : (
                <>
                  <CheckCircle className="w-4 h-4 mr-2" />
                  {tier ? 'Update Tier' : 'Create Tier'}
                </>
              )}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
};
