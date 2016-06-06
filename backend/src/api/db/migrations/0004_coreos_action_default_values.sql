-- +migrate Up

ALTER TABLE ONLY coreos_action ALTER COLUMN event SET DEFAULT 'postinstall';
ALTER TABLE ONLY coreos_action ALTER COLUMN chromeos_version SET DEFAULT '';
ALTER TABLE ONLY coreos_action ALTER COLUMN needs_admin SET DEFAULT 'f';
ALTER TABLE ONLY coreos_action ALTER COLUMN is_delta SET DEFAULT 'f';
ALTER TABLE ONLY coreos_action ALTER COLUMN disable_payload_backoff SET DEFAULT 't';
ALTER TABLE ONLY coreos_action ALTER COLUMN metadata_signature_rsa SET DEFAULT '';
ALTER TABLE ONLY coreos_action ALTER COLUMN metadata_size SET DEFAULT '';
ALTER TABLE ONLY coreos_action ALTER COLUMN deadline SET DEFAULT '';

-- +migrate Down

ALTER TABLE ONLY coreos_action ALTER COLUMN event DROP DEFAULT;
ALTER TABLE ONLY coreos_action ALTER COLUMN chromeos_version DROP DEFAULT;
ALTER TABLE ONLY coreos_action ALTER COLUMN needs_admin DROP DEFAULT;
ALTER TABLE ONLY coreos_action ALTER COLUMN is_delta DROP DEFAULT;
ALTER TABLE ONLY coreos_action ALTER COLUMN disable_payload_backoff DROP DEFAULT;
ALTER TABLE ONLY coreos_action ALTER COLUMN metadata_signature_rsa DROP DEFAULT;
ALTER TABLE ONLY coreos_action ALTER COLUMN metadata_size DROP DEFAULT;
ALTER TABLE ONLY coreos_action ALTER COLUMN deadline DROP DEFAULT;
