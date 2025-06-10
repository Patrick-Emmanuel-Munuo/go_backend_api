package helpers

import (
	"crypto/des"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"
)

var BaseDate = time.Date(2025, 5, 5, 0, 0, 0, 0, time.UTC)

// ---------- SQL HELPERS ----------
func Where(cond map[string]interface{}) (string, []interface{}) {
	var parts []string
	var params []interface{}
	for k, v := range cond {
		parts = append(parts, fmt.Sprintf("%s = ?", k))
		params = append(params, v)
	}
	return strings.Join(parts, " AND "), params
}

func WhereOr(cond map[string]interface{}) (string, []interface{}) {
	var parts []string
	var params []interface{}
	for k, v := range cond {
		parts = append(parts, fmt.Sprintf("%s = ?", k))
		params = append(params, v)
	}
	return strings.Join(parts, " OR "), params
}

func JoinFields(fields []string) string {
	var wrapped []string
	for _, f := range fields {
		wrapped = append(wrapped, "`"+f+"`")
	}
	return strings.Join(wrapped, ", ")
}

func Like(like map[string]interface{}) (string, []interface{}) {
	var conditions []string
	var params []interface{}
	for key, val := range like {
		conditions = append(conditions, fmt.Sprintf("%s LIKE ?", key))
		params = append(params, fmt.Sprintf("%%%v%%", val))
	}
	return strings.Join(conditions, " AND "), params
}

func UpdateSet(set map[string]interface{}) string {
	var parts []string
	for key := range set {
		parts = append(parts, key+" = ?")
	}
	return strings.Join(parts, ", ")
}

// ---------- UTILITY HELPERS ----------

func GenerateUniqueID() string {
	b := make([]byte, 12)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", b)
}

func DecToBin(n int, length int) string {
	return fmt.Sprintf("%0*b", length, n)
}

// BinStrToDecimal converts a binary string to a decimal int64
func BinStrToDecimal(binStr string) int64 {
	dec, err := strconv.ParseInt(binStr, 2, 64)
	if err != nil {
		log.Printf("Error converting binary string to decimal: %v", err)
		return 0
	}
	return dec
}
func BinToHex(binaryStr string) string {
	n, _ := strconv.ParseInt(binaryStr, 2, 64)
	return strings.ToUpper(fmt.Sprintf("%X", n))
}

func HexToBytes(hexStr string) ([]byte, error) {
	return hex.DecodeString(hexStr)
}

func BinStrToBytes(binStr string) []byte {
	// Pad to multiple of 8 bits
	for len(binStr)%8 != 0 {
		binStr = "0" + binStr
	}
	bytes := make([]byte, len(binStr)/8)
	for i := 0; i < len(binStr); i += 8 {
		b, _ := strconv.ParseUint(binStr[i:i+8], 2, 8)
		bytes[i/8] = byte(b)
	}
	return bytes
}

func BytesToBinStr(data []byte) string {
	var result strings.Builder
	for _, b := range data {
		result.WriteString(fmt.Sprintf("%08b", b))
	}
	return result.String()
}

// ---------- ENCRYPTION / DECODING HELPERS ----------
func GenerateDecoderKey() map[string]interface{} {
	var (
		key_type          = 21
		supplyGroupCode   = 12345
		tariffIndex       = 3
		keyRevisionNumber = 0
		decoderRefNumber  = uint64(1234567890)
		secretKey         = 12345
		additionalSeed    = 67890 // new part to pad up to 192 bits
	)
	keyTypeBin := DecToBin(key_type, 8)              // 8 bits
	supplyGroupBin := DecToBin(supplyGroupCode, 16)  // 16 bits
	tariffIndexBin := DecToBin(tariffIndex, 8)       // 8 bits
	keyRevisionBin := DecToBin(keyRevisionNumber, 8) // 8 bits
	drnBin := DecToBin(int(decoderRefNumber), 64)    // 64 bits
	secretKeyBin := DecToBin(secretKey, 24)          // 24 bits
	additionalBin := DecToBin(additionalSeed, 64)    // 64 bits — padding for 192 bits
	dataBlock := keyTypeBin + supplyGroupBin + tariffIndexBin + keyRevisionBin + drnBin + secretKeyBin + additionalBin
	if len(dataBlock) != 192 {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("decoder key must be 192 bits, got %d", len(dataBlock)),
		}
	}
	return map[string]interface{}{
		"success": true,
		"message": dataBlock,
	}
}

func CalculateCRC16(data []byte) map[string]interface{} {
	crc := 0xFFFF
	for _, b := range data {
		crc ^= int(b) << 8
		for i := 0; i < 8; i++ {
			if crc&0x8000 != 0 {
				crc = (crc << 1) ^ 0x1021
			} else {
				crc <<= 1
			}
			crc &= 0xFFFF
		}
	}
	return map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("%016b", crc),
	}
}

func DecodeUnits(packedBin string) map[string]interface{} {
	if len(packedBin) != 23 {
		return map[string]interface{}{
			"success": false,
			"message": "binary string must be 23 bits",
		}
	}
	numPart := packedBin[:16]
	decPart := packedBin[16:]

	numVal, err := strconv.ParseInt(numPart, 2, 64)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": "failed to parse integer part: " + err.Error(),
		}
	}
	decVal, err := strconv.ParseInt(decPart, 2, 64)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": "failed to parse decimal part: " + err.Error(),
		}
	}

	value := float64(numVal) + float64(decVal)/100.0

	return map[string]interface{}{
		"success": true,
		"message": value,
	}
}

func Decrypt3DES(data, key []byte) map[string]interface{} {
	if len(data) != 8 {
		return map[string]interface{}{
			"success": false,
			"message": "data must be 8 bytes for 3DES",
		}
	}
	if len(key) != 24 {
		return map[string]interface{}{
			"success": false,
			"message": "key must be 24 bytes for 3DES",
		}
	}
	cipherBlock, err := des.NewTripleDESCipher(key)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}
	decrypted := make([]byte, 8)
	cipherBlock.Decrypt(decrypted, data)
	return map[string]interface{}{
		"success": true,
		"message": decrypted,
	}
}

func TranspositionAndRemoveClassBits(tokenBin string) map[string]interface{} {
	if len(tokenBin) < 66 {
		return map[string]interface{}{
			"success": true,
			"message": map[string]interface{}{
				"data":  "token binary must be at least 66 bits",
				"class": "token binary must be at least 66 bits",
			},
		}
	}
	bits := strings.Split(tokenBin, "")
	pos65 := len(bits) - 1 - 65
	pos64 := len(bits) - 1 - 64
	pos28 := len(bits) - 1 - 28
	pos27 := len(bits) - 1 - 27

	saved65 := bits[pos65]
	saved64 := bits[pos64]

	bits[pos28] = saved65
	bits[pos27] = saved64

	tokenClass := bits[0:2]
	restored := bits[2:]
	return map[string]interface{}{
		"success": true,
		"message": map[string]interface{}{
			"data":  strings.Join(restored, ""),
			"class": strings.Join(tokenClass, ""),
		},
	}
}

// Checks if all characters are digits
func IsAllDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
