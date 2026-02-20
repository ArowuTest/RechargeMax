import React, { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Sheet, SheetContent, SheetTrigger } from '@/components/ui/sheet';
import { Badge } from '@/components/ui/badge';
import { useAuthContext } from '@/contexts/AuthContext';
import { displayPhoneNumber } from '@/lib/utils';
import { useNavigate, useLocation } from 'react-router-dom';
import { 
  Menu, 
  User, 
  LogOut, 
  Smartphone, 
  Trophy, 
  Users, 
  Gift,
  Settings,
  Home
} from 'lucide-react';

export const Header: React.FC = () => {
  const { user, isAuthenticated, logout } = useAuthContext();
  const navigate = useNavigate();
  const location = useLocation();
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);

  const currentPage = location.pathname;

  const navigationItems = [
    { id: '/', label: 'Home', icon: Home },
    { id: '/recharge', label: 'Recharge', icon: Smartphone },
    { id: '/subscription', label: 'Daily ₦20', icon: Gift },
    { id: '/draws', label: 'Draws', icon: Trophy },
    { id: '/affiliate', label: 'Affiliate', icon: Users },
    ...(isAuthenticated ? [{ id: '/dashboard', label: 'Dashboard', icon: User }] : [])
  ];

  const handleNavigation = (path: string) => {
    if (path === 'profile') {
      navigate('/profile');
    } else if (path === 'home') {
      navigate('/');
    } else {
      navigate(path);
    }
    setIsMobileMenuOpen(false);
  };

  const handleLogin = () => {
    navigate('/login');
  };

  const handleLogout = () => {
    logout();
    navigate('/');
  };

  return (
    <>
      <header className="sticky top-0 z-40 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
        <div className="container flex h-16 items-center justify-between px-4">
          {/* Logo */}
          <div 
            className="flex items-center gap-2 cursor-pointer" 
            onClick={() => handleNavigation('home')}
          >
            <div className="w-8 h-8 bg-gradient-to-r from-blue-600 to-purple-600 rounded-lg flex items-center justify-center">
              <Smartphone className="w-5 h-5 text-white" />
            </div>
            <div>
              <h1 className="text-xl font-bold bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
                RechargeMax
              </h1>
              <p className="text-xs text-muted-foreground -mt-1">Rewards</p>
            </div>
          </div>

          {/* Desktop Navigation */}
          <nav className="hidden md:flex items-center gap-6">
            {navigationItems.map((item) => {
              const Icon = item.icon;
              return (
                <Button
                  key={item.id}
                  variant={currentPage === item.id ? "default" : "ghost"}
                  onClick={() => handleNavigation(item.id)}
                  className="flex items-center gap-2"
                >
                  <Icon className="w-4 h-4" />
                  {item.label}
                </Button>
              );
            })}
          </nav>

          {/* User Section */}
          <div className="flex items-center gap-3">
            {isAuthenticated ? (
              <div className="hidden md:flex items-center gap-3">
                <div className="text-right">
                  <p className="text-sm font-medium">
                    {user?.full_name || 'User'}
                  </p>
                  <p className="text-xs text-muted-foreground">
                    {displayPhoneNumber(user?.msisdn || '')}
                  </p>
                </div>
                <Badge variant="secondary">{user?.loyalty_tier}</Badge>
                <Button variant="outline" size="sm" onClick={() => handleNavigation('profile')}>
                  <User className="w-4 h-4 mr-2" />
                  Profile
                </Button>
                <Button variant="outline" size="sm" onClick={handleLogout}>
                  <LogOut className="w-4 h-4 mr-2" />
                  Logout
                </Button>
              </div>
            ) : (
              <Button onClick={handleLogin} className="hidden md:flex">
                <User className="w-4 h-4 mr-2" />
                Login
              </Button>
            )}

            {/* Mobile Menu */}
            <Sheet open={isMobileMenuOpen} onOpenChange={setIsMobileMenuOpen}>
              <SheetTrigger asChild>
                <Button variant="outline" size="icon" className="md:hidden">
                  <Menu className="w-5 h-5" />
                </Button>
              </SheetTrigger>
              <SheetContent side="right" className="w-80">
                <div className="flex flex-col h-full">
                  {/* User Info */}
                  {isAuthenticated ? (
                    <div className="border-b pb-4 mb-4">
                      <div className="flex items-center gap-3">
                        <div className="w-12 h-12 bg-gradient-to-r from-blue-600 to-purple-600 rounded-full flex items-center justify-center">
                          <User className="w-6 h-6 text-white" />
                        </div>
                        <div>
                          <p className="font-medium">
                            {user?.full_name || 'User'}
                          </p>
                          <p className="text-sm text-muted-foreground">
                            {displayPhoneNumber(user?.msisdn || '')}
                          </p>
                          <Badge variant="secondary" className="mt-1">
                            {user?.loyalty_tier}
                          </Badge>
                        </div>
                      </div>
                    </div>
                  ) : (
                    <div className="border-b pb-4 mb-4">
                      <Button onClick={handleLogin} className="w-full">
                        <User className="w-4 h-4 mr-2" />
                        Login to Account
                      </Button>
                    </div>
                  )}

                  {/* Navigation */}
                  <nav className="flex-1 space-y-2">
                    {navigationItems.map((item) => {
                      const Icon = item.icon;
                      return (
                        <Button
                          key={item.id}
                          variant={currentPage === item.id ? "default" : "ghost"}
                          onClick={() => handleNavigation(item.id)}
                          className="w-full justify-start"
                        >
                          <Icon className="w-4 h-4 mr-3" />
                          {item.label}
                        </Button>
                      );
                    })}
                  </nav>

                  {/* Profile & Logout */}
                  {isAuthenticated && (
                    <div className="border-t pt-4 space-y-2">
                      <Button 
                        variant="outline" 
                        onClick={() => handleNavigation('profile')}
                        className="w-full"
                      >
                        <User className="w-4 h-4 mr-2" />
                        My Profile
                      </Button>
                      <Button 
                        variant="outline" 
                        onClick={handleLogout}
                        className="w-full"
                      >
                        <LogOut className="w-4 h-4 mr-2" />
                        Logout
                      </Button>
                    </div>
                  )}
                </div>
              </SheetContent>
            </Sheet>
          </div>
        </div>
      </header>
    </>
  );
};