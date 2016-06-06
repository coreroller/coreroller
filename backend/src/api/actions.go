package api

import "time"

// CoreosAction represents an Omaha action with some CoreOS specific fields.
type CoreosAction struct {
	ID                    string    `db:"id" json:"id"`
	Event                 string    `db:"event" json:"event"`
	ChromeOSVersion       string    `db:"chromeos_version" json:"chromeos_version"`
	Sha256                string    `db:"sha256" json:"sha256"`
	NeedsAdmin            bool      `db:"needs_admin" json:"needs_admin"`
	IsDelta               bool      `db:"is_delta" json:"is_delta"`
	DisablePayloadBackoff bool      `db:"disable_payload_backoff" json:"disable_payload_backoff"`
	MetadataSignatureRsa  string    `db:"metadata_signature_rsa" json:"metadata_signature_rsa"`
	MetadataSize          string    `db:"metadata_size" json:"metadata_size"`
	Deadline              string    `db:"deadline" json:"deadline"`
	CreatedTs             time.Time `db:"created_ts" json:"created_ts"`
	PackageID             string    `db:"package_id" json:"-"`
}

// AddCoreosAction registers the provided Omaha CoreOS action.
func (api *API) AddCoreosAction(action *CoreosAction) (*CoreosAction, error) {
	err := api.dbR.
		InsertInto("coreos_action").
		Whitelist("event", "chromeos_version", "sha256", "needs_admin", "is_delta", "disable_payload_backoff", "metadata_signature_rsa", "metadata_size", "deadline", "package_id").
		Record(action).
		Returning("*").
		QueryStruct(action)

	return action, err
}

// GetCoreosAction returns the CoreOS action entry associated to the package id
// provided.
func (api *API) GetCoreosAction(packageID string) (*CoreosAction, error) {
	var action CoreosAction

	err := api.dbR.SelectDoc("*").
		From("coreos_action").
		Where("package_id = $1", packageID).
		QueryStruct(&action)

	if err != nil {
		return nil, err
	}

	return &action, nil
}
