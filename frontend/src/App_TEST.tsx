import { Toaster } from "@/components/ui/toaster";
import { Toaster as Sonner } from "@/components/ui/sonner";
import { TooltipProvider } from "@/components/ui/tooltip";
import { QueryClient, QueryClientProvider } from "@tantml:parameter>
import { HashRouter, Routes, Route } from "react-router-dom";
import { AuthProvider } from "./contexts/AuthContext";
import { AdminProvider } from "./contexts/AdminContext";
import { Header } from "./components/Header";
import Index from "./pages/Index";
import NotFound from "./pages/NotFound";
import LoginPage from "./components/pages/LoginPage";
import RechargePage from "./components/pages/RechargePage";
import DrawsPage from "./components/pages/DrawsPage";
import { AffiliatePage } from "./components/pages/AffiliatePage";
import { UserDashboard } from "./components/dashboard/UserDashboard";
import { AffiliateDashboard } from "./components/affiliate/AffiliateDashboard";
import { AdminLoginPage } from "./components/pages/AdminLoginPage";
import { AdminDashboardPage } from "./components/pages/AdminDashboardPage";
import ComprehensiveAdminPortal from "./components/admin/ComprehensiveAdminPortal";
import DrawIntegrationDashboard from "./components/admin/DrawIntegrationDashboard";
import WinnerClaimProcessing from "./components/admin/WinnerClaimProcessing";
import SubscriptionTierManagement from "./components/admin/SubscriptionTierManagement";
import SubscriptionPricingConfig from "./components/admin/SubscriptionPricingConfig";
import DailySubscriptionMonitoring from "./components/admin/DailySubscriptionMonitoring";
import USSDRechargeMonitoring from "./components/admin/USSDRechargeMonitoring";
import StrategicAffiliateAdminDashboard from "./components/admin/StrategicAffiliateAdminDashboard";
import DrawCSVManagement from "./components/admin/DrawCSVManagement";
import SystemMonitoringDashboard from "./components/admin/SystemMonitoringDashboard";
import { ProtectedRoute } from "./components/ProtectedRoute";

const queryClient = new QueryClient();

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <TooltipProvider>
        <AuthProvider>
          <AdminProvider>
            <Toaster />
            <Sonner />
            <HashRouter>
              <Header />
              <Routes>
                <Route path="/" element={<Index />} />
                <Route path="/login" element={<LoginPage />} />
                <Route path="/recharge" element={<RechargePage />} />
                <Route path="/draws" element={<DrawsPage />} />
                <Route path="/dashboard" element={<UserDashboard />} />
                <Route path="/affiliate" element={<AffiliatePage />} />
                <Route path="/affiliate/dashboard" element={<AffiliateDashboard />} />
                
                {/* Admin Routes - TEMPORARILY WITHOUT PROTECTION FOR TESTING */}
                <Route path="/admin/login" element={<AdminLoginPage />} />
                <Route path="/admin/dashboard" element={<AdminDashboardPage />} />
                <Route path="/admin/comprehensive" element={<ComprehensiveAdminPortal />} />
                <Route path="/admin/draws" element={<DrawIntegrationDashboard />} />
                <Route path="/admin/winners" element={<WinnerClaimProcessing />} />
                <Route path="/admin/wheel-prizes" element={<div>Wheel Prize Management (Coming Soon)</div>} />
                <Route path="/admin/subscriptions" element={<SubscriptionTierManagement />} />
                <Route path="/admin/pricing" element={<SubscriptionPricingConfig />} />
                <Route path="/admin/daily-subscriptions" element={<DailySubscriptionMonitoring />} />
                <Route path="/admin/ussd" element={<USSDRechargeMonitoring />} />
                <Route path="/admin/affiliates" element={<StrategicAffiliateAdminDashboard />} />
                <Route path="/admin/csv" element={<DrawCSVManagement />} />
                <Route path="/admin/monitoring" element={<SystemMonitoringDashboard />} />
                
                <Route path="*" element={<NotFound />} />
              </Routes>
            </HashRouter>
          </AdminProvider>
        </AuthProvider>
      </TooltipProvider>
    </QueryClientProvider>
  );
}

export default App;
