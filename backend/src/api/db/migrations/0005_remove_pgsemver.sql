-- +migrate Up

ALTER TABLE package ALTER COLUMN version TYPE varchar(255);
ALTER TABLE package ADD CHECK (version <> '');
ALTER TABLE instance_application ALTER COLUMN version TYPE varchar(255);
ALTER TABLE instance_application ADD CHECK (version <> '');
ALTER TABLE instance_application ALTER COLUMN last_update_version TYPE varchar(255);
ALTER TABLE instance_application ADD CHECK (last_update_version <> '');
ALTER TABLE instance_status_history ALTER COLUMN version TYPE varchar(255);
ALTER TABLE instance_status_history ADD CHECK (version <> '');
ALTER TABLE activity ALTER COLUMN version TYPE varchar(255);
ALTER TABLE activity ADD CHECK (version <> '');
ALTER TABLE event ALTER COLUMN previous_version TYPE varchar(255);
ALTER TABLE coreos_action ALTER COLUMN chromeos_version TYPE varchar(255);

-- +migrate Down

ALTER TABLE package ALTER COLUMN version TYPE SEMVER;
ALTER TABLE package DROP CHECK (version <> '');
ALTER TABLE instance_application ALTER COLUMN version TYPE SEMVER;
ALTER TABLE instance_application DROP CHECK (version <> '');
ALTER TABLE instance_application ALTER COLUMN last_update_version TYPE SEMVER;
ALTER TABLE instance_application DROP CHECK (last_update_version <> '');
ALTER TABLE instance_status_history ALTER COLUMN version TYPE SEMVER;
ALTER TABLE instance_status_history DROP CHECK (version <> '');
ALTER TABLE activity ALTER COLUMN version TYPE SEMVER;
ALTER TABLE activity DROP CHECK (version <> '');
ALTER TABLE event ALTER COLUMN previous_version TYPE varchar(10);
ALTER TABLE coreos_action ALTER COLUMN chromeos_version TYPE varchar(20);
