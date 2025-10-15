CREATE TABLE IF NOT EXISTS exchange_rate_update 
(
	id TEXT NOT NULL PRIMARY KEY,
	from_currency TEXT NOT NULL,
	to_currency TEXT NOT NULL,
	status INTEGER NOT NULL,
	rate_value DECIMAL(18, 6),
	update_time TIMESTAMP
);

CREATE TABLE IF NOT EXISTS exchange_rate
(
	from_currency TEXT NOT NULL,
	to_currency TEXT NOT NULL,
	rate_value DECIMAL(18, 6),
	update_time TIMESTAMP,
	PRIMARY KEY (from_currency, to_currency)
);

CREATE INDEX IF NOT EXISTS exchange_rate_update_idempotency_index
ON exchange_rate_update(from_currency, to_currency, status);

CREATE INDEX IF NOT EXISTS exchange_rate_get_index
ON exchange_rate(from_currency, to_currency, update_time);