import React, { useState, useEffect, useCallback, useRef } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import apiClient from '@/lib/api-client'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useToast } from '@/hooks/useToast'
import { processDailySubscription } from '@/lib/api'
import {
  Phone, CreditCard, Loader2, Trophy, Target, Sparkles,
  Plus, Minus, Layers, XCircle, TrendingUp, Zap, Star,
  CheckCircle, Clock, ChevronRight
} from 'lucide-react'

// ─── Types ───────────────────────────────────────────────────────────────────

interface SubscriptionConfig {
  amount: number
  draw_entries_earned: number
  is_active: boolean
}

interface ActiveLine {
  id: string
  code: string
  entries: number
  daily_amount_ngn: number
  status: string
  next_billing: string
}

interface ActiveLinesData {
  lines: ActiveLine[]
  total_active_lines: number
  total_daily_entries: number
  total_daily_cost_ngn: number
}

// ─── Bundle tiers ─────────────────────────────────────────────────────────────

const BUNDLES = [
  { entries: 1,  label: 'Starter',  color: '#3b82f6', bg: '#eff6ff', border: '#bfdbfe' },
  { entries: 5,  label: 'Plus',     color: '#7c3aed', bg: '#f5f3ff', border: '#ddd6fe', popular: true },
  { entries: 10, label: 'Power',    color: '#d97706', bg: '#fffbeb', border: '#fde68a' },
  { entries: 20, label: 'Champion', color: '#059669', bg: '#ecfdf5', border: '#a7f3d0' },
]

// ─── Main component ───────────────────────────────────────────────────────────

