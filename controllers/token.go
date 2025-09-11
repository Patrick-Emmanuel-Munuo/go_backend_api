package controllers

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"
	"vartrick/helpers"
)

// decript token
func DecriptToken(options map[string]interface{}) map[string]interface{} {
	// Validate token field
	tokenRaw, ok := options["token"]
	if !ok {
		return map[string]interface{}{
			"success": false,
			"message": "Token field is required.",
		}
	}
	tokenStr, ok := tokenRaw.(string)
	if !ok || tokenStr == "" {
		return map[string]interface{}{
			"success": false,
			"message": "Token must be a string.",
		}
	}
	// Remove dashes
	tokenStr = strings.ReplaceAll(tokenStr, "-", "")
	// Validate format: must be 20 digits
	if len(tokenStr) != 20 || !helpers.IsAllDigits(tokenStr) {
		return map[string]interface{}{
			"success": false,
			"message": "Invalid token format. Must be 20 digits (dashes allowed).",
		}
	}
	// Convert token string to integer
	// Convert to *big.Int
	tokenBigInt := new(big.Int)
	_, success := tokenBigInt.SetString(tokenStr, 10)
	if !success {
		return map[string]interface{}{
			"success": false,
			"message": "Failed to parse token as big integer.",
		}
	}
	// Convert to binary with 66 bits
	tokenBin := fmt.Sprintf("%066b", tokenBigInt)
	//tokenBin := helpers.DecToBin(token,)
	// Generate decoder key (already returns bin string)
	keyRes := helpers.GenerateDecoderKey()
	if !keyRes["success"].(bool) {
		return map[string]interface{}{
			"success": false,
			"message": keyRes["message"].(string),
		}
	}
	keyBin := keyRes["message"].(string)
	keyBytes := helpers.BinStrToBytes(keyBin)
	// Perform transposition and extract class bits
	tokRes := helpers.TranspositionAndRemoveClassBits(tokenBin)
	if !tokRes["success"].(bool) {
		return map[string]interface{}{
			"success": false,
			"message": "Failed to extract token blocks",
		}
	}
	tokData := tokRes["message"].(map[string]interface{})
	restored := tokData["data"].(string)
	classBits := tokData["class"].(string)
	if len(restored) < 64 {
		return map[string]interface{}{
			"success": false,
			"message": "Token block must be at least 64 bits.",
		}
	}
	// Decrypt the first 64-bit block (8 bytes)
	encBytes := helpers.BinStrToBytes(restored[:64])
	decRes := helpers.Decrypt3DES(encBytes, keyBytes)
	if !decRes["success"].(bool) {
		return map[string]interface{}{
			"success": false,
			"message": "Decryption failed: " + decRes["message"].(string),
		}
	}
	decBytes := decRes["message"].([]byte)
	// Parse decrypted binary string
	binStr := helpers.BytesToBinStr(decBytes)
	if len(binStr) < 64 {
		return map[string]interface{}{
			"success": false,
			"message": "Decrypted data is less than 64 bits.",
		}
	}
	rndBlock := binStr[0:3]
	tidBlock := binStr[3:25]
	amtBlock := binStr[25:48]
	crcBlock := binStr[48:64]
	//fmt.Println("amtBlock:", amtBlock)
	//fmt.Println("tidBlock:", tidBlock)
	//fmt.Println("binStr:", binStr)
	//crc validation
	dataBin := binStr[:48]
	crcInTokenBin := binStr[48:64]
	// Convert dataBin to hex string, padded to 14 hex chars (48 bits = 12 bytes = 14 hex chars with leading zeros)
	dataHex := helpers.BinToHex(dataBin)
	if len(dataHex) < 14 {
		dataHex = fmt.Sprintf("%014s", dataHex)
	}
	// Convert hex string to bytes
	dataBytes, err := helpers.HexToBytes(dataHex)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": "Failed to convert data hex to bytes: " + err.Error(),
		}
	}
	// Calculate CRC16 of dataBytes
	crcRes := helpers.CalculateCRC16(dataBytes)
	if !crcRes["success"].(bool) {
		return map[string]interface{}{
			"success": false,
			"message": "CRC calculation failed: " + crcRes["message"].(string),
		}
	}
	crcCalcBin := crcRes["message"].(string)
	// Pad CRC binary string to 16 bits if needed
	if len(crcCalcBin) < 16 {
		crcCalcBin = fmt.Sprintf("%016s", crcCalcBin)
	}
	// Validate CRC match
	if crcCalcBin != crcInTokenBin {
		return map[string]interface{}{
			"success": false,
			"message": "CRC mismatch - invalid token data",
		}
	}
	// Time validations
	// Parse timestamp (in minutes since base date)
	tidMinutes, err := strconv.ParseInt(tidBlock, 2, 64)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": "Failed to parse timestamp block: " + err.Error(),
		}
	}
	timeNow := time.Now()
	// Assuming you have parsed tidMinutes from earlier as int64
	// Calculate issue time from BaseDate and tidMinutes
	issueTime := helpers.BaseDate.Add(time.Duration(tidMinutes) * time.Minute)

	// Calculate expiry time (1 year after issue time)
	expiryTime := issueTime.AddDate(1, 0, 0)

	// Check if the token is expired
	if timeNow.After(expiryTime) {
		return map[string]interface{}{
			"success": false,
			"message": "Token issue date has expired",
		}
	}
	// Check if the issue time is invalid (before BaseDate or too far in the future)
	if issueTime.Before(helpers.BaseDate) || issueTime.After(timeNow.Add(24*time.Hour)) {
		return map[string]interface{}{
			"success": false,
			"message": "update Change meter base date for",
		}
	}
	// Decode units block (23 bits)
	unitsRes := helpers.DecodeUnits(amtBlock)
	if !unitsRes["success"].(bool) {
		return map[string]interface{}{
			"success": false,
			"message": "Failed to decode units: " + unitsRes["message"].(string),
		}
	}
	units := unitsRes["message"]
	// Assemble result
	result := map[string]interface{}{
		"crc":                helpers.BinStrToDecimal(crcBlock),
		"class":              helpers.BinStrToDecimal(classBits),
		"identifier_minutes": tidMinutes,
		"units":              units,
		"issued_date":        issueTime.Format(time.RFC3339),
		"expiry_date":        expiryTime.Format(time.RFC3339),
		"base_date":          helpers.BaseDate,
		"random":             helpers.BinStrToDecimal(rndBlock),
		"status":             "Token successfully decrypted and parsed.",
	}
	return map[string]interface{}{
		"success": true,
		"message": result,
	}
}

