import React, { useState, useEffect } from 'react';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Switch } from '@/components/ui/switch';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { CheckCircle2, AlertCircle, Settings, Save } from 'lucide-react';

const API_BASE = '/api/v1';


interface PrizeFulfillmentConfig {
  id: string;
  prizeType: string;
  fulfillmentMode: 'auto_provision' | 'manual_claim';
  autoProvisionEnabled: boolean;
  requireLoginToClaim: boolean;
  claimDeadlineDays: number;
  allowRetryOnFailure: boolean;
  maxRetryAttempts: number;
  isActive: boolean;
}

const PrizeFulfillmentConfigPanel: React.FC = () => {
  const [configs, setConfigs] = useState<PrizeFulfillmentConfig[]>([]);
  const [selectedConfig, setSelectedConfig] = useState<PrizeFulfillmentConfig | null>(null);
  const [loading, setLoading] = useState(false);
  const [saveSuccess, setSaveSuccess] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const prizeTypes = [
    { value: 'airtime', label: 'Airtime', icon: '📱' },
    { value: 'data', label: 'Data', icon: '📶' },
    { value: 'cash', label: 'Cash', icon: '💰' },
    { value: 'goods', label: 'Goods/Physical', icon: '🎁' },
    { value: 'points', label: 'Loyalty Points', icon: '⭐' }
  ];

  useEffect(() => {
    fetchConfigs();
  }, []);

  const fetchConfigs = async () => {
    setLoading(true);
    try {
      const response = await fetch(`${API_BASE}/admin/spin/claims`,  { credentials: 'include' });
      
      if (!response.ok) throw new Error('Failed to fetch configurations');
      
      const data = await response.json();
      setConfigs(data.configs || []);
      
      // Select first config by default
      if (data.configs && data.configs.length > 0) {
        setSelectedConfig(data.configs[0]);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load configurations');
    } finally {
      setLoading(false);
    }
  };

  const handleSaveConfig = async () => {
    if (!selectedConfig) return;
    
    setLoading(true);
    setSaveSuccess(false);
    setError(null);
    
    try {
      const response = await fetch(`${API_BASE}/admin/spin/claims/${selectedConfig.id}`, {
        method: 'PUT',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(selectedConfig)
      });
      
      if (!response.ok) throw new Error('Failed to save configuration');
      
      setSaveSuccess(true);
      setTimeout(() => setSaveSuccess(false), 3000);
      
      // Refresh configs
      await fetchConfigs();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save configuration');
    } finally {
      setLoading(false);
    }
  };

  const updateSelectedConfig = (updates: Partial<PrizeFulfillmentConfig>) => {
    if (!selectedConfig) return;
    setSelectedConfig({ ...selectedConfig, ...updates });
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-3xl font-bold flex items-center gap-2">
            <Settings className="w-8 h-8" />
            Prize Fulfillment Configuration
          </h2>
          <p className="text-gray-600 mt-1">
            Configure how prizes are delivered to winners
          </p>
        </div>
      </div>

      {/* Success/Error Alerts */}
      {saveSuccess && (
        <Alert className="bg-green-50 border-green-200">
          <CheckCircle2 className="w-4 h-4 text-green-600" />
          <AlertDescription className="text-green-800">
            Configuration saved successfully!
          </AlertDescription>
        </Alert>
      )}
      
      {error && (
        <Alert className="bg-red-50 border-red-200">
          <AlertCircle className="w-4 h-4 text-red-600" />
          <AlertDescription className="text-red-800">{error}</AlertDescription>
        </Alert>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Prize Type Selector */}
        <Card>
          <CardHeader>
            <CardTitle>Prize Types</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              {prizeTypes.map((type) => {
                const config = configs.find(c => c.prizeType === type.value);
                const isSelected = selectedConfig?.prizeType === type.value;
                
                return (
                  <button
                    key={type.value}
                    onClick={() => config && setSelectedConfig(config)}
                    className={`w-full p-4 rounded-lg border-2 transition-all ${
                      isSelected
                        ? 'border-blue-500 bg-blue-50'
                        : 'border-gray-200 hover:border-gray-300'
                    }`}
                  >
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-3">
                        <span className="text-2xl">{type.icon}</span>
                        <div className="text-left">
                          <div className="font-semibold">{type.label}</div>
                          {config && (
                            <div className="text-xs text-gray-500">
                              {config.fulfillmentMode === 'auto_provision' 
                                ? 'Auto-Provision' 
                                : 'Manual Claim'}
                            </div>
                          )}
                        </div>
                      </div>
                      {config?.isActive && (
                        <CheckCircle2 className="w-5 h-5 text-green-500" />
                      )}
                    </div>
                  </button>
                );
              })}
            </div>
          </CardContent>
        </Card>

        {/* Configuration Panel */}
        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle>
              {selectedConfig && (
                <>
                  {prizeTypes.find(t => t.value === selectedConfig.prizeType)?.icon}{' '}
                  {prizeTypes.find(t => t.value === selectedConfig.prizeType)?.label} Configuration
                </>
              )}
            </CardTitle>
          </CardHeader>
          <CardContent>
            {selectedConfig ? (
              <div className="space-y-6">
                {/* Fulfillment Mode */}
                <div className="space-y-3">
                  <Label className="text-base font-semibold">Fulfillment Mode</Label>
                  <RadioGroup
                    value={selectedConfig.fulfillmentMode}
                    onValueChange={(value) => 
                      updateSelectedConfig({ 
                        fulfillmentMode: value as 'auto_provision' | 'manual_claim' 
                      })
                    }
                  >
                    <div className="flex items-start space-x-3 p-4 border rounded-lg">
                      <RadioGroupItem value="auto_provision" id="auto" />
                      <div className="flex-1">
                        <Label htmlFor="auto" className="font-semibold cursor-pointer">
                          Auto-Provision (Instant Delivery)
                        </Label>
                        <p className="text-sm text-gray-600 mt-1">
                          Prize is delivered immediately after winning. No user action required.
                          Best for instant gratification and small prizes.
                        </p>
                      </div>
                    </div>
                    
                    <div className="flex items-start space-x-3 p-4 border rounded-lg">
                      <RadioGroupItem value="manual_claim" id="manual" />
                      <div className="flex-1">
                        <Label htmlFor="manual" className="font-semibold cursor-pointer">
                          Manual Claim (User Must Login)
                        </Label>
                        <p className="text-sm text-gray-600 mt-1">
                          User must login and click "Claim" button. Drives engagement,
                          provides upsell opportunities, and verifies ownership.
                        </p>
                      </div>
                    </div>
                  </RadioGroup>
                </div>

                {/* Settings Grid */}
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  {/* Auto-Provision Enabled */}
                  <div className="flex items-center justify-between p-4 border rounded-lg">
                    <div>
                      <Label className="font-semibold">Auto-Provision Enabled</Label>
                      <p className="text-sm text-gray-600 mt-1">
                        Allow automatic provisioning via VTPass
                      </p>
                    </div>
                    <Switch
                      checked={selectedConfig.autoProvisionEnabled}
                      onCheckedChange={(checked) => 
                        updateSelectedConfig({ autoProvisionEnabled: checked })
                      }
                    />
                  </div>

                  {/* Require Login to Claim */}
                  <div className="flex items-center justify-between p-4 border rounded-lg">
                    <div>
                      <Label className="font-semibold">Require Login to Claim</Label>
                      <p className="text-sm text-gray-600 mt-1">
                        User must be logged in to claim prize
                      </p>
                    </div>
                    <Switch
                      checked={selectedConfig.requireLoginToClaim}
                      onCheckedChange={(checked) => 
                        updateSelectedConfig({ requireLoginToClaim: checked })
                      }
                    />
                  </div>

                  {/* Allow Retry on Failure */}
                  <div className="flex items-center justify-between p-4 border rounded-lg">
                    <div>
                      <Label className="font-semibold">Allow Retry on Failure</Label>
                      <p className="text-sm text-gray-600 mt-1">
                        Retry failed provisions automatically
                      </p>
                    </div>
                    <Switch
                      checked={selectedConfig.allowRetryOnFailure}
                      onCheckedChange={(checked) => 
                        updateSelectedConfig({ allowRetryOnFailure: checked })
                      }
                    />
                  </div>

                  {/* Is Active */}
                  <div className="flex items-center justify-between p-4 border rounded-lg">
                    <div>
                      <Label className="font-semibold">Configuration Active</Label>
                      <p className="text-sm text-gray-600 mt-1">
                        Enable/disable this configuration
                      </p>
                    </div>
                    <Switch
                      checked={selectedConfig.isActive}
                      onCheckedChange={(checked) => 
                        updateSelectedConfig({ isActive: checked })
                      }
                    />
                  </div>
                </div>

                {/* Numeric Settings */}
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  {/* Claim Deadline Days */}
                  <div className="space-y-2">
                    <Label htmlFor="deadline">Claim Deadline (Days)</Label>
                    <Input
                      id="deadline"
                      type="number"
                      min="1"
                      max="365"
                      value={selectedConfig.claimDeadlineDays}
                      onChange={(e) => 
                        updateSelectedConfig({ 
                          claimDeadlineDays: parseInt(e.target.value) || 30 
                        })
                      }
                    />
                    <p className="text-sm text-gray-600">
                      Number of days user has to claim prize
                    </p>
                  </div>

                  {/* Max Retry Attempts */}
                  <div className="space-y-2">
                    <Label htmlFor="retries">Max Retry Attempts</Label>
                    <Input
                      id="retries"
                      type="number"
                      min="0"
                      max="10"
                      value={selectedConfig.maxRetryAttempts}
                      onChange={(e) => 
                        updateSelectedConfig({ 
                          maxRetryAttempts: parseInt(e.target.value) || 3 
                        })
                      }
                      disabled={!selectedConfig.allowRetryOnFailure}
                    />
                    <p className="text-sm text-gray-600">
                      Maximum number of retry attempts for failed provisions
                    </p>
                  </div>
                </div>

                {/* Save Button */}
                <div className="flex justify-end pt-4 border-t">
                  <Button
                    onClick={handleSaveConfig}
                    disabled={loading}
                    className="flex items-center gap-2"
                  >
                    <Save className="w-4 h-4" />
                    {loading ? 'Saving...' : 'Save Configuration'}
                  </Button>
                </div>
              </div>
            ) : (
              <div className="text-center py-12 text-gray-500">
                Select a prize type to configure
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
};

export default PrizeFulfillmentConfigPanel;
