package auth

import "context"

type UserRepository interface {
	List(ctx context.Context) ([]User, error)
	GetByID(ctx context.Context, id string) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByAPIKey(ctx context.Context, key string) (*User, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
	IncrementLoginAttempts(ctx context.Context, id string) error
	ResetLoginAttempts(ctx context.Context, id string) error
	SetLocked(ctx context.Context, id string, until *interface{}) error
}

type GroupRepository interface {
	List(ctx context.Context) ([]Group, error)
	GetByID(ctx context.Context, id string) (*Group, error)
	Create(ctx context.Context, group *Group) error
	Update(ctx context.Context, group *Group) error
	Delete(ctx context.Context, id string) error
}

type SessionRepository interface {
	Create(ctx context.Context, session *Session) error
	GetByToken(ctx context.Context, token string) (*Session, error)
	Delete(ctx context.Context, id string) error
	DeleteByUserID(ctx context.Context, userID string) error
	DeleteExpired(ctx context.Context) error
}

type AuditLogRepository interface {
	Create(ctx context.Context, log *AuditLog) error
	List(ctx context.Context, filter AuditFilter) ([]AuditLog, error)
}

type AuditFilter struct {
	UserID    string
	Action    string
	Resource  string
	StartTime string
	EndTime   string
	Limit     int
	Offset    int
}

type CertificateRepository interface {
	List(ctx context.Context) ([]Certificate, error)
	GetByID(ctx context.Context, id string) (*Certificate, error)
	Create(ctx context.Context, cert *Certificate) error
	Update(ctx context.Context, cert *Certificate) error
	Delete(ctx context.Context, id string) error
	ListCAs(ctx context.Context) ([]Certificate, error)
}

// ExternalAuthProvider interfaces with LDAP/RADIUS.
type ExternalAuthProvider interface {
	Authenticate(ctx context.Context, username, password string) (*User, error)
	GetGroups(ctx context.Context, username string) ([]string, error)
	TestConnection(ctx context.Context) error
}

// TOTPProvider handles TOTP-based MFA.
type TOTPProvider interface {
	GenerateSecret(username string) (secret string, qrURL string, err error)
	ValidateCode(secret string, code string) bool
	GenerateBackupCodes(count int) ([]string, error)
}

// ACMEProvider handles Let's Encrypt certificate issuance.
type ACMEProvider interface {
	IssueCertificate(ctx context.Context, domains []string, challenge string) (*Certificate, error)
	RenewCertificate(ctx context.Context, certID string) (*Certificate, error)
}
