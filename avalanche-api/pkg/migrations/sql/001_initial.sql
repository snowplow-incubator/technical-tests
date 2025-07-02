-- +goose Up
CREATE TABLE IF NOT EXISTS attributes (
    id SERIAL UNIQUE,
    name VARCHAR(255) NOT NULL,
    value TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY(name, value)
);

CREATE INDEX idx_attributes_id ON attributes(id);
CREATE INDEX idx_attributes_name ON attributes(name);
CREATE INDEX idx_attributes_created_at ON attributes(created_at);

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS profiles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_profiles_created_at ON profiles(created_at);

CREATE TABLE IF NOT EXISTS profile_attributes (
    id SERIAL PRIMARY KEY,
    profile_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    attribute_id INTEGER NOT NULL REFERENCES attributes(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(profile_id, attribute_id)
);

CREATE INDEX idx_profile_attributes_profile_id ON profile_attributes(profile_id);
CREATE INDEX idx_profile_attributes_attribute_id ON profile_attributes(attribute_id);
CREATE INDEX idx_profile_attributes_created_at ON profile_attributes(created_at);


-- +goose Down
DROP TABLE IF EXISTS events;

DROP TABLE IF EXISTS attributes;

DROP TABLE IF EXISTS profiles;

DROP TABLE IF EXISTS profile_attributes;
