import React, { useState, useEffect } from 'react';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Switch } from '@/components/ui/switch';
import { Badge } from '@/components/ui/badge';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Loader2, Gift, DollarSign, Phone, Wifi, Star, Ticket, AlertTriangle } from 'lucide-react';
import type { WheelPrize } from '@/types/admin-api.types';

interface WheelPrizeDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  prize?: WheelPrize | null;
  existingPrizes: WheelPrize[];
  onSave: (prizeData: Omit<WheelPrize, 'id' | 'created_at' | 'updated_at'>) => Promise<void>;
  loading?: boolean;
}

const PRIZE_TYPES = [
  { value: 'CASH', label: 'Cash Prize', icon: DollarSign, color: 'text-green-600' },
  { value: 'AIRTIME', label: 'Airtime', icon: Phone, color: 'text-blue-600' },
  { value: 'DATA', label: 'Data Bundle', icon: Wifi, color: 'text-purple-600' },
  { value: 'POINTS', label: 'Loyalty Points', icon: Star, color: 'text-yellow-600' },
  { value: 'TICKETS', label: 'Draw Tickets', icon: Ticket, color: 'text-orange-600' }
];

const COLOR_SCHEMES = [
  { value: 'green', label: 'Green', bg: 'bg-green-500', text: 'text-green-600' },
  { value: 'blue', label: 'Blue', bg: 'bg-blue-500', text: 'text-blue-600' },
  { value: 'purple', label: 'Purple', bg: 'bg-purple-500', text: 'text-purple-600' },
  { value: 'yellow', label: 'Yellow', bg: 'bg-yellow-500', text: 'text-yellow-600' },
  { value: 'red', label: 'Red', bg: 'bg-red-500', text: 'text-red-600' },
  { value: 'orange', label: 'Orange', bg: 'bg-orange-500', text: 'text-orange-600' },
  { value: 'pink', label: 'Pink', bg: 'bg-pink-500', text: 'text-pink-600' },
  { value: 'indigo', label: 'Indigo', bg: 'bg-indigo-500', text: 'text-indigo-600' }
];

