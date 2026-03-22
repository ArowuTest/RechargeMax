import React, { useState, useEffect, useCallback, useRef } from 'react'
import apiClient from '@/lib/api-client'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import { useToast } from '@/hooks/useToast'
import { processDailySubscription, validatePhoneNetwork } from '@/lib/api'
import {
  Calendar, Gift, Star, Clock, CheckCircle, Phone,
  CreditCard, Loader2, Trophy, Target, Sparkles, Plus,
  Zap, Layers, XCircle, TrendingUp, ChevronRight
} from 'lucide-react'

// ─── Types ───────────────────────────────────────────────────────────────────

interface SubscriptionConfig {
  amount: number           // naira per entry per day
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
  created_at: string
}

interface ActiveLinesData {
  lines: ActiveLine[]
  total_active_lines: number
  total_daily_entries: number
  total_daily_cost_ngn: number
}

// ─── Bundle quick-picks ───────────────────────────────────────────────────────
// Recalculated dynamically from config.amount, these are just the entry counts.
const BUNDLE_PRESETS = [
  { entries: 1,  label: 'Starter',    description: '1 draw entry/day',    color: 'blue' },
  { entries: 5,  label: 'Plus',       description: '5 draw entries/day',   color: 'purple', popular: true },
  { entries: 10, label: 'Power',      description: '10 draw entries/day',  color: 'amber' },
  { entries: 20, label: 'Champion',   description: '20 draw entries/day',  color: 'green' },
]

// ─── Component ───────────────────────────────────────────────────────────────

