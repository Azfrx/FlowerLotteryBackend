-- Local development only. This resets one user's complete state for one activity.
-- It also removes exchange history and restores the wallet to the seed balances.
USE flower_lottery;

SET @user_no := 'demo';
SET @activity_code := 'flower-wish';
SET @seed_coin_balance := 10000000;
SET @seed_petal_balance := 0;
SET @user_id := NULL;
SET @activity_id := NULL;

SELECT id INTO @user_id
FROM users
WHERE user_no = @user_no AND deleted_at IS NULL
LIMIT 1;

SELECT id INTO @activity_id
FROM activities
WHERE code = @activity_code AND deleted_at IS NULL
LIMIT 1;

START TRANSACTION;

DELETE candidate
FROM user_chest_candidates AS candidate
JOIN user_chest_opportunities AS opportunity
    ON opportunity.id = candidate.opportunity_id
WHERE opportunity.user_id = @user_id
  AND opportunity.activity_id = @activity_id;

DELETE FROM user_stage_reward_claims
WHERE user_id = @user_id AND activity_id = @activity_id;

DELETE FROM flower_light_records
WHERE user_id = @user_id AND activity_id = @activity_id;

DELETE FROM user_chest_opportunities
WHERE user_id = @user_id AND activity_id = @activity_id;

DELETE lottery_draw
FROM lottery_draws AS lottery_draw
JOIN lottery_orders AS lottery_order
    ON lottery_order.id = lottery_draw.lottery_order_id
WHERE lottery_order.user_id = @user_id
  AND lottery_order.activity_id = @activity_id;

DELETE FROM user_rewards
WHERE user_id = @user_id AND activity_id = @activity_id;

DELETE FROM asset_transactions
WHERE user_id = @user_id AND activity_id = @activity_id;

DELETE FROM exchange_orders
WHERE user_id = @user_id AND activity_id = @activity_id;

DELETE FROM lottery_orders
WHERE user_id = @user_id AND activity_id = @activity_id;

DELETE FROM leaderboard_snapshots
WHERE user_id = @user_id AND activity_id = @activity_id;

DELETE FROM leaderboard_entries
WHERE user_id = @user_id AND activity_id = @activity_id;

DELETE FROM user_activity_rounds
WHERE user_id = @user_id AND activity_id = @activity_id;

UPDATE user_wallets
SET coin_balance = @seed_coin_balance,
    petal_balance = @seed_petal_balance,
    version = version + 1
WHERE user_id = @user_id;

COMMIT;

SELECT u.user_no, w.coin_balance, w.petal_balance, w.version
FROM users AS u
JOIN user_wallets AS w ON w.user_id = u.id
WHERE u.id = @user_id;
