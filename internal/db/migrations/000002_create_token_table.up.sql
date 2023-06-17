CREATE TABLE IF NOT EXISTS "tokens" (
  "token" uuid PRIMARY KEY NOT NULL UNIQUE,
  "user" uuid NOT NULL,
  "type" varchar NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL
);

ALTER TABLE "tokens" ADD CONSTRAINT "fk_token_users" FOREIGN KEY ("user") REFERENCES "users" ("id") ON DELETE CASCADE;
