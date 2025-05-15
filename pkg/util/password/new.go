package password

import "encoding/base64"

// New Create a new password object
func New(p string, salts ...string) Password {
	saltBytes, _ := GenerateSalt(16)
	salt := base64.StdEncoding.EncodeToString(saltBytes)
	if len(salts) > 0 {
		salt = salts[0]
	}
	return &password{p: p, salt: salt}
}

type (
	password struct {
		p, enP, salt string
	}

	Password interface {
		EQ(hashedPassword string) bool
		PValue() string
		EnValue() (string, error)
		Salt() string
	}
)

func (p *password) EQ(hashedPassword string) bool {
	if p == nil || len(hashedPassword) == 0 {
		return false
	}
	return CheckPassword(ObfuscatePassword(p.PValue(), p.Salt()), hashedPassword)
}

func (p *password) PValue() string {
	return p.p
}

func (p *password) EnValue() (string, error) {
	var err error
	if p.enP == "" {
		p.enP, err = HashPassword(ObfuscatePassword(p.PValue(), p.Salt()))
	}
	return p.enP, err
}

func (p *password) Salt() string {
	return p.salt
}
