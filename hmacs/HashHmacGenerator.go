package hmacs
import(
	"crypto/sha1"
	"crypto/hmac"
	"fmt"
)

type HashHmacGenerator struct {
	EncryptionKey []byte
}

func (this *HashHmacGenerator) Sha1(data []byte) []byte {
	hmac := hmac.New(sha1.New, this.EncryptionKey)
	hmac.Write(data)
	return hmac.Sum(nil)
}

// shows the same behavior as the PHP function hash_hmac('sha1', data, this.EncryptionKey)
func (this *HashHmacGenerator) Sha1String(data string) string {
	hmacBytes := this.Sha1([]byte(data))
	return fmt.Sprintf("%x", hmacBytes)
}


