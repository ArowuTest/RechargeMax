import React, { useState, useEffect, useCallback } from 'react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Switch } from '@/components/ui/switch';
import { Label } from '@/components/ui/label';
import {
  Wifi,
  Plus,
  Pencil,
  Trash2,
  Loader2,
  AlertCircle,
  CheckCircle2,
  ToggleLeft,
  ToggleRight,
  Phone,
  Database,
} from 'lucide-react';
import { NetworkDialog } from './NetworkDialog';
import apiClient from '@/lib/api-client';

// ─────────────────────────────────────────────────────────────────────────────
// Types
// ─────────────────────────────────────────────────────────────────────────────

interface NetworkConfig {
  id: string;
  network_name: string;
  network_code: string;
  is_active: boolean;
  airtime_enabled: boolean;
  data_enabled: boolean;
  commission_rate: number;
  minimum_amount: number;   // kobo
  maximum_amount: number;   // kobo
  logo_url?: string;
  brand_color?: string;
  sort_order?: number;
}

// ─────────────────────────────────────────────────────────────────────────────
// Network logos (colour + SVG text fallback)
// ─────────────────────────────────────────────────────────────────────────────

const NETWORK_BRAND: Record<string, { color: string; initials: string }> = {
  MTN:     { color: '#FFD700', initials: 'MTN'    },
  GLO:     { color: '#00A651', initials: 'GLO'    },
  AIRTEL:  { color: '#E40000', initials: 'ATL'    },
  '9MOBILE': { color: '#006E51', initials: '9MB'  },
};

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

const naira = (kobo: number) =>
  kobo > 0 ? `₦${(kobo / 100).toLocaleString()}` : '—';

// ─────────────────────────────────────────────────────────────────────────────
// Component
// ─────────────────────────────────────────────────────────────────────────────

