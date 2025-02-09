CREATE TABLE "feed_items"
(
    feed_id VARCHAR NOT NULL,
    item_id VARCHAR NOT NULL,
    PRIMARY KEY ("feed_id", "item_id"),
    FOREIGN KEY ("feed_id") REFERENCES "feeds"("id") ON DELETE CASCADE
);
