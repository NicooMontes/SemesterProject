-- Create database
CREATE DATABASE cloud_storage_app CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE cloud_storage_app;

-- ============================================
-- Table: Users
-- ============================================
CREATE TABLE Users (
    user_id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    lastname VARCHAR(100) NOT NULL,
    email VARCHAR(150) NOT NULL UNIQUE,
    password_hash VARCHAR(64) NOT NULL,
    creation_date DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_email (email)
) ENGINE=InnoDB;

-- ============================================
-- Table: Files
-- ============================================
CREATE TABLE Files (
    file_id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    size BIGINT NOT NULL,
    upload_date DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_date DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    version INT NOT NULL DEFAULT 1,
    FOREIGN KEY (user_id) REFERENCES Users(user_id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id)
    priority ENUM('low', 'medium', 'high') NOT NULL DEFAULT 'low',
    storage_policy ENUM('public', 'private', 'hybrid') NOT NULL DEFAULT 'public',
    region ENUM('EU', 'GLOBAL') NOT NULL DEFAULT 'GLOBAL',
) ENGINE=InnoDB;

-- ============================================
-- Table: Logs
-- ============================================
CREATE TABLE Logs (
    log_id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    action ENUM('upload', 'download', 'share', 'delete') NOT NULL,
    timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES Users(user_id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id),
    INDEX idx_timestamp (timestamp),
    INDEX idx_action (action)
) ENGINE=InnoDB;

-- ============================================
-- Table: Permissions
-- ============================================
CREATE TABLE Permissions (
    permission_id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    file_id INT NOT NULL,
    access_type ENUM('read', 'write', 'share') NOT NULL,
    grant_date DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES Users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (file_id) REFERENCES Files(file_id) ON DELETE CASCADE,
    UNIQUE KEY unique_user_file_access (user_id, file_id, access_type),
    INDEX idx_user_id (user_id),
    INDEX idx_file_id (file_id)
) ENGINE=InnoDB;

-- ============================================
-- Table: Chunks
-- ============================================
CREATE TABLE Chunks (
    chunk_id INT AUTO_INCREMENT PRIMARY KEY,
    file_id INT NOT NULL,
    chunk_index INT NOT NULL,
    hash VARCHAR(64) NOT NULL,
    storage_path VARCHAR(200) NOT NULL,
    replica_count INT NOT NULL DEFAULT 1,
    pod_location VARCHAR(50) NOT NULL,
    creation_date DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (file_id) REFERENCES Files(file_id) ON DELETE CASCADE,
    UNIQUE KEY unique_file_chunk (file_id, chunk_index),
    INDEX idx_file_id (file_id),
    INDEX idx_hash (hash),
    INDEX idx_pod_location (pod_location)
) ENGINE=InnoDB;
