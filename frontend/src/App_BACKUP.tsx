import { Toaster } from "@/components/ui/toaster";
import { Toaster as Sonner } from "@/components/ui/sonner";
import { TooltipProvider } from "@/components/ui/tooltip";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
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
                
                {/* Admin Routes */}
                <Route path="/admin/login" element={<AdminLoginPage />} />
                <Route path="/admin/dashboard" element={<ProtectedRoute><AdminDashboardPage /></ProtectedRoute>} />
                <Route path="/admin/comprehensive" element={<ProtectedRoute><ComprehensiveAdminPortal /></ProtectedRoute>} />
                <Route path="/admin/draws" element={<ProtectedRoute><DrawIntegrationDashboard /></ProtectedRoute>} />
                <Route path="/admin/winners" element={<ProtectedRoute><WinnerClaimProcessing /></ProtectedRoute>} />
                <Route path="/admin/wheel-prizes" element={<ProtectedRoute><div>Wheel Prize Management (Coming Soon)</div></ProtectedRoute>} />
                <Route path="/admin/subscriptions" element={<ProtectedRoute><SubscriptionTierManagement /></ProtectedRoute>} />
                <Route path="/admin/pricing" element={<ProtectedRoute><SubscriptionPricingConfig /></ProtectedRoute>} />
                <Route path="/admin/daily-subscriptions" element={<ProtectedRoute><DailySubscriptionMonitoring /></ProtectedRoute>} />
                <Route path="/admin/ussd" element={<ProtectedRoute><USSDRechargeMonitoring /></ProtectedRoute>} />
                <Route path="/admin/affiliates" element={<ProtectedRoute><StrategicAffiliateAdminDashboard /></ProtectedRoute>} />
                <Route path="/admin/csv" element={<ProtectedRoute><DrawCSVManagement /></ProtectedRoute>} />
                <Route path="/admin/monitoring" element={<ProtectedRoute><SystemMonitoringDashboard /></ProtectedRoute>} />
                
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
