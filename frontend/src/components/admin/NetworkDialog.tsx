import React, { useState, useEffect } from 'react';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import { Loader2, Network } from 'lucide-react';

interface NetworkConfig {
  id?: string;
  network_name: string;
  network_code: string;
  is_active: boolean;
  airtime_enabled: boolean;
  data_enabled: boolean;
  commission_rate: number;
  minimum_amount: number;
  maximum_amount: number;
}

interface NetworkDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  network?: NetworkConfig | null;
  onSave: (networkData: Omit<NetworkConfig, 'id'>) => Promise<void>;
  loading?: boolean;
}

export const NetworkDialog: React.FC<NetworkDialogProps> = ({
  open,
  onOpenChange,
  network,
  onSave,
  loading = false
}) => {
  const [formData, setFormData] = useState<Omit<NetworkConfig, 'id'>>({
    network_name: '',
    network_code: '',
    is_active: true,
    airtime_enabled: true,
    data_enabled: true,
    commission_rate: 2.5,
    minimum_amount: 50,
    maximum_amount: 50000
  });

  const [errors, setErrors] = useState<Record<string, string>>({});

  useEffect(() => {
    if (network) {
      setFormData({
        network_name: network.network_name,
        network_code: network.network_code,
        is_active: network.is_active,
        airtime_enabled: network.airtime_enabled,
        data_enabled: network.data_enabled,
        commission_rate: network.commission_rate,
        minimum_amount: network.minimum_amount,
        maximum_amount: network.maximum_amount
      });
    } else {
      setFormData({
        network_name: '',
        network_code: '',
        is_active: true,
        airtime_enabled: true,
        data_enabled: true,
        commission_rate: 2.5,
        minimum_amount: 50,
        maximum_amount: 50000
      });
    }
    setErrors({});
  }, [network, open]);

  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.network_name.trim()) {
      newErrors.network_name = 'Network name is required';
    }

    if (!formData.network_code.trim()) {
      newErrors.network_code = 'Network code is required';
    } else if (!/^[A-Z_]+$/.test(formData.network_code)) {
      newErrors.network_code = 'Network code must contain only uppercase letters and underscores';
    }

    if (formData.commission_rate < 0 || formData.commission_rate > 100) {
      newErrors.commission_rate = 'Commission rate must be between 0 and 100';
    }

    if (formData.minimum_amount < 1) {
      newErrors.minimum_amount = 'Minimum amount must be at least ₦1';
    }

    if (formData.maximum_amount < formData.minimum_amount) {
      newErrors.maximum_amount = 'Maximum amount must be greater than minimum amount';
    }

    if (!formData.airtime_enabled && !formData.data_enabled) {
      newErrors.services = 'At least one service (Airtime or Data) must be enabled';
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
      console.error('Failed to save network:', error);
    }
  };

  const handleInputChange = (field: keyof typeof formData, value: any) => {
    setFormData(prev => ({ ...prev, [field]: value }));
    
    // Clear error for this field when user starts typing
    if (errors[field]) {
      setErrors(prev => ({ ...prev, [field]: '' }));
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Network className="w-5 h-5" />
            {network ? 'Edit Network' : 'Add New Network'}
          </DialogTitle>
          <DialogDescription>
            Configure network provider settings and commission rates
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          {/* Network Name */}
          <div>
            <Label htmlFor="network_name">Network Name</Label>
            <Input
              id="network_name"
              value={formData.network_name}
              onChange={(e) => handleInputChange('network_name', e.target.value)}
              placeholder="e.g., MTN Nigeria"
              className={errors.network_name ? 'border-red-500' : ''}
            />
            {errors.network_name && (
              <p className="text-red-500 text-sm mt-1">{errors.network_name}</p>
            )}
          </div>

          {/* Network Code */}
          <div>
            <Label htmlFor="network_code">Network Code</Label>
            <Input
              id="network_code"
              value={formData.network_code}
              onChange={(e) => handleInputChange('network_code', e.target.value.toUpperCase())}
              placeholder="e.g., MTN"
              className={errors.network_code ? 'border-red-500' : ''}
            />
            {errors.network_code && (
              <p className="text-red-500 text-sm mt-1">{errors.network_code}</p>
            )}
            <p className="text-xs text-gray-500 mt-1">
              Use uppercase letters and underscores only
            </p>
          </div>

          {/* Commission Rate */}
          <div>
            <Label htmlFor="commission_rate">Commission Rate (%)</Label>
            <Input
              id="commission_rate"
              type="number"
              min="0"
              max="100"
              step="0.1"
              value={formData.commission_rate}
              onChange={(e) => handleInputChange('commission_rate', parseFloat(e.target.value) || 0)}
              className={errors.commission_rate ? 'border-red-500' : ''}
            />
            {errors.commission_rate && (
              <p className="text-red-500 text-sm mt-1">{errors.commission_rate}</p>
            )}
          </div>

          {/* Amount Limits */}
          <div className="grid grid-cols-2 gap-4">
            <div>
              <Label htmlFor="minimum_amount">Minimum Amount (₦)</Label>
              <Input
                id="minimum_amount"
                type="number"
                min="1"
                value={formData.minimum_amount}
                onChange={(e) => handleInputChange('minimum_amount', parseInt(e.target.value) || 0)}
                className={errors.minimum_amount ? 'border-red-500' : ''}
              />
              {errors.minimum_amount && (
                <p className="text-red-500 text-sm mt-1">{errors.minimum_amount}</p>
              )}
            </div>
            <div>
              <Label htmlFor="maximum_amount">Maximum Amount (₦)</Label>
              <Input
                id="maximum_amount"
                type="number"
                min="1"
                value={formData.maximum_amount}
                onChange={(e) => handleInputChange('maximum_amount', parseInt(e.target.value) || 0)}
                className={errors.maximum_amount ? 'border-red-500' : ''}
              />
              {errors.maximum_amount && (
                <p className="text-red-500 text-sm mt-1">{errors.maximum_amount}</p>
              )}
            </div>
          </div>

          {/* Service Toggles */}
          <div className="space-y-3">
            <Label>Available Services</Label>
            
            <div className="flex items-center justify-between p-3 border rounded-lg">
              <div>
                <div className="font-medium">Airtime Recharge</div>
                <div className="text-sm text-gray-500">Enable airtime top-up services</div>
              </div>
              <Switch
                checked={formData.airtime_enabled}
                onCheckedChange={(checked) => handleInputChange('airtime_enabled', checked)}
              />
            </div>

            <div className="flex items-center justify-between p-3 border rounded-lg">
              <div>
                <div className="font-medium">Data Bundle</div>
                <div className="text-sm text-gray-500">Enable data bundle purchases</div>
              </div>
              <Switch
                checked={formData.data_enabled}
                onCheckedChange={(checked) => handleInputChange('data_enabled', checked)}
              />
            </div>

            {errors.services && (
              <p className="text-red-500 text-sm">{errors.services}</p>
            )}
          </div>

          {/* Network Status */}
          <div className="flex items-center justify-between p-3 border rounded-lg">
            <div>
              <div className="font-medium">Network Active</div>
              <div className="text-sm text-gray-500">Enable this network for transactions</div>
            </div>
            <Switch
              checked={formData.is_active}
              onCheckedChange={(checked) => handleInputChange('is_active', checked)}
            />
          </div>

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
              {network ? 'Update Network' : 'Create Network'}
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