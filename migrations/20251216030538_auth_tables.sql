-- Modify "users" table
ALTER TABLE "public"."users" ADD COLUMN "email_verified" boolean NULL DEFAULT false, ADD COLUMN "last_login_at" timestamptz NULL;
-- Create "api_keys" table
CREATE TABLE "public"."api_keys" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "user_id" bigint NOT NULL,
  "name" character varying(100) NOT NULL,
  "key" character varying(64) NOT NULL,
  "last_used_at" timestamptz NULL,
  "expires_at" timestamptz NULL,
  "is_active" boolean NULL DEFAULT true,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_users_api_keys" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_api_keys_deleted_at" to table: "api_keys"
CREATE INDEX "idx_api_keys_deleted_at" ON "public"."api_keys" ("deleted_at");
-- Create index "idx_api_keys_key" to table: "api_keys"
CREATE UNIQUE INDEX "idx_api_keys_key" ON "public"."api_keys" ("key");
-- Create index "idx_api_keys_user_id" to table: "api_keys"
CREATE INDEX "idx_api_keys_user_id" ON "public"."api_keys" ("user_id");
-- Create "tokens" table
CREATE TABLE "public"."tokens" (
  "id" character varying(100) NOT NULL,
  "user_id" bigint NOT NULL,
  "type" character varying(20) NOT NULL,
  "expires_at" timestamptz NOT NULL,
  "created_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_users_tokens" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_tokens_expires_at" to table: "tokens"
CREATE INDEX "idx_tokens_expires_at" ON "public"."tokens" ("expires_at");
-- Create index "idx_tokens_user_id" to table: "tokens"
CREATE INDEX "idx_tokens_user_id" ON "public"."tokens" ("user_id");
