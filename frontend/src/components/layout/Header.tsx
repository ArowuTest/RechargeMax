import React, { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { useAuthContext } from '@/contexts/AuthContext';
import { LoginModal } from '@/components/auth/LoginModal';
import { 
  Zap, 
  User, 
  LogOut, 
  Menu, 
  X, 
  Home, 
  Smartphone, 
  Trophy, 
  Users, 
  Calendar,
  BarChart3,
  Settings
} from 'lucide-react';

export const Header: React.FC = () => {
  const { isAuthenticated, user, logout } = useAuthContext();
  const [showLoginModal, setShowLoginModal] = useState(false);
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);

  const handleLogout = () => {
    logout();
    window.location.href = '/#/';
  };

  const navigationItems = [
    { href: '/#/', label: 'Home', icon: Home },
    { href: '/#/recharge', label: 'Recharge', icon: Smartphone },
    { href: '/#/draws', label: 'Draws', icon: Trophy },
    { href: '/#/subscription', label: 'Subscribe', icon: Calendar },
    { href: '/#/affiliate', label: 'Affiliate', icon: Users },
  ];

  const userMenuItems = isAuthenticated ? [
    { href: '/#/dashboard', label: 'Dashboard', icon: BarChart3 },
    { href: '/#/profile', label: 'Profile', icon: Settings },
  ] : [];

  return (
    <>
      <header className="sticky top-0 z-50 w-full border-b bg-white/95 backdrop-blur supports-[backdrop-filter]:bg-white/60">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex h-16 items-center justify-between">
            {/* Logo */}
            <div className="flex items-center gap-2">
              <div className="p-2 bg-blue-600 rounded-lg">
                <Zap className="w-6 h-6 text-white" />
              </div>
              <h1 className="text-2xl font-bold text-gray-900">
                Recharge<span className="text-blue-600">Max</span>
              </h1>
            </div>

            {/* Desktop Navigation */}
            <nav className="hidden md:flex items-center space-x-1">
              {navigationItems.map((item) => (
                <Button
                  key={item.href}
                  variant="ghost"
                  className="text-gray-600 hover:text-gray-900"
                  onClick={() => window.location.href = item.href}
                >
                  <item.icon className="w-4 h-4 mr-2" />
                  {item.label}
                </Button>
              ))}
            </nav>

            {/* User Section */}
            <div className="flex items-center gap-4">
              {isAuthenticated ? (
                <div className="hidden md:flex items-center gap-4">
                  {/* User Info */}
                  <div className="text-right">
                    <p className="text-sm font-medium text-gray-900">
                      {user?.full_name || 'User'}
                    </p>
                    <div className="flex items-center gap-2">
                      <Badge variant="secondary" className="text-xs">
                        {user?.loyalty_tier || 'Bronze'}
                      </Badge>
                      <span className="text-xs text-gray-500">
                        {user?.total_points || 0} pts
                      </span>
                    </div>
                  </div>

                  {/* User Menu */}
                  <div className="flex items-center gap-2">
                    {userMenuItems.map((item) => (
                      <Button
                        key={item.href}
                        variant="ghost"
                        size="sm"
                        onClick={() => window.location.href = item.href}
                      >
                        <item.icon className="w-4 h-4" />
                      </Button>
                    ))}
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={handleLogout}
                      className="text-red-600 hover:text-red-700"
                    >
                      <LogOut className="w-4 h-4" />
                    </Button>
                  </div>
                </div>
              ) : (
                <div className="hidden md:flex items-center gap-2">
                  <Button
                    variant="ghost"
                    onClick={() => setShowLoginModal(true)}
                  >
                    <User className="w-4 h-4 mr-2" />
                    Login
                  </Button>
                  <Button
                    onClick={() => window.location.href = '/#/recharge'}
                    className="bg-blue-600 hover:bg-blue-700"
                  >
                    <Smartphone className="w-4 h-4 mr-2" />
                    Recharge Now
                  </Button>
                </div>
              )}

              {/* Mobile Menu Button */}
              <Button
                variant="ghost"
                size="sm"
                className="md:hidden"
                onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
              >
                {isMobileMenuOpen ? (
                  <X className="w-5 h-5" />
                ) : (
                  <Menu className="w-5 h-5" />
                )}
              </Button>
            </div>
          </div>
        </div>

        {/* Mobile Menu */}
        {isMobileMenuOpen && (
          <div className="md:hidden border-t bg-white">
            <div className="px-4 py-4 space-y-2">
              {/* Navigation Items */}
              {navigationItems.map((item) => (
                <Button
                  key={item.href}
                  variant="ghost"
                  className="w-full justify-start text-gray-600 hover:text-gray-900"
                  onClick={() => {
                    window.location.href = item.href;
                    setIsMobileMenuOpen(false);
                  }}
                >
                  <item.icon className="w-4 h-4 mr-2" />
                  {item.label}
                </Button>
              ))}

              {/* User Section */}
              {isAuthenticated ? (
                <div className="pt-4 border-t space-y-2">
                  <div className="px-3 py-2">
                    <p className="font-medium text-gray-900">
                      {user?.full_name || 'User'}
                    </p>
                    <div className="flex items-center gap-2 mt-1">
                      <Badge variant="secondary" className="text-xs">
                        {user?.loyalty_tier || 'Bronze'}
                      </Badge>
                      <span className="text-xs text-gray-500">
                        {user?.total_points || 0} points
                      </span>
                    </div>
                  </div>

                  {userMenuItems.map((item) => (
                    <Button
                      key={item.href}
                      variant="ghost"
                      className="w-full justify-start"
                      onClick={() => {
                        window.location.href = item.href;
                        setIsMobileMenuOpen(false);
                      }}
                    >
                      <item.icon className="w-4 h-4 mr-2" />
                      {item.label}
                    </Button>
                  ))}

                  <Button
                    variant="ghost"
                    className="w-full justify-start text-red-600 hover:text-red-700"
                    onClick={() => {
                      handleLogout();
                      setIsMobileMenuOpen(false);
                    }}
                  >
                    <LogOut className="w-4 h-4 mr-2" />
                    Logout
                  </Button>
                </div>
              ) : (
                <div className="pt-4 border-t space-y-2">
                  <Button
                    variant="ghost"
                    className="w-full justify-start"
                    onClick={() => {
                      setShowLoginModal(true);
                      setIsMobileMenuOpen(false);
                    }}
                  >
                    <User className="w-4 h-4 mr-2" />
                    Login
                  </Button>
                  <Button
                    className="w-full bg-blue-600 hover:bg-blue-700"
                    onClick={() => {
                      window.location.href = '/#/recharge';
                      setIsMobileMenuOpen(false);
                    }}
                  >
                    <Smartphone className="w-4 h-4 mr-2" />
                    Recharge Now
                  </Button>
                </div>
              )}
            </div>
          </div>
        )}
      </header>

      {/* Login Modal */}
      <LoginModal 
        isOpen={showLoginModal}
        onClose={() => setShowLoginModal(false)}
        onSuccess={() => {
          setShowLoginModal(false);
          // Refresh to show updated user state
          window.location.reload();
        }}
      />
    </>
  );
};

export default Header;