export const WheelPrizeDialog: React.FC<WheelPrizeDialogProps> = ({
  open,
  onOpenChange,
  prize,
  existingPrizes,
  onSave,
  loading = false
}) => {
  const [formData, setFormData] = useState<Omit<WheelPrize, 'id' | 'created_at' | 'updated_at'>>({
    prize_name: '',
    prize_type: 'CASH',
    prize_value: 0,
    probability_weight: 0,
    probability: 0,
    minimum_recharge: 1000,
    is_active: true,
    display_order: 1,
    sort_order: 1,
    color: 'green',
    color_scheme: 'green',
    icon: 'gift',
    icon_name: 'gift'
  });

  const [errors, setErrors] = useState<Record<string, string>>({});
  const [probabilityWarning, setProbabilityWarning] = useState<string>('');

  useEffect(() => {
    if (prize) {
      setFormData({
        prize_name: prize.prize_name,
        prize_type: prize.prize_type,
        prize_value: prize.prize_value,
        probability_weight: prize.probability_weight,
        probability: prize.probability,
        minimum_recharge: prize.minimum_recharge,
        is_active: prize.is_active,
        display_order: prize.display_order,
        sort_order: prize.sort_order,
        color: prize.color,
        color_scheme: prize.color_scheme,
        icon: prize.icon,
        icon_name: prize.icon_name
      });
    } else {
      setFormData({
        prize_name: '',
        prize_type: 'CASH',
        prize_value: 0,
        probability_weight: 0,
        probability: 0,
        minimum_recharge: 1000,
        is_active: true,
        display_order: 1,
        sort_order: 1,
        color: 'green',
        color_scheme: 'green',
        icon: 'gift',
        icon_name: 'gift'
      });
    }
    setErrors({});
  }, [prize, open]);

  useEffect(() => {
    // Check probability totals when probability changes
    if ((formData.probability ?? 0) > 0) {
      const otherPrizes = existingPrizes.filter(p => p.id !== prize?.id && p.is_active);
      const totalOtherProbability = otherPrizes.reduce((sum, p) => sum + (p.probability ?? 0), 0);
      const newTotal = totalOtherProbability + (formData.probability ?? 0);
      
      if (newTotal > 100) {
        setProbabilityWarning(`Total probability will be ${newTotal.toFixed(1)}% (exceeds 100%)`);
      } else if (newTotal < 99.9) {
        setProbabilityWarning(`Total probability will be ${newTotal.toFixed(1)}% (should equal 100%)`);
      } else {
        setProbabilityWarning('');
      }
    }
  }, [formData.probability, existingPrizes, prize]);

  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.prize_name.trim()) {
      newErrors.prize_name = 'Prize name is required';
    }

    if (formData.prize_value <= 0) {
      newErrors.prize_value = 'Prize value must be greater than 0';
    }

    if ((formData.probability ?? 0) <= 0 || (formData.probability ?? 0) > 100) {
      newErrors.probability = 'Probability must be between 0.1 and 100';
    }

    if ((formData.minimum_recharge ?? 0) < 0) {
      newErrors.minimum_recharge = 'Minimum recharge cannot be negative';
    }

    if ((formData.sort_order ?? 0) < 1) {
      newErrors.sort_order = 'Sort order must be at least 1';
    }

    // Validate prize value based on type
    if (formData.prize_type === 'CASH' && formData.prize_value > 500000) {
      newErrors.prize_value = 'Cash prizes cannot exceed ₦500,000';
    }

    if (formData.prize_type === 'POINTS' && formData.prize_value > 100000) {
      newErrors.prize_value = 'Points prizes cannot exceed 100,000 points';
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
      await onSave(formData);
      onOpenChange(false);
    } catch (error) {
      console.error('Failed to save wheel prize:', error);
    }
  };

  const handleInputChange = (field: keyof typeof formData, value: any) => {
    setFormData(prev => ({ ...prev, [field]: value }));
    
    // Clear error for this field when user starts typing
    if (errors[field]) {
      setErrors(prev => ({ ...prev, [field]: '' }));
    }
  };

  const getSelectedPrizeType = () => {
    return PRIZE_TYPES.find(type => type.value === formData.prize_type);
  };

  const getSelectedColorScheme = () => {
    return COLOR_SCHEMES.find(color => color.value === formData.color_scheme);
  };

  const formatPrizeValue = (type: string, value: number) => {
    switch (type) {
      case 'CASH':
      case 'AIRTIME':
        return `₦${value.toLocaleString()}`;
      case 'DATA':
        return `${value}MB`;
      case 'POINTS':
        return `${value.toLocaleString()} points`;
      case 'TICKETS':
        return `${value} tickets`;
      default:
        return value.toString();
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Gift className="w-5 h-5" />
            {prize ? 'Edit Wheel Prize' : 'Add New Wheel Prize'}
          </DialogTitle>
          <DialogDescription>
            Configure prizes for the spin wheel with probabilities and requirements
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          {/* Prize Name */}
          <div>
            <Label htmlFor="prize_name">Prize Name</Label>
            <Input
              id="prize_name"
              value={formData.prize_name}
              onChange={(e) => handleInputChange('prize_name', e.target.value)}
              placeholder="e.g., ₦100 Cash Prize"
              className={errors.prize_name ? 'border-red-500' : ''}
            />
            {errors.prize_name && (
              <p className="text-red-500 text-sm mt-1">{errors.prize_name}</p>
            )}
          </div>

          {/* Prize Type */}
          <div>
            <Label htmlFor="prize_type">Prize Type</Label>
            <Select 
              value={formData.prize_type} 
              onValueChange={(value) => handleInputChange('prize_type', value)}
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {PRIZE_TYPES.map((type) => {
                  const IconComponent = type.icon;
                  return (
                    <SelectItem key={type.value} value={type.value}>
                      <div className="flex items-center gap-2">
                        <IconComponent className={`w-4 h-4 ${type.color}`} />
                        {type.label}
                      </div>
                    </SelectItem>
                  );
                })}
              </SelectContent>
            </Select>
          </div>

          {/* Prize Value */}
          <div>
            <Label htmlFor="prize_value">
              Prize Value 
              {formData.prize_type === 'CASH' && ' (₦)'}
              {formData.prize_type === 'AIRTIME' && ' (₦)'}
              {formData.prize_type === 'DATA' && ' (MB)'}
              {formData.prize_type === 'POINTS' && ' (Points)'}
            </Label>
            <Input
              id="prize_value"
              type="number"
              min="1"
              step={formData.prize_type === 'CASH' || formData.prize_type === 'AIRTIME' ? '0.01' : '1'}
              value={formData.prize_value}
              onChange={(e) => handleInputChange('prize_value', parseFloat(e.target.value) || 0)}
              className={errors.prize_value ? 'border-red-500' : ''}
            />
            {errors.prize_value && (
              <p className="text-red-500 text-sm mt-1">{errors.prize_value}</p>
            )}
            {formData.prize_value > 0 && (
              <p className="text-sm text-gray-600 mt-1">
                Display: {formatPrizeValue(formData.prize_type, formData.prize_value)}
              </p>
            )}
          </div>

          {/* Probability and Minimum Recharge */}
          <div className="grid grid-cols-2 gap-4">
            <div>
              <Label htmlFor="probability">Probability (%)</Label>
              <Input
                id="probability"
                type="number"
                min="0.1"
                max="100"
                step="0.1"
                value={formData.probability}
                onChange={(e) => handleInputChange('probability', parseFloat(e.target.value) || 0)}
                className={errors.probability ? 'border-red-500' : ''}
              />
              {errors.probability && (
                <p className="text-red-500 text-sm mt-1">{errors.probability}</p>
              )}
            </div>
            <div>
              <Label htmlFor="minimum_recharge">Min Recharge (₦)</Label>
              <Input
                id="minimum_recharge"
                type="number"
                min="0"
                value={formData.minimum_recharge}
                onChange={(e) => handleInputChange('minimum_recharge', parseInt(e.target.value) || 0)}
                className={errors.minimum_recharge ? 'border-red-500' : ''}
              />
              {errors.minimum_recharge && (
                <p className="text-red-500 text-sm mt-1">{errors.minimum_recharge}</p>
              )}
            </div>
          </div>

          {/* Probability Warning */}
          {probabilityWarning && (
            <Alert className="border-yellow-200 bg-yellow-50">
              <AlertTriangle className="h-4 w-4 text-yellow-600" />
              <AlertDescription className="text-yellow-800">
                {probabilityWarning}
              </AlertDescription>
            </Alert>
          )}

          {/* Color Scheme */}
          <div>
            <Label htmlFor="color_scheme">Color Scheme</Label>
            <Select 
              value={formData.color_scheme} 
              onValueChange={(value) => handleInputChange('color_scheme', value)}
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {COLOR_SCHEMES.map((color) => (
                  <SelectItem key={color.value} value={color.value}>
                    <div className="flex items-center gap-2">
                      <div className={`w-4 h-4 rounded-full ${color.bg}`}></div>
                      {color.label}
                    </div>
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {/* Sort Order */}
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
            <p className="text-xs text-gray-500 mt-1">
              Lower numbers appear first on the wheel
            </p>
          </div>

          {/* Prize Status */}
          <div className="flex items-center justify-between p-3 border rounded-lg">
            <div>
              <div className="font-medium">Prize Active</div>
              <div className="text-sm text-gray-500">Include this prize in the wheel</div>
            </div>
            <Switch
              checked={formData.is_active}
              onCheckedChange={(checked) => handleInputChange('is_active', checked)}
            />
          </div>

          {/* Preview */}
          {formData.prize_name && formData.prize_value > 0 && (
            <div className="p-3 bg-gray-50 rounded-lg">
              <Label className="text-sm font-medium text-gray-700">Preview</Label>
              <div className="mt-2 p-3 bg-white border rounded-lg">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    {getSelectedPrizeType() && (() => {
                      const PrizeIcon = getSelectedPrizeType()!.icon;
                      return (
                        <div className={`p-2 rounded-full ${getSelectedColorScheme()?.bg} bg-opacity-20`}>
                          <PrizeIcon className={`w-4 h-4 ${getSelectedPrizeType()!.color}`} />
                        </div>
                      );
                    })()}
                    <div>
                      <div className="font-medium">{formData.prize_name}</div>
                      <div className="text-sm text-gray-500">
                        {formatPrizeValue(formData.prize_type, formData.prize_value)}
                      </div>
                    </div>
                  </div>
                  <div className="text-right">
                    <Badge variant="outline">
                      {formData.probability}% chance
                    </Badge>
                    {(formData.minimum_recharge ?? 0) > 0 && (
                      <div className="text-xs text-gray-500 mt-1">
                        Min: ₦{(formData.minimum_recharge ?? 0).toLocaleString()}
                      </div>
                    )}
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* Action Buttons */}
          <div className="flex gap-2 pt-4">
            <Button
              type="submit"
              disabled={loading}
              className="flex-1"
            >
              {loading ? (
                <Loader2 className="w-4 h-4 animate-spin mr-2" />
              ) : null}
              {prize ? 'Update Prize' : 'Create Prize'}
            </Button>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={loading}
            >
              Cancel
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
};