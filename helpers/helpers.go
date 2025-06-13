package helpers

import (
	"crypto/des"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	BaseDate         = time.Date(2025, 5, 5, 0, 0, 0, 0, time.UTC)
	DatabaseHost     string
	DatabaseUser     string
	DatabasePassword string
	DatabaseName     string
	Mailsender       string
	Mailhost         string
	Mailusername     string
	Mailpassword     string
	Mailport         int
)

func UpdateEnvVars() {
	DatabaseHost = os.Getenv("DATABASE_HOST")
	DatabaseUser = os.Getenv("DATABASE_USER")
	DatabasePassword = os.Getenv("DATABASE_PASSWORD")
	DatabaseName = os.Getenv("DATABASE_NAME")
	Mailsender = os.Getenv("MAIL_SENDER")
	Mailhost = os.Getenv("MAIL_HOST")
	Mailusername = os.Getenv("MAIL_ADDRESS")
	Mailpassword = os.Getenv("MAIL_PASSWORD")
	portStr := os.Getenv("MAIL_PORT")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Printf("Invalid MAIL_PORT value: %v", portStr)
		port = 587 // default fallback
	}
	Mailport = port
}

func PrintEnvVars() {
	fmt.Println("DatabaseHost:", DatabaseHost)
	fmt.Println("DatabaseUser:", DatabaseUser)
	fmt.Println("DatabasePassword:", DatabasePassword)
	fmt.Println("DatabaseName:", DatabaseName)
	fmt.Println("MailSender:", Mailsender)
	fmt.Println("MailHost:", Mailhost)
	fmt.Println("MailUsername:", Mailusername)
	fmt.Println("MailPassword:", Mailpassword)
	fmt.Println("MailPort:", Mailport)
}

func CleanupOldBackups(dir string, olderThan time.Duration) {
	files, _ := ioutil.ReadDir(dir)
	now := time.Now()
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".sql" {
			fullPath := filepath.Join(dir, file.Name())
			if now.Sub(file.ModTime()) > olderThan {
				os.Remove(fullPath)
				log.Println("Deleted old backup:", fullPath)
			}
		}
	}
}

// getServerIPAddress tries to find a non-loopback IP address
func GetServerIPAddress() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "localhost"
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "localhost" // fallback
}

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

// SortedKeys returns keys of a map in sorted order
func SortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
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
func EncodeUnits(units float64) map[string]interface{} {
	const maxAmount = 65530.0
	if units < 0 {
		return map[string]interface{}{
			"success": false,
			"message": "units value containing Negative values is not supported.",
		}
	}
	if units > maxAmount {
		return map[string]interface{}{
			"success": false,
			"message": "units value too large to be represented.",
		}
	}
	number := int(units)
	decimal := int(math.Round((units - float64(number)) * 100))
	numberBin := DecToBin(number, 16)     // 16 bits integer part
	decimalBin := DecToBin(decimal, 7)    // 7 bits decimal part
	amountBlock := numberBin + decimalBin // 23 bits total
	return map[string]interface{}{
		"success": true,
		"message": amountBlock,
		//map[string]string{
		//"number":       numberBin,
		//"decimal":      decimalBin,
		//"amount_block": amountBlock,
		//},
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

// Encrypt3DES encrypts 8 bytes of data using a 24-byte 3DES key.
func Encrypt3DES(data, key []byte) map[string]interface{} {
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
	encrypted := make([]byte, 8)
	cipherBlock.Encrypt(encrypted, data)
	return map[string]interface{}{
		"success": true,
		"message": encrypted,
	}
}

// GenerateClassBits generates 2-bit class bits based on the token class integer (0-3).
func GenerateClassBits(class int) map[string]interface{} {
	if class < 0 || class > 3 {
		return map[string]interface{}{
			"success": false,
			"message": "class must be between 0 and 3",
		}
	}
	classBits := DecToBin(class, 2)
	return map[string]interface{}{
		"success": true,
		"message": classBits,
	}
}

// TranspositionAndAddClassBits inserts the 2 class bits into the token binary string at positions 28 and 27.
// tokenBin must 64 bits.
func TranspositionAndAddClassBits(tokenBin, classBits string) map[string]interface{} {

	if len(tokenBin) < 64 {
		return map[string]interface{}{
			"success": false,
			"message": "token binary must be at least 64 bits",
		}
	}
	if len(classBits) != 2 {
		return map[string]interface{}{
			"success": false,
			"message": "classBits must be exactly 2 bits",
		}
	}

	// Prepend class bits to the token binary
	withClassBits := classBits + tokenBin
	bits := strings.Split(withClassBits, "")
	length := len(bits)

	if length < 66 {
		return map[string]interface{}{
			"success": false,
			"message": "Input token block must be at least 66 bits after prepending class bits",
		}
	}

	// Perform the bit swaps (transposition)
	// Swap pos28 and 65, 27 and 64
	bits[length-1-65] = bits[length-1-28]
	bits[length-1-64] = bits[length-1-27]
	bits[length-1-28] = string(classBits[0])
	bits[length-1-27] = string(classBits[1])

	return map[string]interface{}{
		"success": true,
		"message": strings.Join(bits, ""),
	}
}

// FormatTokenDisplay formats a binary token string into groups for display, grouping every 8 bits separated by a space.
func FormatTokenDisplay(tokenBin string) map[string]interface{} {
	// Convert binary string to integer
	n := new(big.Int)
	_, ok := n.SetString(tokenBin, 2)
	if !ok {
		return map[string]interface{}{
			"success": false,
			"message": "invalid binary token",
		}
	}

	// Format to 20-digit string with leading zeros if needed
	tokenStr := fmt.Sprintf("%020s", n.String())

	// Split into 5 groups of 4 digits
	var parts []string
	for i := 0; i < len(tokenStr); i += 4 {
		parts = append(parts, tokenStr[i:i+4])
	}

	// Join with hyphens
	formattedToken := strings.Join(parts, "-")
	return map[string]interface{}{
		"success": true,
		"message": formattedToken,
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

// GenerateRandomBits returns a random binary string of specified length as "message".
func GenerateRandomBits(length int) map[string]interface{} {
	if length <= 0 {
		return map[string]interface{}{
			"success": false,
			"message": "Length must be positive.",
		}
	}
	max := new(big.Int).Lsh(big.NewInt(1), uint(length)) // 2^length
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": "Random generation failed: " + err.Error(),
		}
	}
	binStr := fmt.Sprintf("%0*b", length, n) // binary zero-padded to length
	return map[string]interface{}{
		"success": true,
		"message": binStr,
	}
}
