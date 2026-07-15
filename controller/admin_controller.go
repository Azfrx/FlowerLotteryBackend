package controller

import (
	"flower-lottery-backend/config"
	"flower-lottery-backend/middleware"
	"flower-lottery-backend/model"
	tokenjwt "flower-lottery-backend/pkg/jwt"
	"flower-lottery-backend/response"
	"flower-lottery-backend/service"
	"flower-lottery-backend/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"strings"
	"time"
)

type AdminController struct {
	db          *gorm.DB
	tokens      *tokenjwt.Manager
	poolConfigs *service.PoolConfigHub
}

type adminUserRow struct {
	ID           uint64     `gorm:"column:id" json:"id"`
	UserID       string     `gorm:"column:user_id" json:"user_id"`
	Nickname     string     `gorm:"column:nickname" json:"nickname"`
	AvatarURL    string     `gorm:"column:avatar_url" json:"avatar_url"`
	CoinBalance  int64      `gorm:"column:coin_balance" json:"coin_balance"`
	PetalBalance int64      `gorm:"column:petal_balance" json:"petal_balance"`
	Status       uint8      `gorm:"column:status" json:"status"`
	Remark       string     `gorm:"column:remark" json:"remark"`
	LastLoginAt  *time.Time `gorm:"column:last_login_at" json:"last_login_at"`
	CreatedAt    time.Time  `gorm:"column:created_at" json:"created_at"`
}

