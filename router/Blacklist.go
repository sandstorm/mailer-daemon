package router
import (
	"encoding/json"
	"strings"
)

type Blacklist struct {
	Emails Set
}

func (this *Blacklist) IsEmpty() bool {
	return this.Emails.IsEmpty()
}

func (this *Blacklist) Contains(recipient string) bool {
	if this.Emails == nil || this.Emails.IsEmpty() { return false }

	var dehydratedRecipient map[string]interface{}
	err := json.Unmarshal([]byte(recipient), &dehydratedRecipient)
	if err != nil { return false }

	emailField := dehydratedRecipient["email"]
	if emailField == nil { return false }

	email, emailIsString := emailField.(string)
	if !emailIsString { return false }

	return this.Emails.Contains(strings.ToLower(email))
}

func CreateBlacklistFromFile(path string) (blacklist Blacklist, err error) {
	if path == "" {
		return Blacklist{
			Emails: Set{},
		}, nil
	}

	emailHashes, err := CreateSetFromFile(path, addFirstColumnLowerCasesFromCsvLine)
	if err != nil { return }
	blacklist = Blacklist{
		Emails: emailHashes,
	}
	return
}

func addFirstColumnLowerCasesFromCsvLine(set *Set, line string) {
	email := line
	indexOfDelimiter := strings.IndexAny(line, ";	,|")
	if indexOfDelimiter >= 0 {
		email = line[:indexOfDelimiter]
	}

	email = strings.TrimSpace(email)
	if email == "" { return }

	(*set).Add(strings.ToLower(email))
}
