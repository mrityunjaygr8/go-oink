CREATE TABLE "oinks" (
  "name" varchar NOT NULL,
  "id" uuid PRIMARY KEY NOT NULL,
  "description" varchar,
  "creator" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL
);

ALTER TABLE "oinks" ADD CONSTRAINT "fk_oinks_user" FOREIGN KEY ("creator") REFERENCES "users" ("id") ON DELETE CASCADE;
