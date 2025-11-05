-- Create dictionary table for storing attribute metadata
CREATE TABLE IF NOT EXISTS dictionary_attributes (
    attribute_id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    long_description TEXT,
    group_id VARCHAR(100),
    mask VARCHAR(50),
    domain VARCHAR(100),
    vector TEXT,
    source JSONB,
    sink JSONB,
    derivation JSONB,
    constraints TEXT[],
    default_value TEXT,
    tags TEXT[],
    sensitivity VARCHAR(50),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    version INTEGER DEFAULT 1
);

-- Create index for efficient lookups
CREATE INDEX idx_dictionary_attributes_name ON dictionary_attributes(name);
CREATE INDEX idx_dictionary_attributes_domain ON dictionary_attributes(domain);
CREATE INDEX idx_dictionary_attributes_group ON dictionary_attributes(group_id);

-- Function to auto-update the updated_at timestamp
CREATE OR REPLACE FUNCTION update_modified_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    NEW.version = OLD.version + 1;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to automatically update updated_at and version
CREATE TRIGGER update_dictionary_attributes_modtime
BEFORE UPDATE ON dictionary_attributes
FOR EACH ROW
EXECUTE FUNCTION update_modified_column();
