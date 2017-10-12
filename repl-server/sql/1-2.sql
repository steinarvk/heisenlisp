CREATE TABLE request_log (
  request_id BIGSERIAL PRIMARY KEY,
  timestamp_utcnano BIGINT NOT NULL,
  server_hostname TEXT NOT NULL,
  client_hostname TEXT NOT NULL,
  http_headers JSONB,
  expr TEXT NOT NULL
);

CREATE TABLE result_log (
  result_id BIGSERIAL PRIMARY KEY,
  timestamp_utcnano BIGINT NOT NULL,
  duration_nanos BIGINT NOT NULL,
  request_id BIGINT NOT NULL REFERENCES request_log (request_id) ON DELETE CASCADE,
  result TEXT NULL,
  error TEXT NULL
);
