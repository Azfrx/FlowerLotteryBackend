CREATE DATABASE IF NOT EXISTS flower_lottery CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
USE flower_lottery;

CREATE TABLE users (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '内部主键',
  user_no VARCHAR(64) NOT NULL COMMENT '业务用户ID/登录账号',
  nickname VARCHAR(64) NOT NULL DEFAULT '' COMMENT '昵称',
  avatar_url VARCHAR(512) NOT NULL DEFAULT '' COMMENT '头像地址',
  password_hash VARCHAR(255) NOT NULL COMMENT 'bcrypt密码哈希',
  status TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '状态:1正常,2禁用',
  last_login_at DATETIME(3) NULL COMMENT '最后登录时间',
  remark VARCHAR(500) NOT NULL DEFAULT '' COMMENT '后台备注',
  extra JSON NULL COMMENT '扩展字段',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  UNIQUE KEY uk_users_user_no (user_no),
  KEY idx_users_status_created (status, created_at),
  KEY idx_users_deleted_at (deleted_at)
) ENGINE=InnoDB COMMENT='用户';

CREATE TABLE user_wallets (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  coin_balance BIGINT NOT NULL DEFAULT 0 COMMENT '金币余额',
  petal_balance BIGINT NOT NULL DEFAULT 0 COMMENT '花瓣余额',
  version BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '乐观锁版本',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  UNIQUE KEY uk_wallet_user (user_id),
  CONSTRAINT fk_wallet_user FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB COMMENT='用户钱包';

CREATE TABLE activities (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  code VARCHAR(64) NOT NULL COMMENT '活动编码',
  name VARCHAR(128) NOT NULL COMMENT '活动名称',
  status TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '0草稿,1预热,2进行中,3结束,4关闭',
  starts_at DATETIME(3) NOT NULL COMMENT '开始时间',
  ends_at DATETIME(3) NOT NULL COMMENT '结束时间',
  leaderboard_freezes_at DATETIME(3) NOT NULL COMMENT '榜单冻结时间',
  timezone VARCHAR(64) NOT NULL DEFAULT 'Asia/Shanghai' COMMENT '业务时区',
  rules_json JSON NULL COMMENT '玩法规则与前端文案',
  resource_json JSON NULL COMMENT '页面与资源扩展',
  remark VARCHAR(500) NOT NULL DEFAULT '',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  UNIQUE KEY uk_activities_code (code),
  KEY idx_activities_status_time (status, starts_at, ends_at)
) ENGINE=InnoDB COMMENT='活动';

CREATE TABLE reward_items (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  item_code VARCHAR(64) NOT NULL COMMENT '道具ID，如1001',
  name VARCHAR(128) NOT NULL COMMENT '奖励名称',
  item_type VARCHAR(32) NOT NULL COMMENT 'coin,petal,item,choice',
  image_url VARCHAR(512) NOT NULL DEFAULT '',
  animation_url VARCHAR(512) NOT NULL DEFAULT '',
  rarity VARCHAR(32) NOT NULL DEFAULT '' COMMENT '品质',
  status TINYINT UNSIGNED NOT NULL DEFAULT 1,
  extra JSON NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  UNIQUE KEY uk_reward_item_code (item_code),
  KEY idx_reward_items_type_status (item_type, status)
) ENGINE=InnoDB COMMENT='奖励道具目录';

CREATE TABLE exchange_options (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  activity_id BIGINT UNSIGNED NOT NULL,
  petal_amount BIGINT UNSIGNED NOT NULL COMMENT '获得花瓣',
  coin_cost BIGINT UNSIGNED NOT NULL COMMENT '消耗金币',
  sort_no INT NOT NULL DEFAULT 0,
  status TINYINT UNSIGNED NOT NULL DEFAULT 1,
  remark VARCHAR(255) NOT NULL DEFAULT '',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  UNIQUE KEY uk_exchange_activity_petal (activity_id, petal_amount),
  CONSTRAINT fk_exchange_activity FOREIGN KEY (activity_id) REFERENCES activities(id)
) ENGINE=InnoDB COMMENT='金币兑换花瓣档位';

CREATE TABLE prize_pools (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  activity_id BIGINT UNSIGNED NOT NULL,
  code VARCHAR(32) NOT NULL COMMENT 'day/night',
  name VARCHAR(64) NOT NULL,
  petal_cost_per_draw BIGINT UNSIGNED NOT NULL,
  coin_value_per_draw BIGINT UNSIGNED NOT NULL COMMENT '用于花朵保底累计',
  supported_draw_counts JSON NOT NULL COMMENT '如[1,10,30]',
  status TINYINT UNSIGNED NOT NULL DEFAULT 1,
  sort_no INT NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  UNIQUE KEY uk_pool_activity_code (activity_id, code),
  CONSTRAINT fk_pool_activity FOREIGN KEY (activity_id) REFERENCES activities(id)
) ENGINE=InnoDB COMMENT='奖池';

CREATE TABLE prize_pool_versions (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  prize_pool_id BIGINT UNSIGNED NOT NULL,
  version_no INT UNSIGNED NOT NULL,
  status TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '0草稿,1已发布,2已停用',
  effective_at DATETIME(3) NULL,
  total_weight BIGINT UNSIGNED NOT NULL DEFAULT 1000000,
  published_by BIGINT UNSIGNED NULL,
  remark VARCHAR(500) NOT NULL DEFAULT '',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  UNIQUE KEY uk_pool_version (prize_pool_id, version_no),
  KEY idx_pool_version_status_effective (prize_pool_id, status, effective_at),
  CONSTRAINT fk_pool_version_pool FOREIGN KEY (prize_pool_id) REFERENCES prize_pools(id)
) ENGINE=InnoDB COMMENT='奖池配置版本';

CREATE TABLE prize_pool_rewards (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  version_id BIGINT UNSIGNED NOT NULL,
  reward_item_id BIGINT UNSIGNED NOT NULL,
  quantity BIGINT UNSIGNED NOT NULL DEFAULT 1,
  weight BIGINT UNSIGNED NOT NULL COMMENT '整数权重',
  choice_group_code VARCHAR(64) NOT NULL DEFAULT '' COMMENT '选择型奖励分组',
  snapshot JSON NULL,
  sort_no INT NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  KEY idx_pool_rewards_version_sort (version_id, sort_no),
  CONSTRAINT fk_pool_rewards_version FOREIGN KEY (version_id) REFERENCES prize_pool_versions(id),
  CONSTRAINT fk_pool_rewards_item FOREIGN KEY (reward_item_id) REFERENCES reward_items(id)
) ENGINE=InnoDB COMMENT='奖池奖励权重';

CREATE TABLE flower_light_rules (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  activity_id BIGINT UNSIGNED NOT NULL,
  flower_position TINYINT UNSIGNED NOT NULL COMMENT '1-18',
  day_probability_ppm INT UNSIGNED NOT NULL COMMENT '白昼百万分比',
  night_probability_ppm INT UNSIGNED NOT NULL COMMENT '星夜百万分比',
  guarantee_coin_total BIGINT UNSIGNED NOT NULL COMMENT '轮次累计金币保底',
  status TINYINT UNSIGNED NOT NULL DEFAULT 1,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  UNIQUE KEY uk_flower_rule_position (activity_id, flower_position),
  CONSTRAINT fk_flower_rule_activity FOREIGN KEY (activity_id) REFERENCES activities(id)
) ENGINE=InnoDB COMMENT='花朵点亮规则';

CREATE TABLE stage_reward_rules (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  activity_id BIGINT UNSIGNED NOT NULL,
  required_flowers TINYINT UNSIGNED NOT NULL,
  reward_item_id BIGINT UNSIGNED NOT NULL,
  quantity BIGINT UNSIGNED NOT NULL DEFAULT 1,
  status TINYINT UNSIGNED NOT NULL DEFAULT 1,
  sort_no INT NOT NULL DEFAULT 0,
  extra JSON NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  UNIQUE KEY uk_stage_activity_threshold (activity_id, required_flowers),
  CONSTRAINT fk_stage_activity FOREIGN KEY (activity_id) REFERENCES activities(id),
  CONSTRAINT fk_stage_reward_item FOREIGN KEY (reward_item_id) REFERENCES reward_items(id)
) ENGINE=InnoDB COMMENT='阶段奖励规则';

CREATE TABLE chest_reward_rules (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  activity_id BIGINT UNSIGNED NOT NULL,
  chest_no TINYINT UNSIGNED NOT NULL COMMENT '1,2,3',
  reward_item_id BIGINT UNSIGNED NOT NULL,
  quantity BIGINT UNSIGNED NOT NULL DEFAULT 1,
  weight BIGINT UNSIGNED NOT NULL DEFAULT 1,
  status TINYINT UNSIGNED NOT NULL DEFAULT 1,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  KEY idx_chest_rule_activity_no (activity_id, chest_no, status),
  CONSTRAINT fk_chest_rule_activity FOREIGN KEY (activity_id) REFERENCES activities(id),
  CONSTRAINT fk_chest_rule_item FOREIGN KEY (reward_item_id) REFERENCES reward_items(id)
) ENGINE=InnoDB COMMENT='宝箱候选奖励规则';

CREATE TABLE asset_transactions (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  activity_id BIGINT UNSIGNED NULL,
  asset_type VARCHAR(16) NOT NULL COMMENT 'coin/petal',
  change_amount BIGINT NOT NULL,
  balance_before BIGINT NOT NULL,
  balance_after BIGINT NOT NULL,
  reason_code VARCHAR(64) NOT NULL,
  biz_type VARCHAR(32) NOT NULL,
  biz_id BIGINT UNSIGNED NULL,
  request_id VARCHAR(64) NOT NULL DEFAULT '',
  remark VARCHAR(255) NOT NULL DEFAULT '',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  KEY idx_asset_user_time (user_id, created_at),
  KEY idx_asset_activity_reason (activity_id, reason_code, created_at),
  KEY idx_asset_biz (biz_type, biz_id),
  CONSTRAINT fk_asset_user FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB COMMENT='不可变资产流水';

CREATE TABLE exchange_orders (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  order_no VARCHAR(32) NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  activity_id BIGINT UNSIGNED NOT NULL,
  exchange_option_id BIGINT UNSIGNED NOT NULL,
  coin_cost BIGINT UNSIGNED NOT NULL,
  petal_amount BIGINT UNSIGNED NOT NULL,
  status TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '1成功,2失败',
  request_id VARCHAR(64) NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  UNIQUE KEY uk_exchange_order_no (order_no),
  UNIQUE KEY uk_exchange_idempotent (user_id, request_id),
  KEY idx_exchange_activity_time (activity_id, created_at),
  CONSTRAINT fk_exchange_order_user FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB COMMENT='花瓣兑换订单';

CREATE TABLE user_activity_rounds (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  activity_id BIGINT UNSIGNED NOT NULL,
  round_no INT UNSIGNED NOT NULL,
  lit_flower_count TINYINT UNSIGNED NOT NULL DEFAULT 0,
  cumulative_coin_value BIGINT UNSIGNED NOT NULL DEFAULT 0,
  chest_granted_count TINYINT UNSIGNED NOT NULL DEFAULT 0,
  chest_processed_count TINYINT UNSIGNED NOT NULL DEFAULT 0,
  status TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '1进行中,2待收尾,3完成',
  completed_at DATETIME(3) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  UNIQUE KEY uk_user_activity_round (user_id, activity_id, round_no),
  KEY idx_round_activity_status (activity_id, status, updated_at),
  CONSTRAINT fk_round_user FOREIGN KEY (user_id) REFERENCES users(id),
  CONSTRAINT fk_round_activity FOREIGN KEY (activity_id) REFERENCES activities(id)
) ENGINE=InnoDB COMMENT='用户活动轮次';

CREATE TABLE lottery_orders (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  order_no VARCHAR(32) NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  activity_id BIGINT UNSIGNED NOT NULL,
  prize_pool_id BIGINT UNSIGNED NOT NULL,
  pool_version_id BIGINT UNSIGNED NOT NULL,
  round_id BIGINT UNSIGNED NOT NULL,
  order_type VARCHAR(16) NOT NULL DEFAULT 'normal' COMMENT 'normal/preview',
  requested_draw_count INT UNSIGNED NOT NULL,
  executed_draw_count INT UNSIGNED NOT NULL DEFAULT 0,
  petal_cost BIGINT UNSIGNED NOT NULL DEFAULT 0,
  petal_refund BIGINT UNSIGNED NOT NULL DEFAULT 0,
  coin_payment BIGINT UNSIGNED NOT NULL DEFAULT 0,
  flowers_before TINYINT UNSIGNED NOT NULL DEFAULT 0,
  flowers_after TINYINT UNSIGNED NOT NULL DEFAULT 0,
  leaderboard_score_added BIGINT UNSIGNED NOT NULL DEFAULT 0,
  status TINYINT UNSIGNED NOT NULL COMMENT '0处理中,1成功,2待支付,3取消,4过期,5失败',
  request_id VARCHAR(64) NOT NULL,
  expires_at DATETIME(3) NULL,
  paid_at DATETIME(3) NULL,
  result_snapshot JSON NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  UNIQUE KEY uk_lottery_order_no (order_no),
  UNIQUE KEY uk_lottery_idempotent (user_id, order_type, request_id),
  KEY idx_lottery_user_time (user_id, created_at),
  KEY idx_lottery_activity_pool_time (activity_id, prize_pool_id, created_at),
  KEY idx_lottery_status_expire (status, expires_at),
  CONSTRAINT fk_lottery_user FOREIGN KEY (user_id) REFERENCES users(id),
  CONSTRAINT fk_lottery_round FOREIGN KEY (round_id) REFERENCES user_activity_rounds(id)
) ENGINE=InnoDB COMMENT='抽奖与先抽后付订单';

CREATE TABLE lottery_draws (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  lottery_order_id BIGINT UNSIGNED NOT NULL,
  draw_index INT UNSIGNED NOT NULL,
  reward_item_id BIGINT UNSIGNED NOT NULL,
  reward_quantity BIGINT UNSIGNED NOT NULL,
  reward_snapshot JSON NOT NULL,
  random_value BIGINT UNSIGNED NOT NULL,
  flower_lit TINYINT UNSIGNED NOT NULL DEFAULT 0,
  flower_position TINYINT UNSIGNED NULL,
  flower_random_value INT UNSIGNED NULL,
  flower_probability_ppm INT UNSIGNED NULL,
  guarantee_triggered TINYINT UNSIGNED NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  UNIQUE KEY uk_draw_order_index (lottery_order_id, draw_index),
  CONSTRAINT fk_draw_order FOREIGN KEY (lottery_order_id) REFERENCES lottery_orders(id),
  CONSTRAINT fk_draw_reward FOREIGN KEY (reward_item_id) REFERENCES reward_items(id)
) ENGINE=InnoDB COMMENT='单抽明细';

CREATE TABLE flower_light_records (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  activity_id BIGINT UNSIGNED NOT NULL,
  round_id BIGINT UNSIGNED NOT NULL,
  lottery_draw_id BIGINT UNSIGNED NOT NULL,
  flower_position TINYINT UNSIGNED NOT NULL,
  trigger_type VARCHAR(16) NOT NULL COMMENT 'probability/guarantee',
  cumulative_coin_value BIGINT UNSIGNED NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  UNIQUE KEY uk_light_round_position (round_id, flower_position),
  KEY idx_light_user_time (user_id, created_at),
  CONSTRAINT fk_light_round FOREIGN KEY (round_id) REFERENCES user_activity_rounds(id)
) ENGINE=InnoDB COMMENT='花朵点亮记录';

CREATE TABLE user_chest_opportunities (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  activity_id BIGINT UNSIGNED NOT NULL,
  round_id BIGINT UNSIGNED NOT NULL,
  chest_no TINYINT UNSIGNED NOT NULL,
  unlock_flower_count TINYINT UNSIGNED NOT NULL,
  status TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '0可开启,1待选择,2已完成,3放弃',
  opened_at DATETIME(3) NULL,
  selected_at DATETIME(3) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  UNIQUE KEY uk_chest_round_no (round_id, chest_no),
  KEY idx_chest_user_status (user_id, status),
  CONSTRAINT fk_chest_round FOREIGN KEY (round_id) REFERENCES user_activity_rounds(id)
) ENGINE=InnoDB COMMENT='用户宝箱机会';

CREATE TABLE user_chest_candidates (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  opportunity_id BIGINT UNSIGNED NOT NULL,
  reward_item_id BIGINT UNSIGNED NOT NULL,
  quantity BIGINT UNSIGNED NOT NULL DEFAULT 1,
  reward_snapshot JSON NOT NULL,
  selected TINYINT UNSIGNED NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  KEY idx_candidate_opportunity (opportunity_id),
  CONSTRAINT fk_candidate_opportunity FOREIGN KEY (opportunity_id) REFERENCES user_chest_opportunities(id)
) ENGINE=InnoDB COMMENT='宝箱候选奖励快照';

CREATE TABLE user_stage_reward_claims (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  activity_id BIGINT UNSIGNED NOT NULL,
  round_id BIGINT UNSIGNED NOT NULL,
  stage_reward_rule_id BIGINT UNSIGNED NOT NULL,
  status TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '1已领取,2发放失败',
  request_id VARCHAR(64) NOT NULL,
  claimed_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  UNIQUE KEY uk_stage_claim (round_id, stage_reward_rule_id),
  UNIQUE KEY uk_stage_claim_request (user_id, request_id),
  CONSTRAINT fk_stage_claim_round FOREIGN KEY (round_id) REFERENCES user_activity_rounds(id)
) ENGINE=InnoDB COMMENT='阶段奖励领取';

CREATE TABLE user_rewards (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  activity_id BIGINT UNSIGNED NOT NULL,
  reward_item_id BIGINT UNSIGNED NOT NULL,
  quantity BIGINT UNSIGNED NOT NULL DEFAULT 1,
  source_type VARCHAR(32) NOT NULL COMMENT 'lottery/chest/stage/leaderboard/admin',
  source_id BIGINT UNSIGNED NULL,
  status TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '0待选择,1已发放,2已使用,3取消,4失败',
  reward_snapshot JSON NOT NULL,
  granted_at DATETIME(3) NULL,
  expires_at DATETIME(3) NULL,
  remark VARCHAR(255) NOT NULL DEFAULT '',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  KEY idx_user_rewards_user_status_time (user_id, status, created_at),
  KEY idx_user_rewards_source (source_type, source_id),
  CONSTRAINT fk_user_reward_user FOREIGN KEY (user_id) REFERENCES users(id),
  CONSTRAINT fk_user_reward_item FOREIGN KEY (reward_item_id) REFERENCES reward_items(id)
) ENGINE=InnoDB COMMENT='用户奖励';

CREATE TABLE leaderboard_entries (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  activity_id BIGINT UNSIGNED NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  score BIGINT UNSIGNED NOT NULL DEFAULT 0,
  reached_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  UNIQUE KEY uk_leaderboard_activity_user (activity_id, user_id),
  KEY idx_leaderboard_rank (activity_id, score DESC, reached_at ASC, user_id ASC),
  CONSTRAINT fk_leaderboard_user FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB COMMENT='实时榜单';

CREATE TABLE leaderboard_snapshots (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  activity_id BIGINT UNSIGNED NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  rank_no INT UNSIGNED NOT NULL,
  score BIGINT UNSIGNED NOT NULL,
  reached_at DATETIME(3) NOT NULL,
  reward_status TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '0未发,1成功,2失败',
  frozen_at DATETIME(3) NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  UNIQUE KEY uk_snapshot_activity_rank (activity_id, rank_no),
  UNIQUE KEY uk_snapshot_activity_user (activity_id, user_id)
) ENGINE=InnoDB COMMENT='榜单冻结快照';

CREATE TABLE admin_users (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  username VARCHAR(64) NOT NULL,
  display_name VARCHAR(64) NOT NULL DEFAULT '',
  password_hash VARCHAR(255) NOT NULL,
  status TINYINT UNSIGNED NOT NULL DEFAULT 1,
  last_login_at DATETIME(3) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  UNIQUE KEY uk_admin_username (username)
) ENGINE=InnoDB COMMENT='管理员';

CREATE TABLE roles (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  code VARCHAR(64) NOT NULL,
  name VARCHAR(64) NOT NULL,
  status TINYINT UNSIGNED NOT NULL DEFAULT 1,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  UNIQUE KEY uk_role_code (code)
) ENGINE=InnoDB COMMENT='角色';

CREATE TABLE permissions (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  code VARCHAR(128) NOT NULL,
  name VARCHAR(128) NOT NULL,
  method VARCHAR(16) NOT NULL DEFAULT '',
  path VARCHAR(255) NOT NULL DEFAULT '',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  UNIQUE KEY uk_permission_code (code)
) ENGINE=InnoDB COMMENT='权限';

CREATE TABLE admin_user_roles (
  admin_user_id BIGINT UNSIGNED NOT NULL,
  role_id BIGINT UNSIGNED NOT NULL,
  PRIMARY KEY (admin_user_id, role_id)
) ENGINE=InnoDB COMMENT='管理员角色关联';

CREATE TABLE role_permissions (
  role_id BIGINT UNSIGNED NOT NULL,
  permission_id BIGINT UNSIGNED NOT NULL,
  PRIMARY KEY (role_id, permission_id)
) ENGINE=InnoDB COMMENT='角色权限关联';

CREATE TABLE refresh_tokens (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  subject_type VARCHAR(16) NOT NULL COMMENT 'user/admin',
  subject_id BIGINT UNSIGNED NOT NULL,
  token_hash CHAR(64) NOT NULL,
  expires_at DATETIME(3) NOT NULL,
  revoked_at DATETIME(3) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  UNIQUE KEY uk_refresh_token_hash (token_hash),
  KEY idx_refresh_subject (subject_type, subject_id, expires_at)
) ENGINE=InnoDB COMMENT='刷新令牌';

CREATE TABLE system_configs (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  config_key VARCHAR(128) NOT NULL,
  config_value JSON NOT NULL,
  description VARCHAR(255) NOT NULL DEFAULT '',
  updated_by BIGINT UNSIGNED NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  UNIQUE KEY uk_system_config_key (config_key)
) ENGINE=InnoDB COMMENT='系统配置';

CREATE TABLE admin_operation_logs (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  admin_user_id BIGINT UNSIGNED NOT NULL,
  request_id VARCHAR(64) NOT NULL DEFAULT '',
  method VARCHAR(16) NOT NULL,
  path VARCHAR(255) NOT NULL,
  action VARCHAR(128) NOT NULL,
  target_type VARCHAR(64) NOT NULL DEFAULT '',
  target_id VARCHAR(64) NOT NULL DEFAULT '',
  request_body JSON NULL,
  response_code INT NOT NULL,
  ip VARCHAR(64) NOT NULL DEFAULT '',
  user_agent VARCHAR(512) NOT NULL DEFAULT '',
  duration_ms INT UNSIGNED NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  KEY idx_admin_log_admin_time (admin_user_id, created_at),
  KEY idx_admin_log_target (target_type, target_id),
  KEY idx_admin_log_request (request_id)
) ENGINE=InnoDB COMMENT='后台操作日志';
