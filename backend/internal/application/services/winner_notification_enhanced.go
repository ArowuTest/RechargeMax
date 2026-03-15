package services

import (
	"context"
	"fmt"
	"log"

	"rechargemax/internal/domain/entities"
)

// NotifyWinnerMultiChannel sends winner notification via all channels (SMS, Email, Push, In-App)
func (s *NotificationService) NotifyWinnerMultiChannel(ctx context.Context, winner *entities.Winner) error {
	// 1. Send SMS
	smsMessage := s.composeWinnerSMS(winner)
	err := s.SendSMS(ctx, winner.MSISDN, smsMessage)
	if err != nil {
		log.Printf("Failed to send winner SMS: %v\n", err)
	}
	
	// 2. Send Email (if user has email)
	user, err := s.userRepo.FindByMSISDN(ctx, winner.MSISDN)
	if err == nil && user.Email != "" {
		emailSubject, emailBody := s.composeWinnerEmail(winner)
		err = s.SendEmail(ctx, user.Email, emailSubject, emailBody)
		if err != nil {
			log.Printf("Failed to send winner email: %v\n", err)
		}
	}
	
	// 3. Send Push Notification
	pushTitle, pushBody := s.composeWinnerPush(winner)
	err = s.SendPushNotification(ctx, winner.MSISDN, pushTitle, pushBody)
	if err != nil {
		log.Printf("Failed to send winner push: %v\n", err)
	}
	
	// 4. Create In-Platform Notification
	notifTitle := fmt.Sprintf("🎉 You Won %s!", winner.PrizeDescription)
	notifMessage := s.composeInPlatformMessage(winner)
	err = s.CreateNotification(ctx, winner.MSISDN, "draw_winner", notifTitle, notifMessage, map[string]interface{}{
		"winner_id":   winner.ID.String(),
		"prize_type":  winner.PrizeType,
		"prize_amount": winner.PrizeAmount,
	})
	if err != nil {
		log.Printf("Failed to create in-platform notification: %v\n", err)
	}
	
	return nil
}

// composeWinnerSMS creates SMS message based on prize type
func (s *NotificationService) composeWinnerSMS(winner *entities.Winner) string {
	switch winner.PrizeType {
	case "airtime", "data":
		return fmt.Sprintf("Congratulations! You won %s in the RechargeMax draw! Your prize has been credited to %s. Enjoy! 🎉",
			winner.PrizeDescription, winner.MSISDN)
	
	case "cash":
		deadline := "soon"
		if winner.ClaimDeadline != nil {
			deadline = winner.ClaimDeadline.Format("Jan 02, 2006")
		}
		return fmt.Sprintf("Congratulations! You won ₦%.2f in the RechargeMax draw! Login to RechargeMax to submit your bank details and claim your prize. Claim deadline: %s.",
			float64(*winner.PrizeAmount)/100, deadline)
	
	case "goods", "physical":
		deadline := "soon"
		if winner.ClaimDeadline != nil {
			deadline = winner.ClaimDeadline.Format("Jan 02, 2006")
		}
		return fmt.Sprintf("Congratulations! You won %s in the RechargeMax draw! Login to RechargeMax to submit your shipping address. Claim deadline: %s.",
			winner.PrizeDescription, deadline)
	
	default:
		return fmt.Sprintf("Congratulations! You won %s in the RechargeMax draw! Login for details.",
			winner.PrizeDescription)
	}
}

// composeWinnerEmail creates email subject and HTML body
func (s *NotificationService) composeWinnerEmail(winner *entities.Winner) (string, string) {
	subject := fmt.Sprintf("🎉 Congratulations! You Won %s in the RechargeMax Draw!", winner.PrizeDescription)
	
	deadline := "soon"
	if winner.ClaimDeadline != nil {
		deadline = winner.ClaimDeadline.Format("January 02, 2006")
	}
	
	// Determine prize display
	prizeDisplay := winner.PrizeDescription
	if winner.PrizeType == "cash" && winner.PrizeAmount != nil {
		prizeDisplay = fmt.Sprintf("₦%.2f Cash", float64(*winner.PrizeAmount)/100)
	}
	
	// Determine claim instructions
	claimInstructions := ""
	if winner.PrizeType == "cash" {
		claimInstructions = `
		<h3>How to Claim Your Prize:</h3>
		<ol>
			<li>Login to your RechargeMax account</li>
			<li>Go to "My Wins" section</li>
			<li>Submit your bank account details</li>
			<li>Our team will process your payout within 3-5 business days</li>
		</ol>`
	} else if winner.PrizeType == "goods" || winner.PrizeType == "physical" {
		claimInstructions = `
		<h3>How to Claim Your Prize:</h3>
		<ol>
			<li>Login to your RechargeMax account</li>
			<li>Go to "My Wins" section</li>
			<li>Submit your shipping address and phone number</li>
			<li>We'll ship your prize within 3-5 business days</li>
		</ol>`
	} else {
		claimInstructions = `
		<p style="background: #e8f5e9; padding: 15px; border-radius: 8px;">
			<strong>Good news!</strong> Your prize has been automatically credited to your account. No action needed!
		</p>`
	}
	
	body := fmt.Sprintf(`
<html>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
	<div style="background: linear-gradient(135deg, #0055FF, #0044CC); padding: 40px; text-align: center;">
		<h1 style="color: white; margin: 0;">🎉 YOU WON! 🎉</h1>
	</div>
	
	<div style="padding: 40px; background: white;">
		<h2>Congratulations!</h2>
		
		<p>We're thrilled to inform you that you won <strong>%s</strong> in the RechargeMax Draw!</p>
		
		<div style="background: #f5f5f5; padding: 20px; border-radius: 8px; margin: 20px 0;">
			<h3>Prize Details:</h3>
			<ul style="list-style: none; padding: 0;">
				<li>🏆 <strong>Prize:</strong> %s</li>
				<li>🎯 <strong>Position:</strong> %d</li>
				<li>⏰ <strong>Claim Deadline:</strong> %s</li>
			</ul>
		</div>
		
		%s
		
		<div style="text-align: center; margin: 30px 0;">
			<a href="https://rechargemax.com/wins" 
			   style="background: #0055FF; color: white; padding: 15px 40px; 
					  text-decoration: none; border-radius: 8px; display: inline-block;">
				View My Wins
			</a>
		</div>
		
		<p style="color: #666; font-size: 14px;">
			<strong>Important:</strong> You must claim your prize by %s. 
			Unclaimed prizes will be forfeited after the deadline.
		</p>
		
		<p>If you have any questions, contact our support team at support@rechargemax.com</p>
		
		<p>Congratulations again! 🎊</p>
		
		<p>Best regards,<br>The RechargeMax Team</p>
	</div>
	
	<div style="background: #f5f5f5; padding: 20px; text-align: center; font-size: 12px; color: #666;">
		<p>© 2026 RechargeMax. All rights reserved.</p>
	</div>
</body>
</html>
`, prizeDisplay, prizeDisplay, winner.Position, deadline, claimInstructions, deadline)
	
	return subject, body
}

