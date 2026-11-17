package db

// SystemUserID is written to created_by/updated_by when no acting user exists
// (self-registration).
const SystemUserID int64 = 1

// CreateUserParams holds the fields for a new user. Empty Name, Email or Phone
// is stored as NULL.
type CreateUserParams struct {
	Name      string
	Email     string
	Phone     string
	Password  string // already hashed
	LoginType string // "email" | "phone"
	CreatedBy int64  // 0 means SystemUserID; also used for updated_by
}

// UpdateProfileParams holds a partial profile update. A nil field is left unchanged.
type UpdateProfileParams struct {
	Name      *string
	Phone     *string
	UpdatedBy int64
}
