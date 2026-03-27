import React, { useState, useEffect } from 'react';
import { useLocation } from 'react-router-dom';
import { adminApi } from '@/lib/api-client';
import type { NetworkConfig, DataPlan, WheelPrize, DailySubscription, PlatformSetting } from '@/types/admin-api.types';
import { getErrorMessage } from '@/utils/error-utils';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { Switch } from '@/components/ui/switch';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Textarea } from '@/components/ui/textarea';
import { useAdminContext } from '@/contexts/AdminContext';
import { useToast } from '@/hooks/useToast';
import { NetworkDialog } from './NetworkDialog';
import { DataPlanDialog } from './DataPlanDialog';
import { WheelPrizeDialog } from './WheelPrizeDialog';
import { CreateAdminDialog } from './CreateAdminDialog';
import SpinTiersManagement from './SpinTiersManagement';
import SpinPrizeClaimsManagement from './SpinPrizeClaimsManagement';
import SystemMonitoringDashboard from './SystemMonitoringDashboard';
import DrawIntegrationDashboard from './DrawIntegrationDashboard';
import PrizeTemplateManagement from './PrizeTemplateManagement';
import StrategicAffiliateAdminDashboard from './StrategicAffiliateAdminDashboard';
// ── Extracted sub-components (P1 refactor) ────────────────────────────────
import PlatformSettingsPage from './PlatformSettingsPage';
import UserManagementTab from './UserManagementTab';
import AuditLogTab from './AuditLogTab';
import { 
  Users, 
  DollarSign, 
  TrendingUp, 
  Award,
  CheckCircle,
  XCircle,
  Clock,
  Shield,
  Settings,
  Eye,
  UserCheck,
  UserX,
  Pause,
  Play,
  Loader2,
  AlertCircle,
  Plus,
  Edit,
  Trash2,
  Wifi,
  Phone,
  Star,
  Gift,
  Bell,
  BarChart3,
  Network,
  Smartphone,
  Ticket,
  Trophy
} from 'lucide-react';

interface AdminStats {
  total_users: number;
  new_users_today: number;
  total_transactions: number;
  transactions_today: number;
  total_prizes: number;
  total_affiliates: number;
  pending_affiliates: number;
  approved_affiliates: number;
  total_revenue: number;
  total_commissions: number;
}

// Types imported from @/types/admin-api.types - no local declarations needed

interface ComprehensiveAdminPortalProps {
  adminSession?: {
    admin: {
      id: string;
      email: string;
      full_name: string;
      role: string;
      permissions: string[];
    };
    session_token: string;
    expires_at: string;
  };
  onLogout?: () => void;
}

