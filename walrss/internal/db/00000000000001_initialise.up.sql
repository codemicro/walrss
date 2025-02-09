CREATE TABLE "users"
(
    "id"            VARCHAR NOT NULL,
    "email"         VARCHAR NOT NULL,
    "password"      BLOB,
    "salt"          BLOB,
    "active"        BOOLEAN NOT NULL,
    "schedule_day"  INTEGER,
    "schedule_hour" INTEGER,
    PRIMARY KEY ("id"),
    UNIQUE ("email")
)

--bun:split

CREATE TABLE "feeds"
(
    "id"             VARCHAR NOT NULL,
    "url"            VARCHAR NOT NULL,
    "name"           VARCHAR NOT NULL,
    "user_id"        VARCHAR NOT NULL,
    PRIMARY KEY ("id")
)