// composeWinnerPush creates push notification title and body
func (s *NotificationService) composeWinnerPush(winner *entities.Winner) (string, string) {
	title := "🎉 You Won!"
	
	var body string
	if winner.PrizeType == "cash" && winner.PrizeAmount != nil {
		body = fmt.Sprintf("Congratulations! You won ₦%.2f in the RechargeMax draw!", float64(*winner.PrizeAmount)/100)
	} else {
		body = fmt.Sprintf("Congratulations! You won %s in the RechargeMax draw!", winner.PrizeDescription)
	}
	
	return title, body
}

// composeInPlatformMessage creates in-platform notification message
func (s *NotificationService) composeInPlatformMessage(winner *entities.Winner) string {
	deadline := "soon"
	if winner.ClaimDeadline != nil {
		deadline = winner.ClaimDeadline.Format("Jan 02, 2006")
	}
	
	return fmt.Sprintf("Congratulations! You won %d place in the RechargeMax Draw. Claim your prize by %s.",
		winner.Position, deadline)
}

// SendClaimReminder sends reminder notification before deadline
func (s *NotificationService) SendClaimReminder(ctx context.Context, winner *entities.Winner, daysLeft int) error {
	var urgency string
	var emoji string
	
	switch {
	case daysLeft <= 1:
		urgency = "FINAL REMINDER"
		emoji = "⚠️"
	case daysLeft <= 3:
		urgency = "URGENT"
		emoji = "🚨"
	case daysLeft <= 7:
		urgency = "Reminder"
		emoji = "⏰"
	default:
		urgency = "Reminder"
		emoji = "⏰"
	}
	
	prizeDisplay := winner.PrizeDescription
	if winner.PrizeType == "cash" && winner.PrizeAmount != nil {
		prizeDisplay = fmt.Sprintf("₦%.2f", float64(*winner.PrizeAmount)/100)
	}
	
	// SMS
	smsMessage := fmt.Sprintf("%s: %s You have %d days left to claim your %s prize! Login to RechargeMax now.",
		urgency, emoji, daysLeft, prizeDisplay)
	s.SendSMS(ctx, winner.MSISDN, smsMessage)
	
	// Push
	pushTitle := fmt.Sprintf("%s %d days left to claim your prize!", emoji, daysLeft)
	pushBody := fmt.Sprintf("Don't miss out on your %s prize!", prizeDisplay)
	s.SendPushNotification(ctx, winner.MSISDN, pushTitle, pushBody)
	
	// In-platform
	notifTitle := fmt.Sprintf("%s Claim Reminder", emoji)
	notifMessage := fmt.Sprintf("You have %d days left to claim your %s prize. Don't miss the deadline!",
		daysLeft, prizeDisplay)
	s.CreateNotification(ctx, winner.MSISDN, "claim_reminder", notifTitle, notifMessage, map[string]interface{}{
		"winner_id":  winner.ID.String(),
		"days_left":  daysLeft,
	})
	
	return nil
}

// ProcessClaimReminders checks for winners nearing deadline and sends reminders
func (s *NotificationService) ProcessClaimReminders(ctx context.Context) error {
	// Get all unclaimed winners
	// This would typically be called by a cron job
	
	// Find winners with upcoming deadlines (7, 3, 1 days)
	// This is a simplified version - in production, you'd query the database
	// for winners where claim_status = 'pending' and deadline is approaching
	
	log.Println("[CRON] Processing claim reminders...")
	
	// Example: Query would look like:
	// SELECT * FROM winners 
	// WHERE claim_status = 'pending' 
	// AND claim_deadline IS NOT NULL
	// AND claim_deadline > NOW()
	// AND (
	//   DATE_PART('day', claim_deadline - NOW()) IN (7, 3, 1)
	// )
	
	return nil
}