export const NetworkManagementPage: React.FC = () => {
  const [networks, setNetworks]       = useState<NetworkConfig[]>([]);
  const [loading, setLoading]         = useState(true);
  const [saving, setSaving]           = useState<string | null>(null);
  const [error, setError]             = useState<string | null>(null);
  const [success, setSuccess]         = useState<string | null>(null);

  // Dialog state
  const [dialogOpen, setDialogOpen]       = useState(false);
  const [dialogLoading, setDialogLoading] = useState(false);
  const [editingNetwork, setEditingNetwork] = useState<NetworkConfig | null>(null);

  // ── Fetch ─────────────────────────────────────────────────────────────────

  const fetchNetworks = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await apiClient.get('/admin/networks');
      const raw: any[] = res.data?.data ?? res.data ?? [];
      const mapped: NetworkConfig[] = raw.map((n: any) => ({
        id:              n.id ?? n.network_code ?? n.NetworkCode,
        network_name:    n.network_name ?? n.NetworkName ?? '',
        network_code:    n.network_code ?? n.NetworkCode ?? '',
        is_active:       n.is_active ?? true,
        airtime_enabled: n.airtime_enabled ?? true,
        data_enabled:    n.data_enabled    ?? true,
        commission_rate: n.commission_rate ?? 0,
        minimum_amount:  n.minimum_amount  ?? 0,
        maximum_amount:  n.maximum_amount  ?? 0,
        logo_url:        n.logo_url        ?? '',
        brand_color:     n.brand_color     ?? '',
        sort_order:      n.sort_order      ?? 0,
      }));
      setNetworks(mapped);
    } catch (err: any) {
      setError(err?.response?.data?.error ?? err?.message ?? 'Failed to load networks');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { fetchNetworks(); }, [fetchNetworks]);

  // ── Quick toggles ─────────────────────────────────────────────────────────

  const toggle = async (
    network: NetworkConfig,
    field: 'is_active' | 'airtime_enabled' | 'data_enabled',
  ) => {
    setSaving(network.id + ':' + field);
    setError(null);
    try {
      await apiClient.put(`/admin/networks/${network.id}`, {
        [field]: !network[field],
      });
      setNetworks((prev) =>
        prev.map((n) => (n.id === network.id ? { ...n, [field]: !n[field] } : n)),
      );
      const labels: Record<string, string> = {
        is_active:       network[field] ? 'disabled' : 'enabled',
        airtime_enabled: network[field] ? 'airtime disabled' : 'airtime enabled',
        data_enabled:    network[field] ? 'data disabled'    : 'data enabled',
      };
      showSuccess(`${network.network_name} ${labels[field]}`);
    } catch (err: any) {
      setError(err?.response?.data?.error ?? err?.message ?? 'Update failed');
    } finally {
      setSaving(null);
    }
  };

  // ── Create / Update ───────────────────────────────────────────────────────

  const handleSave = async (data: Omit<NetworkConfig, 'id'>) => {
    setDialogLoading(true);
    setError(null);
    try {
      if (editingNetwork) {
        await apiClient.put(`/admin/networks/${editingNetwork.id}`, data);
        showSuccess(`${data.network_name} updated`);
      } else {
        await apiClient.post('/admin/networks', data);
        showSuccess(`${data.network_name} created`);
      }
      setDialogOpen(false);
      setEditingNetwork(null);
      await fetchNetworks();
    } catch (err: any) {
      setError(err?.response?.data?.error ?? err?.message ?? 'Save failed');
    } finally {
      setDialogLoading(false);
    }
  };

  // ── Delete ────────────────────────────────────────────────────────────────

  const handleDelete = async (network: NetworkConfig) => {
    if (!window.confirm(`Delete "${network.network_name}"? This cannot be undone.`)) return;
    setSaving(network.id + ':delete');
    setError(null);
    try {
      await apiClient.delete(`/admin/networks/${network.id}`);
      setNetworks((prev) => prev.filter((n) => n.id !== network.id));
      showSuccess(`${network.network_name} deleted`);
    } catch (err: any) {
      setError(err?.response?.data?.error ?? err?.message ?? 'Delete failed');
    } finally {
      setSaving(null);
    }
  };

  // ── Helpers ───────────────────────────────────────────────────────────────

  const showSuccess = (msg: string) => {
    setSuccess(msg);
    setTimeout(() => setSuccess(null), 4000);
  };

  const openCreate = () => {
    setEditingNetwork(null);
    setDialogOpen(true);
  };

  const openEdit = (n: NetworkConfig) => {
    setEditingNetwork(n);
    setDialogOpen(true);
  };

  // ─────────────────────────────────────────────────────────────────────────
  // Render
  // ─────────────────────────────────────────────────────────────────────────

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold flex items-center gap-2">
            <Wifi className="w-7 h-7 text-primary" />
            Network Management
          </h1>
          <p className="text-sm text-gray-500 mt-1">
            Control which networks and product types are available to users.
            Disabled networks are hidden from the recharge form immediately.
          </p>
        </div>
        <Button onClick={openCreate} className="flex items-center gap-2">
          <Plus className="w-4 h-4" />
          Add Network
        </Button>
      </div>

      {/* Feedback */}
      {error && (
        <Alert variant="destructive">
          <AlertCircle className="w-4 h-4" />
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}
      {success && (
        <Alert className="bg-green-50 border-green-200">
          <CheckCircle2 className="w-4 h-4 text-green-600" />
          <AlertDescription className="text-green-800">{success}</AlertDescription>
        </Alert>
      )}

      {/* Stats bar */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
        {[
          { label: 'Total Networks',    value: networks.length },
          { label: 'Active',            value: networks.filter((n) => n.is_active).length, color: 'text-green-600' },
          { label: 'Airtime Enabled',   value: networks.filter((n) => n.airtime_enabled).length, color: 'text-blue-600' },
          { label: 'Data Enabled',      value: networks.filter((n) => n.data_enabled).length, color: 'text-purple-600' },
        ].map((s) => (
          <Card key={s.label}>
            <CardContent className="pt-4 pb-3">
              <div className={`text-2xl font-bold ${s.color ?? ''}`}>{s.value}</div>
              <div className="text-xs text-gray-500">{s.label}</div>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Network Cards */}
      {loading ? (
        <div className="flex justify-center py-16">
          <Loader2 className="w-8 h-8 animate-spin text-primary" />
        </div>
      ) : networks.length === 0 ? (
        <Card>
          <CardContent className="py-16 text-center text-gray-500">
            <Wifi className="w-12 h-12 mx-auto mb-3 text-gray-300" />
            <p className="font-medium">No networks configured yet.</p>
            <p className="text-sm mt-1">Click "Add Network" to add your first network.</p>
          </CardContent>
        </Card>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-2 gap-4">
          {networks
            .slice()
            .sort((a, b) => (a.sort_order ?? 99) - (b.sort_order ?? 99))
            .map((network) => {
              const brand = NETWORK_BRAND[network.network_code.toUpperCase()] ?? {
                color: '#6b7280',
                initials: network.network_code.slice(0, 3).toUpperCase(),
              };

              return (
                <Card
                  key={network.id}
                  className={`border-2 transition-all ${
                    network.is_active ? 'border-gray-200' : 'border-gray-100 opacity-60'
                  }`}
                >
                  <CardHeader className="pb-2">
                    <div className="flex items-center justify-between">
                      {/* Brand avatar */}
                      <div className="flex items-center gap-3">
                        <div
                          className="w-12 h-12 rounded-full flex items-center justify-center font-bold text-white text-sm shadow"
                          style={{ backgroundColor: brand.color }}
                        >
                          {brand.initials}
                        </div>
                        <div>
                          <CardTitle className="text-base">{network.network_name}</CardTitle>
                          <CardDescription className="text-xs font-mono">
                            {network.network_code}
                          </CardDescription>
                        </div>
                      </div>

                      {/* Status badge */}
                      <Badge
                        className={
                          network.is_active
                            ? 'bg-green-100 text-green-800 border-green-200'
                            : 'bg-gray-100 text-gray-600 border-gray-200'
                        }
                        variant="outline"
                      >
                        {network.is_active ? 'Active' : 'Disabled'}
                      </Badge>
                    </div>
                  </CardHeader>

                  <CardContent className="space-y-4">
                    {/* Amount limits */}
                    {(network.minimum_amount > 0 || network.maximum_amount > 0) && (
                      <div className="flex gap-4 text-xs text-gray-500">
                        {network.minimum_amount > 0 && (
                          <span>Min: <strong>{naira(network.minimum_amount)}</strong></span>
                        )}
                        {network.maximum_amount > 0 && (
                          <span>Max: <strong>{naira(network.maximum_amount)}</strong></span>
                        )}
                        {network.commission_rate > 0 && (
                          <span>Commission: <strong>{network.commission_rate}%</strong></span>
                        )}
                      </div>
                    )}

                    {/* Product toggles */}
                    <div className="space-y-3 border-t pt-3">
                      <p className="text-xs font-semibold text-gray-500 uppercase tracking-wide">
                        Product Availability
                      </p>

                      {/* Master: Network on/off */}
                      <div className="flex items-center justify-between">
                        <Label
                          htmlFor={`active-${network.id}`}
                          className="flex items-center gap-2 cursor-pointer"
                        >
                          {network.is_active ? (
                            <ToggleRight className="w-4 h-4 text-green-500" />
                          ) : (
                            <ToggleLeft className="w-4 h-4 text-gray-400" />
                          )}
                          <span className="text-sm font-medium">Network visible to users</span>
                        </Label>
                        <Switch
                          id={`active-${network.id}`}
                          checked={network.is_active}
                          disabled={saving === network.id + ':is_active'}
                          onCheckedChange={() => toggle(network, 'is_active')}
                        />
                      </div>

                      {/* Airtime */}
                      <div className="flex items-center justify-between pl-2">
                        <Label
                          htmlFor={`airtime-${network.id}`}
                          className="flex items-center gap-2 cursor-pointer"
                        >
                          <Phone className="w-4 h-4 text-blue-500" />
                          <span className="text-sm">Airtime recharge</span>
                        </Label>
                        <Switch
                          id={`airtime-${network.id}`}
                          checked={network.airtime_enabled}
                          disabled={!network.is_active || saving === network.id + ':airtime_enabled'}
                          onCheckedChange={() => toggle(network, 'airtime_enabled')}
                        />
                      </div>

                      {/* Data */}
                      <div className="flex items-center justify-between pl-2">
                        <Label
                          htmlFor={`data-${network.id}`}
                          className="flex items-center gap-2 cursor-pointer"
                        >
                          <Database className="w-4 h-4 text-purple-500" />
                          <span className="text-sm">Data bundle recharge</span>
                        </Label>
                        <Switch
                          id={`data-${network.id}`}
                          checked={network.data_enabled}
                          disabled={!network.is_active || saving === network.id + ':data_enabled'}
                          onCheckedChange={() => toggle(network, 'data_enabled')}
                        />
                      </div>
                    </div>

                    {/* Action buttons */}
                    <div className="flex gap-2 border-t pt-3">
                      <Button
                        size="sm"
                        variant="outline"
                        className="flex-1"
                        onClick={() => openEdit(network)}
                      >
                        <Pencil className="w-3 h-3 mr-1" />
                        Edit
                      </Button>
                      <Button
                        size="sm"
                        variant="destructive"
                        className="flex-1"
                        disabled={saving === network.id + ':delete'}
                        onClick={() => handleDelete(network)}
                      >
                        {saving === network.id + ':delete' ? (
                          <Loader2 className="w-3 h-3 mr-1 animate-spin" />
                        ) : (
                          <Trash2 className="w-3 h-3 mr-1" />
                        )}
                        Delete
                      </Button>
                    </div>
                  </CardContent>
                </Card>
              );
            })}
        </div>
      )}

      {/* Future product types section */}
      <Card className="border-dashed border-2 border-gray-200">
        <CardContent className="py-8 text-center text-gray-400">
          <div className="flex justify-center gap-6 text-3xl mb-3">
            📺 ⛽ 💡 🌊
          </div>
          <p className="font-medium text-gray-500">More product types coming soon</p>
          <p className="text-sm mt-1">
            Cable TV, Gas, Electricity, and Water recharges will appear here once enabled.
          </p>
        </CardContent>
      </Card>

      {/* Create / Edit Dialog */}
      <NetworkDialog
        open={dialogOpen}
        onOpenChange={(open) => {
          setDialogOpen(open);
          if (!open) setEditingNetwork(null);
        }}
        network={editingNetwork}
        onSave={handleSave}
        loading={dialogLoading}
      />
    </div>
  );
};

export default NetworkManagementPage;
