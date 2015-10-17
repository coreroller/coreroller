-- +migrate Up

-- CoreRoller schema

CREATE TABLE team (
	id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
	name varchar(25) NOT NULL CHECK (name <> '') UNIQUE,
	created_ts timestamp WITHOUT TIME ZONE DEFAULT (now() at time zone 'utc') NOT NULL
);

CREATE TABLE users (
	id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
	username varchar(25) NOT NULL CHECK (username <> '') UNIQUE,
	secret varchar(50) NOT NULL CHECK (secret <> ''),
	created_ts timestamp WITHOUT TIME ZONE DEFAULT (now() at time zone 'utc') NOT NULL,
	team_id UUID NOT NULL REFERENCES team (id)
);

CREATE TABLE application (
	id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
	name varchar(50) NOT NULL CHECK (name <> ''),
	description text,
	created_ts timestamp WITHOUT TIME ZONE DEFAULT (now() at time zone 'utc') NOT NULL,
	team_id UUID NOT NULL REFERENCES team (id),
	UNIQUE (name, team_id)
);

CREATE TABLE package (
	id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
	type int NOT NULL CHECK (type > 0),
	version SEMVER NOT NULL,
	url varchar(256) NOT NULL CHECK (url <> ''),
	filename varchar(100),
	description text,
	size varchar(20),
	hash varchar(64),
	created_ts timestamp WITHOUT TIME ZONE DEFAULT (now() at time zone 'utc') NOT NULL,
	application_id UUID NOT NULL REFERENCES application (id) ON DELETE CASCADE,
	UNIQUE(version, application_id)
);

-- TODO: review fields types and lengths in this table !!!
CREATE TABLE coreos_action (
	id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
	event varchar(20), 
	chromeos_version varchar(20), 
	sha256 varchar(64),
	needs_admin boolean,
	is_delta boolean,
	disable_payload_backoff boolean,
	metadata_signature_rsa varchar(256),
	metadata_size varchar(100),
	deadline varchar(100),
	created_ts timestamp WITHOUT TIME ZONE DEFAULT (now() at time zone 'utc') NOT NULL,
	package_id UUID NOT NULL REFERENCES package (id) ON DELETE CASCADE
);

CREATE TABLE channel (
	id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
	name varchar(25) NOT NULL CHECK (name <> ''),
	color varchar(25) NOT NULL,
	created_ts timestamp WITHOUT TIME ZONE DEFAULT (now() at time zone 'utc') NOT NULL,
	application_id UUID NOT NULL REFERENCES application (id) ON DELETE CASCADE,
	package_id UUID REFERENCES package (id) ON DELETE SET NULL,
	UNIQUE (name, application_id)
);

CREATE TABLE groups (
	id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
	name varchar(50) NOT NULL CHECK (name <> ''),
	description text NOT NULL,
	rollout_in_progress boolean DEFAULT false NOT NULL,
	policy_updates_enabled boolean DEFAULT true NOT NULL,
	policy_safe_mode boolean DEFAULT true NOT NULL,
	policy_office_hours boolean DEFAULT false NOT NULL,
	policy_timezone varchar(40),
	policy_period_interval varchar(20) NOT NULL CHECK (policy_period_interval <> ''),
	policy_max_updates_per_period integer NOT NULL CHECK (policy_max_updates_per_period > 0),
	policy_update_timeout varchar(20) NOT NULL CHECK (policy_update_timeout <> ''),
	created_ts timestamp WITHOUT TIME ZONE DEFAULT (now() at time zone 'utc') NOT NULL,
	application_id UUID NOT NULL REFERENCES application (id) ON DELETE CASCADE,
	channel_id UUID REFERENCES channel (id) ON DELETE SET NULL,
	UNIQUE (name, application_id)
);

CREATE TABLE instance (
	id varchar(50) PRIMARY KEY CHECK (id <> ''),
	ip inet NOT NULL,
	created_ts timestamp WITHOUT TIME ZONE DEFAULT (now() at time zone 'utc') NOT NULL
);

CREATE TABLE instance_status (
	id serial PRIMARY KEY,
	name varchar(20) NOT NULL,
	color varchar(25) NOT NULL,
	icon varchar(20) NOT NULL
);

