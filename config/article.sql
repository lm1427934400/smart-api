-- 创建文章内容表
CREATE TABLE IF NOT EXISTS `article_contents` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `title` varchar(255) NOT NULL COMMENT '标题',
  `content` longtext COMMENT 'Markdown内容',
  `html_content` longtext COMMENT 'HTML内容',
  `status` tinyint(4) DEFAULT 1 COMMENT '状态：1-正常，0-禁用',
  `create_by` int(11) DEFAULT NULL COMMENT '创建者ID',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_by` int(11) DEFAULT NULL COMMENT '更新者ID',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  `cover` varchar(500) DEFAULT NULL COMMENT '封面图',
  `summary` varchar(500) DEFAULT NULL COMMENT '摘要',
  `view_count` int(11) DEFAULT 0 COMMENT '浏览次数',
  `parent_id` int(11) DEFAULT NULL COMMENT '父节点ID，用于树形结构',
  `type` varchar(20) DEFAULT 'article' COMMENT '类型：article-文章，directory-目录',
  PRIMARY KEY (`id`),
  KEY `idx_create_by` (`create_by`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_deleted_at` (`deleted_at`),
  KEY `idx_parent_id` (`parent_id`),
  KEY `idx_type` (`type`),
  CONSTRAINT `fk_article_parent` FOREIGN KEY (`parent_id`) REFERENCES `article_contents` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文章内容表(支持树形结构)';

-- 如果存在旧的sys_content表，可以选择重命名或删除
-- RENAME TABLE IF EXISTS `sys_content` TO `article_contents`;
-- 或者
-- DROP TABLE IF EXISTS `sys_content`;