export function DailySubscription() {
  const [loading, setLoading] = useState(false)
  const [configLoading, setConfigLoading] = useState(true)
  const [linesLoading, setLinesLoading] = useState(false)
  const [cancellingId, setCancellingId] = useState<string | null>(null)

  const [config, setConfig] = useState<SubscriptionConfig>({ amount: 20, draw_entries_earned: 1, is_active: true })
  const [activeLines, setActiveLines] = useState<ActiveLinesData | null>(null)

  const [phone, setPhone] = useState('')
  const [network, setNetwork] = useState('')
  const [entries, setEntries] = useState(1)
  const [customEntries, setCustomEntries] = useState(false)

  const loadedPhoneRef = useRef(false)
  const { toast } = useToast()

  // ── Load config ─────────────────────────────────────────────────────────────
  useEffect(() => {
    apiClient.get<{ success: boolean; config: SubscriptionConfig }>('/subscription/config')
      .then(r => {
        if (r.data?.success && r.data?.config) {
          setConfig(r.data.config)
        }
      })
      .catch(() => {})
      .finally(() => setConfigLoading(false))
  }, [])

  // ── Load active lines ────────────────────────────────────────────────────────
  const fetchActiveLines = useCallback(async (msisdn: string) => {
    if (!msisdn || msisdn.replace(/\s/g, '').length < 10) return
    setLinesLoading(true)
    try {
      const r = await apiClient.get<{ success: boolean; data: ActiveLinesData }>(
        `/subscription/active-lines?msisdn=${encodeURIComponent(msisdn.replace(/\s/g, ''))}`
      )
      if (r.data?.success) setActiveLines(r.data.data)
    } catch {
      setActiveLines(null)
    } finally {
      setLinesLoading(false)
    }
  }, [])

  // Reload lines whenever phone reaches 11 digits
  useEffect(() => {
    const digits = phone.replace(/\D/g, '')
    if (digits.length === 11) {
      fetchActiveLines(digits)
    } else {
      setActiveLines(null)
    }
  }, [phone, fetchActiveLines])

  // ── Handle return from Paystack payment ────────────────────────────────────
  useEffect(() => {
    const params = new URLSearchParams(window.location.search)
    const status = params.get('status')
    const type   = params.get('type')
    const ref    = params.get('ref')

    if (!status || type !== 'subscription') return
    window.history.replaceState({}, '', '/subscription')

    if (status === 'success') {
      toast({
        title: '🎉 Payment Successful!',
        description: 'Your subscription line is being activated. Entries will appear shortly.',
        duration: 8000,
      })

      // Poll once we know the MSISDN
      if (ref) {
        let attempts = 0
        const poll = setInterval(async () => {
          attempts++
          const digits = phone.replace(/\D/g, '') || ''
          if (digits.length >= 10) {
            try {
              const r = await apiClient.get<{ success: boolean; data: ActiveLinesData }>(
                `/subscription/active-lines?msisdn=${digits}`
              )
              if (r.data?.success && r.data.data.total_active_lines > 0) {
                clearInterval(poll)
                setActiveLines(r.data.data)
                const d = r.data.data
                toast({
                  title: '✅ Subscription Active!',
                  description: `Line added! You now have ${d.total_daily_entries} entr${d.total_daily_entries === 1 ? 'y' : 'ies'}/day (₦${d.total_daily_cost_ngn}/day total).`,
                  duration: 10000,
                })
              }
            } catch {}
          }
          if (attempts >= 5) clearInterval(poll)
        }, 2000)
      }
    } else {
      const err = params.get('error')
      toast({
        title: 'Payment Failed',
        description: err || 'There was an issue with your payment. Please try again.',
        variant: 'destructive',
        duration: 8000,
      })
    }
  }, [])  // eslint-disable-line react-hooks/exhaustive-deps

  // ── Cancel a specific line ──────────────────────────────────────────────────
  const handleCancelLine = async (lineId: string, lineCode: string) => {
    if (!confirm(`Cancel subscription line ${lineCode}?\n\nThis will stop future daily charges for this line.`)) return
    setCancellingId(lineId)
    try {
      await apiClient.post(`/subscription/cancel/${lineId}`, { msisdn: phone.replace(/\s/g, '') })
      toast({ title: 'Line Cancelled', description: `Subscription ${lineCode} has been cancelled.` })
      await fetchActiveLines(phone.replace(/\s/g, ''))
    } catch {
      toast({ title: 'Cancel Failed', description: 'Could not cancel line. Please try again.', variant: 'destructive' })
    } finally {
      setCancellingId(null)
    }
  }

  // ── Subscribe ───────────────────────────────────────────────────────────────
  const handleSubscribe = async () => {
    const digits = phone.replace(/\s/g, '')

    if (digits.length < 11) {
      toast({ title: 'Invalid Phone Number', description: 'Please enter a valid 11-digit Nigerian number.', variant: 'destructive' })
      return
    }
    if (!network) {
      toast({ title: 'Network Required', description: 'Please select your network provider.', variant: 'destructive' })
      return
    }
    if (entries < 1 || entries > 100) {
      toast({ title: 'Invalid Entries', description: 'Choose between 1 and 100 entries.', variant: 'destructive' })
      return
    }

    // Network validation
    try {
      const nv = await validatePhoneNetwork(digits, network)
      if (!nv.success) throw new Error('Validation failed')
      if ((nv as any).detectedNetwork !== network) {
        toast({
          title: 'Network Mismatch',
          description: `${digits} is on ${(nv as any).detectedNetwork}, not ${network}.`,
          variant: 'destructive',
        })
        return
      }
    } catch {
      toast({ title: 'Validation Error', description: 'Could not validate number. Please check and retry.', variant: 'destructive' })
      return
    }

    setLoading(true)
    try {
      const amount = entries * config.amount
      const res = await processDailySubscription({
        action: 'INITIALIZE_PAYMENT',
        msisdn: digits,
        entries,
        amount,
        subscription_amount: config.amount,
      })

      if (!res.success) throw new Error(res.error || 'Payment initialisation failed')

      const d = res.data || res
      const url = d.authorization_url || d.payment_url
      if (!url) throw new Error('Payment URL not received')
      window.location.href = url
    } catch (e: any) {
      toast({ title: 'Subscription Failed', description: e.message || 'Unable to start payment.', variant: 'destructive' })
    } finally {
      setLoading(false)
    }
  }

  // ── Phone formatter ──────────────────────────────────────────────────────────
  const formatPhone = (v: string) => {
    const d = v.replace(/\D/g, '')
    if (d.length <= 4) return d
    if (d.length <= 7) return `${d.slice(0, 4)} ${d.slice(4)}`
    if (d.length <= 11) return `${d.slice(0, 4)} ${d.slice(4, 7)} ${d.slice(7)}`
    return `${d.slice(0, 4)} ${d.slice(4, 7)} ${d.slice(7, 11)}`
  }

  const pricePerEntry = config.amount   // ₦20 (or whatever is configured)
  const totalAmount   = entries * pricePerEntry

  // ── Render ───────────────────────────────────────────────────────────────────
  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 via-purple-50 to-indigo-50 p-4">
      <div className="max-w-4xl mx-auto space-y-8">

        {/* ── Hero ─────────────────────────────────────────────────────── */}
        <div className="text-center space-y-3 pt-4">
          <div className="inline-flex items-center gap-2 bg-blue-100 text-blue-700 rounded-full px-4 py-1 text-sm font-semibold mb-2">
            <Zap className="w-4 h-4" /> Daily Subscription
          </div>
          <h1 className="text-4xl font-extrabold text-gray-900 tracking-tight">
            ₦{pricePerEntry}<span className="text-xl font-normal text-gray-500"> per entry / day</span>
          </h1>
          <p className="text-gray-600 max-w-xl mx-auto">
            Add as many daily entry lines as you want. Each line renews automatically —
            entries and points are only awarded when payment is confirmed.
          </p>
        </div>

        {/* ── Active lines panel ───────────────────────────────────────── */}
        {activeLines && activeLines.total_active_lines > 0 && (
          <Card className="border-2 border-green-200 bg-green-50/60 shadow-sm">
            <CardHeader className="pb-3">
              <CardTitle className="flex items-center gap-2 text-green-800">
                <Layers className="w-5 h-5" />
                Your Active Subscription Lines
              </CardTitle>
              <div className="flex gap-6 mt-1">
                <span className="text-sm text-gray-600">
                  <span className="font-bold text-green-700 text-lg">{activeLines.total_daily_entries}</span> total entries/day
                </span>
                <span className="text-sm text-gray-600">
                  <span className="font-bold text-green-700 text-lg">₦{activeLines.total_daily_cost_ngn}</span> total/day
                </span>
                <span className="text-sm text-gray-500">
                  {activeLines.total_active_lines} line{activeLines.total_active_lines > 1 ? 's' : ''}
                </span>
              </div>
            </CardHeader>
            <CardContent className="space-y-2">
              {activeLines.lines.map(line => (
                <div key={line.id}
                  className="flex items-center justify-between bg-white rounded-xl px-4 py-3 border border-green-100 shadow-xs">
                  <div className="flex items-center gap-3">
                    <div className="w-9 h-9 rounded-full bg-green-100 flex items-center justify-center">
                      <Trophy className="w-4 h-4 text-green-600" />
                    </div>
                    <div>
                      <p className="font-semibold text-gray-800 text-sm">
                        {line.entries} {line.entries === 1 ? 'entry' : 'entries'}/day
                        <span className="text-gray-400 font-normal ml-1">— ₦{line.daily_amount_ngn}/day</span>
                      </p>
                      <p className="text-xs text-gray-400 font-mono">{line.code}</p>
                    </div>
                  </div>
                  <div className="flex items-center gap-3">
                    <span className="text-xs text-gray-500">
                      Next: {new Date(line.next_billing).toLocaleDateString('en-NG', { month: 'short', day: 'numeric' })}
                    </span>
                    <Badge className="bg-emerald-100 text-emerald-700 border-0 text-xs">active</Badge>
                    <button
                      onClick={() => handleCancelLine(line.id, line.code)}
                      disabled={cancellingId === line.id}
                      className="text-red-400 hover:text-red-600 transition-colors p-1 rounded"
                      title="Cancel this line"
                    >
                      {cancellingId === line.id
                        ? <Loader2 className="w-4 h-4 animate-spin" />
                        : <XCircle className="w-4 h-4" />}
                    </button>
                  </div>
                </div>
              ))}

              {/* Totals bar */}
              <div className="flex items-center justify-between bg-green-700 text-white rounded-xl px-4 py-3 mt-2">
                <div className="flex items-center gap-2">
                  <TrendingUp className="w-4 h-4" />
                  <span className="font-bold">Total daily guaranteed</span>
                </div>
                <div className="flex items-center gap-4">
                  <span className="font-bold">{activeLines.total_daily_entries} entries</span>
                  <span className="opacity-75">·</span>
                  <span className="font-bold">₦{activeLines.total_daily_cost_ngn}/day</span>
                </div>
              </div>
            </CardContent>
          </Card>
        )}

        {/* ── Bundle presets ───────────────────────────────────────────── */}
        <div>
          <h2 className="text-lg font-bold text-gray-800 mb-3 flex items-center gap-2">
            <Plus className="w-5 h-5 text-blue-600" />
            {activeLines && activeLines.total_active_lines > 0
              ? 'Add Another Subscription Line'
              : 'Choose Your Daily Entries'}
          </h2>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
            {BUNDLE_PRESETS.map(preset => {
              const amount = preset.entries * pricePerEntry
              const isSelected = !customEntries && entries === preset.entries
              return (
                <button
                  key={preset.entries}
                  onClick={() => { setEntries(preset.entries); setCustomEntries(false) }}
                  className={`relative rounded-2xl border-2 p-4 text-left transition-all ${
                    isSelected
                      ? 'border-blue-500 bg-blue-50 shadow-md'
                      : 'border-gray-200 bg-white hover:border-blue-300 hover:shadow-sm'
                  }`}
                >
                  {preset.popular && (
                    <span className="absolute -top-2 left-1/2 -translate-x-1/2 bg-purple-600 text-white text-xs font-bold px-2 py-0.5 rounded-full">
                      Popular
                    </span>
                  )}
                  <p className="text-2xl font-extrabold text-gray-900">{preset.entries}</p>
                  <p className="text-xs font-semibold text-gray-500 uppercase tracking-wide">{preset.label}</p>
                  <p className="text-sm text-gray-600 mt-1">{preset.description}</p>
                  <p className="text-blue-600 font-bold mt-2">₦{amount}/day</p>
                </button>
              )
            })}
          </div>

          {/* Custom entry count */}
          <button
            onClick={() => setCustomEntries(true)}
            className={`mt-3 w-full rounded-2xl border-2 px-4 py-3 text-sm font-medium transition-all flex items-center justify-between ${
              customEntries ? 'border-blue-500 bg-blue-50' : 'border-dashed border-gray-300 text-gray-500 hover:border-blue-400'
            }`}
          >
            <span className="flex items-center gap-2">
              <Target className="w-4 h-4" /> Enter a custom number of entries (1–100)
            </span>
            <ChevronRight className="w-4 h-4" />
          </button>
          {customEntries && (
            <div className="mt-2 flex items-center gap-3">
              <Input
                type="number"
                min={1}
                max={100}
                value={entries}
                onChange={e => setEntries(Math.min(100, Math.max(1, parseInt(e.target.value) || 1)))}
                className="max-w-[120px] text-lg font-bold text-center"
                autoFocus
              />
              <span className="text-gray-600 text-sm">entries/day = <strong className="text-blue-600">₦{entries * pricePerEntry}/day</strong></span>
            </div>
          )}
        </div>

        {/* ── Subscription form ────────────────────────────────────────── */}
        <Card className="shadow-md border border-gray-100">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Gift className="w-5 h-5 text-blue-600" />
              Subscribe Now
            </CardTitle>
            <CardDescription>
              Each subscription is an independent daily line — you can stack as many as you like.
              Points and entries are awarded only when that day's payment is confirmed.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-5">

            {/* Phone */}
            <div className="space-y-1.5">
              <Label htmlFor="phone" className="flex items-center gap-1.5 text-sm font-medium">
                <Phone className="w-4 h-4" /> Phone Number *
              </Label>
              <Input
                id="phone"
                type="tel"
                value={phone}
                onChange={e => setPhone(formatPhone(e.target.value))}
                placeholder="0801 234 5678"
                className="text-lg"
                maxLength={13}
              />
            </div>

            {/* Network */}
            <div className="space-y-1.5">
              <Label className="flex items-center gap-1.5 text-sm font-medium">
                <Sparkles className="w-4 h-4" /> Network Provider *
              </Label>
              <Select value={network} onValueChange={setNetwork}>
                <SelectTrigger className="text-base">
                  <SelectValue placeholder="Select your network" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="MTN">MTN</SelectItem>
                  <SelectItem value="Airtel">Airtel</SelectItem>
                  <SelectItem value="Glo">Glo</SelectItem>
                  <SelectItem value="9mobile">9mobile</SelectItem>
                </SelectContent>
              </Select>
            </div>

            {/* Order summary */}
            <div className="rounded-xl bg-gradient-to-r from-blue-600 to-purple-600 text-white p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-blue-100 text-sm">Subscribing for</p>
                  <p className="text-3xl font-extrabold">
                    {entries} {entries === 1 ? 'entry' : 'entries'}/day
                  </p>
                  <p className="text-blue-200 text-xs mt-0.5">
                    ₦{pricePerEntry} × {entries} = ₦{totalAmount} first payment
                  </p>
                </div>
                <div className="text-right">
                  <p className="text-blue-100 text-sm">Daily cost</p>
                  <p className="text-2xl font-bold">₦{totalAmount}</p>
                  {activeLines && activeLines.total_active_lines > 0 && (
                    <p className="text-blue-200 text-xs mt-0.5">
                      +{entries} = {activeLines.total_daily_entries + entries} total entries/day
                    </p>
                  )}
                </div>
              </div>
            </div>

            {/* CTA */}
            <Button
              onClick={handleSubscribe}
              disabled={loading || !phone || !network || entries < 1}
              className="w-full py-6 text-lg font-bold bg-blue-600 hover:bg-blue-700"
            >
              {loading
                ? <><Loader2 className="w-5 h-5 mr-2 animate-spin" /> Processing…</>
                : <><CreditCard className="w-5 h-5 mr-2" /> Pay ₦{totalAmount} &amp; Subscribe</>}
            </Button>

            <p className="text-center text-xs text-gray-400">
              Renews automatically at 08:00 WAT daily. Cancel any line anytime.
              Points &amp; entries are only awarded when each day's payment is confirmed.
            </p>
          </CardContent>
        </Card>

        {/* ── How it works ─────────────────────────────────────────────── */}
        <Card className="border border-gray-100 shadow-sm">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-lg">
              <Star className="w-5 h-5 text-amber-500" /> How Multi-Line Subscriptions Work
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {[
                {
                  icon: <Plus className="w-5 h-5 text-blue-600" />,
                  title: 'Stack subscriptions',
                  desc: 'Add N20/1 entry now, N200/10 entries later — they run independently side by side.',
                },
                {
                  icon: <CheckCircle className="w-5 h-5 text-green-600" />,
                  title: 'Points on confirmed payment',
                  desc: 'Each line awards entries & points only after that day\'s charge succeeds.',
                },
                {
                  icon: <Clock className="w-5 h-5 text-orange-500" />,
                  title: 'Auto-retry on failure',
                  desc: 'If a daily charge fails, we retry 3 times (+1h, +3h, +8h) before marking it failed.',
                },
                {
                  icon: <XCircle className="w-5 h-5 text-red-500" />,
                  title: 'Cancel any line anytime',
                  desc: 'Each subscription line is independent — cancel one without affecting the others.',
                },
              ].map((item, i) => (
                <div key={i} className="flex gap-3">
                  <div className="w-9 h-9 rounded-full bg-gray-100 flex items-center justify-center shrink-0">
                    {item.icon}
                  </div>
                  <div>
                    <p className="font-semibold text-gray-800 text-sm">{item.title}</p>
                    <p className="text-gray-500 text-xs mt-0.5">{item.desc}</p>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

      </div>
    </div>
  )
}