CREATE TABLE instance_application (
	version SEMVER NOT NULL,
	created_ts timestamp WITHOUT TIME ZONE DEFAULT (now() at time zone 'utc') NOT NULL,
	status integer,
	last_check_for_updates timestamp WITHOUT TIME ZONE DEFAULT (now() at time zone 'utc') NOT NULL,
	last_update_granted_ts timestamp WITHOUT TIME ZONE,
	last_update_version SEMVER,
	update_in_progress boolean DEFAULT false NOT NULL,
	instance_id varchar(50) NOT NULL REFERENCES instance (id) ON DELETE CASCADE,
	application_id UUID NOT NULL REFERENCES application (id) ON DELETE CASCADE,
	group_id UUID REFERENCES groups (id) ON DELETE SET NULL,
	PRIMARY KEY (instance_id, application_id)
);

CREATE TABLE instance_status_history (
	id serial PRIMARY KEY,
	status integer,
	version SEMVER,
	created_ts timestamp WITHOUT TIME ZONE DEFAULT (now() at time zone 'utc') NOT NULL,
	instance_id varchar(50) NOT NULL REFERENCES instance (id) ON DELETE CASCADE,
	application_id UUID NOT NULL REFERENCES application (id) ON DELETE CASCADE,
	group_id UUID REFERENCES groups (id) ON DELETE CASCADE
);

CREATE TABLE event_type (
	id serial PRIMARY KEY,
	type integer NOT NULL,
	result integer NOT NULL,
	description varchar(100) NOT NULL
);

CREATE TABLE event (
	id serial PRIMARY KEY,
	created_ts timestamp WITHOUT TIME ZONE DEFAULT (now() at time zone 'utc') NOT NULL,
	previous_version varchar(10),
	error_code varchar(100),
	instance_id varchar(50) NOT NULL REFERENCES instance (id) ON DELETE CASCADE,
	application_id UUID NOT NULL REFERENCES application (id) ON DELETE CASCADE,
	event_type_id integer NOT NULL REFERENCES event_type (id) 
);

CREATE TABLE activity (
	id serial PRIMARY KEY,
	created_ts timestamp WITHOUT TIME ZONE DEFAULT (now() at time zone 'utc') NOT NULL,
	class integer NOT NULL,
	severity integer NOT NULL,
	version SEMVER NOT NULL,
	application_id UUID NOT NULL REFERENCES application (id) ON DELETE CASCADE,
	group_id UUID REFERENCES groups (id) ON DELETE CASCADE,
	channel_id UUID REFERENCES channel (id) ON DELETE CASCADE,
	instance_id varchar(50) REFERENCES instance (id) ON DELETE CASCADE
);

-- Initial data

-- Default team and user (admin/admin)
INSERT INTO team (id, name) VALUES ('d89342dc-9214-441d-a4af-bdd837a3b239', 'default');
INSERT INTO users (username, secret, team_id) VALUES ('admin', '8b31292d4778582c0e5fa96aee5513f1', 'd89342dc-9214-441d-a4af-bdd837a3b239');

-- Event types
INSERT INTO event_type (type, result, description) VALUES (3, 0, 'Instance reported an error during an update step.');
INSERT INTO event_type (type, result, description) VALUES (3, 1, 'Updater has processed and applied package.');
INSERT INTO event_type (type, result, description) VALUES (3, 2, 'Instances upgraded to current channel version.');
INSERT INTO event_type (type, result, description) VALUES (13, 1, 'Downloading latest version.');
INSERT INTO event_type (type, result, description) VALUES (14, 1, 'Update package arrived successfully.');
INSERT INTO event_type (type, result, description) VALUES (800, 1, 'Install success. Update completion prevented by instance.');

