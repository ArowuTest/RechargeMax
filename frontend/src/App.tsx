import { Toaster } from "@/components/ui/toaster";
import { Toaster as Sonner } from "@/components/ui/sonner";
import { TooltipProvider } from "@/components/ui/tooltip";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { BrowserRouter, Routes, Route } from "react-router-dom";
import { AuthProvider } from "./contexts/AuthContext";
import { AdminProvider } from "./contexts/AdminContext";
import { Header } from "./components/Header";
import { lazy, Suspense } from "react";
import { ProtectedRoute } from "./components/ProtectedRoute";

// ── Eager-loaded (public / critical path) ─────────────────────────────────
import Index from "./pages/Index";
import NotFound from "./pages/NotFound";
import LoginPage from "./pages/LoginPage";
import RechargePage from "./pages/RechargePage";
import DrawsPage from "./pages/DrawsPage";
import { AffiliatePage } from "./pages/AffiliatePage";
import DailySubscriptionPage from "./pages/DailySubscriptionPage";
import { UserDashboard } from "./components/dashboard/UserDashboard";
import { AffiliateDashboard } from "./components/affiliate/AffiliateDashboard";
import { AdminLoginPage } from "./pages/AdminLoginPage";

// ── Lazy-loaded (admin — only fetched after /admin/login) ─────────────────
const AdminDashboardPage            = lazy(() => import("./pages/AdminDashboardPage").then(m => ({ default: m.AdminDashboardPage })));
const ComprehensiveAdminPortal      = lazy(() => import("./components/admin/ComprehensiveAdminPortal"));
const DrawIntegrationDashboard      = lazy(() => import("./components/admin/DrawIntegrationDashboard"));
const WinnerClaimProcessing         = lazy(() => import("./components/admin/WinnerClaimProcessing"));
const SubscriptionTierManagement    = lazy(() => import("./components/admin/SubscriptionTierManagement"));
const SubscriptionPricingConfig     = lazy(() => import("./components/admin/SubscriptionPricingConfig"));
const DailySubscriptionMonitoring   = lazy(() => import("./components/admin/DailySubscriptionMonitoring"));
const USSDRechargeMonitoring        = lazy(() => import("./components/admin/USSDRechargeMonitoring"));
const StrategicAffiliateAdminDashboard = lazy(() => import("./components/admin/StrategicAffiliateAdminDashboard"));
const DrawCSVManagement             = lazy(() => import("./components/admin/DrawCSVManagement"));
const SystemMonitoringDashboard     = lazy(() => import("./components/admin/SystemMonitoringDashboard"));
const RechargeMonitoringDashboard   = lazy(() => import("./components/admin/RechargeMonitoringDashboard"));
const CommissionReconciliationDashboard = lazy(() => import("./components/admin/CommissionReconciliationDashboard"));
const FailedProvisionsDashboard     = lazy(() => import("./components/admin/FailedProvisionsDashboard"));
const UnclaimedPrizesDashboard      = lazy(() => import("./components/admin/UnclaimedPrizesDashboard"));
const ValidationStatsDashboard      = lazy(() => import("./components/admin/ValidationStatsDashboard"));
const SpinTiersManagement           = lazy(() => import("./components/admin/SpinTiersManagement"));
const PrizeFulfillmentConfig        = lazy(() => import("./components/admin/PrizeFulfillmentConfig"));

const queryClient = new QueryClient();

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <TooltipProvider>
        <AuthProvider>
          <AdminProvider>
            <Toaster />
            <Sonner />
            <BrowserRouter>
              <Header />
              <Suspense fallback={<div className="flex items-center justify-center h-screen text-muted-foreground">Loading…</div>}>
                <Routes>
                  <Route path="/" element={<Index />} />
                  <Route path="/login" element={<LoginPage />} />
                  <Route path="/recharge" element={<RechargePage />} />
                  <Route path="/draws" element={<DrawsPage />} />
                  <Route path="/dashboard" element={<UserDashboard />} />
                  <Route path="/affiliate" element={<AffiliatePage />} />
                  <Route path="/affiliate/dashboard" element={<AffiliateDashboard />} />
                  <Route path="/subscription" element={<DailySubscriptionPage />} />

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
                  <Route path="/admin/recharge-monitoring" element={<ProtectedRoute><RechargeMonitoringDashboard /></ProtectedRoute>} />
                  <Route path="/admin/commissions" element={<ProtectedRoute><CommissionReconciliationDashboard /></ProtectedRoute>} />
                  <Route path="/admin/failed-provisions" element={<ProtectedRoute><FailedProvisionsDashboard /></ProtectedRoute>} />
                  <Route path="/admin/unclaimed-prizes" element={<ProtectedRoute><UnclaimedPrizesDashboard /></ProtectedRoute>} />
                  <Route path="/admin/spin-tiers" element={<ProtectedRoute><SpinTiersManagement /></ProtectedRoute>} />
                  <Route path="/admin/prize-fulfillment" element={<ProtectedRoute><PrizeFulfillmentConfig /></ProtectedRoute>} />
                  <Route path="/admin/validation-stats" element={<ProtectedRoute><ValidationStatsDashboard /></ProtectedRoute>} />

                  <Route path="*" element={<NotFound />} />
                </Routes>
              </Suspense>
            </BrowserRouter>
          </AdminProvider>
        </AuthProvider>
      </TooltipProvider>
    </QueryClientProvider>
  );
}

export default App;
