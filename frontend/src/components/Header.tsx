import React, { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { useAuthContext } from '@/contexts/AuthContext';
import { displayPhoneNumber } from '@/lib/utils';
import { useNavigate, useLocation } from 'react-router-dom';
import { motion, AnimatePresence } from 'framer-motion';
import {
  Menu, X, User, LogOut, Smartphone, Trophy,
  Users, Gift, Home, LayoutDashboard, Zap, ChevronRight, Award,
} from 'lucide-react';

const NAV_ITEMS = [
  { path: '/',             label: 'Home',        icon: Home },
  { path: '/recharge',     label: 'Recharge',    icon: Smartphone },
  { path: '/subscription', label: 'Daily ₦20',   icon: Gift },
  { path: '/draws',        label: 'Prize Draws',  icon: Trophy },
  { path: '/winners',      label: 'Winners',      icon: Award },
  { path: '/affiliate',    label: 'Affiliate',    icon: Users },
];

const TIER_STYLES: Record<string, string> = {
  BRONZE:   'tier-bronze',
  SILVER:   'tier-silver',
  GOLD:     'tier-gold',
  PLATINUM: 'tier-platinum',
};

export const Header: React.FC = () => {
  const { user, isAuthenticated, logout } = useAuthContext();
  const navigate = useNavigate();
  const location = useLocation();
  const [menuOpen, setMenuOpen] = useState(false);
  const [scrolled, setScrolled]   = useState(false);

  if (location.pathname.startsWith('/admin')) return null;

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 12);
    window.addEventListener('scroll', onScroll, { passive: true });
    return () => window.removeEventListener('scroll', onScroll);
  }, []);

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

  const tierKey = (user?.loyalty_tier || 'BRONZE').toUpperCase();
  const tierClass = TIER_STYLES[tierKey] || 'tier-bronze';

  return (
    <>
      {/* ── Fixed header ─────────────────────────────────────────────── */}
      <header
        className={`fixed top-0 left-0 right-0 z-50 transition-all duration-300 ${
          scrolled
            ? 'bg-white/80 backdrop-blur-xl shadow-lg shadow-black/5 border-b border-white/60'
            : 'bg-white/95 backdrop-blur-sm border-b border-gray-100/80'
        }`}
      >
        <div className="max-w-screen-xl mx-auto px-4 h-16 flex items-center justify-between gap-4">

          {/* Logo */}
          <motion.button
            onClick={() => go('/')}
            className="flex items-center gap-2.5 flex-shrink-0 focus:outline-none"
            whileHover={{ scale: 1.02 }}
            whileTap={{ scale: 0.97 }}
          >
            <div className="w-9 h-9 rounded-xl gradient-brand flex items-center justify-center shadow-md glow-brand">
              <Zap className="w-5 h-5 text-white" strokeWidth={2.5} />
            </div>
            <div className="leading-tight">
              <span className="block text-[17px] font-extrabold tracking-tight text-gradient-brand" style={{ fontFamily: "'Plus Jakarta Sans', sans-serif" }}>
                RechargeMax
              </span>
              <span className="block text-[9px] text-gray-400 font-semibold -mt-0.5 uppercase tracking-widest">
                Recharge & Win
              </span>
            </div>
          </motion.button>

          {/* Desktop nav */}
          <nav className="hidden md:flex items-center gap-0.5">
            {navItems.map(({ path, label, icon: Icon }) => {
              const active = isActive(path);
              return (
                <motion.button
                  key={path}
                  onClick={() => go(path)}
                  className={`relative flex items-center gap-1.5 px-3.5 py-2 rounded-xl text-sm font-semibold transition-colors duration-150 ${
                    active
                      ? 'text-white nav-pill-active'
                      : 'text-gray-600 hover:text-purple-700 hover:bg-purple-50'
                  }`}
                  whileHover={{ scale: active ? 1 : 1.03 }}
                  whileTap={{ scale: 0.97 }}
                >
                  <Icon className="w-3.5 h-3.5" strokeWidth={active ? 2.5 : 2} />
                  {label}
                  {active && (
                    <motion.span
                      layoutId="nav-underline"
                      className="absolute inset-0 rounded-xl"
                      style={{ zIndex: -1 }}
                    />
                  )}
                </motion.button>
              );
            })}
          </nav>

          {/* Desktop right section */}
          <div className="hidden md:flex items-center gap-2 flex-shrink-0">
            {isAuthenticated ? (
              <>
                {/* Tier badge */}
                {user?.loyalty_tier && (
                  <motion.span
                    className={`text-[10px] font-bold px-2.5 py-1 rounded-full ${tierClass} shadow-sm`}
                    initial={{ scale: 0.8, opacity: 0 }}
                    animate={{ scale: 1, opacity: 1 }}
                    transition={{ type: 'spring', stiffness: 300, damping: 20 }}
                  >
                    {user.loyalty_tier}
                  </motion.span>
                )}

                {/* User info */}
                <div className="text-right">
                  <p className="text-xs font-bold text-gray-800 leading-tight">
                    {user?.full_name || 'User'}
                  </p>
                  <p className="text-[10px] text-gray-400 leading-tight">
                    {displayPhoneNumber(user?.msisdn || '')}
                  </p>
                </div>

                {/* Avatar */}
                <motion.button
                  onClick={() => go('/profile')}
                  className="w-8 h-8 rounded-full gradient-brand flex items-center justify-center text-white shadow-md"
                  whileHover={{ scale: 1.1 }}
                  whileTap={{ scale: 0.9 }}
                >
                  <User className="w-4 h-4" />
                </motion.button>

                {/* Logout */}
                <motion.button
                  onClick={() => { logout(); navigate('/'); }}
                  className="flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-sm text-gray-500 hover:text-red-500 hover:bg-red-50 transition-all font-medium"
                  whileHover={{ scale: 1.02 }}
                  whileTap={{ scale: 0.97 }}
                >
                  <LogOut className="w-3.5 h-3.5" />
                  <span>Logout</span>
                </motion.button>
              </>
            ) : (
              <motion.div whileHover={{ scale: 1.04 }} whileTap={{ scale: 0.97 }}>
                <Button
                  size="sm"
                  onClick={() => go('/login')}
                  className="gradient-brand text-white border-0 shadow-md hover:shadow-lg font-semibold"
                >
                  <User className="w-3.5 h-3.5 mr-1.5" />
                  Login
                </Button>
              </motion.div>
            )}
          </div>

          {/* Mobile right */}
          <div className="flex md:hidden items-center gap-2">
            {isAuthenticated ? (
              <motion.button
                onClick={() => go('/profile')}
                className="w-8 h-8 rounded-full gradient-brand flex items-center justify-center text-white shadow-md"
                whileTap={{ scale: 0.9 }}
              >
                <User className="w-4 h-4" />
              </motion.button>
            ) : (
              <Button size="sm" onClick={() => go('/login')} className="gradient-brand text-white text-xs px-3 h-8 border-0 font-semibold">
                Login
              </Button>
            )}
            <motion.button
              onClick={() => setMenuOpen((v) => !v)}
              className="w-9 h-9 rounded-xl border border-gray-200 flex items-center justify-center text-gray-600 hover:bg-purple-50 hover:border-purple-200 hover:text-purple-600 transition-colors"
              aria-label={menuOpen ? 'Close menu' : 'Open menu'}
              whileTap={{ scale: 0.9 }}
            >
              {menuOpen ? <X className="w-5 h-5" /> : <Menu className="w-5 h-5" />}
            </motion.button>
          </div>
        </div>
      </header>

      {/* ── Mobile overlay ───────────────────────────────────────────── */}
      <AnimatePresence>
        {menuOpen && (
          <motion.div
            className="fixed inset-0 z-40 bg-black/50 backdrop-blur-sm md:hidden"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            onClick={() => setMenuOpen(false)}
          />
        )}
      </AnimatePresence>

      {/* ── Mobile drawer ────────────────────────────────────────────── */}
      <AnimatePresence>
        {menuOpen && (
          <motion.div
            className="fixed top-16 right-0 bottom-0 z-50 w-72 bg-white/95 backdrop-blur-xl shadow-2xl flex flex-col md:hidden border-l border-gray-100"
            initial={{ x: '100%' }}
            animate={{ x: 0 }}
            exit={{ x: '100%' }}
            transition={{ type: 'spring', damping: 26, stiffness: 280 }}
          >
            {/* User banner */}
            <div className="px-5 py-4 border-b" style={{ background: 'linear-gradient(135deg, #faf5ff, #f3e8ff)' }}>
              {isAuthenticated ? (
                <div className="flex items-center gap-3">
                  <div className="w-11 h-11 rounded-full gradient-brand flex items-center justify-center flex-shrink-0 shadow-md">
                    <User className="w-5 h-5 text-white" />
                  </div>
                  <div className="min-w-0">
                    <p className="font-bold text-gray-900 truncate">{user?.full_name || 'User'}</p>
                    <p className="text-xs text-gray-500 truncate">{displayPhoneNumber(user?.msisdn || '')}</p>
                    {user?.loyalty_tier && (
                      <span className={`inline-block mt-1 text-[10px] font-bold px-2 py-0.5 rounded-full ${tierClass}`}>
                        {user.loyalty_tier}
                      </span>
                    )}
                  </div>
                </div>
              ) : (
                <div className="space-y-2">
                  <p className="text-sm text-gray-600 font-semibold">Start winning today!</p>
                  <Button className="w-full gradient-brand text-white border-0 font-semibold" onClick={() => go('/login')}>
                    <User className="w-4 h-4 mr-2" />
                    Login / Register
                  </Button>
                </div>
              )}
            </div>

            {/* Nav links */}
            <nav className="flex-1 overflow-y-auto py-3 px-3 space-y-1">
              {navItems.map(({ path, label, icon: Icon }, i) => {
                const active = isActive(path);
                return (
                  <motion.button
                    key={path}
                    onClick={() => go(path)}
                    className={`w-full flex items-center justify-between px-4 py-3 rounded-xl text-sm font-semibold transition-all ${
                      active ? 'nav-pill-active' : 'text-gray-700 hover:bg-purple-50 hover:text-purple-700'
                    }`}
                    initial={{ opacity: 0, x: 20 }}
                    animate={{ opacity: 1, x: 0 }}
                    transition={{ delay: i * 0.04 }}
                  >
                    <span className="flex items-center gap-3">
                      <Icon className="w-4 h-4" />
                      {label}
                    </span>
                    <ChevronRight className="w-4 h-4 opacity-40" />
                  </motion.button>
                );
              })}
            </nav>

            {/* Footer */}
            {isAuthenticated && (
              <div className="px-4 pb-6 border-t pt-4 space-y-1">
                <button
                  onClick={() => go('/profile')}
                  className="w-full flex items-center gap-3 px-4 py-3 rounded-xl text-sm font-semibold text-gray-700 hover:bg-gray-50 transition-colors"
                >
                  <User className="w-4 h-4" />
                  My Profile
                </button>
                <button
                  onClick={() => { logout(); navigate('/'); setMenuOpen(false); }}
                  className="w-full flex items-center gap-3 px-4 py-3 rounded-xl text-sm font-semibold text-red-500 hover:bg-red-50 transition-colors"
                >
                  <LogOut className="w-4 h-4" />
                  Logout
                </button>
              </div>
            )}
          </motion.div>
        )}
      </AnimatePresence>

      {/* ── Height spacer ───────────────────────────────────────────── */}
      <div className="h-16" />
    </>
  );
};
