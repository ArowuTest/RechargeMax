import React, { useState, useEffect } from 'react'
import apiClient from '@/lib/api-client'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import { useToast } from '@/hooks/useToast'
import { processDailySubscription, validatePhoneNetwork } from '@/lib/api'
import { 
  Calendar, 
  Gift, 
  Star, 
  TrendingUp, 
  Clock, 
  CheckCircle, 
  Phone,
  CreditCard,
  Loader2,
  Trophy,
  Target,
  Sparkles
} from 'lucide-react'

interface SubscriptionFormData {
  phoneNumber: string
  networkProvider: string
  entries: number
  amount: number
}

interface SubscriptionConfig {
  amount: number
  draw_entries_earned: number
  is_active: boolean
}

export function DailySubscription() {
  const [loading, setLoading] = useState(false)
  const [processingPayment, setProcessingPayment] = useState(false)
  const [configLoading, setConfigLoading] = useState(true)
  const [subscriptionConfig, setSubscriptionConfig] = useState<SubscriptionConfig>({
    amount: 20,
    draw_entries_earned: 1,
    is_active: true
  })
  const [formData, setFormData] = useState<SubscriptionFormData>({
    phoneNumber: '',
    networkProvider: '',
    entries: 1,
    amount: 20
  })
  const { toast } = useToast()

  // Fetch subscription configuration on component mount
  useEffect(() => {
    const fetchSubscriptionConfig = async () => {
      try {
        setConfigLoading(true)
        
        // Use Go backend API instead of Supabase
        
        const response = await apiClient.get<{ success: boolean; config: typeof subscriptionConfig }>('/subscription/config')
        
        const result = response.data
        
        if (result.success && result.config) {
          setSubscriptionConfig(result.config)
          // Update form data with new pricing
          setFormData(prev => ({
            ...prev,
            amount: prev.entries * result.config.amount
          }))
        } else {
        }
      } catch (error) {
        console.error('Error fetching subscription config:', error)
      } finally {
        setConfigLoading(false)
      }
    }
    
    fetchSubscriptionConfig()
  }, [])

  // Calculate amount based on entries and current config
  useEffect(() => {
    setFormData(prev => ({
      ...prev,
      amount: prev.entries * subscriptionConfig.amount
    }))
  }, [formData.entries, subscriptionConfig.amount])

  // Handle success status from URL parameters
  useEffect(() => {
    // Extract parameters from URL search string (BrowserRouter uses standard ?key=val)
    const urlParams = new URLSearchParams(window.location.search)
    
    const status = urlParams.get('status')
    const amount = urlParams.get('amount')
    const entries = urlParams.get('entries')
    const totalEntries = urlParams.get('totalEntries')
    const totalPoints = urlParams.get('totalPoints')
    const points = urlParams.get('points')
    const ref = urlParams.get('ref')
    const type = urlParams.get('type')

    // Debug: Log URL parameters

    // Only handle subscription success, not recharge success
    if (status === 'success' && type === 'subscription') {
      const isAdditional = totalEntries && parseInt(totalEntries) > parseInt(entries || '0')
      
      toast({
        title: isAdditional ? "Subscription Added! 🎉" : "Daily Subscription Activated! 🎉",
        description: isAdditional 
          ? `Successfully added ${entries} entries for ₦${amount}! Your total daily entries: ${totalEntries} (${totalPoints} points). Good luck in today's draw!`
          : `Your daily draw subscription is now active! You have ${entries} entries for ₦${amount} and earned ${points} points. Good luck in today's draw!`,
        duration: 10000,
      })
      
      // Clear URL parameters by updating hash
      window.history.replaceState({}, '', '#/subscription')
    } else if (status === 'error') {
      const error = urlParams.get('error')
      toast({
        title: "Subscription Failed",
        description: error || "There was an issue with your subscription. Please try again.",
        variant: "destructive",
        duration: 8000,
      })
      
      // Clear URL parameters by updating hash
      window.history.replaceState({}, '', '#/subscription')
    }
  }, [])

  const handleSubscribe = async () => {
    try {
      // Validate phone number
      if (!formData.phoneNumber || formData.phoneNumber.length < 11) {
        toast({
          title: "Invalid Phone Number",
          description: "Please enter a valid Nigerian phone number",
          variant: "destructive"
        })
        return
      }

      // Validate network selection
      if (!formData.networkProvider) {
        toast({
          title: "Network Required",
          description: "Please select your network provider",
          variant: "destructive"
        })
        return
      }

      // Validate phone number with network
      const networkValidationResult = await validatePhoneNetwork(
        formData.phoneNumber.replace(/\s/g, ''),
        formData.networkProvider
      );
      
      if (!networkValidationResult.success) {
        toast({
          title: "Validation Failed",
          description: "Failed to validate phone number",
          variant: "destructive"
        })
        return
      }
      
      if ((networkValidationResult as any).detectedNetwork !== formData.networkProvider) {
        toast({
          title: "Network Mismatch",
          description: `Phone number ${formData.phoneNumber} belongs to ${(networkValidationResult as any).detectedNetwork}, but you selected ${formData.networkProvider}. Please select the correct network.`,
          variant: "destructive"
        })
        return
      }

      // Validate entries
      if (formData.entries < 1 || formData.entries > 100) {
        toast({
          title: "Invalid Entries",
          description: "Please enter between 1 and 100 entries",
          variant: "destructive"
        })
        return
      }

      setLoading(true)
      setProcessingPayment(true)

      // Initialize subscription payment with Paystack
      const subscriptionData = {

        action: 'INITIALIZE_PAYMENT',
        msisdn: formData.phoneNumber.replace(/\s/g, ''),
        entries: formData.entries,
        amount: formData.amount,
        subscription_amount: subscriptionConfig.amount // Dynamic subscription amount per entry
      };
      
      
      const response = await processDailySubscription(subscriptionData);

      // Remove diagnostic block (was leftover debug code)

      if (!response.success) {
        throw new Error(response.error || 'Subscription payment initialization failed')
      }

      // Backend wraps the DTO inside response.data.
      // Try all known field names in case of future API changes.
      const subData = response.data || response
      const payURL = subData.authorization_url || subData.payment_url

      // Redirect to Paystack payment page
      if (payURL) {
        window.location.href = payURL
      } else {
        throw new Error('Payment URL not received from server')
      }

    } catch (error: any) {
      console.error('Subscription error:', error)
      toast({
        title: "Subscription Failed",
        description: error.message || "Failed to initialize subscription payment",
        variant: "destructive"
      })
    } finally {
      setLoading(false)
      setProcessingPayment(false)
    }
  }

  const formatPhoneNumber = (value: string) => {
    // Remove all non-digits
    const digits = value.replace(/\D/g, '')
    
    // Format as Nigerian number
    if (digits.length <= 4) return digits
    if (digits.length <= 7) return `${digits.slice(0, 4)} ${digits.slice(4)}`
    if (digits.length <= 11) return `${digits.slice(0, 4)} ${digits.slice(4, 7)} ${digits.slice(7)}`
    return `${digits.slice(0, 4)} ${digits.slice(4, 7)} ${digits.slice(7, 11)}`
  }

  const handlePhoneChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const formatted = formatPhoneNumber(e.target.value)
    setFormData(prev => ({ ...prev, phoneNumber: formatted }))
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-purple-50 p-4">
      <div className="max-w-4xl mx-auto space-y-8">
        {/* Hero Section */}
        <div className="text-center space-y-4">
          <div className="flex items-center justify-center gap-2 mb-4">
            <Calendar className="w-12 h-12 text-blue-600" />
          </div>
          <h1 className="text-4xl font-bold text-gray-900">Daily Subscription</h1>
          <p className="text-xl text-gray-600 max-w-2xl mx-auto">
            {configLoading ? 'Loading configuration...' : `Subscribe for guaranteed daily draw entries. Only ₦${subscriptionConfig.amount} per entry!`}
          </p>
          {/* Debug info */}
          <div className="text-xs text-gray-400 mt-2">
            Debug: Config={JSON.stringify(subscriptionConfig)} | Loading={configLoading.toString()}
          </div>
        </div>

        {/* Benefits Cards */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
          <Card className="text-center">
            <CardContent className="p-6">
              <CheckCircle className="w-12 h-12 text-green-600 mx-auto mb-4" />
              <h3 className="text-lg font-semibold mb-2">Guaranteed Entry</h3>
              <p className="text-gray-600">Every subscription gives you confirmed draw entries</p>
            </CardContent>
          </Card>
          
          <Card className="text-center">
            <CardContent className="p-6">
              <Trophy className="w-12 h-12 text-yellow-600 mx-auto mb-4" />
              <h3 className="text-lg font-semibold mb-2">Win Big</h3>
              <p className="text-gray-600">Cash prizes up to ₦500,000 in daily draws</p>
            </CardContent>
          </Card>
          
          <Card className="text-center">
            <CardContent className="p-6">
              <Clock className="w-12 h-12 text-blue-600 mx-auto mb-4" />
              <h3 className="text-lg font-semibold mb-2">Daily Draws</h3>
              <p className="text-gray-600">Multiple draws every day, more chances to win</p>
            </CardContent>
          </Card>
        </div>

        {/* Subscription Form */}
        <Card className="max-w-2xl mx-auto">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Gift className="w-6 h-6" />
              Subscribe Now
            </CardTitle>
            <CardDescription>
              Enter your phone number and select how many entries you want
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            {/* Phone Number Input */}
            <div className="space-y-2">
              <Label htmlFor="phoneNumber" className="flex items-center gap-2">
                <Phone className="w-4 h-4" />
                Phone Number *
              </Label>
              <Input
                id="phoneNumber"
                type="tel"
                value={formData.phoneNumber}
                onChange={handlePhoneChange}
                placeholder="0801 234 5678"
                className="text-lg"
                maxLength={13}
              />
              <p className="text-sm text-gray-500">
                Enter the phone number you want to subscribe with
              </p>
            </div>

            {/* Network Selection */}
            <div className="space-y-2">
              <Label htmlFor="networkProvider" className="flex items-center gap-2">
                <Sparkles className="w-4 h-4" />
                Network Provider *
              </Label>
              <Select 
                value={formData.networkProvider} 
                onValueChange={(value) => setFormData(prev => ({ ...prev, networkProvider: value }))}
              >
                <SelectTrigger className="text-lg">
                  <SelectValue placeholder="Select your network" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="MTN">MTN</SelectItem>
                  <SelectItem value="Airtel">Airtel</SelectItem>
                  <SelectItem value="Glo">Glo</SelectItem>
                  <SelectItem value="9mobile">9mobile</SelectItem>
                </SelectContent>
              </Select>
              <p className="text-sm text-gray-500">
                Select the network provider for your phone number
              </p>
            </div>

            {/* Entries Selection */}
            <div className="space-y-2">
              <Label htmlFor="entries" className="flex items-center gap-2">
                <Target className="w-4 h-4" />
                Number of Entries *
              </Label>
              <Input
                id="entries"
                type="number"
                min="1"
                max="100"
                value={formData.entries}
                onChange={(e) => setFormData(prev => ({ ...prev, entries: parseInt(e.target.value) || 1 }))}
                placeholder="Enter number of entries"
                className="text-lg"
              />
              <p className="text-sm text-gray-500">
                Each entry costs ₦{subscriptionConfig.amount}. Enter 1-100 entries (e.g., 50 entries = ₦{50 * subscriptionConfig.amount})
              </p>
            </div>

            {/* Amount Display */}
            <div className="p-4 bg-blue-50 rounded-lg border border-blue-200">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-gray-600">Total Amount</p>
                  <p className="text-2xl font-bold text-blue-600">₦{formData.amount}</p>
                </div>
                <div className="text-right">
                  <p className="text-sm text-gray-600">Entries</p>
                  <p className="text-2xl font-bold text-green-600">{formData.entries}</p>
                </div>
              </div>
            </div>

            {/* Subscribe Button */}
            <Button 
              onClick={handleSubscribe}
              disabled={loading || !formData.phoneNumber || !formData.networkProvider || formData.entries < 1}
              className="w-full bg-blue-600 hover:bg-blue-700 text-white py-6 text-lg"
            >
              {processingPayment ? (
                <>
                  <Loader2 className="w-5 h-5 mr-2 animate-spin" />
                  Processing Payment...
                </>
              ) : (
                <>
                  <CreditCard className="w-5 h-5 mr-2" />
                  Pay ₦{formData.amount} - Subscribe Now
                </>
              )}
            </Button>

            <div className="text-center text-sm text-gray-500">
              <p>After payment, login with your phone number to view your subscription and entries</p>
            </div>
          </CardContent>
        </Card>

        {/* How It Works */}
        <Card className="max-w-4xl mx-auto">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Sparkles className="w-6 h-6" />
              How It Works
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
              <div className="text-center space-y-2">
                <div className="w-12 h-12 bg-blue-100 rounded-full flex items-center justify-center mx-auto">
                  <span className="text-blue-600 font-bold">1</span>
                </div>
                <h4 className="font-semibold">Enter Phone</h4>
                <p className="text-sm text-gray-600">Enter your phone number</p>
              </div>
              
              <div className="text-center space-y-2">
                <div className="w-12 h-12 bg-blue-100 rounded-full flex items-center justify-center mx-auto">
                  <span className="text-blue-600 font-bold">2</span>
                </div>
                <h4 className="font-semibold">Select Entries</h4>
                <p className="text-sm text-gray-600">Choose how many entries</p>
              </div>
              
              <div className="text-center space-y-2">
                <div className="w-12 h-12 bg-blue-100 rounded-full flex items-center justify-center mx-auto">
                  <span className="text-blue-600 font-bold">3</span>
                </div>
                <h4 className="font-semibold">Pay Securely</h4>
                <p className="text-sm text-gray-600">Complete payment via Paystack</p>
              </div>
              
              <div className="text-center space-y-2">
                <div className="w-12 h-12 bg-blue-100 rounded-full flex items-center justify-center mx-auto">
                  <span className="text-blue-600 font-bold">4</span>
                </div>
                <h4 className="font-semibold">Win Prizes</h4>
                <p className="text-sm text-gray-600">Participate in daily draws</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}