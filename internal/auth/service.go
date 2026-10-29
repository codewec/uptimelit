package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"regexp"
	"time"
	"uptimelit/internal/db"
)

var ErrCredentials = errors.New("invalid credentials")
var ErrDenied = errors.New("account is not provisioned")
var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9._-]{1,64}$`)

type User struct {
	ID              string         `json:"id"`
	Username        string         `json:"username"`
	Role            string         `json:"role"`
	PasswordHash    string         `json:"-"`
	Issuer, Subject sql.NullString `json:"-"`
	HasPassword     bool           `json:"hasPassword"`
	HasOIDC         bool           `json:"hasOIDC"`
	AuthMethod      string         `json:"authMethod,omitempty"`
}

func ValidateUsername(v string) error {
	if !usernamePattern.MatchString(v) {
		return errors.New("username must contain 1–64 letters, digits, dots, underscores or dashes")
	}
	return nil
}
func HashPassword(v string) (string, error) {
	if len(v) < 8 || len(v) > 72 {
		return "", errors.New("password must contain 8–72 bytes")
	}
	b, e := bcrypt.GenerateFromPassword([]byte(v), bcrypt.DefaultCost)
	return string(b), e
}
func PasswordMatches(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
func Random() string {
	b := make([]byte, 32)
	if _, e := rand.Read(b); e != nil {
		panic(e)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

type Service struct {
	DB     *sql.DB
	secret []byte
	dummy  string
}

func New(d *db.DB, c Config) (*Service, error) {
	secret := c.JWTSecret
	if secret == "" {
		if _, e := d.SQL().Exec(`INSERT OR IGNORE INTO settings(key,value) VALUES('jwt_secret',?)`, Random()); e != nil {
			return nil, e
		}
		if e := d.SQL().QueryRow(`SELECT value FROM settings WHERE key='jwt_secret'`).Scan(&secret); e != nil {
			return nil, e
		}
	}
	dummy, e := HashPassword(Random())
	if e != nil {
		return nil, e
	}
	return &Service{DB: d.SQL(), secret: []byte(secret), dummy: dummy}, nil
}

type scanner interface{ Scan(...any) error }

func scanUser(row scanner) (*User, error) {
	u := new(User)
	e := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.Issuer, &u.Subject)
	u.HasPassword = u.PasswordHash != ""
	u.HasOIDC = u.Issuer.Valid && u.Subject.Valid
	return u, e
}

const columns = `id,username,password_hash,role,oidc_issuer,oidc_subject`

func (s *Service) User(ctx context.Context, id string) (*User, error) {
	return scanUser(s.DB.QueryRowContext(ctx, `SELECT `+columns+` FROM users WHERE id=?`, id))
}
func (s *Service) Empty(ctx context.Context) (bool, error) {
	var n int
	e := s.DB.QueryRowContext(ctx, `SELECT count(*) FROM users`).Scan(&n)
	return n == 0, e
}
func (s *Service) Login(ctx context.Context, name, password string) (*User, error) {
	if e := ValidateUsername(name); e != nil {
		return nil, ErrCredentials
	}
	// Hash outside the write transaction; the transaction serializes bootstrap.
	empty, e := s.Empty(ctx)
	if e != nil {
		return nil, e
	}
	hash := ""
	if empty {
		hash, e = HashPassword(password)
		if e != nil {
			return nil, e
		}
	}
	tx, e := s.DB.BeginTx(ctx, nil)
	if e != nil {
		return nil, e
	}
	defer tx.Rollback()
	var n int
	if e = tx.QueryRowContext(ctx, `SELECT count(*) FROM users`).Scan(&n); e != nil {
		return nil, e
	}
	if n == 0 {
		u := &User{ID: uuid.NewString(), Username: name, Role: "admin", PasswordHash: hash, HasPassword: true}
		if hash == "" {
			return nil, ErrCredentials
		}
		_, e = tx.ExecContext(ctx, `INSERT INTO users(id,username,password_hash,role) VALUES(?,?,?,'admin')`, u.ID, name, hash)
		if e != nil {
			return nil, e
		}
		return u, tx.Commit()
	}
	u, e := scanUser(tx.QueryRowContext(ctx, `SELECT `+columns+` FROM users WHERE username=?`, name))
	if e != nil {
		PasswordMatches(s.dummy, password)
		return nil, ErrCredentials
	}
	if !PasswordMatches(u.PasswordHash, password) {
		return nil, ErrCredentials
	}
	return u, tx.Commit()
}
func (s *Service) OIDCUser(ctx context.Context, i *OIDCIdentity) (*User, error) {
	tx, e := s.DB.BeginTx(ctx, nil)
	if e != nil {
		return nil, e
	}
	defer tx.Rollback()
	u, e := scanUser(tx.QueryRowContext(ctx, `SELECT `+columns+` FROM users WHERE oidc_issuer=? AND oidc_subject=?`, i.Issuer, i.Subject))
	if e == nil {
		return u, tx.Commit()
	}
	if !errors.Is(e, sql.ErrNoRows) {
		return nil, e
	}
	name := oidcUsername(i)
	u, e = scanUser(tx.QueryRowContext(ctx, `SELECT `+columns+` FROM users WHERE username=? AND oidc_issuer IS NULL AND oidc_subject IS NULL`, name))
	if e == nil {
		_, e = tx.ExecContext(ctx, `UPDATE users SET oidc_issuer=?,oidc_subject=? WHERE id=?`, i.Issuer, i.Subject, u.ID)
		if e != nil {
			return nil, e
		}
		if _, e = tx.ExecContext(ctx, `DELETE FROM sessions WHERE user_id=?`, u.ID); e != nil {
			return nil, e
		}
		u.HasOIDC = true
		return u, tx.Commit()
	}
	if !errors.Is(e, sql.ErrNoRows) {
		return nil, e
	}
	var n int
	if e = tx.QueryRowContext(ctx, `SELECT count(*) FROM users`).Scan(&n); e != nil {
		return nil, e
	}
	if n != 0 {
		return nil, ErrDenied
	}
	u = &User{ID: uuid.NewString(), Username: name, Role: "admin", HasOIDC: true}
	_, e = tx.ExecContext(ctx, `INSERT INTO users(id,username,role,oidc_issuer,oidc_subject) VALUES(?,?,'admin',?,?)`, u.ID, name, i.Issuer, i.Subject)
	if e != nil {
		return nil, e
	}
	return u, tx.Commit()
}

type Claims struct {
	Method string `json:"method"`
	jwt.RegisteredClaims
}

func (s *Service) Session(ctx context.Context, u *User, method string) (string, error) {
	now := time.Now()
	id := Random()
	exp := now.Add(7 * 24 * time.Hour)
	claims := Claims{Method: method, RegisteredClaims: jwt.RegisteredClaims{Issuer: "uptimelit", Audience: jwt.ClaimStrings{"uptimelit"}, Subject: u.ID, ID: id, IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(exp)}}
	token, e := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
	if e != nil {
		return "", e
	}
	if _, e = s.DB.ExecContext(ctx, `DELETE FROM sessions WHERE expires_at<=?`, now.Unix()); e != nil {
		return "", e
	}
	_, e = s.DB.ExecContext(ctx, `INSERT INTO sessions VALUES(?,?,?,?)`, id, u.ID, method, exp.Unix())
	return token, e
}
func (s *Service) Authenticate(ctx context.Context, raw string) (*User, string, error) {
	c := new(Claims)
	_, e := jwt.ParseWithClaims(raw, c, func(t *jwt.Token) (any, error) { return s.secret, nil }, jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired(), jwt.WithIssuedAt(), jwt.WithIssuer("uptimelit"), jwt.WithAudience("uptimelit"))
	if e != nil {
		return nil, "", ErrCredentials
	}
	var method string
	e = s.DB.QueryRowContext(ctx, `SELECT auth_method FROM sessions WHERE id=? AND user_id=? AND expires_at>?`, c.ID, c.Subject, time.Now().Unix()).Scan(&method)
	if e != nil || method != c.Method {
		return nil, "", ErrCredentials
	}
	u, e := s.User(ctx, c.Subject)
	if e != nil {
		return nil, "", ErrCredentials
	}
	u.AuthMethod = method
	return u, c.ID, nil
}
func (s *Service) Logout(ctx context.Context, id string) error {
	_, e := s.DB.ExecContext(ctx, `DELETE FROM sessions WHERE id=?`, id)
	return e
}
func (s *Service) Users(ctx context.Context) ([]*User, error) {
	rows, e := s.DB.QueryContext(ctx, `SELECT `+columns+` FROM users ORDER BY username`)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []*User{}
	for rows.Next() {
		u, e := scanUser(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, u)
	}
	return out, rows.Err()
}
func (s *Service) CreateUser(ctx context.Context, name, password, role string) (*User, error) {
	if e := ValidateUsername(name); e != nil {
		return nil, e
	}
	if role != "admin" && role != "user" {
		return nil, errors.New("invalid role")
	}
	hash := ""
	var e error
	if password != "" {
		hash, e = HashPassword(password)
		if e != nil {
			return nil, e
		}
	}
	u := &User{ID: uuid.NewString(), Username: name, Role: role, HasPassword: hash != ""}
	_, e = s.DB.ExecContext(ctx, `INSERT INTO users(id,username,password_hash,role) VALUES(?,?,?,?)`, u.ID, name, hash, role)
	return u, e
}

// UpdateUser and DeleteUser serialize checks for the last administrator with their writes.
func (s *Service) UpdateUser(ctx context.Context, id, name, role, password string, resetOIDC, passwordEnabled bool) error {
	if e := ValidateUsername(name); e != nil {
		return e
	}
	if role != "admin" && role != "user" {
		return errors.New("invalid role")
	}
	hash := ""
	var e error
	if password != "" {
		hash, e = HashPassword(password)
		if e != nil {
			return e
		}
	}
	tx, e := s.DB.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	u, e := scanUser(tx.QueryRowContext(ctx, `SELECT `+columns+` FROM users WHERE id=?`, id))
	if e != nil {
		return e
	}
	if u.Role == "admin" && role != "admin" {
		var n int
		if e = tx.QueryRowContext(ctx, `SELECT count(*) FROM users WHERE role='admin'`).Scan(&n); e != nil {
			return e
		}
		if n <= 1 {
			return errors.New("cannot demote the last administrator")
		}
	}
	if resetOIDC && (hash == "" && u.PasswordHash == "" || !passwordEnabled) {
		return errors.New("cannot remove the only available login method")
	}
	if hash == "" {
		hash = u.PasswordHash
	}
	_, e = tx.ExecContext(ctx, `UPDATE users SET username=?,role=?,password_hash=?,oidc_issuer=CASE WHEN ? THEN NULL ELSE oidc_issuer END,oidc_subject=CASE WHEN ? THEN NULL ELSE oidc_subject END WHERE id=?`, name, role, hash, resetOIDC, resetOIDC, id)
	if e != nil {
		return e
	}
	if _, e = tx.ExecContext(ctx, `DELETE FROM sessions WHERE user_id=?`, id); e != nil {
		return e
	}
	return tx.Commit()
}
func (s *Service) DeleteUser(ctx context.Context, id, current string) ([]int64, error) {
	if id == current {
		return nil, errors.New("cannot delete your own account")
	}
	tx, e := s.DB.BeginTx(ctx, nil)
	if e != nil {
		return nil, e
	}
	defer tx.Rollback()
	u, e := scanUser(tx.QueryRowContext(ctx, `SELECT `+columns+` FROM users WHERE id=?`, id))
	if e != nil {
		return nil, e
	}
	if u.Role == "admin" {
		var n int
		if e = tx.QueryRowContext(ctx, `SELECT count(*) FROM users WHERE role='admin'`).Scan(&n); e != nil {
			return nil, e
		}
		if n <= 1 {
			return nil, errors.New("cannot delete the last administrator")
		}
	}
	rows, e := tx.QueryContext(ctx, `SELECT id FROM endpoints WHERE owner_id=?`, id)
	if e != nil {
		return nil, e
	}
	ids := []int64{}
	for rows.Next() {
		var v int64
		if e = rows.Scan(&v); e != nil {
			rows.Close()
			return nil, e
		}
		ids = append(ids, v)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return nil, e
	}
	if _, e = tx.ExecContext(ctx, `DELETE FROM users WHERE id=?`, id); e != nil {
		return nil, e
	}
	return ids, tx.Commit()
}
func IsConflict(e error) bool {
	return e != nil && regexp.MustCompile(`(?i)unique constraint`).MatchString(e.Error())
}
