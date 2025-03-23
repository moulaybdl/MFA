CREATE TABLE IF NOT EXISTS users (
  id SERIAL PRIMARY KEY ,
  name text NOT NULL,
  email citext UNIQUE NOT NULL,
  phone_number text UNIQUE,
  password_hash bytea NOT NULL,
  otp_activated bool NOT NULL,
  biometric_public_key bytea
)
