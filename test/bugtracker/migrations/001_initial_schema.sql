CREATE TABLE tenants (
  id UUID NOT NULL PRIMARY KEY,
  "key" VARCHAR(32) NOT NULL,
  "name" VARCHAR(100) NOT NULL,
  created TIMESTAMPTZ NOT NULL,
  updated TIMESTAMPTZ NULL,
  CONSTRAINT uq_tenant_key UNIQUE ("key")
);

CREATE TABLE issues (
  tid UUID NOT NULL,
  id VARCHAR(16) NOT NULL,
  "name" VARCHAR(200) NOT NULL,
  project_key VARCHAR(4) NOT NULL,
  fields JSONB NOT NULL,
  created TIMESTAMPTZ NOT NULL,
  modified TIMESTAMPTZ NULL,
  PRIMARY KEY(tid, id)
);

CREATE TABLE tags (
  tid UUID NOT NULL,
  "key" VARCHAR(16) NOT NULL,
  created TIMESTAMPTZ NOT NULL,
  PRIMARY KEY(tid, "key")
);

CREATE TABLE issue_tags (
  tid UUID NOT NULL,
  issue_id VARCHAR(16) NOT NULL,
  tag_key VARCHAR(16) NOT NULL,
  created TIMESTAMPTZ NOT NULL,
  PRIMARY KEY(tid, issue_id, tag_key)
);

CREATE TABLE projects (
  tid UUID NOT NULL,
  "key" VARCHAR(4) NOT NULL,
  "name" VARCHAR(100) NOT NULL,
  created TIMESTAMPTZ NOT NULL,
  modified TIMESTAMPTZ NULL,
  PRIMARY KEY(tid, "key")
);

CREATE TABLE issue_type (
  tid UUID NOT NULL,
  id UUID NOT NULL,
  "key" VARCHAR(32) NOT NULL,
  PRIMARY KEY(tid, id),
  CONSTRAINT uq_issue_type_key UNIQUE (tid, "key")
);

CREATE TABLE statuses (
  tid UUID NOT NULL,
  issue_type_id UUID NOT NULL,
  "key" VARCHAR(32) NOT NULL,
  created TIMESTAMPTZ NOT NULL,
  PRIMARY KEY(tid, issue_type_id, "key")
);