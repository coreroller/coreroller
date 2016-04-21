-- +migrate Up

CREATE TABLE package_channel_blacklist (
	package_id UUID NOT NULL REFERENCES package (id) ON DELETE CASCADE,
	channel_id UUID NOT NULL REFERENCES channel (id) ON DELETE CASCADE,
	PRIMARY KEY (package_id, channel_id)
);

-- +migrate Down

DROP TABLE IF EXISTS package_channel_blacklist CASCADE;