export function DailySubscription() {
  const [config, setConfig]         = useState<SubscriptionConfig>({ amount: 20, draw_entries_earned: 1, is_active: true })
  const [configLoading, setConfigLoading] = useState(true)
  const [activeLines, setActiveLines] = useState<ActiveLinesData | null>(null)
  const [linesLoading, setLinesLoading] = useState(false)
  const [cancellingId, setCancellingId] = useState<string | null>(null)
  const [loading, setLoading]       = useState(false)
  const [phone, setPhone]           = useState('')
  const [entries, setEntries]       = useState(1)
  const [selectedBundle, setSelectedBundle] = useState<number | null>(1)
  const phoneRef = useRef<HTMLInputElement>(null)
  const { toast } = useToast()

  // ── Load config ─────────────────────────────────────────────────────────────
  useEffect(() => {
    apiClient.get<{ success: boolean; config: SubscriptionConfig }>('/subscription/config')
      .then(r => { if (r.data?.success && r.data?.config) setConfig(r.data.config) })
      .catch(() => {})
      .finally(() => setConfigLoading(false))
  }, [])

  // ── Load active lines when phone is complete ─────────────────────────────────
  const fetchActiveLines = useCallback(async (msisdn: string) => {
    setLinesLoading(true)
    try {
      const r = await apiClient.get<{ success: boolean; data: ActiveLinesData }>(
        `/subscription/active-lines?msisdn=${encodeURIComponent(msisdn)}`
      )
      if (r.data?.success) setActiveLines(r.data.data)
      else setActiveLines(null)
    } catch { setActiveLines(null) }
    finally { setLinesLoading(false) }
  }, [])

  useEffect(() => {
    const digits = phone.replace(/\D/g, '')
    if (digits.length === 11) fetchActiveLines(digits)
    else setActiveLines(null)
  }, [phone, fetchActiveLines])

  // ── Handle Paystack return ──────────────────────────────────────────────────
  useEffect(() => {
    const params = new URLSearchParams(window.location.search)
    const status = params.get('status')
    const type   = params.get('type')
    const ref    = params.get('ref')
    if (!status || type !== 'subscription') return
    window.history.replaceState({}, '', '/subscription')

    if (status === 'success') {
      // Restore phone from sessionStorage (saved before Paystack redirect)
      const savedPhone = sessionStorage.getItem('sub_phone') || ''
      if (savedPhone && phoneRef.current) {
        setPhone(savedPhone)
      }
      sessionStorage.removeItem('sub_phone')
      sessionStorage.removeItem('sub_entries')

      toast({ title: '🎉 Payment Successful!', description: 'Activating your subscription line…', duration: 7000 })

      // Poll /subscription/active-lines using the auth cookie (JWT) — no phone param needed.
      // The backend resolves MSISDN from the JWT when no query param is provided.
      let attempts = 0
      const poll = setInterval(async () => {
        attempts++
        try {
          // Try auth-cookie route first (logged-in users)
          const r = await apiClient.get<{ success: boolean; data: ActiveLinesData }>(
            '/subscription/active-lines'
          )
          if (r.data?.success && r.data.data.total_active_lines > 0) {
            clearInterval(poll)
            setActiveLines(r.data.data)
            const d = r.data.data
            toast({
              title: '✅ Subscription Active!',
              description: `You now have ${d.total_daily_entries} entr${d.total_daily_entries === 1 ? 'y' : 'ies'}/day · ₦${d.total_daily_cost_ngn}/day`,
              duration: 10000,
            })
            return
          }
        } catch {
          // If auth-cookie call fails, fall back to phone param
          const digits = (phoneRef.current?.value || savedPhone).replace(/\D/g, '')
          if (digits.length >= 10) {
            try {
              const r2 = await apiClient.get<{ success: boolean; data: ActiveLinesData }>(
                `/subscription/active-lines?msisdn=${digits}`
              )
              if (r2.data?.success && r2.data.data.total_active_lines > 0) {
                clearInterval(poll)
                setActiveLines(r2.data.data)
                const d = r2.data.data
                toast({
                  title: '✅ Subscription Active!',
                  description: `You now have ${d.total_daily_entries} entr${d.total_daily_entries === 1 ? 'y' : 'ies'}/day · ₦${d.total_daily_cost_ngn}/day`,
                  duration: 10000,
                })
              }
            } catch {}
          }
        }
        if (attempts >= 8) clearInterval(poll)
      }, 2000)
    } else {
      toast({ title: 'Payment Failed', description: params.get('error') || 'Please try again.', variant: 'destructive', duration: 8000 })
    }
  }, []) // eslint-disable-line

  // ── Cancel a line ───────────────────────────────────────────────────────────
  const handleCancelLine = async (line: ActiveLine) => {
    if (!confirm(`Cancel subscription line ${line.code}?\n\nThis stops future daily charges for this line only.`)) return
    setCancellingId(line.id)
    try {
      await apiClient.post(`/subscription/cancel/${line.id}`, { msisdn: phone.replace(/\s/g, '') })
      toast({ title: 'Line Cancelled', description: `${line.code} cancelled.` })
      await fetchActiveLines(phone.replace(/\D/g, ''))
    } catch {
      toast({ title: 'Cancel Failed', variant: 'destructive' })
    } finally { setCancellingId(null) }
  }

  // ── Entry stepper helpers ───────────────────────────────────────────────────
  const increment = () => {
    const next = Math.min(100, entries + 1)
    setEntries(next)
    setSelectedBundle(BUNDLES.find(b => b.entries === next)?.entries ?? null)
  }
  const decrement = () => {
    const next = Math.max(1, entries - 1)
    setEntries(next)
    setSelectedBundle(BUNDLES.find(b => b.entries === next)?.entries ?? null)
  }
  const selectBundle = (b: typeof BUNDLES[0]) => {
    setEntries(b.entries)
    setSelectedBundle(b.entries)
  }

  // ── Subscribe ───────────────────────────────────────────────────────────────
  const handleSubscribe = async () => {
    const digits = phone.replace(/\D/g, '')
    if (digits.length < 11) {
      toast({ title: 'Invalid Phone Number', description: 'Please enter a valid 11-digit Nigerian phone number.', variant: 'destructive' })
      return
    }
    if (entries < 1 || entries > 100) {
      toast({ title: 'Invalid Entries', description: 'Choose between 1 and 100 entries.', variant: 'destructive' })
      return
    }
    setLoading(true)
    try {
      const amount = entries * config.amount
      const res = await processDailySubscription({ msisdn: digits, entries, amount, subscription_amount: config.amount })
      if (!res.success) throw new Error(res.error || 'Payment initialisation failed')
      const d = res.data || res
      const url = d.authorization_url || d.payment_url
      if (!url) throw new Error('Payment URL not received from server')
      // Save phone + entries before leaving the page so we can restore them on return
      sessionStorage.setItem('sub_phone', phone)
      sessionStorage.setItem('sub_entries', String(entries))
      window.location.href = url
    } catch (e: any) {
      toast({ title: 'Subscription Failed', description: e.message || 'Unable to start payment.', variant: 'destructive' })
    } finally { setLoading(false) }
  }

  // ── Phone formatter ──────────────────────────────────────────────────────────
  const formatPhone = (v: string) => {
    const d = v.replace(/\D/g, '')
    if (d.length <= 4) return d
    if (d.length <= 7) return `${d.slice(0,4)} ${d.slice(4)}`
    if (d.length <= 11) return `${d.slice(0,4)} ${d.slice(4,7)} ${d.slice(7)}`
    return `${d.slice(0,4)} ${d.slice(4,7)} ${d.slice(7,11)}`
  }

  const price      = config.amount        // ₦20 per entry
  const total      = entries * price
  const hasActive  = activeLines && activeLines.total_active_lines > 0
  const newTotal   = (activeLines?.total_daily_entries ?? 0) + entries

  // ─────────────────────────────────────────────────────────────────────────────
  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 via-purple-50 to-indigo-50 p-4">
      <div className="max-w-2xl mx-auto space-y-6 py-6">

        {/* ── Hero ─────────────────────────────────────────────────────── */}
        <div className="text-center space-y-2">
          <div className="inline-flex items-center gap-2 bg-blue-100 text-blue-700 rounded-full px-4 py-1.5 text-sm font-semibold">
            <Zap className="w-4 h-4" /> Daily Subscription
          </div>
          <h1 className="text-3xl font-extrabold text-gray-900">
            ₦{price}<span className="text-base font-normal text-gray-500"> per entry · per day</span>
          </h1>
          <p className="text-gray-500 text-sm max-w-md mx-auto">
            Choose how many daily draw entries you want. Add more lines anytime — each runs independently.
          </p>
        </div>

        {/* ── Active lines ─────────────────────────────────────────────── */}
        <AnimatePresence>
          {hasActive && (
            <motion.div
              initial={{ opacity: 0, y: -10 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -10 }}
              className="rounded-2xl border-2 border-green-200 bg-green-50 overflow-hidden"
            >
              <div className="flex items-center justify-between px-5 py-3 border-b border-green-200">
                <div className="flex items-center gap-2">
                  <Layers className="w-4 h-4 text-green-700" />
                  <span className="font-bold text-green-900 text-sm">Active Subscription Lines</span>
                </div>
                <span className="text-xs text-green-700 font-semibold bg-green-200 px-2 py-0.5 rounded-full">
                  {activeLines!.total_daily_entries} entries/day · ₦{activeLines!.total_daily_cost_ngn}/day
                </span>
              </div>
              <div className="divide-y divide-green-100">
                {activeLines!.lines.map(line => (
                  <div key={line.id} className="flex items-center justify-between px-5 py-2.5">
                    <div>
                      <span className="font-semibold text-gray-800 text-sm">
                        {line.entries} {line.entries === 1 ? 'entry' : 'entries'}/day
                      </span>
                      <span className="text-gray-400 text-xs ml-2 font-mono">{line.code}</span>
                    </div>
                    <div className="flex items-center gap-3">
                      <span className="text-xs text-gray-400">
                        Next {new Date(line.next_billing).toLocaleDateString('en-NG', { month: 'short', day: 'numeric' })}
                      </span>
                      <button
                        onClick={() => handleCancelLine(line)}
                        disabled={cancellingId === line.id}
                        className="text-red-400 hover:text-red-600 transition-colors"
                        title="Cancel this line"
                      >
                        {cancellingId === line.id
                          ? <Loader2 className="w-4 h-4 animate-spin" />
                          : <XCircle className="w-4 h-4" />}
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            </motion.div>
          )}
        </AnimatePresence>

        {/* ── ENTRY SELECTOR ───────────────────────────────────────────── */}
        <Card className="shadow-sm border border-gray-100">
          <CardHeader className="pb-3">
            <CardTitle className="text-base flex items-center gap-2">
              <Target className="w-4 h-4 text-blue-600" />
              How many daily entries do you want?
            </CardTitle>
            <CardDescription className="text-xs">
              Each entry = one slot in every daily prize draw. More entries = more chances to win.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">

            {/* Bundle tiles */}
            <div className="grid grid-cols-4 gap-2">
              {BUNDLES.map(b => {
                const isSelected = selectedBundle === b.entries
                return (
                  <motion.button
                    key={b.entries}
                    onClick={() => selectBundle(b)}
                    whileTap={{ scale: 0.95 }}
                    className="relative rounded-2xl border-2 p-3 text-left transition-all focus:outline-none"
                    style={{
                      borderColor: isSelected ? b.color : '#e5e7eb',
                      background: isSelected ? b.bg : '#fff',
                      boxShadow: isSelected ? `0 0 0 3px ${b.color}22` : undefined,
                    }}
                  >
                    {b.popular && (
                      <span className="absolute -top-2 left-1/2 -translate-x-1/2 text-white text-[9px] font-bold px-1.5 py-0.5 rounded-full whitespace-nowrap"
                        style={{ background: b.color }}>
                        Popular
                      </span>
                    )}
                    <p className="text-xl font-extrabold text-gray-900">{b.entries}</p>
                    <p className="text-[10px] font-bold uppercase tracking-wide mt-0.5" style={{ color: b.color }}>{b.label}</p>
                    <p className="text-xs text-gray-500 mt-1">₦{b.entries * price}/day</p>
                    {isSelected && (
                      <motion.div
                        layoutId="bundle-check"
                        className="absolute top-2 right-2 w-4 h-4 rounded-full flex items-center justify-center"
                        style={{ background: b.color }}
                      >
                        <CheckCircle className="w-3 h-3 text-white" />
                      </motion.div>
                    )}
                  </motion.button>
                )
              })}
            </div>

            {/* ── Stepper ── */}
            <div className="rounded-2xl border-2 border-dashed border-gray-200 bg-gray-50 p-4">
              <p className="text-xs font-semibold text-gray-500 uppercase tracking-wide mb-3 text-center">
                Or pick a custom number
              </p>
              <div className="flex items-center justify-center gap-4">
                {/* Minus */}
                <motion.button
                  whileTap={{ scale: 0.85 }}
                  onClick={decrement}
                  disabled={entries <= 1}
                  className="w-11 h-11 rounded-full bg-white border-2 border-gray-200 flex items-center justify-center shadow-sm disabled:opacity-30 hover:border-blue-400 hover:text-blue-600 transition-colors"
                >
                  <Minus className="w-5 h-5" />
                </motion.button>

                {/* Counter display */}
                <div className="flex flex-col items-center w-28">
                  <AnimatePresence mode="popLayout">
                    <motion.div
                      key={entries}
                      initial={{ y: -12, opacity: 0 }}
                      animate={{ y: 0, opacity: 1 }}
                      exit={{ y: 12, opacity: 0 }}
                      transition={{ type: 'spring', stiffness: 400, damping: 30 }}
                      className="text-5xl font-extrabold text-gray-900 leading-none text-center"
                    >
                      {entries}
                    </motion.div>
                  </AnimatePresence>
                  <span className="text-xs text-gray-400 mt-1">
                    {entries === 1 ? 'entry' : 'entries'}/day
                  </span>
                  <span className="text-sm font-bold text-blue-600 mt-0.5">
                    ₦{total}/day
                  </span>
                </div>

                {/* Plus */}
                <motion.button
                  whileTap={{ scale: 0.85 }}
                  onClick={increment}
                  disabled={entries >= 100}
                  className="w-11 h-11 rounded-full bg-blue-600 text-white flex items-center justify-center shadow-sm disabled:opacity-30 hover:bg-blue-700 transition-colors"
                >
                  <Plus className="w-5 h-5" />
                </motion.button>
              </div>

              {/* Scrubber bar */}
              <div className="mt-4 px-2">
                <input
                  type="range"
                  min={1}
                  max={100}
                  value={entries}
                  onChange={e => {
                    const v = Number(e.target.value)
                    setEntries(v)
                    setSelectedBundle(BUNDLES.find(b => b.entries === v)?.entries ?? null)
                  }}
                  className="w-full accent-blue-600 cursor-pointer"
                />
                <div className="flex justify-between text-[10px] text-gray-400 mt-1">
                  <span>1 entry · ₦{price}</span>
                  <span>100 entries · ₦{100 * price}</span>
                </div>
              </div>
            </div>

            {/* Animated cost summary */}
            <AnimatePresence mode="wait">
              <motion.div
                key={entries}
                initial={{ scale: 0.97, opacity: 0.6 }}
                animate={{ scale: 1, opacity: 1 }}
                transition={{ type: 'spring', stiffness: 300, damping: 25 }}
                className="rounded-xl p-4 text-white"
                style={{ background: 'linear-gradient(135deg, #2563eb, #7c3aed)' }}
              >
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-blue-200 text-xs">You're subscribing for</p>
                    <p className="text-3xl font-extrabold">
                      {entries} {entries === 1 ? 'entry' : 'entries'}<span className="text-lg font-normal opacity-70">/day</span>
                    </p>
                    <p className="text-blue-200 text-xs mt-0.5">₦{price} × {entries} = ₦{total} first charge</p>
                  </div>
                  <div className="text-right">
                    <p className="text-blue-200 text-xs">Daily cost</p>
                    <p className="text-2xl font-extrabold">₦{total}</p>
                    {hasActive && (
                      <p className="text-blue-200 text-xs mt-0.5">
                        +{entries} → <span className="font-bold text-white">{newTotal} total/day</span>
                      </p>
                    )}
                  </div>
                </div>
              </motion.div>
            </AnimatePresence>
          </CardContent>
        </Card>

        {/* ── SUBSCRIBE FORM ────────────────────────────────────────────── */}
        <Card className="shadow-sm border border-gray-100">
          <CardHeader className="pb-3">
            <CardTitle className="text-base flex items-center gap-2">
              <Phone className="w-4 h-4 text-blue-600" /> Your Phone Number
            </CardTitle>
            <CardDescription className="text-xs">
              This is the number that will be subscribed. Payment is via your debit/credit card — no airtime charged.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <Input
              ref={phoneRef}
              type="tel"
              value={phone}
              onChange={e => setPhone(formatPhone(e.target.value))}
              placeholder="0801 234 5678"
              className="text-lg h-12"
              maxLength={13}
            />

            {/* Loading indicator for active lines */}
            {linesLoading && (
              <p className="text-xs text-gray-400 flex items-center gap-1">
                <Loader2 className="w-3 h-3 animate-spin" /> Checking existing subscriptions…
              </p>
            )}

            <Button
              onClick={handleSubscribe}
              disabled={loading || phone.replace(/\D/g,'').length < 11 || entries < 1}
              className="w-full h-14 text-lg font-bold bg-blue-600 hover:bg-blue-700 rounded-xl"
            >
              {loading
                ? <><Loader2 className="w-5 h-5 mr-2 animate-spin" /> Processing…</>
                : <><CreditCard className="w-5 h-5 mr-2" /> Pay ₦{total} &amp; Subscribe</>}
            </Button>

            <p className="text-center text-xs text-gray-400">
              Renews daily at 08:00 WAT via card. Cancel any line anytime.
              Points &amp; entries awarded only when each day's payment is confirmed.
            </p>
          </CardContent>
        </Card>

        {/* ── HOW IT WORKS ─────────────────────────────────────────────── */}
        <div className="grid grid-cols-2 gap-3">
          {[
            { icon: <Plus className="w-5 h-5" />, color: '#2563eb', title: 'Stack lines', desc: 'Add ₦20/1 entry now, ₦200/10 later — all run in parallel.' },
            { icon: <CheckCircle className="w-5 h-5" />, color: '#059669', title: 'Pay to earn', desc: 'Entries & points only awarded when each day\'s charge succeeds.' },
            { icon: <Clock className="w-5 h-5" />, color: '#d97706', title: 'Auto-retry', desc: 'Failed charges retried at +1h, +3h, +8h before marking failed.' },
            { icon: <XCircle className="w-5 h-5" />, color: '#dc2626', title: 'Cancel anytime', desc: 'Each line is independent — cancel one without affecting others.' },
          ].map((item, i) => (
            <div key={i} className="bg-white rounded-2xl border border-gray-100 p-4 flex gap-3 shadow-sm">
              <div className="w-9 h-9 rounded-full flex items-center justify-center shrink-0 mt-0.5"
                style={{ background: `${item.color}18`, color: item.color }}>
                {item.icon}
              </div>
              <div>
                <p className="font-bold text-gray-800 text-sm">{item.title}</p>
                <p className="text-gray-500 text-xs mt-0.5 leading-relaxed">{item.desc}</p>
              </div>
            </div>
          ))}
        </div>

      </div>
    </div>
  )
}
