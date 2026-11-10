package admin_module

import (
	"time"

	"gorm.io/datatypes"
)

type AuditAction string

const (
	ActionUserDeactivated AuditAction = "USER_DEACTIVATED"
	ActionUserReactivated AuditAction = "USER_REACTIVATED"
	ActionUserRoleChanged AuditAction = "USER_ROLE_CHANGED"
	ActionUserDeleted     AuditAction = "USER_DELETED"
	ActionUserErased      AuditAction = "USER_ERASED"
	ActionReviewDeleted   AuditAction = "REVIEW_DELETED"
)

type AdminAuditLog struct {
	ID uint `gorm:"primarykey;autoIncrement" json:"id"`
	// AdminID becomes NULL when the admin is hard-deleted (FK fk_admin_audit_logs_admin, SET NULL).
	// No relation tag: the FK is created by a SQL migration, not AutoMigrate.
	AdminID    *uint          `gorm:"index" json:"admin_id"`
	Action     AuditAction    `gorm:"type:varchar(50);not null;index" json:"action"`
	TargetType string         `gorm:"type:varchar(50);not null;index" json:"target_type"`
	TargetID   uint           `gorm:"not null;index" json:"target_id"`
	Reason     *string        `gorm:"type:text" json:"reason"`
	MetaData   datatypes.JSON `gorm:"type:jsonb" json:"meta_data"`
	CreatedAt  time.Time      `gorm:"not null;index" json:"created_at"`
}

func (AdminAuditLog) TableName() string {
	return "admin_audit_logs"
}
