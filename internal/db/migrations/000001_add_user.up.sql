CREATE TABLE IF NOT EXISTS "users" (
  "email" varchar UNIQUE NOT NULL,
  "id" uuid PRIMARY KEY UNIQUE NOT NULL,
  "password" varchar NOT NULL,
  "username" varchar UNIQUE NOT NULL,
  "created_at" TIMESTAMPTZ NOT NULL,
  "updated_at" TIMESTAMPTZ NOT NULL
);

