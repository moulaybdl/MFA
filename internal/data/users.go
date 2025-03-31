package data

import (
	"database/sql"
	"errors"

	"mfa.moulay/internal/validator"

	"golang.org/x/crypto/bcrypt"
)

// Errors:
var (
	ErrDuplicateEmail = errors.New("this email is already taken")
	ErrDuplicatePhoneNumber = errors.New("this phone number is already taken")
)

// the User struct
type User struct {
	ID int `json:"id"`
	Name string `json:"name"`
	Email string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Password password `json:"-"`
	OTPActivated bool `json:"otp_actiavated"`
	BiometricPublicKey string `json:"biometric_public_key"`
}

// the UserModel
type UsersModel struct {
	DB *sql.DB
}

// local password struct to hold the passowrd info
type password struct {
	plainText *string 
	hashed []byte
}

func (p *password) CreatePassowrd(plain_text *string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(*plain_text), 12)
	if err != nil {
		return err
	}
	
	p.plainText = plain_text
	p.hashed = hashed

	return nil

}

func (p *password) VerifyMatch(plaintext_password string) (bool, error)  {
	err := bcrypt.CompareHashAndPassword(p.hashed, []byte(plaintext_password))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

// validation

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRx), "email", "must be a valid email address")
}

func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}


func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.Name != "", "name", "must be provided")
	v.Check(len(user.Name) <= 500, "name", "must not be more than 500 bytes long")

	ValidateEmail(v, user.Email)

	if user.Password.plainText != nil {
	ValidatePasswordPlaintext(v, *user.Password.plainText)
	}

	if user.Password.hashed == nil {
	panic("missing password hash for user")
	}
}
// the userModel methods
func (m *UsersModel) Insert(user *User) error {
	query := `
	INSERT INTO users (name, email, phone_number, password_hash, otp_activated, biometric_public_key)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id
`
	args := []interface{}{user.Name, user.Email, user.PhoneNumber, user.Password.hashed, user.OTPActivated, user.BiometricPublicKey}

	err := m.DB.QueryRow(query, args...).Scan(&user.ID)
	if err != nil {
		if err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"` {
			return ErrDuplicateEmail
		} else if err.Error() == `pq: duplicate key value violates unique constraint "users_phone_number_key"` {
			return ErrDuplicatePhoneNumber
		}
		return err

}
return nil
}

func (m *UsersModel) ChangeOTPSate(userID int) error {
	query := `
	UPDATE users
	SET otp_activated = true
	WHERE id = $1
	`
	args := []interface{}{userID}

	_, err := m.DB.Exec(query, args...)
	if err != nil {
		return err
	}

	return nil

}