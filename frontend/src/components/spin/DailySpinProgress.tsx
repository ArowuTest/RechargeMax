import React, { useState, useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Progress } from '@/components/ui/progress'
import { Badge } from '@/components/ui/badge'
import { Zap, Target, Trophy, TrendingUp } from 'lucide-react'
import { getTierProgress, getSpinTiers } from '@/lib/api'

interface TierProgressData {
  currentAmount: number
  currentTier: string
  currentSpins: number
  nextTier: string | null
  nextTierAmount: number | null
  nextTierSpins: number | null
  progressToNext: number
  spinsAvailable: number
}

interface SpinTier {
  tier_name: string
  min_amount: number
  max_amount: number | null
  spins_awarded: number
}

interface DailySpinProgressProps {
  msisdn: string
  onSpinsUpdate?: (spins: number) => void
}

export function DailySpinProgress({ msisdn, onSpinsUpdate }: DailySpinProgressProps) {
  const [progress, setProgress] = useState<TierProgressData | null>(null)
  const [tiers, setTiers] = useState<SpinTier[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetchProgress = async () => {
    try {
      setLoading(true)
      setError(null)

      const [progressResponse, tiersResponse] = await Promise.all([
        getTierProgress(msisdn),
        getSpinTiers()
      ])

      if ('success' in progressResponse && progressResponse.success) {
        setProgress(progressResponse.data as unknown as TierProgressData)
        onSpinsUpdate?.((progressResponse.data as unknown as TierProgressData).spinsAvailable)
      } else {
        setError('Failed to load progress')
      }

      if ('success' in tiersResponse && tiersResponse.success) {
        setTiers(tiersResponse.data)
      }
    } catch (err) {
      console.error('Error fetching spin progress:', err)
      setError('Failed to load spin progress')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (msisdn) {
      fetchProgress()
    }
  }, [msisdn])

  const getTierColor = (tierName: string) => {
    switch (tierName.toLowerCase()) {
      case 'bronze': return 'bg-amber-600'
      case 'silver': return 'bg-gray-400'
      case 'gold': return 'bg-yellow-500'
      case 'platinum': return 'bg-purple-600'
      case 'diamond': return 'bg-blue-600'
      default: return 'bg-gray-500'
    }
  }

  const getTierIcon = (tierName: string) => {
    switch (tierName.toLowerCase()) {
      case 'bronze': return <Target className="w-4 h-4" />
      case 'silver': return <Zap className="w-4 h-4" />
      case 'gold': return <Trophy className="w-4 h-4" />
      case 'platinum': return <TrendingUp className="w-4 h-4" />
      case 'diamond': return <Trophy className="w-4 h-4" />
      default: return <Target className="w-4 h-4" />
    }
  }

  if (loading) {
    return (
      <Card className="w-full">
        <CardContent className="p-6">
          <div className="animate-pulse space-y-4">
            <div className="h-4 bg-gray-200 rounded w-3/4"></div>
            <div className="h-8 bg-gray-200 rounded"></div>
            <div className="h-4 bg-gray-200 rounded w-1/2"></div>
          </div>
        </CardContent>
      </Card>
    )
  }

  if (error || !progress) {
    return (
      <Card className="w-full">
        <CardContent className="p-6">
          <p className="text-red-500 text-center">{error || 'No progress data available'}</p>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card className="w-full bg-gradient-to-br from-blue-50 to-purple-50 border-2 border-blue-200">
      <CardHeader className="pb-3">
        <CardTitle className="flex items-center justify-between text-lg">
          <span className="flex items-center gap-2">
            <Zap className="w-5 h-5 text-blue-600" />
            Daily Spin Progress
          </span>
          <Badge variant="secondary" className="text-lg px-3 py-1">
            {progress.spinsAvailable} spins
          </Badge>
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Current Tier */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            {getTierIcon(progress.currentTier)}
            <span className="font-semibold">{progress.currentTier} Tier</span>
          </div>
          <Badge className={`${getTierColor(progress.currentTier)} text-white`}>
            {progress.currentSpins} spins
          </Badge>
        </div>

        {/* Current Amount */}
        <div className="text-center py-2">
          <div className="text-2xl font-bold text-blue-600">
            ₦{progress.currentAmount.toLocaleString()}
          </div>
          <div className="text-sm text-gray-600">Today's Total Recharge</div>
        </div>

        {/* Progress to Next Tier */}
        {progress.nextTier && progress.nextTierAmount && (
          <div className="space-y-2">
            <div className="flex justify-between text-sm">
              <span>Progress to {progress.nextTier}</span>
              <span>₦{progress.nextTierAmount.toLocaleString()}</span>
            </div>
            <Progress 
              value={Math.min(progress.progressToNext, 100)} 
              className="h-3"
            />
            <div className="flex justify-between text-xs text-gray-600">
              <span>₦{(progress.nextTierAmount - progress.currentAmount).toLocaleString()} more needed</span>
              <span>+{progress.nextTierSpins! - progress.currentSpins} spins</span>
            </div>
          </div>
        )}

        {/* Tier Overview */}
        <div className="mt-4 p-3 bg-white rounded-lg border">
          <h4 className="font-semibold text-sm mb-2">Spin Tiers</h4>
          <div className="grid grid-cols-2 gap-2 text-xs">
            {tiers.filter(tier => tier.tier_name !== 'No Spin').map((tier) => (
              <div 
                key={tier.tier_name}
                className={`p-2 rounded flex justify-between ${
                  tier.tier_name === progress.currentTier 
                    ? 'bg-blue-100 border-2 border-blue-300' 
                    : 'bg-gray-50'
                }`}
              >
                <span className="font-medium">{tier.tier_name}</span>
                <span>{tier.spins_awarded} spins</span>
              </div>
            ))}
          </div>
        </div>

        {/* Call to Action */}
        {progress.nextTier && (
          <div className="text-center p-3 bg-gradient-to-r from-blue-500 to-purple-600 text-white rounded-lg">
            <div className="text-sm font-semibold">
              Recharge ₦{(progress.nextTierAmount! - progress.currentAmount).toLocaleString()} more
            </div>
            <div className="text-xs opacity-90">
              to unlock {progress.nextTier} tier and get {progress.nextTierSpins} total spins!
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  )
}

export default DailySpinProgress