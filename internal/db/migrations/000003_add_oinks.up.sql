CREATE TABLE "oinks" (
  "name" varchar NOT NULL,
  "id" uuid PRIMARY KEY NOT NULL,
  "description" varchar,
  "creator" uuid NOT NULL,
  "created_at" timestamptz,
  "updated_at" timestamptz
);

ALTER TABLE "oinks" ADD CONSTRAINT "fk_oinks_user" FOREIGN KEY ("creator") REFERENCES "users" ("id") ON DELETE CASCADE;