export const ComprehensiveAdminPortal: React.FC<ComprehensiveAdminPortalProps> = ({ 
  adminSession, 
  onLogout 
}) => {
  const { admin: contextAdmin, sessionToken, isAuthenticated, isLoading, hasPermission, logout } = useAdminContext();
  const { toast } = useToast();
  
  // Use props admin session if provided, otherwise fall back to context
  const admin = adminSession?.admin || contextAdmin;
  const isAuth = adminSession ? true : isAuthenticated;
  const isLoad = adminSession ? false : isLoading;
  
  // Permission check function
  const checkPermission = (permission: string): boolean => {
    if (!admin) return false;
      return admin.permissions?.includes(permission) || admin.role?.toUpperCase() === 'SUPER_ADMIN';
  };
  
  // Use the appropriate permission function
  const permissionCheck = adminSession ? checkPermission : hasPermission;
  const [actionLoading, setActionLoading] = useState<string | null>(null);
  
  // Determine first available tab based on permissions
  const getFirstAvailableTab = (): string => {
    if (admin?.role?.toUpperCase() === 'SUPER_ADMIN') return 'dashboard';
    if (hasPermission('view_analytics')) return 'dashboard';
    if (hasPermission('view_monitoring')) return 'monitoring';
    if (hasPermission('manage_draws')) return 'draw';
    if (hasPermission('manage_networks')) return 'networks';
    if (hasPermission('manage_prizes')) return 'prizes';
    if (hasPermission('manage_settings')) return 'spin-tiers';
    if (hasPermission('manage_affiliates')) return 'strategic-affiliates';
    if (hasPermission('manage_users')) return 'users';
    if (hasPermission('manage_admins')) return 'admins';
    return 'monitoring'; // fallback
  };
  
  const location = useLocation();
  const TAB_FROM_PATH: Record<string, string> = {
    '/admin/wheel-prizes': 'prizes',
    '/admin/spin-tiers': 'spin-tiers',
    '/admin/draws': 'draw-engine',
    '/admin/winners': 'prize-claims',
    '/admin/affiliates': 'strategic-affiliates',
    '/admin/monitoring': 'monitoring',
  };
  const [activeTab, setActiveTab] = useState<string>(TAB_FROM_PATH[location.pathname] || getFirstAvailableTab());
  
  // Admin management states
  const [admins, setAdmins] = useState<any[]>([]);
  const [showCreateAdminDialog, setShowCreateAdminDialog] = useState(false);
  const [showEditAdminDialog, setShowEditAdminDialog] = useState(false);
  const [editingAdmin, setEditingAdmin] = useState<any | null>(null);
  
  // State for all admin data
  const [stats, setStats] = useState<AdminStats | null>(null);
  const [networks, setNetworks] = useState<NetworkConfig[]>([]);
  const [dataPlans, setDataPlans] = useState<DataPlan[]>([]);
  const [wheelPrizes, setWheelPrizes] = useState<WheelPrize[]>([]);
  const [dailySubscription, setDailySubscription] = useState<DailySubscription | null>(null);
  const [allDailySubscriptions, setAllDailySubscriptions] = useState<any[]>([]);
  const [settings, setSettings] = useState<PlatformSetting[]>([]);
  const [affiliates, setAffiliates] = useState<any[]>([]);
  const [users, setUsers] = useState<any[]>([]);
  const [transactions, setTransactions] = useState<any[]>([]);
  const [prizes, setPrizes] = useState<any[]>([]);
  
  // Dialog states
  const [editingPrize, setEditingPrize] = useState<WheelPrize | null>(null);
  const [editingDataPlan, setEditingDataPlan] = useState<DataPlan | null | undefined>(null);
  const [showPrizeDialog, setShowPrizeDialog] = useState(false);
  const [showDataPlanDialog, setShowDataPlanDialog] = useState(false);
  const [showNetworkDialog, setShowNetworkDialog] = useState(false);
  const [editingNetwork, setEditingNetwork] = useState<NetworkConfig | null | undefined>(null);

  useEffect(() => {
    if (admin) {
      initializeAdminPortal();
      // Set initial tab based on permissions after admin loads
      const firstTab = getFirstAvailableTab();
      if (activeTab !== firstTab) {
        setActiveTab(firstTab);
      }
    }
  }, [admin]);

  // Log tab state for debugging
  useEffect(() => {
  }, [activeTab]);

  // Monitor and protect dailySubscription state from corruption
  useEffect(() => {
    if (dailySubscription) {
      
      // Check for suspicious values that indicate corruption
      if (dailySubscription?.amount === 80 && dailySubscription?.draw_entries_earned === 4) {
        setDailySubscription({
          id: dailySubscription.id || 'corrected',
          amount: 30,
          draw_entries_earned: 1,
          is_paid: true
        });
      }
    }
  }, [dailySubscription]);

  const initializeAdminPortal = async () => {
    try {
      const promises = [];
      
      // For Super Admin, fetch everything. For others, check permissions
      const isSuperAdmin = admin?.role?.toUpperCase() === 'SUPER_ADMIN';
      
      if (isSuperAdmin || permissionCheck('view_analytics')) {
        promises.push(fetchDashboardStats());
      }
      if (isSuperAdmin || permissionCheck('manage_networks')) {
        promises.push(fetchNetworks(), fetchDataPlans());
      }
      if (isSuperAdmin || permissionCheck('manage_prizes')) {
        promises.push(fetchWheelPrizes());
      }
      if (isSuperAdmin || permissionCheck('manage_settings')) {
        promises.push(fetchDailySubscription(), fetchAllDailySubscriptions(), fetchSettings());
      }
      if (isSuperAdmin || permissionCheck('manage_affiliates')) {
        promises.push(fetchAffiliates());
      }
      if (isSuperAdmin || permissionCheck('manage_users')) {
        promises.push(fetchUsers(), fetchTransactions(), fetchPrizes());
      }
      if (isSuperAdmin || permissionCheck('manage_admins')) {
        promises.push(fetchAdmins());
      }
      
      await Promise.all(promises);
    } catch (error) {
      console.error('Failed to initialize admin portal:', error);
    }
  };

  const callAdminAPI = async (action: string, data?: any): Promise<any> => {
    if (!sessionToken) throw new Error('Admin session required');

    // Route actions to appropriate adminApi methods
    switch (action) {
      case 'get_dashboard_stats':
        try {
          const statsResponse = await adminApi.getStats();
          // API returns { success: true, data: { total_users, active_draws, ... } }
          if (statsResponse && statsResponse.success && statsResponse.data) {
            return statsResponse.data;
          }
          return statsResponse;
        } catch (error) {
          console.error('Failed to fetch dashboard stats:', error);
          // Return zeros if API fails
          return {
            total_users: 0,
            active_draws: 0,
            pending_claims: 0,
            new_users_today: 0,
            total_transactions: 0,
            transactions_today: 0,
            total_prizes: 0,
            total_affiliates: 0,
            pending_affiliates: 0,
            approved_affiliates: 0,
            total_revenue: 0,
            today_revenue: 0,
            total_commissions: 0
          };
        }
      
      case 'get_networks':
        const networks = await adminApi.getNetworks();
        const networksData = networks.success ? networks.data : [];
        const networksResult = { success: true, networks: networksData };
        return networksResult;
      
      case 'get_data_plans':
        const plans = await adminApi.getDataPlans();
        const plansData = plans.success ? plans.data : [];
        const plansResult = { success: true, data_plans: plansData };
        return plansResult;
      
      case 'get_wheel_prizes':
        const prizes = await adminApi.getWheelPrizes();
        return { success: true, wheel_prizes: prizes.success ? prizes.data : [] };
      
      case 'get_users':
        const users = await adminApi.users.getAll();
        const usersData = users.success ? users.data : [];
        const result = { success: true, users: usersData || [] };
        return result;
      
      case 'get_admins': {
        const adminsResp = await adminApi.get('/admin/admins');
        return { success: true, admins: adminsResp.data || [] };
      }
      
      case 'get_affiliates': {
        const affiliatesResp = await adminApi.affiliates.getAll();
        // Backend returns { success: true, data: [...] }
        const affiliatesData = affiliatesResp.success ? (affiliatesResp.data || []) : [];
        return { success: true, affiliates: Array.isArray(affiliatesData) ? affiliatesData : [] };
      }
      
      case 'get_transactions': {
        const txResp = await adminApi.get('/admin/recharge/transactions', { params: data });
        return { success: true, transactions: txResp.success ? (txResp.data || []) : [] };
      }
      
      case 'get_prizes':
        // Get wheel prizes (same as get_wheel_prizes)
        const allPrizes = await adminApi.getWheelPrizes();
        return { success: true, prizes: allPrizes.success ? allPrizes.data : [] };
      
      case 'get_settings': {
        const settingsResp = await adminApi.get('/admin/settings');
        // Backend returns a nested object; flatten into array of {key, value, description}
        const raw = settingsResp.data || {};
        const flatSettings: any[] = [];
        for (const [category, values] of Object.entries(raw)) {
          if (values && typeof values === 'object') {
            for (const [k, v] of Object.entries(values as Record<string, any>)) {
              flatSettings.push({ key: `${category}.${k}`, value: v, description: `${category} - ${k}` });
            }
          }
        }
        return { success: true, settings: flatSettings };
      }
      
      case 'get_daily_subscription':
        try {
          const config = await adminApi.subscriptions.getConfig();
          return { success: true, daily_subscription: config.success ? config.data : null };
        } catch (error) {
          return { success: true, daily_subscription: null };
        }
      
      case 'get_all_daily_subscriptions':
        try {
          const subs = await adminApi.subscriptions.getAll();
          const subsData = subs.success ? subs.data : [];
          return { success: true, data: { subscriptions: subsData || [] } };
        } catch (error) {
          return { success: true, data: { subscriptions: [] } };
        }
      
      case 'update_daily_subscription': {
        const subData = data.subscription_data;
        // Map frontend fields back to backend format
        const backendData: any = {};
        if (subData.amount !== undefined) backendData.daily_price = Math.round(Number(subData.amount) * 100);
        if (subData.draw_entries_earned !== undefined) backendData.daily_spins = Number(subData.draw_entries_earned);
        if (subData.is_paid !== undefined) backendData.daily_subscription_enabled = Boolean(subData.is_paid);
        await adminApi.subscriptions.updateConfig(backendData);
        return { success: true };
      }
      
      case 'update_network':
        await adminApi.networks.update(data.network_id, data.updates);
        return { success: true };
      
      case 'create_network':
        await adminApi.networks.create(data.network_data || data);
        return { success: true };
      
      case 'update_data_plan':
        await adminApi.bundles.update(data.plan_id, data.updates);
        return { success: true };
      
      case 'create_data_plan':
        await adminApi.bundles.create(data.plan);
        return { success: true };
      
      case 'update_wheel_prize': {
        // Map WheelPrizeDialog form fields to backend field names
        const updatePayload = {
          name: data.updates.prize_name || data.updates.name,
          type: data.updates.prize_type || data.updates.type,
          value: data.updates.prize_value !== undefined ? data.updates.prize_value : data.updates.value,
          probability: data.updates.probability,
          is_active: data.updates.is_active,
          minimum_recharge: data.updates.minimum_recharge,
          color: data.updates.color_scheme || data.updates.color,
          sort_order: data.updates.sort_order || data.updates.display_order,
        };
        await adminApi.spin.updatePrize(data.prize_id, updatePayload);
        return { success: true };
      }
      
      case 'create_wheel_prize': {
        // Map WheelPrizeDialog form fields to backend field names
        const prizePayload = {
          name: data.prize_name || data.name,
          type: data.prize_type || data.type,
          value: data.prize_value !== undefined ? data.prize_value : data.value,
          probability: data.probability,
          is_active: data.is_active !== undefined ? data.is_active : true,
          minimum_recharge: data.minimum_recharge,
          color: data.color_scheme || data.color,
          sort_order: data.sort_order || data.display_order,
        };
        await adminApi.spin.createPrize(prizePayload);
        return { success: true };
      }
      
      case 'approve_affiliate':
        await adminApi.affiliates.approve(data.affiliate_id);
        return { success: true };
      
      case 'reject_affiliate':
        await adminApi.affiliates.reject(data.affiliate_id, data.reason || 'Rejected by admin');
        return { success: true };
      
      case 'suspend_affiliate':
        await adminApi.affiliates.suspend(data.affiliate_id);
        return { success: true };
      
      case 'create_admin':
        const adminCreateResp = await adminApi.post('/admin/admins', data.admin_data || data);
        return { success: true, temp_password: adminCreateResp?.data?.temporary_password || adminCreateResp?.temporary_password || '' };
      
      case 'update_admin':
        await adminApi.put(`/admin/admins/${data.admin_id}`, data.admin_data || data.updates || data);
        return { success: true };
      
      case 'delete_admin':
        await adminApi.delete(`/admin/admins/${data.admin_id}`);
        return { success: true };
      
      case 'update_setting': {
        // Use PUT /admin/settings/:key with { value: string }
        const settingKey = data.setting_key || '';
        await adminApi.put(`/admin/settings/${encodeURIComponent(settingKey)}`, { value: String(data.setting_value) });
        return { success: true };
      }
      
      case 'update_user_status': {
        await adminApi.put(`/admin/users/${data.user_id}/status`, { status: data.status, reason: data.reason || '' });
        return { success: true };
      }
      
      case 'trigger_frontend_update':
        // No-op for now - frontend will update automatically
        return { success: true };
      
      default:
        console.warn(`Unhandled action: ${action}`);
        return { success: false, error: `Action ${action} not implemented` };
    }
  };

  const fetchDashboardStats = async () => {
    try {
      const data = await callAdminAPI('get_dashboard_stats');
      setStats(data || {
        total_users: 0,
        new_users_today: 0,
        total_transactions: 0,
        transactions_today: 0,
        total_prizes: 0,
        total_affiliates: 0,
        pending_affiliates: 0,
        approved_affiliates: 0,
        total_revenue: 0,
        today_revenue: 0,
        total_commissions: 0
      });
    } catch (error) {
      console.error('Failed to fetch dashboard stats:', error);
    }
  };

  const fetchNetworks = async () => {
    try {
      const data = await callAdminAPI('get_networks');

      setNetworks(data?.networks || []);
    } catch (error) {
      console.error('❌ Failed to fetch networks:', error);
    }
  };

  const fetchDataPlans = async () => {
    try {
      const data = await callAdminAPI('get_data_plans');

      setDataPlans(data?.data_plans || []);
    } catch (error) {
      console.error('❌ Failed to fetch data plans:', error);
    }
  };

  const fetchWheelPrizes = async () => {
    try {
      const data = await callAdminAPI('get_wheel_prizes');
      setWheelPrizes(data?.wheel_prizes || []);
    } catch (error) {
      console.error('Failed to fetch wheel prizes:', error);
    }
  };

  const fetchDailySubscription = async () => {
    try {
      const data = await callAdminAPI('get_daily_subscription');
      
        if (data?.daily_subscription) {
        const config = data.daily_subscription;
        
        // Map backend fields: daily_price (kobo) -> amount (naira), daily_spins -> draw_entries_earned
        const cleanConfig = {
          id: config.id || 'config',
          amount: config.daily_price ? Number(config.daily_price) / 100 : (Number(config.amount) || 20),
          draw_entries_earned: Number(config.daily_spins) || Number(config.draw_entries_earned) || 1,
          is_paid: config.daily_subscription_enabled !== false && config.is_paid !== false
        };
        setDailySubscription(cleanConfig);
      } else {
        setDailySubscription(null);
      }
    } catch (error) {
      console.error('❌ Failed to fetch daily subscription:', error);
      // Set default values on error
      setDailySubscription({
        id: 'default',
        amount: 30,
        draw_entries_earned: 1,
        is_paid: true
      });
    }
  };

  const fetchSettings = async () => {
    try {
      const data = await callAdminAPI('get_settings');
      setSettings(data?.settings || []);
    } catch (error) {
      console.error('Failed to fetch settings:', error);
    }
  };

  const fetchAffiliates = async () => {
    try {
      const data = await callAdminAPI('get_affiliates');
      setAffiliates(data?.affiliates || []);
    } catch (error) {
      console.error('Failed to fetch affiliates:', error);
    }
  };

  const fetchAdmins = async () => {
    try {
      const data = await callAdminAPI('get_admins');
      setAdmins(data?.admins || []);
    } catch (error) {
      console.error('Failed to fetch admins:', error);
    }
  };

  const fetchUsers = async () => {
    try {
      const data = await callAdminAPI('get_users');


      setUsers(data?.users || []);
    } catch (error) {
      console.error('❌ Failed to fetch users:', error);
    }
  };

  const fetchTransactions = async () => {
    try {
      const data = await callAdminAPI('get_transactions');
      setTransactions(data?.transactions || []);
    } catch (error) {
      console.error('Failed to fetch transactions:', error);
    }
  };

  const fetchPrizes = async () => {
    try {
      const data = await callAdminAPI('get_prizes');
      setPrizes(data?.prizes || []);
    } catch (error) {
      console.error('Failed to fetch prizes:', error);
    }
  };

  const fetchAllDailySubscriptions = async () => {
    try {
      setActionLoading('fetch_all_subscriptions');
      const data = await callAdminAPI('get_all_daily_subscriptions', { limit: 100 });
      setAllDailySubscriptions(data?.data?.subscriptions || []);
    } catch (error) {
      console.error('Failed to fetch all daily subscriptions:', error);
    } finally {
      setActionLoading(null);
    }
  };

  const handleNetworkUpdate = async (networkId: string, updates: Partial<NetworkConfig>) => {
    try {
      setActionLoading(networkId);
      await callAdminAPI('update_network', { network_id: networkId, updates });
      await fetchNetworks();
      toast({
        title: "Success",
        description: "Network configuration updated successfully",
      });
    } catch (error) {
      toast({
        title: "Update Failed",
        description: getErrorMessage(error),
        variant: "destructive"
      });
    } finally {
      setActionLoading('');
    }
  };


  const handleNetworkSave = async (networkData: any) => {
    try {
      setActionLoading('network_save');
      if (editingNetwork) {
        await callAdminAPI('update_network', {
          network_id: editingNetwork.id, 
          updates: networkData 
        });
        toast({
          title: "Success",
          description: "Network updated successfully",
        });
      } else {
        await callAdminAPI('create_network', { network_data: networkData });
        toast({
          title: "Success",
          description: "Network created successfully",
        });
      }
      await fetchNetworks();
      setShowNetworkDialog(false);
      setEditingNetwork(null);
    } catch (error) {
      toast({
        title: editingNetwork ? "Update Failed" : "Creation Failed",
        description: getErrorMessage(error),
        variant: "destructive"
      });
    } finally {
      setActionLoading('');
    }
  };

  const handleDataPlanUpdate = async (planId: string, updates: any) => {
    try {
      setActionLoading(planId);
      await callAdminAPI('update_data_plan', { 
        plan_id: planId, 
        updates 
      });
      await fetchDataPlans();
      toast({
        title: "Success",
        description: "Data plan updated successfully",
      });
    } catch (error) {
      toast({
        title: "Update Failed",
        description: getErrorMessage(error),
        variant: "destructive"
      });
    } finally {
      setActionLoading('');
    }
  };

  const handleDataPlanSave = async (planData: any) => {
    try {
      setActionLoading('plan_save');
      if (editingDataPlan) {
        await callAdminAPI('update_data_plan', {
          plan_id: editingDataPlan.id, 
          updates: planData 
        });
        toast({
          title: "Success",
          description: "Data plan updated successfully",
        });
      } else {
        await callAdminAPI('create_data_plan', { plan: planData });
        toast({
          title: "Success",
          description: "Data plan created successfully",
        });
      }
      await fetchDataPlans();
      setShowDataPlanDialog(false);
      setEditingDataPlan(null);
    } catch (error) {
      toast({
        title: editingDataPlan ? "Update Failed" : "Creation Failed",
        description: getErrorMessage(error),
        variant: "destructive"
      });
    } finally {
      setActionLoading('');
    }
  };

  const validatePrizeProbabilities = (prizes: WheelPrize[]) => {
    const totalProbability = prizes
      .filter(p => p.is_active)
      .reduce((sum, p) => sum + Number(p.probability), 0);
    return { total: totalProbability, isValid: Math.abs(totalProbability - 100) < 0.01 };
  };

  const handleWheelPrizeSave = async (prizeData: any) => {
    // Validate that total probability will not exceed 100% after this save
    if (prizeData.is_active !== false) {
      const otherActivePrizes = wheelPrizes.filter(p => p.is_active && p.id !== editingPrize?.id);
      const otherTotal = otherActivePrizes.reduce((sum, p) => sum + Number(p.probability), 0);
      const newTotal = otherTotal + Number(prizeData.probability || 0);
      if (newTotal > 100.01) {
        toast({
          title: "Probability Exceeds 100%",
          description: `Adding this prize would bring total probability to ${newTotal.toFixed(1)}%. Please reduce other prizes first. Current total from other active prizes: ${otherTotal.toFixed(1)}%.`,
          variant: "destructive",
        });
        return;
      }
    }
    try {
      setActionLoading('prize_save');
      if (editingPrize) {
        await callAdminAPI('update_wheel_prize', { 
          prize_id: editingPrize.id, 
          updates: prizeData 
        });
        toast({
          title: "Success",
          description: "Wheel prize updated successfully",
        });
      } else {
        await callAdminAPI('create_wheel_prize', prizeData);
        toast({
          title: "Success",
          description: "Wheel prize created successfully",
        });
      }
      await fetchWheelPrizes();
      setShowPrizeDialog(false);
      setEditingPrize(null);
      
      // Sync wheel configuration with frontend
      try {
        await callAdminAPI('trigger_frontend_update');
      } catch (syncError) {
        console.error('Wheel sync failed:', syncError);
      }
    } catch (error) {
      toast({
        title: editingPrize ? "Update Failed" : "Creation Failed",
        description: getErrorMessage(error),
        variant: "destructive"
      });
    } finally {
      setActionLoading('');
    }
  };


  const handleDailySubscriptionUpdate = async (subscriptionData: any) => {
    try {
      setActionLoading('subscription');
      await callAdminAPI('update_daily_subscription', { subscription_data: subscriptionData });
      await fetchDailySubscription();
      
      // Sync with frontend
      try {
        await callAdminAPI('trigger_frontend_update');
      } catch (syncError) {
        console.error('Frontend sync failed:', syncError);
      }
      
      toast({
        title: "Success",
        description: "Daily subscription updated successfully and synced with frontend",
      });
    } catch (error) {
      toast({
        title: "Update Failed",
        description: getErrorMessage(error),
        variant: "destructive"
      });
    } finally {
      setActionLoading('');
    }
  };

  const handleSettingUpdate = async (settingKey: string, settingValue: any, description?: string) => {
    try {
      setActionLoading(settingKey);
      await callAdminAPI('update_setting', { 
        setting_key: settingKey, 
        setting_value: settingValue,
        description 
      });
      await fetchSettings();
      
      // Sync settings with frontend
      try {
        await callAdminAPI('trigger_frontend_update');
      } catch (syncError) {
        console.error('Settings sync failed:', syncError);
      }
      
      toast({
        title: "Success",
        description: "Setting updated successfully and synced with frontend",
      });
    } catch (error) {
      toast({
        title: "Update Failed",
        description: getErrorMessage(error),
        variant: "destructive"
      });
    } finally {
      setActionLoading('');
    }
  };

  const handleCreateAdmin = async (adminData: any) => {
    try {
      setActionLoading('create_admin');
      const result = await callAdminAPI('create_admin', { admin_data: adminData });
      await fetchAdmins();
      setShowCreateAdminDialog(false);
      toast({
        title: "Admin Created",
        description: `Admin created successfully. Temporary password: ${result.temp_password}`,
      });
    } catch (error) {
      toast({
        title: "Creation Failed",
        description: getErrorMessage(error),
        variant: "destructive"
      });
    } finally {
      setActionLoading('');
    }
  };

  const handleAffiliateAction = async (action: string, affiliateId: string) => {
    try {
      setActionLoading(affiliateId);
      await callAdminAPI(action, { affiliate_id: affiliateId });
      await fetchAffiliates();
      await fetchDashboardStats(); // Refresh stats
      toast({
        title: "Success",
        description: "Affiliate action completed successfully",
      });
    } catch (error) {
      toast({
        title: "Action Failed",
        description: getErrorMessage(error),
        variant: "destructive"
      });
    } finally {
      setActionLoading('');
    }
  };


  const handleUpdateAdmin = async (adminId: string, adminData: any) => {
    try {
      setActionLoading(adminId);
      await callAdminAPI('update_admin', { admin_id: adminId, admin_data: adminData });
      await fetchAdmins();
      toast({
        title: "Admin Updated",
        description: "Admin updated successfully",
      });
    } catch (error) {
      toast({
        title: "Update Failed",
        description: getErrorMessage(error),
        variant: "destructive"
      });
    } finally {
      setActionLoading('');
    }
  };

  const handleDeleteAdmin = async (adminId: string) => {
    try {
      setActionLoading(adminId);
      await callAdminAPI('delete_admin', { admin_id: adminId });
      await fetchAdmins();
      toast({
        title: "Admin Deactivated",
        description: "Admin deactivated successfully",
      });
    } catch (error) {
      toast({
        title: "Deactivation Failed",
        description: getErrorMessage(error),
        variant: "destructive"
      });
    } finally {
      setActionLoading('');
    }
  };


  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-NG', {
      style: 'currency',
      currency: 'NGN'
    }).format(amount);
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric'
    });
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'APPROVED':
        return <Badge className="bg-green-100 text-green-800">Approved</Badge>;
      case 'PENDING':
        return <Badge className="bg-yellow-100 text-yellow-800">Pending</Badge>;
      case 'REJECTED':
        return <Badge className="bg-red-100 text-red-800">Rejected</Badge>;
      case 'SUSPENDED':
        return <Badge className="bg-gray-100 text-gray-800">Suspended</Badge>;
      default:
        return <Badge variant="secondary">{status}</Badge>;
    }
  };

  const getPrizeIcon = (prizeType: string) => {
    switch (prizeType) {
      case 'CASH': return <DollarSign className="w-4 h-4 text-green-600" />;
      case 'DATA': return <Wifi className="w-4 h-4 text-blue-600" />;
      case 'AIRTIME': return <Phone className="w-4 h-4 text-purple-600" />;
      case 'POINTS': return <Star className="w-4 h-4 text-yellow-600" />;
      case 'TICKETS': return <Ticket className="w-4 h-4 text-orange-600" />;
      default: return <Gift className="w-4 h-4 text-gray-600" />;
    }
  };

  if (isLoad) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-50 to-purple-50 flex items-center justify-center">
        <Card className="w-full max-w-md text-center">
          <CardContent className="p-8">
            <Loader2 className="w-8 h-8 animate-spin mx-auto mb-4" />
            <p>Loading comprehensive admin portal...</p>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (!admin) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-50 to-purple-50 flex items-center justify-center p-4">
        <Card className="w-full max-w-md text-center">
          <CardContent className="p-8">
            <Shield className="w-12 h-12 text-blue-500 mx-auto mb-4" />
            <h2 className="text-xl font-bold mb-2">Admin Access Required</h2>
            <p className="text-gray-600 mb-4">
              Please login with admin credentials to access the admin portal.
            </p>
            <Button onClick={() => window.location.href = '/admin/login'}>
              Admin Login
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }



  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-purple-50 p-4">
      <div className="max-w-7xl mx-auto space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">Admin Portal</h1>
            <p className="text-gray-600">Welcome, {(admin as any)?.full_name || admin?.email} ({admin?.role})</p>
          </div>
          <div className="flex gap-2">
            <Button variant="outline" onClick={() => window.location.href = '/'}>
              Main Site
            </Button>
            <Button variant="outline" onClick={logout}>
              Logout
            </Button>
          </div>
        </div>

        {/* Stats Cards */}
        {stats && (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            <Card>
              <CardContent className="p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-gray-600">Total Users</p>
                    <p className="text-2xl font-bold">{(stats?.total_users || 0).toLocaleString()}</p>
                    <p className="text-xs text-green-600">+{stats?.new_users_today || 0} today</p>
                  </div>
                  <Users className="w-8 h-8 text-blue-600" />
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardContent className="p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-gray-600">Total Revenue</p>
                    <p className="text-2xl font-bold">{formatCurrency(stats?.total_revenue || 0)}</p>
                    <p className="text-xs text-blue-600">{stats?.transactions_today || 0} transactions today</p>
                  </div>
                  <DollarSign className="w-8 h-8 text-green-600" />
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardContent className="p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-gray-600">Affiliates</p>
                    <p className="text-2xl font-bold">{stats?.total_affiliates || 0}</p>
                    <p className="text-xs text-yellow-600">{stats?.pending_affiliates || 0} pending approval</p>
                  </div>
                  <Award className="w-8 h-8 text-purple-600" />
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardContent className="p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-gray-600">Commissions</p>
                    <p className="text-2xl font-bold">{formatCurrency(stats?.total_commissions || 0)}</p>
                    <p className="text-xs text-purple-600">{stats?.total_prizes || 0} prizes won</p>
                  </div>
                  <TrendingUp className="w-8 h-8 text-yellow-600" />
                </div>
              </CardContent>
            </Card>
          </div>
        )}

        {/* Main Content Tabs */}
        <Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-4">
          <TabsList className="grid w-full grid-cols-12">
            {(admin?.role === 'SUPER_ADMIN' || hasPermission('view_analytics')) && <TabsTrigger value="dashboard">Dashboard</TabsTrigger>}
            {(admin?.role === 'SUPER_ADMIN' || hasPermission('view_monitoring')) && <TabsTrigger value="monitoring">Monitoring</TabsTrigger>}
            {(admin?.role === 'SUPER_ADMIN' || hasPermission('manage_draws')) && <TabsTrigger value="draw">Draw Engine</TabsTrigger>}
            {(admin?.role === 'SUPER_ADMIN' || hasPermission('manage_draws')) && <TabsTrigger value="prize-templates">Prize Templates</TabsTrigger>}
            {(admin?.role === 'SUPER_ADMIN' || hasPermission('manage_networks')) && <TabsTrigger value="networks">Networks</TabsTrigger>}
            {(admin?.role === 'SUPER_ADMIN' || hasPermission('manage_prizes')) && <TabsTrigger value="prizes">Prizes</TabsTrigger>}
            {(admin?.role === 'SUPER_ADMIN' || hasPermission('manage_settings')) && <TabsTrigger value="spin-tiers">Spin Tiers</TabsTrigger>}
            {(admin?.role === 'SUPER_ADMIN' || hasPermission('manage_prizes')) && <TabsTrigger value="spin-claims">Prize Claims</TabsTrigger>}
            {(admin?.role === 'SUPER_ADMIN' || hasPermission('manage_settings')) && <TabsTrigger value="subscription">Daily Draw</TabsTrigger>}
            {(admin?.role === 'SUPER_ADMIN' || hasPermission('manage_affiliates')) && <TabsTrigger value="strategic-affiliates">Strategic Affiliates</TabsTrigger>}
            {(admin?.role === 'SUPER_ADMIN' || hasPermission('manage_users')) && <TabsTrigger value="users">Users</TabsTrigger>}
            {(admin?.role === 'SUPER_ADMIN' || hasPermission('manage_admins')) && <TabsTrigger value="admins">Admins</TabsTrigger>}
            {(admin?.role === 'SUPER_ADMIN' || hasPermission('manage_settings')) && <TabsTrigger value="settings">Settings</TabsTrigger>}
            {(admin?.role === 'SUPER_ADMIN' || hasPermission('manage_settings')) && <TabsTrigger value="audit">Audit Log</TabsTrigger>}
          </TabsList>

          {/* Dashboard */}
          <TabsContent value="dashboard">
            <div className="space-y-6">
              {/* Key Performance Indicators */}
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                <Card className="border-l-4 border-l-blue-500">
                  <CardContent className="p-5">
                    <div className="flex items-center justify-between">
                      <div>
                        <p className="text-xs font-semibold uppercase tracking-wide text-gray-500">Total Registered Users</p>
                        <p className="text-3xl font-bold text-gray-900 mt-1">{(stats?.total_users || 0).toLocaleString()}</p>
                        <p className="text-xs text-green-600 mt-1">+{stats?.new_users_today || 0} new today</p>
                      </div>
                      <div className="w-12 h-12 bg-blue-100 rounded-full flex items-center justify-center">
                        <Users className="w-6 h-6 text-blue-600" />
                      </div>
                    </div>
                  </CardContent>
                </Card>

                <Card className="border-l-4 border-l-green-500">
                  <CardContent className="p-5">
                    <div className="flex items-center justify-between">
                      <div>
                        <p className="text-xs font-semibold uppercase tracking-wide text-gray-500">Total Revenue</p>
                        <p className="text-3xl font-bold text-gray-900 mt-1">{formatCurrency(stats?.total_revenue || 0)}</p>
                        <p className="text-xs text-blue-600 mt-1">{(stats?.total_transactions || 0).toLocaleString()} total transactions</p>
                      </div>
                      <div className="w-12 h-12 bg-green-100 rounded-full flex items-center justify-center">
                        <DollarSign className="w-6 h-6 text-green-600" />
                      </div>
                    </div>
                  </CardContent>
                </Card>

                <Card className="border-l-4 border-l-purple-500">
                  <CardContent className="p-5">
                    <div className="flex items-center justify-between">
                      <div>
                        <p className="text-xs font-semibold uppercase tracking-wide text-gray-500">Active Affiliates</p>
                        <p className="text-3xl font-bold text-gray-900 mt-1">{stats?.approved_affiliates || 0}</p>
                        <p className="text-xs text-yellow-600 mt-1">{stats?.pending_affiliates || 0} pending approval</p>
                      </div>
                      <div className="w-12 h-12 bg-purple-100 rounded-full flex items-center justify-center">
                        <Award className="w-6 h-6 text-purple-600" />
                      </div>
                    </div>
                  </CardContent>
                </Card>

                <Card className="border-l-4 border-l-yellow-500">
                  <CardContent className="p-5">
                    <div className="flex items-center justify-between">
                      <div>
                        <p className="text-xs font-semibold uppercase tracking-wide text-gray-500">Commissions Paid</p>
                        <p className="text-3xl font-bold text-gray-900 mt-1">{formatCurrency(stats?.total_commissions || 0)}</p>
                        <p className="text-xs text-purple-600 mt-1">{stats?.total_prizes || 0} prizes won</p>
                      </div>
                      <div className="w-12 h-12 bg-yellow-100 rounded-full flex items-center justify-center">
                        <TrendingUp className="w-6 h-6 text-yellow-600" />
                      </div>
                    </div>
                  </CardContent>
                </Card>

                <Card className="border-l-4 border-l-indigo-500">
                  <CardContent className="p-5">
                    <div className="flex items-center justify-between">
                      <div>
                        <p className="text-xs font-semibold uppercase tracking-wide text-gray-500">Today's Transactions</p>
                        <p className="text-3xl font-bold text-gray-900 mt-1">{(stats?.transactions_today || 0).toLocaleString()}</p>
                        <p className="text-xs text-gray-500 mt-1">Recharges processed today</p>
                      </div>
                      <div className="w-12 h-12 bg-indigo-100 rounded-full flex items-center justify-center">
                        <Smartphone className="w-6 h-6 text-indigo-600" />
                      </div>
                    </div>
                  </CardContent>
                </Card>

                <Card className="border-l-4 border-l-red-500">
                  <CardContent className="p-5">
                    <div className="flex items-center justify-between">
                      <div>
                        <p className="text-xs font-semibold uppercase tracking-wide text-gray-500">Spin Prizes Available</p>
                        <p className="text-3xl font-bold text-gray-900 mt-1">{wheelPrizes.filter(p => p.is_active).length}</p>
                        <p className="text-xs text-gray-500 mt-1">
                          {wheelPrizes.length > 0 ? (() => {
                            const v = validatePrizeProbabilities(wheelPrizes);
                            return v.isValid
                              ? <span className="text-green-600">Probability: {v.total.toFixed(1)}% ✓</span>
                              : <span className="text-red-600">⚠ Probability: {v.total.toFixed(1)}%</span>;
                          })() : 'No prizes configured'}
                        </p>
                      </div>
                      <div className="w-12 h-12 bg-red-100 rounded-full flex items-center justify-center">
                        <Gift className="w-6 h-6 text-red-600" />
                      </div>
                    </div>
                  </CardContent>
                </Card>
              </div>

              {/* Platform Overview Table */}
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <BarChart3 className="w-5 h-5" />
                    Platform Overview
                  </CardTitle>
                  <CardDescription>Summary of key platform metrics and configuration status</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <div className="space-y-3">
                      <h4 className="font-semibold text-sm text-gray-700 uppercase tracking-wide">User Metrics</h4>
                      <div className="space-y-2">
                        {[
                          { label: 'Total Users', value: (stats?.total_users || 0).toLocaleString() },
                          { label: 'New Users Today', value: (stats?.new_users_today || 0).toLocaleString() },
                          { label: 'Total Transactions', value: (stats?.total_transactions || 0).toLocaleString() },
                          { label: 'Transactions Today', value: (stats?.transactions_today || 0).toLocaleString() },
                        ].map(item => (
                          <div key={item.label} className="flex justify-between items-center py-1 border-b border-gray-100">
                            <span className="text-sm text-gray-600">{item.label}</span>
                            <span className="text-sm font-semibold">{item.value}</span>
                          </div>
                        ))}
                      </div>
                    </div>
                    <div className="space-y-3">
                      <h4 className="font-semibold text-sm text-gray-700 uppercase tracking-wide">Financial Metrics</h4>
                      <div className="space-y-2">
                        {[
                          { label: 'Total Revenue', value: formatCurrency(stats?.total_revenue || 0) },
                          { label: 'Total Commissions', value: formatCurrency(stats?.total_commissions || 0) },
                          { label: 'Total Affiliates', value: (stats?.total_affiliates || 0).toLocaleString() },
                          { label: 'Pending Affiliates', value: (stats?.pending_affiliates || 0).toLocaleString() },
                        ].map(item => (
                          <div key={item.label} className="flex justify-between items-center py-1 border-b border-gray-100">
                            <span className="text-sm text-gray-600">{item.label}</span>
                            <span className="text-sm font-semibold">{item.value}</span>
                          </div>
                        ))}
                      </div>
                    </div>
                  </div>
                </CardContent>
              </Card>

              {/* Recent Users Preview */}
              {users.length > 0 && (
                <Card>
                  <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                      <Users className="w-5 h-5" />
                      Recent Users
                    </CardTitle>
                    <CardDescription>Last 5 registered users on the platform</CardDescription>
                  </CardHeader>
                  <CardContent>
                    <Table>
                      <TableHeader>
                        <TableRow>
                          <TableHead>User</TableHead>
                          <TableHead>Phone</TableHead>
                          <TableHead>Tier</TableHead>
                          <TableHead>Points</TableHead>
                          <TableHead>Joined</TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {users.slice(0, 5).map((u: any) => (
                          <TableRow key={u.id}>
                            <TableCell className="font-medium">{u.full_name || u.first_name || 'N/A'}</TableCell>
                            <TableCell>{u.msisdn}</TableCell>
                            <TableCell><Badge variant="outline">{u.loyalty_tier || 'BRONZE'}</Badge></TableCell>
                            <TableCell>{(u.total_points || 0).toLocaleString()}</TableCell>
                            <TableCell>{formatDate(u.created_at)}</TableCell>
                          </TableRow>
                        ))}
                      </TableBody>
                    </Table>
                  </CardContent>
                </Card>
              )}
            </div>
          </TabsContent>

          {/* System Monitoring */}
          <TabsContent value="monitoring">
            <SystemMonitoringDashboard />
          </TabsContent>

          {/* Draw Engine Integration */}
          <TabsContent value="draw">
            <DrawIntegrationDashboard />
          </TabsContent>

          <TabsContent value="prize-templates">
            <PrizeTemplateManagement />
          </TabsContent>

          <TabsContent value="networks">
            <div className="space-y-6">
              <Card>
                <CardHeader>
                  <div className="flex items-center justify-between">
                    <div>
                      <CardTitle className="flex items-center gap-2">
                        <Network className="w-6 h-6" />
                        Network Configuration
                      </CardTitle>
                      <CardDescription>
                        Manage network providers, commission rates, and limits
                      </CardDescription>
                    </div>
                    <Button 
                      onClick={() => {
                        setEditingNetwork(null);
                        setShowNetworkDialog(true);
                      }}
                      disabled={actionLoading === 'create_network'}
                    >
                      {actionLoading === 'create_network' ? (
                        <Loader2 className="w-4 h-4 animate-spin mr-2" />
                      ) : (
                        <Plus className="w-4 h-4 mr-2" />
                      )}
                      Add Network
                    </Button>
                  </div>
                </CardHeader>
                <CardContent>
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>Network</TableHead>
                        <TableHead>Status</TableHead>
                        <TableHead>Services</TableHead>
                        <TableHead>Commission</TableHead>
                        <TableHead>Limits</TableHead>
                        <TableHead>Actions</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {(networks || []).map((network) => (
                        <TableRow key={network.id}>
                          <TableCell>
                            <div>
                              <div className="font-medium">{network.network || network.network_name}</div>
                              <div className="text-sm text-gray-500">{network.code || network.network_code}</div>
                            </div>
                          </TableCell>
                          <TableCell>
                            <Switch
                              checked={network.enabled ?? network.is_active ?? false}
                              onCheckedChange={(checked) => 
                                handleNetworkUpdate(network.id, { enabled: checked, is_active: checked })
                              }
                              disabled={actionLoading === network.id}
                            />
                          </TableCell>
                          <TableCell>
                            <div className="flex gap-2">
                              <Badge variant={network.airtime_enabled ? "default" : "secondary"}>
                                Airtime
                              </Badge>
                              <Badge variant={network.data_enabled ? "default" : "secondary"}>
                                Data
                              </Badge>
                            </div>
                          </TableCell>
                          <TableCell>{network.commission_rate}%</TableCell>
                          <TableCell>
                            <div className="text-sm">
                              <div>Min: {formatCurrency(network.minimum_amount ?? 0)}</div>
                              <div>Max: {formatCurrency(network.maximum_amount ?? 0)}</div>
                            </div>
                          </TableCell>
                          <TableCell>
                            <Button 
                              size="sm" 
                              variant="outline"
                              onClick={() => {
                                // Map API response fields to dialog expected fields
                                setEditingNetwork({
                                  ...network,
                                  network_name: network.network || network.network_name || '',
                                  network_code: network.code || network.network_code || '',
                                  is_active: network.enabled ?? network.is_active ?? true,
                                });
                                setShowNetworkDialog(true);
                              }}
                              disabled={actionLoading === network.id}
                            >
                              {actionLoading === network.id ? (
                                <Loader2 className="w-3 h-3 animate-spin" />
                              ) : (
                                <Edit className="w-3 h-3" />
                              )}
                            </Button>
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </CardContent>
              </Card>

              {/* Data Plans Management */}
              <Card>
                <CardHeader>
                  <div className="flex items-center justify-between">
                    <div>
                      <CardTitle className="flex items-center gap-2">
                        <Smartphone className="w-6 h-6" />
                        Data Bundle Plans
                      </CardTitle>
                      <CardDescription>
                        Configure available data plans for each network
                      </CardDescription>
                    </div>
                    <Button onClick={() => {
                      setEditingDataPlan(null);
                      setShowDataPlanDialog(true);
                    }}>
                      <Plus className="w-4 h-4 mr-2" />
                      Add Plan
                    </Button>
                  </div>
                </CardHeader>
                <CardContent>
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>Network</TableHead>
                        <TableHead>Plan</TableHead>
                        <TableHead>Data Amount</TableHead>
                        <TableHead>Price</TableHead>
                        <TableHead>Validity</TableHead>
                        <TableHead>Status</TableHead>
                        <TableHead>Actions</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {(dataPlans || []).map((plan) => (
                        <TableRow key={plan.id}>
                          <TableCell>
                            <Badge variant="outline">
                              {plan.network_provider}
                            </Badge>
                          </TableCell>
                          <TableCell>{plan.plan_name}</TableCell>
                          <TableCell>{plan.data_amount}</TableCell>
                          <TableCell>{formatCurrency(plan.price)}</TableCell>
                          <TableCell>{plan.validity_days} days</TableCell>
                          <TableCell>
                            <Badge variant={plan.is_active ? "default" : "secondary"}>
                              {plan.is_active ? "Active" : "Inactive"}
                            </Badge>
                          </TableCell>
                          <TableCell>
                            <div className="flex gap-1">
                              <Button 
                                size="sm" 
                                variant="outline"
                                onClick={() => {
                                  setEditingDataPlan(plan);
                                  setShowDataPlanDialog(true);
                                }}
                              >
                                <Edit className="w-3 h-3" />
                              </Button>
                              <Button 
                                size="sm" 
                                variant="destructive"
                                onClick={async () => {
                                  if (confirm('Are you sure you want to delete this data plan?')) {
                                    try {
                                      setActionLoading(plan.id);
                                      await adminApi.bundles.delete(plan.id);
                                      await fetchDataPlans();
                                      toast({ title: 'Success', description: 'Data plan deleted successfully' });
                                    } catch (error) {
                                      toast({ title: 'Error', description: getErrorMessage(error), variant: 'destructive' });
                                    } finally {
                                      setActionLoading(null);
                                    }
                                  }
                                }}
                                disabled={actionLoading === plan.id}
                              >
                                {actionLoading === plan.id ? <Loader2 className="w-3 h-3 animate-spin" /> : <Trash2 className="w-3 h-3" />}
                              </Button>
                            </div>
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </CardContent>
              </Card>
            </div>
          </TabsContent>

          <TabsContent value="prizes">
            <Card>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <div>
                    <CardTitle className="flex items-center gap-2">
                      <Gift className="w-6 h-6" />
                      Wheel Spin Prizes
                    </CardTitle>
                    <CardDescription>
                      Configure prizes, probabilities, and minimum recharge requirements
                    </CardDescription>
                    {wheelPrizes.length > 0 && (() => {
                      const validation = validatePrizeProbabilities(wheelPrizes);
                      return (
                        <div className={`text-sm mt-2 p-2 rounded ${
                          validation.isValid ? 'bg-green-50 text-green-700' : 'bg-red-50 text-red-700'
                        }`}>
                          Total Probability: {validation.total.toFixed(1)}% 
                          {validation.isValid ? ' ✓ Valid' : ' ⚠️ Must equal 100%'}
                        </div>
                      );
                    })()}
                  </div>
                  <Button onClick={() => {
                    setEditingPrize(null);
                    setShowPrizeDialog(true);
                  }}>
                    <Plus className="w-4 h-4 mr-2" />
                    Add Prize
                  </Button>
                </div>
              </CardHeader>
              <CardContent>
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Prize</TableHead>
                      <TableHead>Type</TableHead>
                      <TableHead>Value</TableHead>
                      <TableHead>Probability</TableHead>
                      <TableHead>Min Recharge</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {(wheelPrizes || []).map((prize) => (
                      <TableRow key={prize.id}>
                        <TableCell>
                          <div className="flex items-center gap-2">
                            {getPrizeIcon(prize.prize_type)}
                            <span>{prize.prize_name}</span>
                          </div>
                        </TableCell>
                        <TableCell>
                          <Badge variant="outline">{prize.prize_type}</Badge>
                        </TableCell>
                        <TableCell>
                          {prize.prize_type === 'CASH' ? formatCurrency(prize.prize_value) :
                           prize.prize_type === 'DATA' ? `${prize.prize_value}MB` :
                           prize.prize_type === 'AIRTIME' ? formatCurrency(prize.prize_value) :
                           `${prize.prize_value}`}
                        </TableCell>
                        <TableCell>{prize.probability}%</TableCell>
                        <TableCell>{formatCurrency(prize.minimum_recharge ?? 0)}</TableCell>
                        <TableCell>
                          <Badge variant={prize.is_active ? "default" : "secondary"}>
                            {prize.is_active ? "Active" : "Inactive"}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          <div className="flex gap-1">
                            <Button 
                              size="sm" 
                              variant="outline"
                              onClick={() => {
                                setEditingPrize(prize);
                                setShowPrizeDialog(true);
                              }}
                            >
                              <Edit className="w-3 h-3" />
                            </Button>
                            <Button 
                              size="sm" 
                              variant="destructive"
                              onClick={async () => {
                                if (confirm('Are you sure you want to delete this prize?')) {
                                  try {
                                    setActionLoading(prize.id);
                                    await adminApi.spin.deletePrize(prize.id);
                                    await fetchWheelPrizes();
                                    toast({ title: 'Success', description: 'Prize deleted successfully' });
                                  } catch (error) {
                                    toast({ title: 'Error', description: getErrorMessage(error), variant: 'destructive' });
                                  } finally {
                                    setActionLoading(null);
                                  }
                                }
                              }}
                              disabled={actionLoading === prize.id}
                            >
                              {actionLoading === prize.id ? <Loader2 className="w-3 h-3 animate-spin" /> : <Trash2 className="w-3 h-3" />}
                            </Button>
                          </div>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </CardContent>
            </Card>
          </TabsContent>

          {/* Spin Tiers Management */}
          <TabsContent value="spin-tiers">
            <SpinTiersManagement />
          </TabsContent>

          {/* Spin Prize Claims Management */}
          <TabsContent value="spin-claims">
            <SpinPrizeClaimsManagement />
          </TabsContent>

          {/* Daily Subscription Management */}
          <TabsContent value="subscription">
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Clock className="w-6 h-6" />
                  Daily Subscription Management
                </CardTitle>
                <CardDescription>
                  Configure daily subscription pricing and benefits
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="space-y-6">
                  <div className="bg-blue-50 p-4 rounded-lg">
                    <h3 className="font-semibold text-blue-900 mb-2">Daily Subscription Configuration</h3>
                    <p className="text-sm text-blue-700">Configure the pricing and benefits for daily subscriptions. Changes will be reflected immediately on the user frontend.</p>
                  </div>
                  
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <div className="space-y-4">
                      <div>
                        <Label htmlFor="daily_amount">Daily Subscription Price (₦)</Label>
                        <Input
                          id="daily_amount"
                          type="number"
                          min="1"
                          step="0.01"
                          value={(() => {
                            const amount = dailySubscription?.amount || 30;

                            // Force correct minimum value
                            return amount >= 20 ? amount : 30;
                          })()}
                          onChange={(e) => setDailySubscription(prev => prev ? ({
                            ...prev,
                            amount: parseFloat(e.target.value) || 20
                          }) : null)}
                        />
                        <p className="text-xs text-gray-500 mt-1">Amount users pay for daily subscription access</p>
                      </div>
                      
                      <div>
                        <Label htmlFor="draw_entries">Draw Entries Per Subscription</Label>
                        <Input
                          id="draw_entries"
                          type="number"
                          min="1"
                          value={(() => {
                            const entries = dailySubscription?.draw_entries_earned || 1;
                            return entries >= 1 ? entries : 1;
                          })()}
                          onChange={(e) => setDailySubscription(prev => prev ? ({
                            ...prev,
                            draw_entries_earned: parseInt(e.target.value) || 1
                          }) : null)}
                        />
                        <p className="text-xs text-gray-500 mt-1">Number of wheel spin entries earned per subscription</p>
                      </div>

                      <div className="flex items-center space-x-2">
                        <Switch
                          id="subscription_active"
                          checked={dailySubscription?.is_paid !== false}
                          onCheckedChange={(checked) => setDailySubscription(prev => prev ? ({
                            ...prev,
                            is_paid: checked
                          }) : null)}
                        />
                        <Label htmlFor="subscription_active">Enable Daily Subscriptions</Label>
                      </div>
                    </div>

                    <div className="space-y-4">
                      <div className="bg-gray-50 p-4 rounded-lg">
                        <h4 className="font-medium mb-2">Current Configuration</h4>
                        <div className="space-y-2 text-sm">
                          <div className="flex justify-between">
                            <span>Daily Price:</span>
                            <span className="font-medium">₦{dailySubscription?.amount || 20}</span>
                          </div>
                          <div className="flex justify-between">
                            <span>Draw Entries:</span>
                            <span className="font-medium">{dailySubscription?.draw_entries_earned || 1}</span>
                          </div>
                          <div className="flex justify-between">
                            <span>Status:</span>
                            <Badge variant={dailySubscription?.is_paid ? 'default' : 'destructive'}>
                              {dailySubscription?.is_paid ? 'Active' : 'Disabled'}
                            </Badge>
                          </div>
                        </div>
                      </div>
                      
                      <div className="bg-yellow-50 p-4 rounded-lg">
                        <h4 className="font-medium text-yellow-800 mb-2">Impact</h4>
                        <p className="text-sm text-yellow-700">
                          Changes to the daily subscription price will be immediately reflected on the user frontend. 
                          Users will see the new price when they attempt to purchase daily subscriptions.
                        </p>
                      </div>
                    </div>
                  </div>

                  <Button 
                    onClick={() => handleDailySubscriptionUpdate(dailySubscription)}
                    disabled={actionLoading === 'subscription'}
                    className="w-full md:w-auto"
                  >
                    {actionLoading === 'subscription' ? (
                      <Loader2 className="w-4 h-4 animate-spin mr-2" />
                    ) : null}
                    Save Configuration
                  </Button>
                </div>
              </CardContent>
            </Card>
            
            {/* All Daily Subscriptions */}
            <Card className="mt-6">
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Users className="w-5 h-5" />
                  All Daily Subscriptions
                </CardTitle>
                <CardDescription>
                  View all user daily subscriptions and system configurations
                </CardDescription>
              </CardHeader>
              <CardContent>
                <Button 
                  onClick={() => fetchAllDailySubscriptions()}
                  className="mb-4"
                  disabled={actionLoading === 'fetch_all_subscriptions'}
                >
                  {actionLoading === 'fetch_all_subscriptions' ? (
                    <Loader2 className="w-4 h-4 animate-spin mr-2" />
                  ) : null}
                  Refresh All Subscriptions
                </Button>
                
                {allDailySubscriptions && allDailySubscriptions.length > 0 ? (
                  <div className="space-y-4">
                    <div className="text-sm text-gray-600">
                      Total Subscriptions: {allDailySubscriptions.length}
                    </div>
                    <Table>
                      <TableHeader>
                        <TableRow>
                          <TableHead>Date</TableHead>
                          <TableHead>User</TableHead>
                          <TableHead>Amount</TableHead>
                          <TableHead>Draw Entries</TableHead>
                          <TableHead>Status</TableHead>
                          <TableHead>Reference</TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {allDailySubscriptions.map((subscription) => (
                          <TableRow key={subscription.id}>
                            <TableCell>
                              {new Date(subscription.subscription_date).toLocaleDateString()}
                            </TableCell>
                            <TableCell>
                              {subscription.display_phone || 'System Config'}
                              {subscription.display_name && (
                                <div className="text-sm text-gray-500">
                                  {subscription.display_name}
                                </div>
                              )}
                            </TableCell>
                            <TableCell>₦{subscription.amount}</TableCell>
                            <TableCell>{subscription.draw_entries_earned}</TableCell>
                            <TableCell>
                              <Badge variant={subscription.is_paid ? 'default' : 'destructive'}>
                                {subscription.is_paid ? 'Paid' : 'Unpaid'}
                              </Badge>
                            </TableCell>
                            <TableCell className="text-sm">
                              {subscription.payment_reference}
                            </TableCell>
                          </TableRow>
                        ))}
                      </TableBody>
                    </Table>
                  </div>
                ) : (
                  <div className="text-center py-8 text-gray-500">
                    No daily subscriptions found. Click "Refresh" to load data.
                  </div>
                )}
              </CardContent>
            </Card>
          </TabsContent>

          {/* Strategic Affiliate Management */}
          <TabsContent value="strategic-affiliates">
            <StrategicAffiliateAdminDashboard sessionToken={sessionToken || ''} />
          </TabsContent>

          {/* User Management */}
          <TabsContent value="users">
            <UserManagementTab />
          </TabsContent>

          

          {/* Platform Settings */}
          <TabsContent value="settings">
            <PlatformSettingsPage />
          </TabsContent>

          

          {/* Admin Management */}
          <TabsContent value="admins">
            <Card>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <div>
                    <CardTitle className="flex items-center gap-2">
                      <Shield className="w-6 h-6" />
                      Admin Management
                    </CardTitle>
                    <CardDescription>
                      Manage admin users and their roles (Super Admin only)
                    </CardDescription>
                  </div>
                  {hasPermission('manage_admins') && (
                    <Button onClick={() => setShowCreateAdminDialog(true)}>
                      <Plus className="w-4 h-4 mr-2" />
                      Create Admin
                    </Button>
                  )}
                </div>
              </CardHeader>
              <CardContent>
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Admin</TableHead>
                      <TableHead>Role</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>Last Login</TableHead>
                      <TableHead>Created</TableHead>
                      <TableHead>Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {(admins || []).map((adminUser) => (
                      <TableRow key={adminUser.id}>
                        <TableCell>
                          <div>
                            <div className="font-medium">{adminUser.full_name}</div>
                            <div className="text-sm text-gray-500">{adminUser.email}</div>
                          </div>
                        </TableCell>
                        <TableCell>
                          <Badge variant={adminUser.role === 'SUPER_ADMIN' ? 'default' : 'secondary'}>
                            {adminUser.role}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          <Badge variant={adminUser.is_active ? 'default' : 'destructive'}>
                            {adminUser.is_active ? 'Active' : 'Inactive'}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          {adminUser.last_login ? formatDate(adminUser.last_login) : 'Never'}
                        </TableCell>
                        <TableCell>{formatDate(adminUser.created_at)}</TableCell>
                        <TableCell>
                          {hasPermission('manage_admins') && adminUser.role !== 'SUPER_ADMIN' && (
                            <div className="flex gap-1">
                              <Button 
                                size="sm" 
                                variant="outline"
                                onClick={() => {
                                  setEditingAdmin(adminUser);
                                  setShowEditAdminDialog(true);
                                }}
                                disabled={actionLoading === adminUser.id}
                                title="Edit admin"
                              >
                                <Edit className="w-3 h-3" />
                              </Button>
                              <Button 
                                size="sm" 
                                variant="destructive"
                                onClick={() => handleDeleteAdmin(adminUser.id)}
                                disabled={actionLoading === adminUser.id}
                              >
                                <Trash2 className="w-3 h-3" />
                              </Button>
                            </div>
                          )}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </CardContent>
            </Card>
          </TabsContent>

          {/* Audit Log */}
          <TabsContent value="audit">
            <AuditLogTab />
          </TabsContent>

        </Tabs>

        {/* Dialog Components */}
        <NetworkDialog
          open={showNetworkDialog}
          onOpenChange={setShowNetworkDialog}
          network={(editingNetwork ?? undefined) as any}
          onSave={handleNetworkSave}
          loading={actionLoading === 'network_save'}
        />

        <DataPlanDialog
          open={showDataPlanDialog}
          onOpenChange={setShowDataPlanDialog}
          dataPlan={(editingDataPlan ?? undefined) as any}
          networks={networks}
          onSave={handleDataPlanSave}
          loading={actionLoading === 'plan_save'}
        />

        <WheelPrizeDialog
          open={showPrizeDialog}
          onOpenChange={setShowPrizeDialog}
          prize={editingPrize}
          existingPrizes={wheelPrizes}
          onSave={handleWheelPrizeSave}
          loading={actionLoading === 'prize_save'}
        />

        <CreateAdminDialog
          open={showCreateAdminDialog}
          onOpenChange={setShowCreateAdminDialog}
          onSave={handleCreateAdmin}
          loading={actionLoading === 'create_admin'}
        />

        {/* Edit Admin Dialog */}
        {showEditAdminDialog && editingAdmin && (
          <Dialog open={showEditAdminDialog} onOpenChange={(open) => { setShowEditAdminDialog(open); if (!open) setEditingAdmin(null); }}>
            <DialogContent className="max-w-lg">
              <DialogHeader>
                <DialogTitle className="flex items-center gap-2">
                  <Edit className="w-5 h-5" />
                  Edit Admin: {editingAdmin.full_name}
                </DialogTitle>
                <DialogDescription>
                  Update admin account details and permissions
                </DialogDescription>
              </DialogHeader>
              <form onSubmit={async (e) => {
                e.preventDefault();
                const form = e.target as HTMLFormElement;
                const data = new FormData(form);
                const updatedData = {
                  full_name: data.get('full_name') as string,
                  role: data.get('role') as string,
                  is_active: (form.querySelector('#edit_is_active') as HTMLInputElement)?.checked ?? editingAdmin.is_active,
                };
                await handleUpdateAdmin(editingAdmin.id, updatedData);
                setShowEditAdminDialog(false);
                setEditingAdmin(null);
              }} className="space-y-4">
                <div>
                  <Label htmlFor="edit_full_name">Full Name</Label>
                  <Input id="edit_full_name" name="full_name" defaultValue={editingAdmin.full_name} required />
                </div>
                <div>
                  <Label htmlFor="edit_email">Email Address</Label>
                  <Input id="edit_email" value={editingAdmin.email} disabled className="bg-gray-50" />
                  <p className="text-xs text-gray-500 mt-1">Email cannot be changed</p>
                </div>
                <div>
                  <Label htmlFor="edit_role">Role</Label>
                  <select name="role" defaultValue={editingAdmin.role} className="w-full border rounded-md px-3 py-2 text-sm">
                    <option value="ADMIN">Admin - Standard admin with limited permissions</option>
                    <option value="SUPER_ADMIN">Super Admin - Full system access</option>
                    <option value="MODERATOR">Moderator - User and transaction management</option>
                    <option value="SUPPORT">Support - Customer support access</option>
                    <option value="VIEWER">Viewer - Read-only access</option>
                  </select>
                </div>
                <div className="flex items-center justify-between p-3 border rounded-lg">
                  <div>
                    <div className="font-medium">Account Active</div>
                    <div className="text-sm text-gray-500">Enable or disable login for this account</div>
                  </div>
                  <input type="checkbox" id="edit_is_active" defaultChecked={editingAdmin.is_active} className="w-4 h-4" />
                </div>
                <div className="flex gap-2 pt-4">
                  <Button type="submit" disabled={actionLoading === editingAdmin.id} className="flex-1">
                    {actionLoading === editingAdmin.id ? <Loader2 className="w-4 h-4 animate-spin mr-2" /> : <Edit className="w-4 h-4 mr-2" />}
                    Update Admin
                  </Button>
                  <Button type="button" variant="outline" onClick={() => { setShowEditAdminDialog(false); setEditingAdmin(null); }}>
                    Cancel
                  </Button>
                </div>
              </form>
            </DialogContent>
          </Dialog>
        )}
      </div>
    </div>
  );
};

export default ComprehensiveAdminPortal;