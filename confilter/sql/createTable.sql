-- 产品线表
CREATE TABLE `products` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
  `product` varchar(64) NOT NULL COMMENT '产品线',
  `is_delete` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否删除',
  `created_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `modified_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '修改时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_p` (`product`),
  KEY `idx_i` (`is_delete`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 产品线的字典表
CREATE TABLE `dics` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
  `product_id` bigint(20) unsigned NOT NULL COMMENT '隶属的产品线id',
  `dic_name` varchar(128) NOT NULL COMMENT '字典名称',
  `match_type` tinyint(1) NOT NULL COMMENT '匹配类型，1表示单模匹配 2表示多模匹配',
  `distance` smallint(5) unsigned NOT NULL COMMENT '多模匹配模式的时候的距离',
  `is_delete` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否删除',
  `created_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `modified_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '修改时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_pn` (`product_id`,`dic_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 所有的词
CREATE TABLE `words` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
  `dic_id` bigint(20) unsigned NOT NULL COMMENT '隶属的dic_id',
  `word` varchar(128) NOT NULL COMMENT '词条',
  `submitter` varchar(16) NOT NULL DEFAULT '' COMMENT '提交人',
  `description` varchar(32) NOT NULL DEFAULT '' COMMENT '描述',
  `created_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `modified_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '修改时间',
  `disable_time` datetime NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '失效时间, 如果是1970-01-01 00:00:00，那么表示永不失效',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_dw` (`dic_id`,`word`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 词典组
CREATE TABLE `groups` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
  `product_id` bigint(20) unsigned NOT NULL COMMENT '隶属的产品线id',
  `group_name` varchar(64) NOT NULL COMMENT '词典组名称',
  `is_delete` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否删除',
  `created_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `modified_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '修改时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_pg` (`product_id`,`group_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 词典组到词典的映射表
CREATE TABLE `group_dic` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
  `group_id` bigint(20) unsigned NOT NULL COMMENT '词典组id',
  `dic_id` bigint(20) unsigned NOT NULL COMMENT '词典id',
  `is_delete` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否删除',
  `created_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `modified_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '修改时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_gd` (`group_id`,`dic_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- todo design
-- table sync_job (一个产品线，一段时间内积累的可以同步的job ， 包括产品线，词典组，词典的新增，删除和修改)
