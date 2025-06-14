package helpers

import (
	"bytes"
	"crypto/des"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/clbanning/mxj"
)

var (
	BaseDate         = time.Date(2025, 5, 5, 0, 0, 0, 0, time.UTC)
	ServerSecurity   string
	ServerDomain     string
	ServerPort       string
	SslCertificate   string
	SslKey           string
	DatabaseHost     string
	DatabaseUser     string
	DatabasePassword string
	DatabaseName     string
	Mailsender       string
	Mailhost         string
	Mailusername     string
	Mailpassword     string
	Mailport         int
	SmsUserName      string
	SmsApiKey        string
	SmsSenderId      string
)

func UpdateEnvVars() {
	//security := os.Getenv("SECURITY") DOMAIN
	ServerSecurity = getEnvValue("SECURITY", "http").(string)
	ServerDomain = getEnvValue("DOMAIN", "localhost").(string)
	ServerPort = getEnvValue("PORT", "2010").(string)
	SslCertificate = getEnvValue("SSL_CERTIFICATE", "").(string)
	SslKey = getEnvValue("SSL_KEY", "").(string)
	DatabaseHost = getEnvValue("DATABASE_HOST", "localhost").(string)
	DatabaseHost = getEnvValue("DATABASE_HOST", "localhost").(string)
	DatabaseUser = getEnvValue("DATABASE_USER", "root").(string)
	DatabasePassword = getEnvValue("DATABASE_PASSWORD", "").(string)
	DatabaseName = getEnvValue("DATABASE_NAME", "trick").(string)
	Mailsender = getEnvValue("MAIL_SENDER", "noreply@example.com").(string)
	Mailhost = getEnvValue("MAIL_HOST", "smtp.example.com").(string)
	Mailusername = getEnvValue("MAIL_ADDRESS", "noreply@example.com").(string)
	Mailpassword = getEnvValue("MAIL_PASSWORD", "").(string)
	Mailport = getEnvValue("MAIL_PORT", 587).(int)
	SmsUserName = getEnvValue("AFRICAS_TALKING_USERNAME", "").(string)
	SmsApiKey = getEnvValue("AFRICAS_TALKING_API_KEY", "").(string)
	SmsSenderId = getEnvValue("AFRICAS_TALKING_SENDER_ID", "").(string)
}

func getEnvValue(key string, fallback interface{}) interface{} {
	val := os.Getenv(key)
	if val == "" {
		log.Printf(`{"warning": "Environment variable %q not set, using default: %v"}`, key, fallback)
		return fallback
	}
	switch fallback.(type) {
	case int:
		intVal, err := strconv.Atoi(val)
		if err != nil {
			log.Printf(`{"warning": "Invalid int for %q: %q, using default: %v"}`, key, val, fallback)
			return fallback
		}
		return intVal
	case string:
		return val
	default:
		log.Printf(`{"warning": "Unsupported type for key %q, using default: %v"}`, key, fallback)
		return fallback
	}
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
		return "127.0.0.1"
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			ip := ipnet.IP.To4()
			if ip != nil && !ip.IsLoopback() && !strings.HasPrefix(ip.String(), "169.254.") {
				return ip.String() // Return first non-loopback, non-link-local IPv4
			}
		}
	}
	return "127.0.0.1" // fallback
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
func DecToBinBigInt(num *big.Int, length int) string {
	binStr := num.Text(2) // binary string
	if len(binStr) > length {
		binStr = binStr[len(binStr)-length:] // truncate left bits
	} else if len(binStr) < length {
		binStr = fmt.Sprintf("%0*s", length, binStr) // pad left zeros
	}
	return binStr
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
		key_type          = 0                    // 8 bits
		keyRevisionNumber = 0                    // 8 bits
		supplyGroupCode   = 12345                // 16 bits
		tariffIndex       = 3                    // 8 bits
		decoderRefNumber  = uint64(123456789012) // 64 bits
		random            = 12345678             // 88 bits
	)

	keyTypeBin := DecToBin(key_type, 8)                        // 8 bits
	supplyGroupBin := DecToBin(supplyGroupCode, 16)            // 16 bits
	tariffIndexBin := DecToBin(tariffIndex, 8)                 // 8 bits
	keyRevisionBin := DecToBin(keyRevisionNumber, 8)           // 8 bits
	decoderRefNumberBin := DecToBin(int(decoderRefNumber), 64) // 64 bits
	padding := DecToBin(random, 88)                            // 88 bits
	// Build data block by concatenating all bits
	dataBlock := keyTypeBin + supplyGroupBin + tariffIndexBin + keyRevisionBin + decoderRefNumberBin + padding
	//fmt.Println("DataBlock: ", dataBlock)
	//fmt.Println("dataBlock length:  ", len(dataBlock)) // Should be 16+8+16+8+8+64 = 120 bits
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

// LuhnCheckDigit calculates the Luhn check digit for a given number string
func LuhnCheckDigit(number string) int {
	sum := 0
	alt := false
	// Loop through the number in reverse
	for i := len(number) - 1; i >= 0; i-- {
		n, _ := strconv.Atoi(string(number[i]))
		if alt {
			n *= 2
			if n > 9 {
				n -= 9
			}
		}
		sum += n
		alt = !alt
	}
	// Compute the check digit
	checkDigit := (10 - (sum % 10)) % 10
	return checkDigit
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
func EncodeUnits(amount float64) map[string]interface{} {
	const maxAmount = 65530.0
	units := amount + 0.00001
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
	decimal := int((units - float64(number)) * 100)
	//log.Println("units:", units)
	//log.Println("number: ", number)
	//log.Println("decimal: ", decimal)
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

// convert xml to json
func XMLtoJSON(resp *http.Response) map[string]interface{} {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": "failed to read response body: " + err.Error(),
		}
	}
	// Close the body after reading
	resp.Body.Close()
	// Convert []byte to io.Reader for mxj
	reader := bytes.NewReader(body)
	// Parse XML to map[string]interface{}
	data, err := mxj.NewMapXmlReader(reader)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": "failed to parse XML: " + err.Error(),
		}
	}
	// Convert mxj.Map to map[string]interface{}
	result := map[string]interface{}(data)

	// Assume data is already parsed from XML and stored in a variable like this:
	return map[string]interface{}{
		"success": true,
		"message": result,
	}
}