-- CoreOS application
INSERT INTO application (id, name, description, team_id) VALUES ('e96281a6-d1af-4bde-9a0a-97b76e56dc57', 'CoreOS', 'Linux for massive server deployments', 'd89342dc-9214-441d-a4af-bdd837a3b239');
INSERT INTO package VALUES ('2ba4c984-5e9b-411e-b7c3-b3eb14f7a261', 1, '766.3.0', 'https://commondatastorage.googleapis.com/update-storage.core-os.net/amd64-usr/766.3.0/', 'update.gz', NULL, '154967458', 'l4Kw7AeBLrVID9JbfyMoJeB5yKg=', '2015-09-20 00:12:37.523938', 'e96281a6-d1af-4bde-9a0a-97b76e56dc57');
INSERT INTO package VALUES ('337b3f7e-ff29-47e8-a052-f0834d25bdb5', 1, '766.4.0', 'https://commondatastorage.googleapis.com/update-storage.core-os.net/amd64-usr/766.4.0/', 'update.gz', NULL, '155018912', 'frkka+B/zTv7OPWgidY+k4SnDSg=', '2015-09-20 06:15:29.108266', 'e96281a6-d1af-4bde-9a0a-97b76e56dc57');
INSERT INTO package VALUES ('c2a36312-b989-403e-ab57-06c055a7eac2', 1, '808.0.0', 'https://commondatastorage.googleapis.com/update-storage.core-os.net/amd64-usr/808.0.0/', 'update.gz', NULL, '177717414', 'bq3fQRHP8xB3RFUjCdAf3wQYC2E=', '2015-09-20 00:09:06.839989', 'e96281a6-d1af-4bde-9a0a-97b76e56dc57');
INSERT INTO package VALUES ('43580892-cad8-468a-a0bb-eb9d0e09eca4', 1, '815.0.0', 'https://commondatastorage.googleapis.com/update-storage.core-os.net/amd64-usr/815.0.0/', 'update.gz', NULL, '178643579', 'kN4amoKYVZUG2WoSdQH1PHPzr5A=', '2015-09-25 13:55:20.741419', 'e96281a6-d1af-4bde-9a0a-97b76e56dc57');
INSERT INTO package VALUES ('284d295b-518f-4d67-999e-94968d0eed90', 1, '829.0.0', 'https://commondatastorage.googleapis.com/update-storage.core-os.net/amd64-usr/829.0.0/', 'update.gz', NULL, '186245514', '2lhoUvvnoY359pi2FnaS/xsgtig=', '2015-10-10 23:11:10.825985', 'e96281a6-d1af-4bde-9a0a-97b76e56dc57');
INSERT INTO channel VALUES ('e06064ad-4414-4904-9a6e-fd465593d1b2', 'stable', '#14b9d6', '2015-09-19 05:09:34.261241', 'e96281a6-d1af-4bde-9a0a-97b76e56dc57', '337b3f7e-ff29-47e8-a052-f0834d25bdb5');
INSERT INTO channel VALUES ('128b8c29-5058-4643-8e67-a1a0e3c641c9', 'beta', '#fc7f33', '2015-09-19 05:09:34.264334', 'e96281a6-d1af-4bde-9a0a-97b76e56dc57', '337b3f7e-ff29-47e8-a052-f0834d25bdb5');
INSERT INTO channel VALUES ('a87a03ad-4984-47a1-8dc4-3507bae91ee1', 'alpha', '#1fbb86', '2015-09-19 05:09:34.265754', 'e96281a6-d1af-4bde-9a0a-97b76e56dc57', '284d295b-518f-4d67-999e-94968d0eed90');
INSERT INTO groups VALUES ('9a2deb70-37be-4026-853f-bfdd6b347bbe', 'Stable', 'For production clusters', false, true, true, false, 'Australia/Sydney', '15 minutes', 2, '60 minutes', '2015-09-19 05:09:34.269062', 'e96281a6-d1af-4bde-9a0a-97b76e56dc57', 'e06064ad-4414-4904-9a6e-fd465593d1b2');
INSERT INTO groups VALUES ('3fe10490-dd73-4b49-b72a-28ac19acfcdc', 'Beta', 'Promoted alpha releases, to catch bugs specific to your configuration', true, true, true, false, 'Australia/Sydney', '15 minutes', 2, '60 minutes', '2015-09-19 05:09:34.273244', 'e96281a6-d1af-4bde-9a0a-97b76e56dc57', '128b8c29-5058-4643-8e67-a1a0e3c641c9');
INSERT INTO groups VALUES ('5b810680-e36a-4879-b98a-4f989e80b899', 'Alpha', 'Tracks current development work and is released frequently', false, true, true, false, 'Australia/Sydney', '15 minutes', 1, '30 minutes', '2015-09-19 05:09:34.274911', 'e96281a6-d1af-4bde-9a0a-97b76e56dc57', 'a87a03ad-4984-47a1-8dc4-3507bae91ee1');
INSERT INTO coreos_action VALUES ('b2b16e2e-57f8-4775-827f-8f0b11ae9bd2', 'postinstall', '', 'k8CB8tMe0M8DyZ5RZwzDLyTdkHjO/YgfKVn2RgUMokc=', false, false, true, '', '', '', '2015-09-20 00:12:37.532281', '2ba4c984-5e9b-411e-b7c3-b3eb14f7a261');
INSERT INTO coreos_action VALUES ('d5a2cbf3-b810-4e8c-88e8-6df91fc264c6', 'postinstall', '', 'QUGnmP51hp7zy+++o5fBIwElInTAms7/njnkxutn/QI=', false, false, true, '', '', '', '2015-09-20 06:15:29.11685', '337b3f7e-ff29-47e8-a052-f0834d25bdb5');
INSERT INTO coreos_action VALUES ('299c54d1-3344-4ae9-8ad2-5c63d56d6c14', 'postinstall', '', 'SCv89GYzx7Ix+TljqbNsd7on65ooWqBzcCrLFL4wChQ=', false, false, true, '', '', '', '2015-09-20 00:09:06.927461', 'c2a36312-b989-403e-ab57-06c055a7eac2');
INSERT INTO coreos_action VALUES ('748df5fc-12a5-4dad-a71e-465cc1668048', 'postinstall', '', '9HUs4whizfyvb4mgl+WaNaW3VLQYwsW1GHNHJNpcFg4=', false, false, true, '', '', '', '2015-09-25 13:55:20.825242', '43580892-cad8-468a-a0bb-eb9d0e09eca4');
INSERT INTO coreos_action VALUES ('9cd474c5-efa3-4989-9992-58ddb852ed84', 'postinstall', '', '1S9zQCLGjmefYnE/aFcpCjL1NsguHhQGj0UCm5f0M98=', false, false, true, '', '', '', '2015-10-10 23:11:10.913778', '284d295b-518f-4d67-999e-94968d0eed90');

