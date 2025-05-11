-- Create extension for UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Drop existing tables if they exist
DROP TABLE IF EXISTS scan_results CASCADE;
DROP TABLE IF EXISTS scan_templates CASCADE;
DROP TABLE IF EXISTS scan_tags CASCADE;
DROP TABLE IF EXISTS templates CASCADE;
DROP TABLE IF EXISTS scans CASCADE;
DROP TYPE IF EXISTS scan_status CASCADE;

-- Create enum for scan status
CREATE TYPE scan_status AS ENUM ('pending', 'running', 'completed', 'failed', 'cancelled');

-- Create scans table
CREATE TABLE IF NOT EXISTS scans (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    target TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    error TEXT,
    options JSONB,
    template_ids TEXT[] NOT NULL DEFAULT '{}',
    tags TEXT[] NOT NULL DEFAULT '{}'
);

-- Create scan_templates table
CREATE TABLE IF NOT EXISTS scan_templates (
    id SERIAL PRIMARY KEY,
    scan_id UUID NOT NULL REFERENCES scans(id) ON DELETE CASCADE,
    template_id TEXT NOT NULL,
    UNIQUE (scan_id, template_id)
);

-- Create scan_tags table
CREATE TABLE IF NOT EXISTS scan_tags (
    id SERIAL PRIMARY KEY,
    scan_id UUID NOT NULL REFERENCES scans(id) ON DELETE CASCADE,
    tag TEXT NOT NULL,
    UNIQUE (scan_id, tag)
);

-- Create templates table
CREATE TABLE IF NOT EXISTS templates (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    author TEXT NOT NULL,
    tags TEXT[] NOT NULL DEFAULT '{}',
    severity TEXT NOT NULL,
    type TEXT NOT NULL,
    description TEXT NOT NULL,
    path TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create scan_results table
CREATE TABLE IF NOT EXISTS scan_results (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    scan_id UUID NOT NULL REFERENCES scans(id) ON DELETE CASCADE,
    template_id TEXT NOT NULL,
    template_name TEXT,
    severity TEXT,
    matched BOOLEAN NOT NULL DEFAULT false,
    host TEXT,
    matched_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    matcher_name TEXT,
    extracted_results JSONB,
    request TEXT,
    response TEXT,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX idx_scans_status ON scans(status);
CREATE INDEX idx_scans_created_at ON scans(created_at);
CREATE INDEX idx_scan_templates_scan_id ON scan_templates(scan_id);
CREATE INDEX idx_scan_templates_template_id ON scan_templates(template_id);
CREATE INDEX idx_scan_tags_scan_id ON scan_tags(scan_id);
CREATE INDEX idx_scan_tags_tag ON scan_tags(tag);
CREATE INDEX idx_templates_tags ON templates USING GIN (tags);
CREATE INDEX idx_templates_severity ON templates (severity);
CREATE INDEX idx_templates_type ON templates (type);
CREATE INDEX idx_templates_author ON templates (author);
CREATE INDEX idx_scan_results_scan_id ON scan_results(scan_id);
CREATE INDEX idx_scan_results_template_id ON scan_results(template_id);
CREATE INDEX idx_scan_results_severity ON scan_results(severity);
CREATE INDEX idx_scan_results_matched_at ON scan_results(matched_at); 