import React, { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { useAuthContext } from '@/contexts/AuthContext';
import { displayPhoneNumber } from '@/lib/utils';
import { useNavigate, useLocation } from 'react-router-dom';
import {
  Menu,
  X,
  User,
  LogOut,
  Smartphone,
  Trophy,
  Users,
  Gift,
  Home,
  LayoutDashboard,
  Zap,
  ChevronRight,
} from 'lucide-react';

const NAV_ITEMS = [
  { path: '/',             label: 'Home',        icon: Home },
  { path: '/recharge',     label: 'Recharge',    icon: Smartphone },
  { path: '/subscription', label: 'Daily ₦20',   icon: Gift },
  { path: '/draws',        label: 'Prize Draws',  icon: Trophy },
  { path: '/affiliate',    label: 'Affiliate',   icon: Users },
];

export const Header: React.FC = () => {
  const { user, isAuthenticated, logout } = useAuthContext();
  const navigate = useNavigate();
  const location = useLocation();
  const [menuOpen, setMenuOpen] = useState(false);
  const [scrolled, setScrolled] = useState(false);

  // Hide on admin routes
  if (location.pathname.startsWith('/admin')) return null;

  // Shadow on scroll
  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 8);
    window.addEventListener('scroll', onScroll, { passive: true });
    return () => window.removeEventListener('scroll', onScroll);
  }, []);

  // Lock body scroll when mobile menu open
  useEffect(() => {
    document.body.style.overflow = menuOpen ? 'hidden' : '';
    return () => { document.body.style.overflow = ''; };
  }, [menuOpen]);

  const go = (path: string) => { navigate(path); setMenuOpen(false); };

  const navItems = [
    ...NAV_ITEMS,
    ...(isAuthenticated ? [{ path: '/dashboard', label: 'Dashboard', icon: LayoutDashboard }] : []),
  ];

  const isActive = (path: string) =>
    path === '/' ? location.pathname === '/' : location.pathname.startsWith(path);

  return (
    <>
      {/* ── Fixed header ───────────────────────────────────────────────── */}
      <header
        className={`fixed top-0 left-0 right-0 z-50 transition-all duration-200 ${
          scrolled
            ? 'bg-white/95 backdrop-blur-md shadow-md border-b border-gray-100'
            : 'bg-white border-b border-gray-100'
        }`}
      >
        <div className="max-w-screen-xl mx-auto px-4 h-16 flex items-center justify-between gap-4">

          {/* Logo */}
          <button
            onClick={() => go('/')}
            className="flex items-center gap-2.5 flex-shrink-0 focus:outline-none"
          >
            <div className="w-9 h-9 rounded-xl bg-gradient-to-br from-blue-600 to-purple-600 flex items-center justify-center shadow-md">
              <Zap className="w-5 h-5 text-white" />
            </div>
            <div className="leading-tight">
              <span className="block text-lg font-extrabold tracking-tight bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
                RechargeMax
              </span>
              <span className="block text-[10px] text-gray-400 font-medium -mt-0.5">
                Recharge & Win
              </span>
            </div>
          </button>

          {/* Desktop nav */}
          <nav className="hidden md:flex items-center gap-1">
            {navItems.map(({ path, label, icon: Icon }) => (
              <button
                key={path}
                onClick={() => go(path)}
                className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm font-medium transition-all duration-150 ${
                  isActive(path)
                    ? 'bg-blue-600 text-white shadow-sm'
                    : 'text-gray-600 hover:text-blue-600 hover:bg-blue-50'
                }`}
              >
                <Icon className="w-3.5 h-3.5" />
                {label}
              </button>
            ))}
          </nav>

          {/* Desktop right section */}
          <div className="hidden md:flex items-center gap-2 flex-shrink-0">
            {isAuthenticated ? (
              <>
                <div className="text-right mr-1">
                  <p className="text-xs font-semibold text-gray-800 leading-tight">
                    {user?.full_name || 'User'}
                  </p>
                  <p className="text-[11px] text-gray-400 leading-tight">
                    {displayPhoneNumber(user?.msisdn || '')}
                  </p>
                </div>
                {user?.loyalty_tier && (
                  <Badge className="bg-gradient-to-r from-yellow-400 to-orange-400 text-white border-0 text-[10px] px-2 py-0.5 font-bold shadow-sm">
                    {user.loyalty_tier}
                  </Badge>
                )}
                <button
                  onClick={() => go('/profile')}
                  className="w-8 h-8 rounded-full bg-gradient-to-br from-blue-500 to-purple-500 flex items-center justify-center text-white hover:scale-105 transition-transform"
                >
                  <User className="w-4 h-4" />
                </button>
                <button
                  onClick={() => { logout(); navigate('/'); }}
                  className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm text-gray-500 hover:text-red-500 hover:bg-red-50 transition-all"
                >
                  <LogOut className="w-3.5 h-3.5" />
                  <span>Logout</span>
                </button>
              </>
            ) : (
              <Button
                size="sm"
                onClick={() => go('/login')}
                className="bg-gradient-to-r from-blue-600 to-purple-600 text-white border-0 shadow-md hover:shadow-lg hover:scale-105 transition-all"
              >
                <User className="w-3.5 h-3.5 mr-1.5" />
                Login
              </Button>
            )}
          </div>

          {/* Mobile: login shortcut + hamburger */}
          <div className="flex md:hidden items-center gap-2">
            {isAuthenticated ? (
              <button
                onClick={() => go('/profile')}
                className="w-8 h-8 rounded-full bg-gradient-to-br from-blue-500 to-purple-500 flex items-center justify-center text-white"
              >
                <User className="w-4 h-4" />
              </button>
            ) : (
              <Button size="sm" onClick={() => go('/login')} className="bg-blue-600 text-white text-xs px-3 h-8">
                Login
              </Button>
            )}
            <button
              onClick={() => setMenuOpen((v) => !v)}
              className="w-9 h-9 rounded-lg border border-gray-200 flex items-center justify-center text-gray-600 hover:bg-gray-50 transition-colors"
              aria-label={menuOpen ? 'Close menu' : 'Open menu'}
            >
              {menuOpen ? <X className="w-5 h-5" /> : <Menu className="w-5 h-5" />}
            </button>
          </div>
        </div>
      </header>

      {/* ── Mobile drawer overlay ────────────────────────────────────────── */}
      {menuOpen && (
        <div
          className="fixed inset-0 z-40 bg-black/40 md:hidden"
          onClick={() => setMenuOpen(false)}
        />
      )}

      {/* ── Mobile drawer panel ──────────────────────────────────────────── */}
      <div
        className={`fixed top-16 right-0 bottom-0 z-50 w-72 bg-white shadow-2xl flex flex-col transition-transform duration-300 md:hidden ${
          menuOpen ? 'translate-x-0' : 'translate-x-full'
        }`}
      >
        {/* User banner */}
        <div className="px-5 py-4 border-b bg-gradient-to-r from-blue-50 to-purple-50">
          {isAuthenticated ? (
            <div className="flex items-center gap-3">
              <div className="w-11 h-11 rounded-full bg-gradient-to-br from-blue-500 to-purple-500 flex items-center justify-center flex-shrink-0">
                <User className="w-5 h-5 text-white" />
              </div>
              <div className="min-w-0">
                <p className="font-semibold text-gray-900 truncate">{user?.full_name || 'User'}</p>
                <p className="text-xs text-gray-500 truncate">{displayPhoneNumber(user?.msisdn || '')}</p>
                {user?.loyalty_tier && (
                  <Badge className="mt-1 bg-gradient-to-r from-yellow-400 to-orange-400 text-white border-0 text-[10px]">
                    {user.loyalty_tier}
                  </Badge>
                )}
              </div>
            </div>
          ) : (
            <div className="space-y-2">
              <p className="text-sm text-gray-600 font-medium">Start winning today!</p>
              <Button
                className="w-full bg-gradient-to-r from-blue-600 to-purple-600 text-white border-0"
                onClick={() => go('/login')}
              >
                <User className="w-4 h-4 mr-2" />
                Login / Register
              </Button>
            </div>
          )}
        </div>

        {/* Nav links */}
        <nav className="flex-1 overflow-y-auto py-3 px-3">
          {navItems.map(({ path, label, icon: Icon }) => (
            <button
              key={path}
              onClick={() => go(path)}
              className={`w-full flex items-center justify-between px-4 py-3 rounded-xl mb-1 text-sm font-medium transition-all ${
                isActive(path)
                  ? 'bg-blue-600 text-white shadow-sm'
                  : 'text-gray-700 hover:bg-blue-50 hover:text-blue-600'
              }`}
            >
              <span className="flex items-center gap-3">
                <Icon className="w-4 h-4" />
                {label}
              </span>
              <ChevronRight className="w-4 h-4 opacity-50" />
            </button>
          ))}
        </nav>

        {/* Footer actions */}
        {isAuthenticated && (
          <div className="px-4 pb-6 border-t pt-4 space-y-2">
            <button
              onClick={() => go('/profile')}
              className="w-full flex items-center gap-3 px-4 py-3 rounded-xl text-sm font-medium text-gray-700 hover:bg-gray-50 transition-colors"
            >
              <User className="w-4 h-4" />
              My Profile
            </button>
            <button
              onClick={() => { logout(); navigate('/'); setMenuOpen(false); }}
              className="w-full flex items-center gap-3 px-4 py-3 rounded-xl text-sm font-medium text-red-500 hover:bg-red-50 transition-colors"
            >
              <LogOut className="w-4 h-4" />
              Logout
            </button>
          </div>
        )}
      </div>

      {/* ── Spacer so content doesn't sit under fixed header ─────────────── */}
      <div className="h-16" />
    </>
  );
};
