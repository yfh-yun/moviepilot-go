#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
MoviePilot 数据迁移工具
用于将 SQLite 数据库迁移到 PostgreSQL 数据库
"""

import os
import sys
import argparse
import logging
import sqlite3
import psycopg2
from psycopg2 import sql
from psycopg2.extras import execute_values

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler('migrate.log'),
        logging.StreamHandler()
    ]
)
logger = logging.getLogger(__name__)

# 表映射关系
table_mappings = {
    # 用户相关表
    'users': 'users',
    'roles': 'roles',
    'permissions': 'permissions',
    'role_permissions': 'role_permissions',
    'user_roles': 'user_roles',
    # 站点相关表
    'sites': 'sites',
    'site_cookies': 'site_cookies',
    'checkin_logs': 'checkin_logs',
    'site_stats': 'site_stats',
    'sync_logs': 'sync_logs',
    # 下载相关表
    'download_tasks': 'download_tasks',
    'download_history': 'download_history',
    # 订阅相关表
    'subscriptions': 'subscriptions',
    'subscription_items': 'subscription_items',
    'subscribe_shares': 'subscribe_shares',
    'rss': 'rss',
    # 媒体相关表
    'media_items': 'media_items',
    'media_versions': 'media_versions',
    'media_files': 'media_files',
    # 刮削相关表
    'metadata_scrapes': 'metadata_scrapes',
    # 认证相关表
    'auth_logs': 'auth_logs',
    # 插件相关表
    'plugin_data': 'plugin_data',
    # 系统配置表
    'system_configs': 'system_configs'
}

class MigrationTool:
    """
    数据迁移工具类
    """
    
    def __init__(self, sqlite_db, pg_conn_str):
        """
        初始化迁移工具
        
        Args:
            sqlite_db: SQLite数据库文件路径
            pg_conn_str: PostgreSQL连接字符串
        """
        self.sqlite_db = sqlite_db
        self.pg_conn_str = pg_conn_str
        self.sqlite_conn = None
        self.pg_conn = None
    
    def connect(self):
        """
        连接到SQLite和PostgreSQL数据库
        """
        try:
            # 连接SQLite数据库
            self.sqlite_conn = sqlite3.connect(self.sqlite_db)
            self.sqlite_conn.row_factory = sqlite3.Row
            logger.info(f"成功连接到SQLite数据库: {self.sqlite_db}")
            
            # 连接PostgreSQL数据库
            self.pg_conn = psycopg2.connect(self.pg_conn_str)
            self.pg_conn.autocommit = False
            logger.info("成功连接到PostgreSQL数据库")
            
            return True
        except sqlite3.Error as e:
            logger.error(f"SQLite连接失败: {e}")
            return False
        except psycopg2.Error as e:
            logger.error(f"PostgreSQL连接失败: {e}")
            return False
    
    def disconnect(self):
        """
        关闭数据库连接
        """
        if self.sqlite_conn:
            self.sqlite_conn.close()
            logger.info("已关闭SQLite连接")
        
        if self.pg_conn:
            self.pg_conn.close()
            logger.info("已关闭PostgreSQL连接")
    
    def get_sqlite_tables(self):
        """
        获取SQLite数据库中的所有表
        
        Returns:
            表名列表
        """
        cursor = self.sqlite_conn.cursor()
        cursor.execute("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%';")
        tables = [row[0] for row in cursor.fetchall()]
        cursor.close()
        return tables
    
    def get_table_columns(self, table_name):
        """
        获取表的列信息
        
        Args:
            table_name: 表名
            
        Returns:
            列名列表
        """
        cursor = self.sqlite_conn.cursor()
        cursor.execute(f"PRAGMA table_info({table_name});")
        columns = [row[1] for row in cursor.fetchall()]
        cursor.close()
        return columns
    
    def migrate_table(self, table_name):
        """
        迁移单个表的数据
        
        Args:
            table_name: 表名
            
        Returns:
            迁移成功的行数
        """
        if table_name not in table_mappings:
            logger.warning(f"忽略未映射的表: {table_name}")
            return 0
        
        pg_table_name = table_mappings[table_name]
        logger.info(f"开始迁移表: {table_name} -> {pg_table_name}")
        
        try:
            # 获取表的列信息
            columns = self.get_table_columns(table_name)
            if not columns:
                logger.warning(f"表 {table_name} 没有列信息")
                return 0
            
            # 构建查询SQL
            select_sql = f"SELECT {', '.join(columns)} FROM {table_name}"
            
            # 获取SQLite数据
            sqlite_cursor = self.sqlite_conn.cursor()
            sqlite_cursor.execute(select_sql)
            rows = sqlite_cursor.fetchall()
            sqlite_cursor.close()
            
            if not rows:
                logger.info(f"表 {table_name} 没有数据")
                return 0
            
            logger.info(f"从 {table_name} 读取到 {len(rows)} 行数据")
            
            # 连接PostgreSQL并执行插入
            pg_cursor = self.pg_conn.cursor()
            
            # 构建插入SQL
            insert_sql = sql.SQL("INSERT INTO {table} ({columns}) VALUES %s ON CONFLICT DO NOTHING").format(
                table=sql.Identifier(pg_table_name),
                columns=sql.SQL(", ").join(map(sql.Identifier, columns))
            )
            
            # 转换数据为元组列表
            data = [tuple(row) for row in rows]
            
            # 执行批量插入
            execute_values(pg_cursor, insert_sql, data)
            
            # 提交事务
            self.pg_conn.commit()
            
            inserted_rows = pg_cursor.rowcount
            logger.info(f"成功迁移 {inserted_rows} 行数据到 {pg_table_name}")
            
            pg_cursor.close()
            return inserted_rows
            
        except sqlite3.Error as e:
            logger.error(f"读取SQLite表 {table_name} 失败: {e}")
            self.pg_conn.rollback()
            return 0
        except psycopg2.Error as e:
            logger.error(f"写入PostgreSQL表 {pg_table_name} 失败: {e}")
            self.pg_conn.rollback()
            return 0
    
    def migrate_all_tables(self):
        """
        迁移所有表的数据
        """
        tables = self.get_sqlite_tables()
        total_rows = 0
        
        for table in tables:
            migrated_rows = self.migrate_table(table)
            total_rows += migrated_rows
        
        logger.info(f"数据迁移完成，总共迁移了 {total_rows} 行数据")
        return total_rows
    
    def create_schema(self):
        """
        在PostgreSQL中创建数据库schema
        """
        try:
            # 读取schema文件
            schema_path = os.path.join(os.path.dirname(__file__), '..', 'database', 'schema.sql')
            if not os.path.exists(schema_path):
                logger.error(f"Schema文件不存在: {schema_path}")
                return False
            
            with open(schema_path, 'r', encoding='utf-8') as f:
                schema_sql = f.read()
            
            # 执行schema创建
            pg_cursor = self.pg_conn.cursor()
            pg_cursor.execute(schema_sql)
            self.pg_conn.commit()
            pg_cursor.close()
            
            logger.info("成功创建PostgreSQL数据库schema")
            return True
        except psycopg2.Error as e:
            logger.error(f"创建PostgreSQL schema失败: {e}")
            self.pg_conn.rollback()
            return False
    
    def run_migration(self, create_schema_first=False):
        """
        执行完整的数据迁移
        
        Args:
            create_schema_first: 是否先创建schema
        """
        if not self.connect():
            return False
        
        try:
            if create_schema_first:
                if not self.create_schema():
                    return False
            
            self.migrate_all_tables()
            return True
        finally:
            self.disconnect()

def main():
    """
    主函数
    """
    parser = argparse.ArgumentParser(description='MoviePilot 数据迁移工具')
    parser.add_argument('--sqlite-db', required=True, help='SQLite数据库文件路径')
    parser.add_argument('--pg-conn', required=True, help='PostgreSQL连接字符串')
    parser.add_argument('--create-schema', action='store_true', help='是否先创建schema')
    
    args = parser.parse_args()
    
    # 检查SQLite数据库文件是否存在
    if not os.path.exists(args.sqlite_db):
        logger.error(f"SQLite数据库文件不存在: {args.sqlite_db}")
        sys.exit(1)
    
    # 创建迁移工具实例
    migrator = MigrationTool(args.sqlite_db, args.pg_conn)
    
    # 执行迁移
    if migrator.run_migration(args.create_schema):
        logger.info("数据迁移成功")
        sys.exit(0)
    else:
        logger.error("数据迁移失败")
        sys.exit(1)

if __name__ == '__main__':
    main()