-- Sample application 1
INSERT INTO application (id, name, description, team_id) VALUES ('b6458005-8f40-4627-b33b-be70a718c48e', 'Sample application', 'Just an application to show how cool CoreRoller is :)', 'd89342dc-9214-441d-a4af-bdd837a3b239');
INSERT INTO package (id, type, url, filename, version, application_id) VALUES ('5195d5a2-5f82-11e5-9d70-feff819cdc9f', 4, 'https://coreroller.org/', 'test_1.0.2', '1.0.2', 'b6458005-8f40-4627-b33b-be70a718c48e');
INSERT INTO package (id, type, url, filename, version, application_id) VALUES ('12697fa4-5f83-11e5-9d70-feff819cdc9f', 4, 'https://coreroller.org/', 'test_1.0.3', '1.0.3', 'b6458005-8f40-4627-b33b-be70a718c48e');
INSERT INTO package (id, type, url, filename, version, application_id) VALUES ('8004bad8-5f97-11e5-9d70-feff819cdc9f', 4, 'https://coreroller.org/', 'test_1.0.4', '1.0.4', 'b6458005-8f40-4627-b33b-be70a718c48e');
INSERT INTO channel (id, name, color, application_id, package_id) VALUES ('bfe32b4a-5f8c-11e5-9d70-feff819cdc9f', 'Master', '#00CC00', 'b6458005-8f40-4627-b33b-be70a718c48e', '8004bad8-5f97-11e5-9d70-feff819cdc9f');
INSERT INTO channel (id, name, color, application_id, package_id) VALUES ('cb2deea8-5f83-11e5-9d70-feff819cdc9f', 'Stable', '#0099FF', 'b6458005-8f40-4627-b33b-be70a718c48e', '12697fa4-5f83-11e5-9d70-feff819cdc9f');
INSERT INTO groups VALUES ('bcaa68bc-5f82-11e5-9d70-feff819cdc9f', 'Prod EC2 us-west-2', 'Production servers, west coast', false, true, true, false, 'Australia/Sydney', '15 minutes', 2, '60 minutes', '2015-09-19 05:09:34.269062', 'b6458005-8f40-4627-b33b-be70a718c48e', 'cb2deea8-5f83-11e5-9d70-feff819cdc9f');
INSERT INTO groups VALUES ('7074264a-2070-4b84-96ed-8a269dba5021', 'Prod EC2 us-east-1', 'Production servers, east coast', false, true, true, false, 'Australia/Sydney', '15 minutes', 2, '60 minutes', '2015-09-19 05:09:34.269062', 'b6458005-8f40-4627-b33b-be70a718c48e', 'cb2deea8-5f83-11e5-9d70-feff819cdc9f');
INSERT INTO groups VALUES ('b110813a-5f82-11e5-9d70-feff819cdc9f', 'Qa-Dev', 'QA and development servers, Sydney', false, true, true, false, 'Australia/Sydney', '15 minutes', 2, '60 minutes', '2015-09-19 05:09:34.269062', 'b6458005-8f40-4627-b33b-be70a718c48e', 'bfe32b4a-5f8c-11e5-9d70-feff819cdc9f');
INSERT INTO instance (id, ip) VALUES ('instance1', '10.0.0.1');
INSERT INTO instance (id, ip) VALUES ('instance2', '10.0.0.2');
INSERT INTO instance (id, ip) VALUES ('instance3', '10.0.0.3');
INSERT INTO instance (id, ip) VALUES ('instance4', '10.0.0.4');
INSERT INTO instance (id, ip) VALUES ('instance5', '10.0.0.5');
INSERT INTO instance (id, ip) VALUES ('instance6', '10.0.0.6');
INSERT INTO instance (id, ip) VALUES ('instance7', '10.0.0.7');
INSERT INTO instance (id, ip) VALUES ('instance8', '10.0.0.8');
INSERT INTO instance (id, ip) VALUES ('instance9', '10.0.0.9');
INSERT INTO instance (id, ip) VALUES ('instance10', '10.0.0.10');
INSERT INTO instance (id, ip) VALUES ('instance11', '10.0.0.11');
INSERT INTO instance_application VALUES ('1.0.3', DEFAULT, 4, DEFAULT, DEFAULT, NULL, DEFAULT, 'instance1', 'b6458005-8f40-4627-b33b-be70a718c48e', 'bcaa68bc-5f82-11e5-9d70-feff819cdc9f');
INSERT INTO instance_application VALUES ('1.0.3', DEFAULT, 4, DEFAULT, DEFAULT, NULL, DEFAULT, 'instance2', 'b6458005-8f40-4627-b33b-be70a718c48e', 'bcaa68bc-5f82-11e5-9d70-feff819cdc9f');
INSERT INTO instance_application VALUES ('1.0.2', DEFAULT, 4, DEFAULT, DEFAULT, NULL, DEFAULT, 'instance3', 'b6458005-8f40-4627-b33b-be70a718c48e', 'bcaa68bc-5f82-11e5-9d70-feff819cdc9f');
INSERT INTO instance_application VALUES ('1.0.3', DEFAULT, 4, DEFAULT, DEFAULT, NULL, DEFAULT, 'instance4', 'b6458005-8f40-4627-b33b-be70a718c48e', '7074264a-2070-4b84-96ed-8a269dba5021');
INSERT INTO instance_application VALUES ('1.0.3', DEFAULT, 4, DEFAULT, DEFAULT, NULL, DEFAULT, 'instance5', 'b6458005-8f40-4627-b33b-be70a718c48e', '7074264a-2070-4b84-96ed-8a269dba5021');
INSERT INTO instance_application VALUES ('1.0.2', DEFAULT, 4, DEFAULT, DEFAULT, NULL, DEFAULT, 'instance6', 'b6458005-8f40-4627-b33b-be70a718c48e', '7074264a-2070-4b84-96ed-8a269dba5021');
INSERT INTO instance_application VALUES ('1.0.1', DEFAULT, 3, DEFAULT, DEFAULT, NULL, DEFAULT, 'instance7', 'b6458005-8f40-4627-b33b-be70a718c48e', '7074264a-2070-4b84-96ed-8a269dba5021');
INSERT INTO instance_application VALUES ('1.0.4', DEFAULT, 4, DEFAULT, DEFAULT, NULL, DEFAULT, 'instance8', 'b6458005-8f40-4627-b33b-be70a718c48e', 'b110813a-5f82-11e5-9d70-feff819cdc9f');
INSERT INTO instance_application VALUES ('1.0.3', DEFAULT, 7, DEFAULT, DEFAULT, NULL, DEFAULT, 'instance9', 'b6458005-8f40-4627-b33b-be70a718c48e', 'b110813a-5f82-11e5-9d70-feff819cdc9f');
INSERT INTO instance_application VALUES ('1.0.2', DEFAULT, 2, DEFAULT, DEFAULT, NULL, DEFAULT, 'instance10', 'b6458005-8f40-4627-b33b-be70a718c48e', 'b110813a-5f82-11e5-9d70-feff819cdc9f');
INSERT INTO instance_application VALUES ('1.0.1', DEFAULT, 3, DEFAULT, DEFAULT, NULL, DEFAULT, 'instance11', 'b6458005-8f40-4627-b33b-be70a718c48e', 'b110813a-5f82-11e5-9d70-feff819cdc9f');
INSERT INTO activity VALUES (DEFAULT, now() at time zone 'utc' - interval '3 hours', 1, 4, '1.0.3', 'b6458005-8f40-4627-b33b-be70a718c48e', '7074264a-2070-4b84-96ed-8a269dba5021', 'cb2deea8-5f83-11e5-9d70-feff819cdc9f', 'instance1');
INSERT INTO activity VALUES (DEFAULT, now() at time zone 'utc' - interval '6 hours', 5, 3, '1.0.3', 'b6458005-8f40-4627-b33b-be70a718c48e', '7074264a-2070-4b84-96ed-8a269dba5021', 'cb2deea8-5f83-11e5-9d70-feff819cdc9f', 'instance1');
INSERT INTO activity VALUES (DEFAULT, now() at time zone 'utc' - interval '12 hours', 3, 1, '1.0.3', 'b6458005-8f40-4627-b33b-be70a718c48e', '7074264a-2070-4b84-96ed-8a269dba5021', 'cb2deea8-5f83-11e5-9d70-feff819cdc9f', 'instance1');
INSERT INTO activity VALUES (DEFAULT, now() at time zone 'utc' - interval '18 hours', 4, 4, '1.0.3', 'b6458005-8f40-4627-b33b-be70a718c48e', '7074264a-2070-4b84-96ed-8a269dba5021', 'cb2deea8-5f83-11e5-9d70-feff819cdc9f', 'instance1');
INSERT INTO activity VALUES (DEFAULT, now() at time zone 'utc' - interval '24 hours', 2, 2, '1.0.3', 'b6458005-8f40-4627-b33b-be70a718c48e', '7074264a-2070-4b84-96ed-8a269dba5021', 'cb2deea8-5f83-11e5-9d70-feff819cdc9f', 'instance1');
INSERT INTO instance_status_history VALUES (DEFAULT, 4, '1.0.3', now() at time zone 'utc' - interval '8 hours', 'instance5', 'b6458005-8f40-4627-b33b-be70a718c48e', '7074264a-2070-4b84-96ed-8a269dba5021');
INSERT INTO instance_status_history VALUES (DEFAULT, 5, '1.0.3', now() at time zone 'utc' - interval '8 hours 5 minutes', 'instance5', 'b6458005-8f40-4627-b33b-be70a718c48e', '7074264a-2070-4b84-96ed-8a269dba5021');
INSERT INTO instance_status_history VALUES (DEFAULT, 6, '1.0.3', now() at time zone 'utc' - interval '9 hours', 'instance5', 'b6458005-8f40-4627-b33b-be70a718c48e', '7074264a-2070-4b84-96ed-8a269dba5021');
INSERT INTO instance_status_history VALUES (DEFAULT, 7, '1.0.3', now() at time zone 'utc' - interval '9 hours 45 minutes', 'instance5', 'b6458005-8f40-4627-b33b-be70a718c48e', '7074264a-2070-4b84-96ed-8a269dba5021');
INSERT INTO instance_status_history VALUES (DEFAULT, 2, '1.0.3', now() at time zone 'utc' - interval '9 hours 45 minutes 10 seconds', 'instance5', 'b6458005-8f40-4627-b33b-be70a718c48e', '7074264a-2070-4b84-96ed-8a269dba5021');
INSERT INTO instance_status_history VALUES (DEFAULT, 4, '1.0.2', now() at time zone 'utc' - interval '36 hours', 'instance5', 'b6458005-8f40-4627-b33b-be70a718c48e', '7074264a-2070-4b84-96ed-8a269dba5021');
INSERT INTO instance_status_history VALUES (DEFAULT, 5, '1.0.2', now() at time zone 'utc' - interval '36 hours 5 minutes', 'instance5', 'b6458005-8f40-4627-b33b-be70a718c48e', '7074264a-2070-4b84-96ed-8a269dba5021');
INSERT INTO instance_status_history VALUES (DEFAULT, 6, '1.0.2', now() at time zone 'utc' - interval '37 hours', 'instance5', 'b6458005-8f40-4627-b33b-be70a718c48e', '7074264a-2070-4b84-96ed-8a269dba5021');
INSERT INTO instance_status_history VALUES (DEFAULT, 7, '1.0.2', now() at time zone 'utc' - interval '37 hours 45 minutes', 'instance5', 'b6458005-8f40-4627-b33b-be70a718c48e', '7074264a-2070-4b84-96ed-8a269dba5021');
INSERT INTO instance_status_history VALUES (DEFAULT, 2, '1.0.2', now() at time zone 'utc' - interval '37 hours 45 minutes 10 seconds', 'instance5', 'b6458005-8f40-4627-b33b-be70a718c48e', '7074264a-2070-4b84-96ed-8a269dba5021');

