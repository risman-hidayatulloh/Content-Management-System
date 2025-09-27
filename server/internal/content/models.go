package content

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type User struct {
	ID           string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Email        string `gorm:"uniqueIndex"`
	PasswordHash string
	Roles        []Role `gorm:"many2many:user_roles;"`
	CreatedAt    time.Time
}

type Role struct {
	ID          string       `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name        string       `gorm:"uniqueIndex"`
	Permissions []Permission `gorm:"many2many:role_permissions;"`
}

type Permission struct {
	ID       string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Action   string // create/read/edit/publish/delete
	Resource string // content:*, media:*, users:*
}

type ContentType struct {
	ID        string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name      string `gorm:"uniqueIndex"`
	Slug      string `gorm:"uniqueIndex"`
	Fields    []FieldDef
	CreatedAt time.Time
	UpdatedAt time.Time
}

type FieldDef struct {
	ID            string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ContentTypeID string `gorm:"index"`
	Name          string
	Key           string `gorm:"index"`
	Type          string
	Required      bool
	IsUnique      bool
	IsRelation    bool
	RelationTo    *string
	Config        datatypes.JSON
	Order         int
}

type Entry struct {
	ID            string         `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ContentTypeID string         `gorm:"index"`
	Slug          *string        `gorm:"uniqueIndex"`
	Status        string         `gorm:"index"` // draft|published
	ScheduledAt   *time.Time     `gorm:"index"`
	PublishedAt   *time.Time     `gorm:"index"`
	Data          datatypes.JSON // JSONB
	Version       int
	CreatedBy     string
	UpdatedBy     string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

type EntryVersion struct {
	ID        string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	EntryID   string `gorm:"index"`
	Version   int
	Data      datatypes.JSON
	Summary   *string
	ActorID   string
	CreatedAt time.Time
}

type AuditLog struct {
	ID        string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ActorID   *string
	Action    string
	Target    string
	Meta      datatypes.JSON
	CreatedAt time.Time
}

type Media struct {
	ID         string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Filename   string
	Mime       string
	Size       int64
	Width      *int
	Height     *int
	Variants   datatypes.JSON
	URL        string
	StorageKey string
	CreatedBy  string
	CreatedAt  time.Time
}
