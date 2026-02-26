import React, { useState, useEffect } from 'react';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Switch } from '@/components/ui/switch';
import { Badge } from '@/components/ui/badge';
import { Loader2, Smartphone } from 'lucide-react';
import type { NetworkConfig } from '@/types/admin-api.types';

interface DataPlan {
  id?: string;
  network_id: string;
  plan_name: string;
  data_amount: string;
  price: number;
  validity_days: number;
  plan_code: string;
  is_active: boolean;
  sort_order: number;
}

interface DataPlanDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  dataPlan?: DataPlan | null | undefined;
  networks: NetworkConfig[];
  onSave: (planData: Omit<DataPlan, 'id'>) => Promise<void>;
  loading?: boolean;
}

export const DataPlanDialog: React.FC<DataPlanDialogProps> = ({
  open,
  onOpenChange,
  dataPlan,
  networks,
  onSave,
  loading = false
}) => {
  const [formData, setFormData] = useState<Omit<DataPlan, 'id'>>({
    network_id: '',
    plan_name: '',
    data_amount: '',
    price: 0,
    validity_days: 30,
    plan_code: '',
    is_active: true,
    sort_order: 1
  });

  const [errors, setErrors] = useState<Record<string, string>>({});

  useEffect(() => {
    if (dataPlan) {
      setFormData({
        network_id: dataPlan.network_id,
        plan_name: dataPlan.plan_name,
        data_amount: dataPlan.data_amount,
        price: dataPlan.price,
        validity_days: dataPlan.validity_days,
        plan_code: dataPlan.plan_code,
        is_active: dataPlan.is_active,
        sort_order: dataPlan.sort_order
      });
    } else {
      setFormData({
        network_id: '',
        plan_name: '',
        data_amount: '',
        price: 0,
        validity_days: 30,
        plan_code: '',
        is_active: true,
        sort_order: 1
      });
    }
    setErrors({});
  }, [dataPlan, open]);

  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.network_id) {
      newErrors.network_id = 'Please select a network';
    }

    if (!formData.plan_name.trim()) {
      newErrors.plan_name = 'Plan name is required';
    }

    if (!formData.data_amount.trim()) {
      newErrors.data_amount = 'Data amount is required';
    }

    if (!formData.plan_code.trim()) {
      newErrors.plan_code = 'Plan code is required';
    }

    if (formData.price <= 0) {
      newErrors.price = 'Price must be greater than 0';
    }

    if (formData.validity_days <= 0) {
      newErrors.validity_days = 'Validity days must be greater than 0';
    }

    if (formData.sort_order < 1) {
      newErrors.sort_order = 'Sort order must be at least 1';
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
      console.error('Failed to save data plan:', error);
    }
  };

  const handleInputChange = (field: keyof typeof formData, value: any) => {
    setFormData(prev => ({ ...prev, [field]: value }));
    
    // Clear error for this field when user starts typing
    if (errors[field]) {
      setErrors(prev => ({ ...prev, [field]: '' }));
    }
  };

  const getSelectedNetwork = () => {
    return networks.find(n => n.id === formData.network_id);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Smartphone className="w-5 h-5" />
            {dataPlan ? 'Edit Data Plan' : 'Add New Data Plan'}
          </DialogTitle>
          <DialogDescription>
            Configure data bundle plans for network providers
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          {/* Network Selection */}
          <div>
            <Label htmlFor="network_id">Network Provider</Label>
            <Select 
              value={formData.network_id} 
              onValueChange={(value) => handleInputChange('network_id', value)}
            >
              <SelectTrigger className={errors.network_id ? 'border-red-500' : ''}>
                <SelectValue placeholder="Select network provider" />
              </SelectTrigger>
              <SelectContent>
                {networks.filter(n => n.is_active).map((network) => (
                  <SelectItem key={network.id} value={network.id}>
                    <div className="flex items-center gap-2">
                      <Badge variant="outline">{network.network_code}</Badge>
                      {network.network_name}
                    </div>
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            {errors.network_id && (
              <p className="text-red-500 text-sm mt-1">{errors.network_id}</p>
            )}
          </div>

          {/* Plan Name */}
          <div>
            <Label htmlFor="plan_name">Plan Name</Label>
            <Input
              id="plan_name"
              value={formData.plan_name}
              onChange={(e) => handleInputChange('plan_name', e.target.value)}
              placeholder="e.g., 1GB Monthly Plan"
              className={errors.plan_name ? 'border-red-500' : ''}
            />
            {errors.plan_name && (
              <p className="text-red-500 text-sm mt-1">{errors.plan_name}</p>
            )}
          </div>

          {/* Data Amount */}
          <div>
            <Label htmlFor="data_amount">Data Amount</Label>
            <Input
              id="data_amount"
              value={formData.data_amount}
              onChange={(e) => handleInputChange('data_amount', e.target.value)}
              placeholder="e.g., 1GB, 500MB, 2.5GB"
              className={errors.data_amount ? 'border-red-500' : ''}
            />
            {errors.data_amount && (
              <p className="text-red-500 text-sm mt-1">{errors.data_amount}</p>
            )}
            <p className="text-xs text-gray-500 mt-1">
              Include unit (MB, GB) for clarity
            </p>
          </div>

          {/* Plan Code */}
          <div>
            <Label htmlFor="plan_code">Plan Code</Label>
            <Input
              id="plan_code"
              value={formData.plan_code}
              onChange={(e) => handleInputChange('plan_code', e.target.value)}
              placeholder="e.g., MTN_1GB_30D"
              className={errors.plan_code ? 'border-red-500' : ''}
            />
            {errors.plan_code && (
              <p className="text-red-500 text-sm mt-1">{errors.plan_code}</p>
            )}
            <p className="text-xs text-gray-500 mt-1">
              Unique identifier for API integration
            </p>
          </div>

          {/* Price and Validity */}
          <div className="grid grid-cols-2 gap-4">
            <div>
              <Label htmlFor="price">Price (₦)</Label>
              <Input
                id="price"
                type="number"
                min="1"
                step="0.01"
                value={formData.price}
                onChange={(e) => handleInputChange('price', parseFloat(e.target.value) || 0)}
                className={errors.price ? 'border-red-500' : ''}
              />
              {errors.price && (
                <p className="text-red-500 text-sm mt-1">{errors.price}</p>
              )}
            </div>
            <div>
              <Label htmlFor="validity_days">Validity (Days)</Label>
              <Input
                id="validity_days"
                type="number"
                min="1"
                value={formData.validity_days}
                onChange={(e) => handleInputChange('validity_days', parseInt(e.target.value) || 0)}
                className={errors.validity_days ? 'border-red-500' : ''}
              />
              {errors.validity_days && (
                <p className="text-red-500 text-sm mt-1">{errors.validity_days}</p>
              )}
            </div>
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
              Lower numbers appear first in the list
            </p>
          </div>

          {/* Plan Status */}
          <div className="flex items-center justify-between p-3 border rounded-lg">
            <div>
              <div className="font-medium">Plan Active</div>
              <div className="text-sm text-gray-500">Make this plan available for purchase</div>
            </div>
            <Switch
              checked={formData.is_active}
              onCheckedChange={(checked) => handleInputChange('is_active', checked)}
            />
          </div>

          {/* Preview */}
          {formData.plan_name && formData.data_amount && formData.price > 0 && (
            <div className="p-3 bg-gray-50 rounded-lg">
              <Label className="text-sm font-medium text-gray-700">Preview</Label>
              <div className="mt-2 p-3 bg-white border rounded-lg">
                <div className="flex justify-between items-center">
                  <div>
                    <div className="font-medium">{formData.plan_name}</div>
                    <div className="text-sm text-gray-500">{formData.data_amount}</div>
                    <div className="text-xs text-gray-400">Valid for {formData.validity_days} days</div>
                  </div>
                  <div className="text-right">
                    <div className="font-bold text-lg">₦{formData.price.toLocaleString()}</div>
                    {getSelectedNetwork() && (
                      <Badge variant="outline" className="text-xs">
                        {getSelectedNetwork()?.network_code}
                      </Badge>
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
              {dataPlan ? 'Update Plan' : 'Create Plan'}
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