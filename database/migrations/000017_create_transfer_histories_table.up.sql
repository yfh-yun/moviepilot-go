-- 创建 transfer_histories 表
CREATE TABLE IF NOT EXISTS transfer_histories (
    id SERIAL PRIMARY KEY,
    
    -- 源信息
    src VARCHAR(1000) NOT NULL,
    src_storage VARCHAR(50),
    src_fileitem JSONB,
    source VARCHAR(500) NOT NULL,
    source_path VARCHAR(1000) NOT NULL,
    
    -- 目标信息
    dest VARCHAR(1000) NOT NULL,
    dest_storage VARCHAR(50),
    dest_fileitem JSONB,
    target VARCHAR(500) NOT NULL,
    target_path VARCHAR(1000) NOT NULL,
    
    -- 转移信息
    mode VARCHAR(20),  -- move, copy, link, hardlink
    type VARCHAR(20) NOT NULL,  -- movie, tv
    category VARCHAR(100),
    
    -- 媒体信息
    title VARCHAR(500) NOT NULL,
    year VARCHAR(10),
    tmdb_id INT,
    imdb_id VARCHAR(20),
    tvdb_id INT,
    douban_id VARCHAR(50),
    seasons VARCHAR(100),
    episodes VARCHAR(500),
    image VARCHAR(500),
    episode_group VARCHAR(100),
    
    -- 下载信息
    downloader VARCHAR(50),
    download_hash VARCHAR(100),
    
    -- 状态信息
    status BOOLEAN DEFAULT true,
    errmsg TEXT,
    date VARCHAR(50),
    
    -- 文件清单
    files JSONB,
    
    -- 用户信息
    userid VARCHAR(100),
    username VARCHAR(100),
    
    -- 其他信息
    note JSONB,
    media_category VARCHAR(100),
    
    -- 时间戳
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_transfer_histories_src ON transfer_histories(src);
CREATE INDEX idx_transfer_histories_title ON transfer_histories(title);
CREATE INDEX idx_transfer_histories_tmdb_id ON transfer_histories(tmdb_id);
CREATE INDEX idx_transfer_histories_tvdb_id ON transfer_histories(tvdb_id);
CREATE INDEX idx_transfer_histories_download_hash ON transfer_histories(download_hash);
CREATE INDEX idx_transfer_histories_date ON transfer_histories(date);
CREATE INDEX idx_transfer_histories_deleted_at ON transfer_histories(deleted_at);

-- 添加注释
COMMENT ON TABLE transfer_histories IS '文件转移历史表';
COMMENT ON COLUMN transfer_histories.id IS '历史ID';
COMMENT ON COLUMN transfer_histories.src IS '源路径';
COMMENT ON COLUMN transfer_histories.dest IS '目标路径';
COMMENT ON COLUMN transfer_histories.mode IS '转移模式';
COMMENT ON COLUMN transfer_histories.type IS '媒体类型';
COMMENT ON COLUMN transfer_histories.title IS '标题';
COMMENT ON COLUMN transfer_histories.status IS '转移状态';
COMMENT ON COLUMN transfer_histories.created_at IS '创建时间';
COMMENT ON COLUMN transfer_histories.updated_at IS '更新时间';
COMMENT ON COLUMN transfer_histories.deleted_at IS '删除时间（软删除）';
