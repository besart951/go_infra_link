-- Project schema migration

-- Phases
CREATE TABLE phases (
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    name VARCHAR(255) NOT NULL
);

CREATE UNIQUE INDEX idx_phases_name ON phases(name);
CREATE INDEX idx_phases_deleted_at ON phases(deleted_at);

-- Projects
CREATE TABLE projects (
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'planned',
    start_date TIMESTAMPTZ,
    phase_id UUID NOT NULL,
    creator_id UUID NOT NULL,
    CONSTRAINT fk_projects_phase FOREIGN KEY (phase_id) REFERENCES phases(id),
    CONSTRAINT fk_projects_creator FOREIGN KEY (creator_id) REFERENCES users(id)
);

CREATE INDEX idx_projects_deleted_at ON projects(deleted_at);
CREATE INDEX idx_projects_phase_id ON projects(phase_id);
CREATE INDEX idx_projects_creator_id ON projects(creator_id);
CREATE INDEX idx_projects_status ON projects(status);

-- Project members (many-to-many)
CREATE TABLE project_users (
    project_id UUID NOT NULL,
    user_id UUID NOT NULL,
    PRIMARY KEY (project_id, user_id),
    CONSTRAINT fk_project_users_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    CONSTRAINT fk_project_users_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_project_users_user_id ON project_users(user_id);