// encript token
func EncriptToken(options map[string]interface{}) map[string]interface{} {
	// Validate required fields
	amountRaw, ok := options["amount"]
	if !ok {
		return map[string]interface{}{
			"success": false,
			"message": "Amount field is required.",
		}
	}
	// First, assert that amountRaw is a string
	amountStr, ok := amountRaw.(string)
	if !ok {
		return map[string]interface{}{
			"success": false,
			"message": "Amount must be a string.",
		}
	}
	// Then parse the string to float64
	amountNumber, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": "Amount must be a valid number.",
		}
	}
	// Use amountNumber as needed
	amount := amountNumber
	issueTime := time.Now()
	if t, ok := options["issued_time"]; ok {
		if tStr, ok := t.(string); ok {
			parsedTime, err := time.Parse(time.RFC3339, tStr)
			if err == nil {
				issueTime = parsedTime
			}
		}
	}
	// Calculate TID (minutes since base date)
	tidMinutes := int64(issueTime.Sub(helpers.BaseDate).Minutes())
	tidBin := fmt.Sprintf("%022b", tidMinutes)
	// Encode units (amount → binary string of 23 bits)
	amtRes := helpers.EncodeUnits(amount)
	if !amtRes["success"].(bool) {
		return map[string]interface{}{
			"success": false,
			"message": "Failed to encode units: " + amtRes["message"].(string),
		}
	}
	amtBlock := amtRes["message"].(string)
	// Generate 3-bit random block
	randRes := helpers.GenerateRandomBits(3)
	if !randRes["success"].(bool) {
		// handle error here
	}
	randomBits := randRes["message"].(string)
	// Construct binary string: random(3) + tid(22) + amount(23) = 48 bits
	dataBin := randomBits + tidBin + amtBlock
	// Convert to hex for CRC
	dataHex := helpers.BinToHex(dataBin)
	if len(dataHex) < 14 {
		dataHex = fmt.Sprintf("%014s", dataHex)
	}
	dataBytes, err := helpers.HexToBytes(dataHex)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": "Failed to convert data to bytes: " + err.Error(),
		}
	}
	// Calculate CRC16 on first 48 bits
	crcRes := helpers.CalculateCRC16(dataBytes)
	if !crcRes["success"].(bool) {
		return map[string]interface{}{
			"success": false,
			"message": "CRC16 calculation failed: " + crcRes["message"].(string),
		}
	}
	crcBin := crcRes["message"].(string)
	if len(crcBin) < 16 {
		crcBin = fmt.Sprintf("%016s", crcBin)
	}
	// Full binary block (64-bit)
	fullBin := dataBin + crcBin
	// Convert to byte array
	fullBytes := helpers.BinStrToBytes(fullBin)
	// Generate decoder key
	keyRes := helpers.GenerateDecoderKey()
	if !keyRes["success"].(bool) {
		return map[string]interface{}{
			"success": false,
			"message": "Failed to generate decoder key: " + keyRes["message"].(string),
		}
	}
	keyBin := keyRes["message"].(string)
	keyBytes := helpers.BinStrToBytes(keyBin)
	// Encrypt with 3DES
	encRes := helpers.Encrypt3DES(fullBytes, keyBytes)
	if !encRes["success"].(bool) {
		return map[string]interface{}{
			"success": false,
			"message": "Encryption failed: " + encRes["message"].(string),
		}
	}
	encBytes := encRes["message"].([]byte)
	// Convert encrypted bytes to binary string
	encBin := helpers.BytesToBinStr(encBytes)
	if len(encBin) < 64 {
		return map[string]interface{}{
			"success": false,
			"message": "Encrypted binary is less than 64 bits.",
		}
	}

	// Add class bits (assume class provided in options or default to 0)
	classInt := 0
	if c, ok := options["class"]; ok {
		if cInt, ok := c.(int); ok {
			classInt = cInt
		} else if cFloat, ok := c.(float64); ok {
			classInt = int(cFloat)
		}
	}
	classBitsRes := helpers.GenerateClassBits(classInt)
	if !classBitsRes["success"].(bool) {
		return map[string]interface{}{
			"success": false,
			"message": "Failed to generate class bits: " + classBitsRes["message"].(string),
		}
	}
	classBits := classBitsRes["message"].(string)

	// Apply transposition
	transpositionRes := helpers.TranspositionAndAddClassBits(encBin, classBits)
	if !transpositionRes["success"].(bool) {
		return map[string]interface{}{
			"success": false,
			"message": "Failed in transposition: " + transpositionRes["message"].(string),
		}
	}
	finalBin := transpositionRes["message"].(string)
	// Format for display (e.g. spaced groups)
	tokenResponce := helpers.FormatTokenDisplay(finalBin)
	if !tokenResponce["success"].(bool) {
		return map[string]interface{}{
			"success": false,
			"message": "Failed to format token display: " + tokenResponce["message"].(string),
		}
	}
	token := tokenResponce["message"].(string)
	// Return response
	return map[string]interface{}{
		"success": true,
		"message": map[string]interface{}{
			//"token":         tokenStr,
			"token":            token,
			"issued_date":      issueTime.Format(time.RFC3339),
			"expired_datetime": issueTime.AddDate(1, 0, 0).Format(time.RFC3339),
			"identifier":       tidMinutes,
			"units":            amount,
			"unitsDecoded":     helpers.DecodeUnits(amtBlock)["message"],
			"random_bits":      randomBits,
			"class_bits":       classBits,
			"crc_block":        crcBin,
		},
	}
}