type adminAssetRow struct {
	ID           uint64    `gorm:"column:id" json:"id"`
	UserID       string    `gorm:"column:user_id" json:"user_id"`
	Nickname     string    `gorm:"column:nickname" json:"nickname"`
	AssetType    string    `gorm:"column:asset_type" json:"asset_type"`
	ChangeAmount int64     `gorm:"column:change_amount" json:"change_amount"`
	BalanceAfter int64     `gorm:"column:balance_after" json:"balance_after"`
	ReasonCode   string    `gorm:"column:reason_code" json:"reason_code"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
}

type adminLotteryRow struct {
	ID             uint64    `gorm:"column:id" json:"id"`
	OrderNo        string    `gorm:"column:order_no" json:"order_no"`
	UserID         string    `gorm:"column:user_id" json:"user_id"`
	Nickname       string    `gorm:"column:nickname" json:"nickname"`
	OrderType      string    `gorm:"column:order_type" json:"order_type"`
	PoolCode       string    `gorm:"column:pool_code" json:"pool_code"`
	PoolName       string    `gorm:"column:pool_name" json:"pool_name"`
	DrawCount      uint      `gorm:"column:draw_count" json:"draw_count"`
	PetalCost      uint64    `gorm:"column:petal_cost" json:"petal_cost"`
	Rewards        string    `gorm:"column:rewards" json:"rewards"`
	LitFlowerCount uint      `gorm:"column:lit_flower_count" json:"lit_flower_count"`
	Status         uint8     `gorm:"column:status" json:"status"`
	CreatedAt      time.Time `gorm:"column:created_at" json:"created_at"`
}

type adminRewardRow struct {
	ID         uint64    `gorm:"column:id" json:"id"`
	UserID     string    `gorm:"column:user_id" json:"user_id"`
	Nickname   string    `gorm:"column:nickname" json:"nickname"`
	ItemCode   string    `gorm:"column:item_code" json:"item_code"`
	RewardName string    `gorm:"column:reward_name" json:"reward_name"`
	Quantity   uint64    `gorm:"column:quantity" json:"quantity"`
	SourceType string    `gorm:"column:source_type" json:"source_type"`
	Status     uint8     `gorm:"column:status" json:"status"`
	ObtainedAt time.Time `gorm:"column:obtained_at" json:"obtained_at"`
}

type adminRoundRow struct {
	ID                 uint64    `gorm:"column:id" json:"id"`
	UserID             string    `gorm:"column:user_id" json:"user_id"`
	Nickname           string    `gorm:"column:nickname" json:"nickname"`
	RoundNo            uint      `gorm:"column:round_no" json:"round_no"`
	LitFlowerCount     uint8     `gorm:"column:lit_flower_count" json:"lit_flower_count"`
	ChestOpenedCount   int64     `gorm:"column:chest_opened_count" json:"chest_opened_count"`
	StageEligibleCount int64     `gorm:"column:stage_eligible_count" json:"stage_eligible_count"`
	StageClaimedCount  int64     `gorm:"column:stage_claimed_count" json:"stage_claimed_count"`
	StageFailedCount   int64     `gorm:"column:stage_failed_count" json:"stage_failed_count"`
	Status             uint8     `gorm:"column:status" json:"status"`
	UpdatedAt          time.Time `gorm:"column:updated_at" json:"updated_at"`
}

type adminLeaderboardRow struct {
	ID        uint64    `gorm:"column:id" json:"id"`
	Rank      int64     `gorm:"column:rank_no" json:"rank"`
	UserID    string    `gorm:"column:user_id" json:"user_id"`
	Nickname  string    `gorm:"column:nickname" json:"nickname"`
	Score     uint64    `gorm:"column:score" json:"score"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

type adminActivityRow struct {
	ID                   uint64    `gorm:"column:id" json:"id"`
	Code                 string    `gorm:"column:code" json:"code"`
	Name                 string    `gorm:"column:name" json:"name"`
	Status               uint8     `gorm:"column:status" json:"status"`
	StartsAt             time.Time `gorm:"column:starts_at" json:"starts_at"`
	EndsAt               time.Time `gorm:"column:ends_at" json:"ends_at"`
	LeaderboardFreezesAt time.Time `gorm:"column:leaderboard_freezes_at" json:"leaderboard_freezes_at"`
	Timezone             string    `gorm:"column:timezone" json:"timezone"`
	UpdatedAt            time.Time `gorm:"column:updated_at" json:"updated_at"`
}

type adminPoolRow struct {
	ID                  uint64     `gorm:"column:id" json:"id"`
	ActivityName        string     `gorm:"column:activity_name" json:"activity_name"`
	Code                string     `gorm:"column:code" json:"code"`
	Name                string     `gorm:"column:name" json:"name"`
	PetalCostPerDraw    uint64     `gorm:"column:petal_cost_per_draw" json:"petal_cost_per_draw"`
	CoinValuePerDraw    uint64     `gorm:"column:coin_value_per_draw" json:"coin_value_per_draw"`
	SupportedDrawCounts string     `gorm:"column:supported_draw_counts" json:"supported_draw_counts"`
	Status              uint8      `gorm:"column:status" json:"status"`
	VersionNo           uint       `gorm:"column:version_no" json:"version_no"`
	EffectiveAt         *time.Time `gorm:"column:effective_at" json:"effective_at"`
	RewardCount         int64      `gorm:"column:reward_count" json:"reward_count"`
}

func NewAdminController(db *gorm.DB, c config.JWT, poolConfigs *service.PoolConfigHub) *AdminController {
	return &AdminController{db: db, tokens: tokenjwt.New(c), poolConfigs: poolConfigs}
}
func (a *AdminController) Login(c *gin.Context) {
	var q struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if c.ShouldBindJSON(&q) != nil {
		response.Error(c, 400, 10001, "参数错误")
		return
	}
	var u model.AdminUser
	if a.db.Where("username=? AND deleted_at IS NULL", q.Username).First(&u).Error != nil || !utils.CheckPassword(u.PasswordHash, q.Password) || u.Status != 1 {
		response.Error(c, 401, 11001, "账号或密码错误")
		return
	}
	pair, _, e := a.tokens.Issue(u.ID, "admin")
	if e != nil {
		writeError(c, e)
		return
	}
	now := time.Now()
	a.db.Model(&u).Update("last_login_at", now)
	u.LastLoginAt = &now
	response.Success(c, gin.H{"access_token": pair.AccessToken, "refresh_token": pair.RefreshToken, "admin": u})
}
func (a *AdminController) Me(c *gin.Context) {
	var u model.AdminUser
	if a.db.Where("id=? AND status=1 AND deleted_at IS NULL", middleware.CurrentAdminID(c)).First(&u).Error != nil {
		response.Error(c, 401, 11003, "登录状态已失效")
		return
	}
	response.Success(c, u)
}
func (a *AdminController) Users(c *gin.Context) {
	var list []adminUserRow
	query := a.db.Table("users AS u").
		Select("u.id AS id,u.user_no AS user_id,u.nickname AS nickname,u.avatar_url AS avatar_url,COALESCE(w.coin_balance,0) AS coin_balance,COALESCE(w.petal_balance,0) AS petal_balance,u.status AS status,u.remark AS remark,u.last_login_at AS last_login_at,u.created_at AS created_at").
		Joins("LEFT JOIN user_wallets AS w ON w.user_id=u.id").
		Where("u.deleted_at IS NULL")
	if keyword := strings.TrimSpace(c.Query("q")); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("u.user_no LIKE ? OR u.nickname LIKE ?", like, like)
	}
	if status := strings.TrimSpace(c.Query("status")); status == "1" || status == "2" {
		query = query.Where("u.status=?", status)
	}
	err := query.
		Order("u.id DESC").
		Limit(500).
		Scan(&list).Error
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, list)
}
func (a *AdminController) Assets(c *gin.Context) {
	var list []adminAssetRow
	err := a.db.Table("asset_transactions AS t").
		Select("t.id AS id,COALESCE(u.user_no,CAST(t.user_id AS CHAR)) AS user_id,COALESCE(u.nickname,'') AS nickname,t.asset_type AS asset_type,t.change_amount AS change_amount,t.balance_after AS balance_after,t.reason_code AS reason_code,t.created_at AS created_at").
		Joins("LEFT JOIN users AS u ON u.id=t.user_id").
		Order("t.id DESC").
		Limit(500).
		Scan(&list).Error
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, list)
}
func (a *AdminController) Orders(c *gin.Context) {
	var list []adminLotteryRow
	err := a.db.Table("lottery_orders AS o").
		Select("o.id AS id,o.order_no AS order_no,COALESCE(u.user_no,CAST(o.user_id AS CHAR)) AS user_id,COALESCE(u.nickname,'') AS nickname,o.order_type AS order_type,p.code AS pool_code,p.name AS pool_name,o.executed_draw_count AS draw_count,o.petal_cost AS petal_cost,COALESCE(reward_summary.rewards,'-') AS rewards,CASE WHEN o.flowers_after>=o.flowers_before THEN o.flowers_after-o.flowers_before ELSE 0 END AS lit_flower_count,o.status AS status,o.created_at AS created_at").
		Joins("LEFT JOIN users AS u ON u.id=o.user_id").
		Joins("LEFT JOIN prize_pools AS p ON p.id=o.prize_pool_id").
		Joins(`LEFT JOIN (
			SELECT reward_total.lottery_order_id,
				GROUP_CONCAT(CONCAT(reward_total.reward_name,' ×',reward_total.quantity) ORDER BY reward_total.first_draw SEPARATOR '、') AS rewards
			FROM (
				SELECT draw.lottery_order_id,item.id AS reward_item_id,item.name AS reward_name,
					SUM(draw.reward_quantity) AS quantity,MIN(draw.draw_index) AS first_draw
				FROM lottery_draws AS draw
				JOIN reward_items AS item ON item.id=draw.reward_item_id
				GROUP BY draw.lottery_order_id,item.id,item.name
			) AS reward_total
			GROUP BY reward_total.lottery_order_id
		) AS reward_summary ON reward_summary.lottery_order_id=o.id`).
		Order("o.id DESC").
		Limit(500).
		Scan(&list).Error
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, list)
}
func (a *AdminController) Rewards(c *gin.Context) {
	var list []adminRewardRow
	err := a.db.Table("user_rewards AS ur").
		Select("ur.id AS id,COALESCE(u.user_no,CAST(ur.user_id AS CHAR)) AS user_id,COALESCE(u.nickname,'') AS nickname,i.item_code AS item_code,i.name AS reward_name,ur.quantity AS quantity,ur.source_type AS source_type,ur.status AS status,COALESCE(ur.granted_at,ur.created_at) AS obtained_at").
		Joins("LEFT JOIN users AS u ON u.id=ur.user_id").
		Joins("JOIN reward_items AS i ON i.id=ur.reward_item_id").
		Order("ur.id DESC").
		Limit(500).
		Scan(&list).Error
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, list)
}
func (a *AdminController) Rounds(c *gin.Context) {
	var list []adminRoundRow
	err := a.db.Table("user_activity_rounds AS r").
		Select(`r.id AS id,COALESCE(u.user_no,CAST(r.user_id AS CHAR)) AS user_id,COALESCE(u.nickname,'') AS nickname,
			r.round_no AS round_no,r.lit_flower_count AS lit_flower_count,
			(SELECT COUNT(*) FROM user_chest_opportunities AS chest WHERE chest.round_id=r.id AND chest.opened_at IS NOT NULL) AS chest_opened_count,
			(SELECT COUNT(*) FROM stage_reward_rules AS rule WHERE rule.activity_id=r.activity_id AND rule.status=1 AND rule.required_flowers<=r.lit_flower_count) AS stage_eligible_count,
			(SELECT COUNT(*) FROM user_stage_reward_claims AS claim WHERE claim.round_id=r.id AND claim.status=1) AS stage_claimed_count,
			(SELECT COUNT(*) FROM user_stage_reward_claims AS claim WHERE claim.round_id=r.id AND claim.status=2) AS stage_failed_count,
			r.status AS status,r.updated_at AS updated_at`).
		Joins("LEFT JOIN users AS u ON u.id=r.user_id").
		Where(`r.id=(
			SELECT r2.id FROM user_activity_rounds AS r2
			WHERE r2.user_id=r.user_id AND r2.activity_id=r.activity_id
			ORDER BY r2.round_no DESC LIMIT 1
		)`).
		Order("r.updated_at DESC").
		Limit(500).
		Scan(&list).Error
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, list)
}
func (a *AdminController) Leaderboard(c *gin.Context) {
	var list []adminLeaderboardRow
	err := a.db.Table("leaderboard_entries AS e").
		Select("e.id AS id,ROW_NUMBER() OVER (PARTITION BY e.activity_id ORDER BY e.score DESC,e.reached_at ASC,e.user_id ASC) AS rank_no,COALESCE(u.user_no,CAST(e.user_id AS CHAR)) AS user_id,COALESCE(u.nickname,'') AS nickname,e.score AS score,e.updated_at AS updated_at").
		Joins("LEFT JOIN users AS u ON u.id=e.user_id").
		Order("e.activity_id ASC,e.score DESC,e.reached_at ASC,e.user_id ASC").
		Limit(200).
		Scan(&list).Error
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, list)
}