-- Sample application 2
INSERT INTO application (id, name, description, team_id) VALUES ('780d6940-9a48-4414-88df-95ba63bbe9cb', 'Sample application 2', 'Another sample application, feel free to remove me', 'd89342dc-9214-441d-a4af-bdd837a3b239');
INSERT INTO package (id, type, url, filename, version, application_id) VALUES ('efb186c9-d5cb-4df2-9382-c4821e4dcc4b', 4, 'http://localhost:8000/', 'demo_v1.0.0', '1.0.0', '780d6940-9a48-4414-88df-95ba63bbe9cb');
INSERT INTO package (id, type, url, filename, version, application_id) VALUES ('ba28af48-b5b9-460e-946a-eba906ce7daf', 4, 'http://localhost:8000/', 'demo_v1.0.1', '1.0.1', '780d6940-9a48-4414-88df-95ba63bbe9cb');
INSERT INTO channel (id, name, color, application_id, package_id) VALUES ('a7c8c9a4-d2a3-475d-be64-911ff8d6e997', 'Master', '#14b9d6', '780d6940-9a48-4414-88df-95ba63bbe9cb', 'efb186c9-d5cb-4df2-9382-c4821e4dcc4b');
INSERT INTO groups VALUES ('51a32aa9-3552-49fc-a28c-6543bccf0069', 'Master - dev', 'The latest stuff will be always here', false, true, true, false, 'Australia/Sydney', '15 minutes', 2, '60 minutes', '2015-09-19 05:09:34.269062', '780d6940-9a48-4414-88df-95ba63bbe9cb', 'a7c8c9a4-d2a3-475d-be64-911ff8d6e997');

-- +migrate Down

DROP TABLE IF EXISTS team CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS application CASCADE;
DROP TABLE IF EXISTS package CASCADE;
DROP TABLE IF EXISTS coreos_action CASCADE;
DROP TABLE IF EXISTS channel CASCADE;
DROP TABLE IF EXISTS groups CASCADE;
DROP TABLE IF EXISTS instance CASCADE;
DROP TABLE IF EXISTS instance_status CASCADE;
DROP TABLE IF EXISTS instance_application CASCADE;
DROP TABLE IF EXISTS instance_status_history CASCADE;
DROP TABLE IF EXISTS event_type CASCADE;
DROP TABLE IF EXISTS event CASCADE;
DROP TABLE IF EXISTS activity CASCADE;
