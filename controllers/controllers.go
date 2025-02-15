package controllers

import (
	"math/rand"
	"time"
)

// Generate OTP
func GenerateOTP() string {
	// Generate 6-digit random number for OTP
	rand.Seed(time.Now().UnixNano())
	var otpValue string
	otpString := "0123456789" // You can expand this to letters if necessary

	for i := 0; i < 6; i++ {
		otpValue += string(otpString[rand.Intn(len(otpString))])
	}

	return otpValue
}
