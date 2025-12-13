-- MoviePilot PostgreSQL Database Schema
-- Version: 2.8.1

-- 1. 用户相关表
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    nickname VARCHAR(100),
    avatar VARCHAR(500),
    status VARCHAR(20) DEFAULT 'active' NOT NULL,
    last_login_at TIMESTAMP,
    last_login_ip VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);

CREATE TABLE IF NOT EXISTS roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    status VARCHAR(20) DEFAULT 'active' NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS permissions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    resource VARCHAR(50) NOT NULL,
    action VARCHAR(20) NOT NULL,
    status VARCHAR(20) DEFAULT 'active' NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS role_permissions (
    role_id INT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id INT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE IF NOT EXISTS user_roles (
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id INT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

-- 2. 站点相关表
CREATE TABLE IF NOT EXISTS sites (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    url VARCHAR(500) NOT NULL,
    type VARCHAR(20) NOT NULL,
    priority INT DEFAULT 5,
    enabled BOOLEAN DEFAULT TRUE NOT NULL,
    cookie TEXT,
    user_agent VARCHAR(500),
    proxy VARCHAR(200),
    checkin_enabled BOOLEAN DEFAULT FALSE NOT NULL,
    checkin_cron VARCHAR(50) DEFAULT '0 8 * * *',
    checkin_url VARCHAR(500),
    upload BIGINT DEFAULT 0,
    download BIGINT DEFAULT 0,
    ratio DECIMAL(10,2) DEFAULT 0,
    status VARCHAR(20) DEFAULT 'active' NOT NULL,
    last_checkin TIMESTAMP,
    last_sync TIMESTAMP,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_sites_user_id ON sites(user_id);
CREATE INDEX IF NOT EXISTS idx_sites_deleted_at ON sites(deleted_at);
CREATE INDEX IF NOT EXISTS idx_sites_enabled ON sites(enabled);

CREATE TABLE IF NOT EXISTS site_cookies (
    id SERIAL PRIMARY KEY,
    site_id INT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    value VARCHAR(1000) NOT NULL,
    domain VARCHAR(100),
    path VARCHAR(100),
    expires_at TIMESTAMP,
    http_only BOOLEAN DEFAULT FALSE,
    secure BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_site_cookies_site_id ON site_cookies(site_id);

CREATE TABLE IF NOT EXISTS checkin_logs (
    id SERIAL PRIMARY KEY,
    site_id INT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL,
    message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_checkin_logs_site_id ON checkin_logs(site_id);

CREATE TABLE IF NOT EXISTS site_stats (
    id SERIAL PRIMARY KEY,
    site_id INT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    upload BIGINT DEFAULT 0,
    download BIGINT DEFAULT 0,
    ratio DECIMAL(10,2) DEFAULT 0,
    checkin_status VARCHAR(20),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_site_stats_site_id ON site_stats(site_id);

CREATE TABLE IF NOT EXISTS sync_logs (
    id SERIAL PRIMARY KEY,
    site_id INT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL,
    message TEXT,
    items_count INT DEFAULT 0,
    duration INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_sync_logs_site_id ON sync_logs(site_id);

-- 3. 下载相关表
CREATE TABLE IF NOT EXISTS download_tasks (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subscription_id INT,
    downloader_type VARCHAR(20) NOT NULL,
    hash VARCHAR(64) UNIQUE NOT NULL,
    name VARCHAR(200) NOT NULL,
    save_path VARCHAR(500),
    size BIGINT DEFAULT 0,
    status VARCHAR(20) DEFAULT 'downloading' NOT NULL,
    progress DECIMAL(5,2) DEFAULT 0 NOT NULL,
    download_speed BIGINT DEFAULT 0,
    upload_speed BIGINT DEFAULT 0,
    downloaded BIGINT DEFAULT 0,
    uploaded BIGINT DEFAULT 0,
    ratio DECIMAL(5,2) DEFAULT 0,
    error_message TEXT,
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_download_tasks_user_id ON download_tasks(user_id);
CREATE INDEX IF NOT EXISTS idx_download_tasks_hash ON download_tasks(hash);
CREATE INDEX IF NOT EXISTS idx_download_tasks_status ON download_tasks(status);
CREATE INDEX IF NOT EXISTS idx_download_tasks_created_at ON download_tasks(created_at);

CREATE TABLE IF NOT EXISTS download_history (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    task_id INT REFERENCES download_tasks(id) ON DELETE SET NULL,
    hash VARCHAR(64) NOT NULL,
    name VARCHAR(200) NOT NULL,
    size BIGINT,
    downloader_type VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL,
    download_time INT,
    average_speed BIGINT,
    final_ratio DECIMAL(5,2),
    save_path VARCHAR(500),
    error_message TEXT,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_download_history_user_id ON download_history(user_id);
CREATE INDEX IF NOT EXISTS idx_download_history_hash ON download_history(hash);
CREATE INDEX IF NOT EXISTS idx_download_history_status ON download_history(status);

-- 4. 订阅相关表
CREATE TABLE IF NOT EXISTS subscriptions (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    media_type VARCHAR(20) NOT NULL, -- movie, tv, anime
    tmdb_id INT NOT NULL,
    title VARCHAR(200) NOT NULL,
    year INT,
    season INT,
    episode INT,
    status VARCHAR(20) DEFAULT 'active' NOT NULL,
    priority INT DEFAULT 5,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_deleted_at ON subscriptions(deleted_at);
CREATE INDEX IF NOT EXISTS idx_subscriptions_tmdb_id ON subscriptions(tmdb_id);

CREATE TABLE IF NOT EXISTS subscription_items (
    id SERIAL PRIMARY KEY,
    subscription_id INT NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE,
    item_type VARCHAR(20) NOT NULL, -- season, episode
    season INT,
    episode INT,
    status VARCHAR(20) DEFAULT 'pending' NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_subscription_items_subscription_id ON subscription_items(subscription_id);

CREATE TABLE IF NOT EXISTS subscribe_shares (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    share_id VARCHAR(100) UNIQUE NOT NULL,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    media_type VARCHAR(20) NOT NULL,
    data JSONB NOT NULL,
    status VARCHAR(20) DEFAULT 'active' NOT NULL,
    expire_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_subscribe_shares_user_id ON subscribe_shares(user_id);
CREATE INDEX IF NOT EXISTS idx_subscribe_shares_share_id ON subscribe_shares(share_id);

CREATE TABLE IF NOT EXISTS rss (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    url VARCHAR(500) NOT NULL,
    enabled BOOLEAN DEFAULT TRUE NOT NULL,
    priority INT DEFAULT 5,
    last_sync_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_rss_user_id ON rss(user_id);
CREATE INDEX IF NOT EXISTS idx_rss_deleted_at ON rss(deleted_at);

-- 5. 媒体相关表
CREATE TABLE IF NOT EXISTS media_items (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tmdb_id INT NOT NULL,
    media_type VARCHAR(20) NOT NULL, -- movie, tv, anime
    title VARCHAR(200) NOT NULL,
    original_title VARCHAR(200),
    year INT,
    overview TEXT,
    poster_path VARCHAR(500),
    backdrop_path VARCHAR(500),
    rating DECIMAL(3,1),
    release_date DATE,
    status VARCHAR(20), -- released, upcoming, in production
    genres TEXT[],
    countries TEXT[],
    languages TEXT[],
    runtime INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_media_items_user_id ON media_items(user_id);
CREATE INDEX IF NOT EXISTS idx_media_items_tmdb_id ON media_items(tmdb_id);
CREATE INDEX IF NOT EXISTS idx_media_items_media_type ON media_items(media_type);

CREATE TABLE IF NOT EXISTS media_versions (
    id SERIAL PRIMARY KEY,
    media_id INT NOT NULL REFERENCES media_items(id) ON DELETE CASCADE,
    season INT,
    episode INT,
    title VARCHAR(200),
    overview TEXT,
    release_date DATE,
    runtime INT,
    episode_count INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_media_versions_media_id ON media_versions(media_id);

CREATE TABLE IF NOT EXISTS media_files (
    id SERIAL PRIMARY KEY,
    media_id INT REFERENCES media_items(id) ON DELETE SET NULL,
    version_id INT REFERENCES media_versions(id) ON DELETE SET NULL,
    file_path VARCHAR(1000) NOT NULL,
    file_name VARCHAR(500) NOT NULL,
    file_size BIGINT NOT NULL,
    file_hash VARCHAR(64),
    video_codec VARCHAR(50),
    audio_codec VARCHAR(50),
    resolution VARCHAR(20),
    duration INT,
    subtitle_languages TEXT[],
    audio_languages TEXT[],
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_media_files_media_id ON media_files(media_id);
CREATE INDEX IF NOT EXISTS idx_media_files_version_id ON media_files(version_id);
CREATE INDEX IF NOT EXISTS idx_media_files_file_path ON media_files(file_path);
CREATE INDEX IF NOT EXISTS idx_media_files_deleted_at ON media_files(deleted_at);

-- 6. 刮削相关表
CREATE TABLE IF NOT EXISTS metadata_scrapes (
    id SERIAL PRIMARY KEY,
    media_id INT NOT NULL REFERENCES media_items(id) ON DELETE CASCADE,
    source VARCHAR(20) NOT NULL, -- tmdb, tvdb, imdb
    data JSONB NOT NULL,
    status VARCHAR(20) NOT NULL,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_metadata_scrapes_media_id ON metadata_scrapes(media_id);

-- 7. 认证相关表
CREATE TABLE IF NOT EXISTS auth_logs (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL,
    ip VARCHAR(50) NOT NULL,
    user_agent VARCHAR(500),
    status VARCHAR(20) NOT NULL,
    message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_auth_logs_created_at ON auth_logs(created_at);

-- 8. 插件相关表
CREATE TABLE IF NOT EXISTS plugin_data (
    id SERIAL PRIMARY KEY,
    plugin_id VARCHAR(50) NOT NULL,
    key VARCHAR(100) NOT NULL,
    value TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    UNIQUE (plugin_id, key)
);

CREATE INDEX IF NOT EXISTS idx_plugin_data_plugin_id ON plugin_data(plugin_id);

-- 9. 系统配置表
CREATE TABLE IF NOT EXISTS system_configs (
    id SERIAL PRIMARY KEY,
    key VARCHAR(100) UNIQUE NOT NULL,
    value TEXT NOT NULL,
    type VARCHAR(20) NOT NULL, -- string, int, bool, json
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- 10. 索引优化
CREATE INDEX IF NOT EXISTS idx_download_tasks_user_id_status ON download_tasks(user_id, status);
CREATE INDEX IF NOT EXISTS idx_media_items_user_id_media_type ON media_items(user_id, media_type);
CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id_status ON subscriptions(user_id, status);
