-- Create "users" table
CREATE TABLE "public"."users" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "name" character varying(255) NOT NULL,
  "email" character varying(255) NOT NULL,
  "password" text NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_users_deleted_at" to table: "users"
CREATE INDEX "idx_users_deleted_at" ON "public"."users" ("deleted_at");
-- Create index "idx_users_email" to table: "users"
CREATE UNIQUE INDEX "idx_users_email" ON "public"."users" ("email");
-- Create "finance_categories" table
CREATE TABLE "public"."finance_categories" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "user_id" bigint NOT NULL,
  "name" character varying(100) NOT NULL,
  "type" character varying(20) NOT NULL,
  "description" character varying(500) NULL,
  "color" character varying(7) NULL DEFAULT '#000000',
  "icon" character varying(50) NULL,
  "is_active" boolean NULL DEFAULT true,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_users_categories" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_finance_categories_deleted_at" to table: "finance_categories"
CREATE INDEX "idx_finance_categories_deleted_at" ON "public"."finance_categories" ("deleted_at");
-- Create index "idx_finance_categories_type" to table: "finance_categories"
CREATE INDEX "idx_finance_categories_type" ON "public"."finance_categories" ("type");
-- Create index "idx_finance_categories_user_id" to table: "finance_categories"
CREATE INDEX "idx_finance_categories_user_id" ON "public"."finance_categories" ("user_id");
-- Create "finance_journals" table
CREATE TABLE "public"."finance_journals" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "user_id" bigint NOT NULL,
  "category_id" bigint NOT NULL,
  "type" character varying(20) NOT NULL,
  "amount" numeric(15,2) NOT NULL,
  "title" character varying(255) NOT NULL,
  "description" text NULL,
  "date" date NOT NULL,
  "payment_method" character varying(50) NULL,
  "location" character varying(255) NULL,
  "is_recurring" boolean NULL DEFAULT false,
  "receipt_url" character varying(500) NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_finance_categories_journals" FOREIGN KEY ("category_id") REFERENCES "public"."finance_categories" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_users_journals" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_finance_journals_category_id" to table: "finance_journals"
CREATE INDEX "idx_finance_journals_category_id" ON "public"."finance_journals" ("category_id");
-- Create index "idx_finance_journals_date" to table: "finance_journals"
CREATE INDEX "idx_finance_journals_date" ON "public"."finance_journals" ("date");
-- Create index "idx_finance_journals_deleted_at" to table: "finance_journals"
CREATE INDEX "idx_finance_journals_deleted_at" ON "public"."finance_journals" ("deleted_at");
-- Create index "idx_finance_journals_type" to table: "finance_journals"
CREATE INDEX "idx_finance_journals_type" ON "public"."finance_journals" ("type");
-- Create index "idx_finance_journals_user_id" to table: "finance_journals"
CREATE INDEX "idx_finance_journals_user_id" ON "public"."finance_journals" ("user_id");
