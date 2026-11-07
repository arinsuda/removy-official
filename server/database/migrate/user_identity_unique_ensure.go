package migrate

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

type identityUnique struct {
	name    string
	columns string
}

var identityUniques = []identityUnique{
	{"uq_provider_subject", "provider, provider_subject"},
	{"uq_user_provider", "user_id, provider"},
}

// EnsureUserIdentityUniques makes sure the identity uniqueness guarantees exist.
// AutoMigrate creates user_identities from the model first, which carries no
// unique constraints, so the raw-SQL definition in google_oauth_user_identity_v1
// is skipped on a fresh database. Additive only: an index is created only when
// absent from the current schema and no duplicate rows exist. Nothing is deleted or
// rewritten, and duplicates never fail the boot: they are logged at ERROR and the
// names of the indexes still missing are returned so the caller can leave the
// migration unrecorded and retry once the data is fixed.
func EnsureUserIdentityUniques(tx *gorm.DB) (missing []string, err error) {
	for _, u := range identityUniques {
		var present int64
		if err := tx.Raw(`SELECT count(*) FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
			WHERE c.relname = ? AND n.nspname = current_schema()`, u.name).Scan(&present).Error; err != nil {
			return nil, err
		}
		if present > 0 {
			continue
		}
		var dupGroups []int64
		if err := tx.Raw(fmt.Sprintf(`SELECT count(*) FROM user_identities GROUP BY %s HAVING count(*) > 1`, u.columns)).Scan(&dupGroups).Error; err != nil {
			return nil, err
		}
		if len(dupGroups) > 0 {
			log.Printf("ERROR migration: %s not created, %d duplicate (%s) groups exist in user_identities; resolve them by hand, the index is retried on every boot", u.name, len(dupGroups), u.columns)
			missing = append(missing, u.name)
			continue
		}
		if err := tx.Exec(fmt.Sprintf(`CREATE UNIQUE INDEX %s ON user_identities (%s)`, u.name, u.columns)).Error; err != nil {
			return nil, err
		}
	}
	return missing, nil
}
