// Path: ./blogX_server/service/email_service/email_address_validation.go

package email_service

import (
	"net"
	"regexp"
	"strings"
)

func IsValidEmail(email string) bool {
	// RFC 5322 简化版正则
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

func IsValidWithDomain(email string) bool {
	valid := IsValidEmail(email)
	if !valid {
		return false
	}
	parts := strings.Split(email, "@")
	domain := parts[1]
	_, err := net.LookupMX(domain)
	return err == nil
}
