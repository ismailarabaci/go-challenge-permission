-- Users table
CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- User groups table
CREATE TABLE IF NOT EXISTS user_groups (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- User to group membership (direct membership only)
CREATE TABLE IF NOT EXISTS user_group_members (
    user_id INT NOT NULL,
    user_group_id INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, user_group_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (user_group_id) REFERENCES user_groups(id) ON DELETE CASCADE,
    INDEX idx_user_group_id (user_group_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- User group hierarchy (groups within groups)
CREATE TABLE IF NOT EXISTS user_group_hierarchy (
    child_group_id INT NOT NULL,
    parent_group_id INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (child_group_id, parent_group_id),
    FOREIGN KEY (child_group_id) REFERENCES user_groups(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_group_id) REFERENCES user_groups(id) ON DELETE CASCADE,
    INDEX idx_parent_group_id (parent_group_id),
    CHECK (child_group_id != parent_group_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Permissions table
CREATE TABLE IF NOT EXISTS permissions (
    source_type ENUM('user', 'group') NOT NULL,
    source_id INT NOT NULL,
    target_type ENUM('user', 'group') NOT NULL,
    target_id INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (source_type, source_id, target_type, target_id),
    INDEX idx_source (source_type, source_id),
    INDEX idx_target (target_type, target_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