func (a *AdminController) Activities(c *gin.Context) {
	var list []adminActivityRow
	err := a.db.Table("activities").
		Select("id,code,name,status,starts_at,ends_at,leaderboard_freezes_at,timezone,updated_at").
		Where("deleted_at IS NULL").
		Order("id DESC").
		Scan(&list).Error
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, list)
}
func (a *AdminController) Pools(c *gin.Context) {
	var list []adminPoolRow
	err := a.db.Table("prize_pools AS p").
		Select(`p.id AS id,a.name AS activity_name,p.code AS code,p.name AS name,
			p.petal_cost_per_draw AS petal_cost_per_draw,p.coin_value_per_draw AS coin_value_per_draw,
			CAST(p.supported_draw_counts AS CHAR) AS supported_draw_counts,p.status AS status,
			COALESCE(v.version_no,0) AS version_no,v.effective_at AS effective_at,COUNT(pr.id) AS reward_count`).
		Joins("JOIN activities AS a ON a.id=p.activity_id").
		Joins(`LEFT JOIN prize_pool_versions AS v ON v.id=(
			SELECT v2.id FROM prize_pool_versions AS v2
			WHERE v2.prize_pool_id=p.id AND v2.status=1
			ORDER BY v2.version_no DESC LIMIT 1
		)`).
		Joins("LEFT JOIN prize_pool_rewards AS pr ON pr.version_id=v.id").
		Where("p.deleted_at IS NULL").
		Group("p.id,a.name,p.code,p.name,p.petal_cost_per_draw,p.coin_value_per_draw,p.supported_draw_counts,p.status,v.version_no,v.effective_at").
		Order("p.activity_id,p.sort_no").
		Scan(&list).Error
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, list)
}
