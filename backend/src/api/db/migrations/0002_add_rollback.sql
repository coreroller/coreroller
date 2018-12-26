-- +migrate Up

-- CoreRoller schema

alter table groups add column policy_rollback_allowed boolean not null default false;