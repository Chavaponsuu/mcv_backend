package services

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"sort"
)
func GenerateSignature(params map[string] string, apiSecret string) string {
	keys := make([]string, 0 , len(params))
	for k := range params{
		keys = append(keys, k)
	} 
	sort.Strings(keys)

	str := ""

	for i,k := range keys{
		str += fmt.Sprintf("%s=%s",k,params[k])
		if i < len(keys)-1{
			str += "&" 
		}

		

		
	}
	str += apiSecret
	hash := sha1.New()
	hash.Write([] byte(str))
	return hex.EncodeToString(hash.Sum(nil))
}