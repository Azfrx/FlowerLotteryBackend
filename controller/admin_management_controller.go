package controller

import (
	"encoding/json"
	"errors"
	"flower-lottery-backend/middleware"
	"flower-lottery-backend/model"
	"flower-lottery-backend/response"
	"flower-lottery-backend/service"
	"flower-lottery-backend/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const adminPoolTotalWeight uint64 = 1_000_000
const adminCoinValuePerPetal uint64 = 60
const adminActivityContentMaxBytes = 256 * 1024
const adminExchangeOptionMaxValue uint64 = 9_223_372_036_854_775_807

type adminCreateUserInput struct {
	UserID              string `json:"user_id"`
	Nickname            string `json:"nickname"`
	Password            string `json:"password"`
	AvatarURL           string `json:"avatar_url"`
	Status              uint8  `json:"status"`
	Remark              string `json:"remark"`
	InitialCoinBalance  int64  `json:"initial_coin_balance"`
	InitialPetalBalance int64  `json:"initial_petal_balance"`
}

type adminUpdateUserInput struct {
	Nickname    string `json:"nickname"`
	AvatarURL   string `json:"avatar_url"`
	Status      uint8  `json:"status"`
	Remark      string `json:"remark"`
	NewPassword string `json:"new_password"`
}

type adminAdjustAssetInput struct {
	AssetType    string `json:"asset_type"`
	ChangeAmount int64  `json:"change_amount"`
	Remark       string `json:"remark"`
}

type adminRewardItemInput struct {
	ItemCode     string `json:"item_code"`
	Name         string `json:"name"`
	ItemType     string `json:"item_type"`
	ImageURL     string `json:"image_url"`
	AnimationURL string `json:"animation_url"`
	Rarity       string `json:"rarity"`
	Status       uint8  `json:"status"`
}

type adminRewardItemRow struct {
	ID           uint64    `gorm:"column:id" json:"id"`
	ItemCode     string    `gorm:"column:item_code" json:"item_code"`
	Name         string    `gorm:"column:name" json:"name"`
	ItemType     string    `gorm:"column:item_type" json:"item_type"`
	ImageURL     string    `gorm:"column:image_url" json:"image_url"`
	AnimationURL string    `gorm:"column:animation_url" json:"animation_url"`
	Rarity       string    `gorm:"column:rarity" json:"rarity"`
	Status       uint8     `gorm:"column:status" json:"status"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updated_at"`
}

type adminPoolRewardInput struct {
	RewardItemID    uint64 `json:"reward_item_id"`
	Quantity        uint64 `json:"quantity"`
	Weight          uint64 `json:"weight"`
	ChoiceGroupCode string `json:"choice_group_code"`
}

type adminUpdatePoolInput struct {
	Name             string                 `json:"name"`
	PetalCostPerDraw uint64                 `json:"petal_cost_per_draw"`
	Status           uint8                  `json:"status"`
	Remark           string                 `json:"remark"`
	Rewards          []adminPoolRewardInput `json:"rewards"`
}

type adminPoolRewardRow struct {
	ID              uint64 `gorm:"column:id" json:"id"`
	RewardItemID    uint64 `gorm:"column:reward_item_id" json:"reward_item_id"`
	ItemCode        string `gorm:"column:item_code" json:"item_code"`
	Name            string `gorm:"column:name" json:"name"`
	ItemType        string `gorm:"column:item_type" json:"item_type"`
	ImageURL        string `gorm:"column:image_url" json:"image_url"`
	Quantity        uint64 `gorm:"column:quantity" json:"quantity"`
	Weight          uint64 `gorm:"column:weight" json:"weight"`
	ChoiceGroupCode string `gorm:"column:choice_group_code" json:"choice_group_code"`
	SortNo          int    `gorm:"column:sort_no" json:"sort_no"`
}

type adminPoolConfig struct {
	ID                  uint64               `json:"id"`
	ActivityID          uint64               `json:"activity_id"`
	ActivityName        string               `json:"activity_name"`
	Code                string               `json:"code"`
	Name                string               `json:"name"`
	PetalCostPerDraw    uint64               `json:"petal_cost_per_draw"`
	CoinValuePerDraw    uint64               `json:"coin_value_per_draw"`
	SupportedDrawCounts []uint               `json:"supported_draw_counts"`
	Status              uint8                `json:"status"`
	VersionID           uint64               `json:"version_id"`
	VersionNo           uint                 `json:"version_no"`
	EffectiveAt         *time.Time           `json:"effective_at"`
	TotalWeight         uint64               `json:"total_weight"`
	Rewards             []adminPoolRewardRow `json:"rewards"`
}

type adminUpdateActivityInput struct {
	Name                 string    `json:"name"`
	Status               uint8     `json:"status"`
	StartsAt             time.Time `json:"starts_at"`
	EndsAt               time.Time `json:"ends_at"`
	LeaderboardFreezesAt time.Time `json:"leaderboard_freezes_at"`
	Timezone             string    `json:"timezone"`
}

type adminExchangeOptionInput struct {
	ActivityID  uint64 `json:"activity_id"`
	PetalAmount uint64 `json:"petal_amount"`
	CoinCost    uint64 `json:"coin_cost"`
	SortNo      int    `json:"sort_no"`
	Status      uint8  `json:"status"`
	Remark      string `json:"remark"`
}

type adminExchangeOptionRow struct {
	ID           uint64    `gorm:"column:id" json:"id"`
	ActivityID   uint64    `gorm:"column:activity_id" json:"activity_id"`
	ActivityCode string    `gorm:"column:activity_code" json:"activity_code"`
	ActivityName string    `gorm:"column:activity_name" json:"activity_name"`
	PetalAmount  uint64    `gorm:"column:petal_amount" json:"petal_amount"`
	CoinCost     uint64    `gorm:"column:coin_cost" json:"coin_cost"`
	SortNo       int       `gorm:"column:sort_no" json:"sort_no"`
	Status       uint8     `gorm:"column:status" json:"status"`
	Remark       string    `gorm:"column:remark" json:"remark"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updated_at"`
}

type adminOperationLogRow struct {
	ID           uint64    `gorm:"column:id" json:"id"`
	AdminName    string    `gorm:"column:admin_name" json:"admin_name"`
	Method       string    `gorm:"column:method" json:"method"`
	Path         string    `gorm:"column:path" json:"path"`
	Action       string    `gorm:"column:action" json:"action"`
	TargetType   string    `gorm:"column:target_type" json:"target_type"`
	TargetID     string    `gorm:"column:target_id" json:"target_id"`
	ResponseCode int       `gorm:"column:response_code" json:"response_code"`
	IP           string    `gorm:"column:ip" json:"ip"`
	DurationMS   uint      `gorm:"column:duration_ms" json:"duration_ms"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
}

func (a *AdminController) CreateUser(c *gin.Context) {
	startedAt := time.Now()
	var input adminCreateUserInput
	if c.ShouldBindJSON(&input) != nil {
		adminRequestError(c, "用户信息格式不正确")
		return
	}
	input.UserID = strings.TrimSpace(input.UserID)
	input.Nickname = strings.TrimSpace(input.Nickname)
	input.AvatarURL = strings.TrimSpace(input.AvatarURL)
	input.Remark = strings.TrimSpace(input.Remark)
	if !utils.ValidAccount(input.UserID) {
		adminRequestError(c, "账号需为3至64位英文或数字")
		return
	}
	if input.Nickname == "" || utf8.RuneCountInString(input.Nickname) > 64 {
		adminRequestError(c, "昵称不能为空且不能超过64个字符")
		return
	}
	if len(input.Password) < 6 || len(input.Password) > 72 {
		adminRequestError(c, "密码需为6至72个字符")
		return
	}
	if input.Status != 1 && input.Status != 2 {
		adminRequestError(c, "用户状态无效")
		return
	}
	if input.InitialCoinBalance < 0 || input.InitialPetalBalance < 0 {
		adminRequestError(c, "初始余额不能为负数")
		return
	}
	if !validAdminResourceURL(input.AvatarURL) {
		adminRequestError(c, "头像链接格式不正确")
		return
	}
	hash, err := utils.HashPassword(input.Password)
	if err != nil {
		writeError(c, err)
		return
	}
	user := model.User{
		UserNo:       input.UserID,
		Nickname:     input.Nickname,
		AvatarURL:    input.AvatarURL,
		PasswordHash: hash,
		Status:       input.Status,
		Remark:       input.Remark,
	}
	err = a.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		wallet := model.UserWallet{
			UserID:       user.ID,
			CoinBalance:  input.InitialCoinBalance,
			PetalBalance: input.InitialPetalBalance,
		}
		if err := tx.Create(&wallet).Error; err != nil {
			return err
		}
		activityID := currentAdminActivityID(tx)
		requestID := adminMutationRequestID(middleware.CurrentAdminID(c))
		transactions := make([]model.AssetTransaction, 0, 2)
		if input.InitialCoinBalance > 0 {
			transactions = append(transactions, model.AssetTransaction{
				UserID: user.ID, ActivityID: activityID, AssetType: "coin",
				ChangeAmount: input.InitialCoinBalance, BalanceBefore: 0, BalanceAfter: input.InitialCoinBalance,
				ReasonCode: "admin_adjustment", BizType: "admin", RequestID: requestID, Remark: "创建用户初始余额",
			})
		}
		if input.InitialPetalBalance > 0 {
			transactions = append(transactions, model.AssetTransaction{
				UserID: user.ID, ActivityID: activityID, AssetType: "petal",
				ChangeAmount: input.InitialPetalBalance, BalanceBefore: 0, BalanceAfter: input.InitialPetalBalance,
				ReasonCode: "admin_adjustment", BizType: "admin", RequestID: requestID, Remark: "创建用户初始余额",
			})
		}
		if len(transactions) > 0 {
			return tx.Create(&transactions).Error
		}
		return nil
	})
	if err != nil {
		if isAdminDuplicateError(err) {
			response.Error(c, 409, 13001, "账号已存在")
			return
		}
		writeError(c, err)
		return
	}
	row, err := a.adminUserByID(user.ID)
	if err != nil {
		writeError(c, err)
		return
	}
	a.recordAdminOperation(c, startedAt, "user.create", "user", strconv.FormatUint(user.ID, 10), gin.H{
		"user_id": input.UserID, "nickname": input.Nickname, "status": input.Status,
	})
	response.Success(c, row)
}

func (a *AdminController) UpdateUser(c *gin.Context) {
	startedAt := time.Now()
	id, ok := adminPathID(c)
	if !ok {
		return
	}
	var input adminUpdateUserInput
	if c.ShouldBindJSON(&input) != nil {
		adminRequestError(c, "用户信息格式不正确")
		return
	}
	input.Nickname = strings.TrimSpace(input.Nickname)
	input.AvatarURL = strings.TrimSpace(input.AvatarURL)
	input.Remark = strings.TrimSpace(input.Remark)
	if input.Nickname == "" || utf8.RuneCountInString(input.Nickname) > 64 {
		adminRequestError(c, "昵称不能为空且不能超过64个字符")
		return
	}
	if input.Status != 1 && input.Status != 2 {
		adminRequestError(c, "用户状态无效")
		return
	}
	if input.NewPassword != "" && (len(input.NewPassword) < 6 || len(input.NewPassword) > 72) {
		adminRequestError(c, "新密码需为6至72个字符")
		return
	}
	if !validAdminResourceURL(input.AvatarURL) {
		adminRequestError(c, "头像链接格式不正确")
		return
	}
	var user model.User
	if err := a.db.Where("id=? AND deleted_at IS NULL", id).First(&user).Error; err != nil {
		adminRecordError(c, err, "用户不存在")
		return
	}
	updates := map[string]any{
		"nickname": input.Nickname, "avatar_url": input.AvatarURL,
		"status": input.Status, "remark": input.Remark,
	}
	passwordChanged := input.NewPassword != ""
	if passwordChanged {
		hash, err := utils.HashPassword(input.NewPassword)
		if err != nil {
			writeError(c, err)
			return
		}
		updates["password_hash"] = hash
	}
	err := a.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.User{}).Where("id=? AND deleted_at IS NULL", id).Updates(updates).Error; err != nil {
			return err
		}
		if passwordChanged || input.Status != 1 {
			now := time.Now()
			return tx.Model(&model.RefreshToken{}).
				Where("subject_type='user' AND subject_id=? AND revoked_at IS NULL", id).
				Update("revoked_at", now).Error
		}
		return nil
	})
	if err != nil {
		writeError(c, err)
		return
	}
	row, err := a.adminUserByID(id)
	if err != nil {
		writeError(c, err)
		return
	}
	a.recordAdminOperation(c, startedAt, "user.update", "user", strconv.FormatUint(id, 10), gin.H{
		"nickname": input.Nickname, "status": input.Status, "password_reset": passwordChanged,
	})
	response.Success(c, row)
}

func (a *AdminController) DeleteUser(c *gin.Context) {
	startedAt := time.Now()
	id, ok := adminPathID(c)
	if !ok {
		return
	}
	var user model.User
	if err := a.db.Where("id=? AND deleted_at IS NULL", id).First(&user).Error; err != nil {
		adminRecordError(c, err, "用户不存在")
		return
	}
	now := time.Now()
	err := a.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.User{}).Where("id=? AND deleted_at IS NULL", id).
			Updates(map[string]any{"status": 2, "deleted_at": now}).Error; err != nil {
			return err
		}
		return tx.Model(&model.RefreshToken{}).
			Where("subject_type='user' AND subject_id=? AND revoked_at IS NULL", id).
			Update("revoked_at", now).Error
	})
	if err != nil {
		writeError(c, err)
		return
	}
	a.recordAdminOperation(c, startedAt, "user.delete", "user", strconv.FormatUint(id, 10), gin.H{"user_id": user.UserNo})
	response.Success(c, gin.H{"id": id})
}

func (a *AdminController) AdjustUserAsset(c *gin.Context) {
	startedAt := time.Now()
	id, ok := adminPathID(c)
	if !ok {
		return
	}
	var input adminAdjustAssetInput
	if c.ShouldBindJSON(&input) != nil {
		adminRequestError(c, "资产调整信息格式不正确")
		return
	}
	input.AssetType = strings.TrimSpace(input.AssetType)
	input.Remark = strings.TrimSpace(input.Remark)
	if input.AssetType != "coin" && input.AssetType != "petal" {
		adminRequestError(c, "资产类型仅支持金币或花瓣")
		return
	}
	if input.ChangeAmount == 0 || input.ChangeAmount > 1_000_000_000_000 || input.ChangeAmount < -1_000_000_000_000 {
		adminRequestError(c, "调整数量必须为非零且绝对值不超过一万亿")
		return
	}
	var result model.UserWallet
	err := a.db.Transaction(func(tx *gorm.DB) error {
		var user model.User
		if err := tx.Where("id=? AND deleted_at IS NULL", id).First(&user).Error; err != nil {
			return err
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id=?", id).First(&result).Error; err != nil {
			return err
		}
		before := result.CoinBalance
		column := "coin_balance"
		if input.AssetType == "petal" {
			before = result.PetalBalance
			column = "petal_balance"
		}
		after := before + input.ChangeAmount
		if after < 0 {
			return errAdminInsufficientBalance
		}
		if err := tx.Model(&result).Updates(map[string]any{column: after, "version": gorm.Expr("version + 1")}).Error; err != nil {
			return err
		}
		if input.AssetType == "coin" {
			result.CoinBalance = after
		} else {
			result.PetalBalance = after
		}
		transaction := model.AssetTransaction{
			UserID: id, ActivityID: currentAdminActivityID(tx), AssetType: input.AssetType,
			ChangeAmount: input.ChangeAmount, BalanceBefore: before, BalanceAfter: after,
			ReasonCode: "admin_adjustment", BizType: "admin",
			RequestID: adminMutationRequestID(middleware.CurrentAdminID(c)), Remark: input.Remark,
		}
		return tx.Create(&transaction).Error
	})
	if errors.Is(err, errAdminInsufficientBalance) {
		response.Error(c, 409, 13002, "调整后余额不能小于零")
		return
	}
	if err != nil {
		adminRecordError(c, err, "用户或钱包不存在")
		return
	}
	a.recordAdminOperation(c, startedAt, "user.asset_adjust", "user", strconv.FormatUint(id, 10), gin.H{
		"asset_type": input.AssetType, "change_amount": input.ChangeAmount, "remark": input.Remark,
	})
	response.Success(c, gin.H{"coin_balance": result.CoinBalance, "petal_balance": result.PetalBalance})
}

func (a *AdminController) RewardItems(c *gin.Context) {
	var list []adminRewardItemRow
	query := a.db.Table("reward_items").
		Select("id,item_code,name,item_type,image_url,animation_url,rarity,status,created_at,updated_at").
		Where("deleted_at IS NULL")
	if keyword := strings.TrimSpace(c.Query("q")); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("item_code LIKE ? OR name LIKE ?", like, like)
	}
	if itemType := strings.TrimSpace(c.Query("type")); itemType != "" {
		query = query.Where("item_type=?", itemType)
	}
	if status := strings.TrimSpace(c.Query("status")); status == "0" || status == "1" {
		query = query.Where("status=?", status)
	}
	if err := query.Order("id DESC").Limit(500).Scan(&list).Error; err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, list)
}

func (a *AdminController) CreateRewardItem(c *gin.Context) {
	startedAt := time.Now()
	var input adminRewardItemInput
	if c.ShouldBindJSON(&input) != nil || !normalizeAndValidateRewardItem(&input, true) {
		adminRequestError(c, "奖品信息不完整或格式不正确")
		return
	}
	item := model.RewardItem{
		ItemCode: input.ItemCode, Name: input.Name, ItemType: input.ItemType,
		ImageURL: input.ImageURL, AnimationURL: input.AnimationURL,
		Rarity: input.Rarity, Status: input.Status,
	}
	if err := a.db.Create(&item).Error; err != nil {
		if isAdminDuplicateError(err) {
			response.Error(c, 409, 13003, "道具ID已存在")
			return
		}
		writeError(c, err)
		return
	}
	a.recordAdminOperation(c, startedAt, "reward.create", "reward_item", strconv.FormatUint(item.ID, 10), gin.H{
		"item_code": item.ItemCode, "name": item.Name, "status": item.Status,
	})
	response.Success(c, adminRewardItemFromModel(item))
}

func (a *AdminController) UpdateRewardItem(c *gin.Context) {
	startedAt := time.Now()
	id, ok := adminPathID(c)
	if !ok {
		return
	}
	var input adminRewardItemInput
	if c.ShouldBindJSON(&input) != nil || !normalizeAndValidateRewardItem(&input, false) {
		adminRequestError(c, "奖品信息不完整或格式不正确")
		return
	}
	var item model.RewardItem
	err := a.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id=? AND deleted_at IS NULL", id).First(&item).Error; err != nil {
			return err
		}
		if rewardItemGoingOffline(item.Status, input.Status) {
			referenceCount, err := onlinePrizePoolRewardReferenceCount(tx, id)
			if err != nil {
				return err
			}
			if referenceCount > 0 {
				return errRewardItemReferencedByOnlinePool
			}
		}
		updates := map[string]any{
			"name": input.Name, "item_type": input.ItemType,
			"image_url": input.ImageURL, "animation_url": input.AnimationURL,
			"rarity": input.Rarity, "status": input.Status,
		}
		if err := tx.Model(&item).Updates(updates).Error; err != nil {
			return err
		}
		return tx.Where("id=?", id).First(&item).Error
	})
	if errors.Is(err, errRewardItemReferencedByOnlinePool) {
		response.Error(c, 409, 13004, "奖品仍被线上奖池引用，请先从奖池中移除并发布新版本")
		return
	}
	if err != nil {
		adminRecordError(c, err, "奖品不存在")
		return
	}
	a.recordAdminOperation(c, startedAt, "reward.update", "reward_item", strconv.FormatUint(id, 10), gin.H{
		"item_code": item.ItemCode, "name": item.Name, "status": item.Status,
	})
	response.Success(c, adminRewardItemFromModel(item))
}

func (a *AdminController) DeleteRewardItem(c *gin.Context) {
	startedAt := time.Now()
	id, ok := adminPathID(c)
	if !ok {
		return
	}
	var item model.RewardItem
	if err := a.db.Where("id=? AND deleted_at IS NULL", id).First(&item).Error; err != nil {
		adminRecordError(c, err, "奖品不存在")
		return
	}
	var referenceCount int64
	if err := a.db.Table("prize_pool_rewards AS pr").
		Joins("JOIN prize_pool_versions AS v ON v.id=pr.version_id AND v.status=1").
		Where("pr.reward_item_id=?", id).Count(&referenceCount).Error; err != nil {
		writeError(c, err)
		return
	}
	if referenceCount == 0 {
		if err := a.db.Model(&model.StageRewardRule{}).Where("reward_item_id=? AND status=1", id).Count(&referenceCount).Error; err != nil {
			writeError(c, err)
			return
		}
	}
	if referenceCount == 0 {
		if err := a.db.Model(&model.ChestRewardRule{}).Where("reward_item_id=? AND status=1", id).Count(&referenceCount).Error; err != nil {
			writeError(c, err)
			return
		}
	}
	if referenceCount > 0 {
		response.Error(c, 409, 13004, "奖品仍被启用中的奖池或奖励规则引用，无法删除")
		return
	}
	now := time.Now()
	if err := a.db.Model(&item).Updates(map[string]any{"status": 0, "deleted_at": now}).Error; err != nil {
		writeError(c, err)
		return
	}
	a.recordAdminOperation(c, startedAt, "reward.delete", "reward_item", strconv.FormatUint(id, 10), gin.H{"item_code": item.ItemCode})
	response.Success(c, gin.H{"id": id})
}

func (a *AdminController) ExchangeOptions(c *gin.Context) {
	var list []adminExchangeOptionRow
	query := a.db.Table("exchange_options AS o").
		Select(`o.id AS id,o.activity_id AS activity_id,a.code AS activity_code,a.name AS activity_name,
			o.petal_amount AS petal_amount,o.coin_cost AS coin_cost,o.sort_no AS sort_no,
			o.status AS status,o.remark AS remark,o.created_at AS created_at,o.updated_at AS updated_at`).
		Joins("JOIN activities AS a ON a.id=o.activity_id AND a.deleted_at IS NULL").
		Where("o.deleted_at IS NULL")
	if rawActivityID := strings.TrimSpace(c.Query("activity_id")); rawActivityID != "" {
		activityID, err := strconv.ParseUint(rawActivityID, 10, 64)
		if err != nil || activityID == 0 {
			adminRequestError(c, "活动ID无效")
			return
		}
		query = query.Where("o.activity_id=?", activityID)
	}
	if status := strings.TrimSpace(c.Query("status")); status == "0" || status == "1" {
		query = query.Where("o.status=?", status)
	}
	if err := query.Order("a.id DESC,o.sort_no ASC,o.id ASC").Limit(500).Scan(&list).Error; err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, list)
}

func (a *AdminController) CreateExchangeOption(c *gin.Context) {
	startedAt := time.Now()
	var input adminExchangeOptionInput
	if c.ShouldBindJSON(&input) != nil || !normalizeAndValidateExchangeOption(&input) {
		adminRequestError(c, "花瓣套餐信息不完整或格式不正确")
		return
	}
	if err := adminEnsureActivity(a.db, input.ActivityID); err != nil {
		adminRecordError(c, err, "所属活动不存在")
		return
	}

	var existing model.ExchangeOption
	err := a.db.Unscoped().Where("activity_id=? AND petal_amount=?", input.ActivityID, input.PetalAmount).First(&existing).Error
	if err == nil {
		if existing.DeletedAt == nil {
			response.Error(c, 409, 13006, "该活动已存在相同花瓣数量的套餐")
			return
		}
		if err := a.db.Unscoped().Model(&existing).Updates(map[string]any{
			"coin_cost": input.CoinCost, "sort_no": input.SortNo, "status": input.Status,
			"remark": input.Remark, "deleted_at": nil,
		}).Error; err != nil {
			writeError(c, err)
			return
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		writeError(c, err)
		return
	} else {
		existing = model.ExchangeOption{
			ActivityID: input.ActivityID, PetalAmount: input.PetalAmount, CoinCost: input.CoinCost,
			SortNo: input.SortNo, Status: input.Status, Remark: input.Remark,
		}
		if err := a.db.Create(&existing).Error; err != nil {
			if isAdminDuplicateError(err) {
				response.Error(c, 409, 13006, "该活动已存在相同花瓣数量的套餐")
				return
			}
			writeError(c, err)
			return
		}
	}

	row, err := a.adminExchangeOptionByID(existing.ID)
	if err != nil {
		writeError(c, err)
		return
	}
	a.recordAdminOperation(c, startedAt, "exchange_option.create", "exchange_option", strconv.FormatUint(existing.ID, 10), gin.H{
		"activity_id": input.ActivityID, "petal_amount": input.PetalAmount,
		"coin_cost": input.CoinCost, "status": input.Status,
	})
	response.Success(c, row)
}

func (a *AdminController) UpdateExchangeOption(c *gin.Context) {
	startedAt := time.Now()
	id, ok := adminPathID(c)
	if !ok {
		return
	}
	var input adminExchangeOptionInput
	if c.ShouldBindJSON(&input) != nil || !normalizeAndValidateExchangeOption(&input) {
		adminRequestError(c, "花瓣套餐信息不完整或格式不正确")
		return
	}
	var option model.ExchangeOption
	if err := a.db.Where("id=? AND deleted_at IS NULL", id).First(&option).Error; err != nil {
		adminRecordError(c, err, "花瓣套餐不存在")
		return
	}
	if input.ActivityID != option.ActivityID {
		adminRequestError(c, "套餐创建后不能修改所属活动")
		return
	}
	if err := a.db.Model(&option).Updates(map[string]any{
		"petal_amount": input.PetalAmount, "coin_cost": input.CoinCost,
		"sort_no": input.SortNo, "status": input.Status, "remark": input.Remark,
	}).Error; err != nil {
		if isAdminDuplicateError(err) {
			response.Error(c, 409, 13006, "该活动已存在相同花瓣数量的套餐")
			return
		}
		writeError(c, err)
		return
	}
	row, err := a.adminExchangeOptionByID(id)
	if err != nil {
		writeError(c, err)
		return
	}
	a.recordAdminOperation(c, startedAt, "exchange_option.update", "exchange_option", strconv.FormatUint(id, 10), gin.H{
		"activity_id": input.ActivityID, "petal_amount": input.PetalAmount,
		"coin_cost": input.CoinCost, "sort_no": input.SortNo, "status": input.Status,
	})
	response.Success(c, row)
}

func (a *AdminController) DeleteExchangeOption(c *gin.Context) {
	startedAt := time.Now()
	id, ok := adminPathID(c)
	if !ok {
		return
	}
	var option model.ExchangeOption
	if err := a.db.Where("id=? AND deleted_at IS NULL", id).First(&option).Error; err != nil {
		adminRecordError(c, err, "花瓣套餐不存在")
		return
	}
	now := time.Now()
	if err := a.db.Model(&option).Updates(map[string]any{"status": 0, "deleted_at": now}).Error; err != nil {
		writeError(c, err)
		return
	}
	a.recordAdminOperation(c, startedAt, "exchange_option.delete", "exchange_option", strconv.FormatUint(id, 10), gin.H{
		"activity_id": option.ActivityID, "petal_amount": option.PetalAmount, "coin_cost": option.CoinCost,
	})
	response.Success(c, gin.H{"id": id})
}

func (a *AdminController) PoolConfig(c *gin.Context) {
	id, ok := adminPathID(c)
	if !ok {
		return
	}
	config, err := a.loadPoolConfig(id)
	if err != nil {
		adminRecordError(c, err, "奖池不存在")
		return
	}
	response.Success(c, config)
}

func (a *AdminController) UpdatePoolConfig(c *gin.Context) {
	startedAt := time.Now()
	id, ok := adminPathID(c)
	if !ok {
		return
	}
	var input adminUpdatePoolInput
	if c.ShouldBindJSON(&input) != nil {
		adminRequestError(c, "奖池配置格式不正确")
		return
	}
	input.Name = strings.TrimSpace(input.Name)
	input.Remark = strings.TrimSpace(input.Remark)
	if input.Name == "" || utf8.RuneCountInString(input.Name) > 64 {
		adminRequestError(c, "奖池名称不能为空且不能超过64个字符")
		return
	}
	coinValuePerDraw, validPetalCost := adminCoinValueForPetalCost(input.PetalCostPerDraw)
	if !validPetalCost {
		adminRequestError(c, "单抽消耗花瓣必须大于零且数值有效")
		return
	}
	if input.Status > 1 || len(input.Rewards) == 0 {
		adminRequestError(c, "奖池状态或奖励列表无效")
		return
	}
	var totalWeight uint64
	itemIDs := make([]uint64, 0, len(input.Rewards))
	uniqueItemIDs := make(map[uint64]struct{}, len(input.Rewards))
	for i := range input.Rewards {
		reward := &input.Rewards[i]
		reward.ChoiceGroupCode = strings.TrimSpace(reward.ChoiceGroupCode)
		if reward.RewardItemID == 0 || reward.Quantity == 0 || reward.Weight == 0 || len(reward.ChoiceGroupCode) > 64 {
			adminRequestError(c, "奖品、数量、概率权重或选择分组无效")
			return
		}
		if totalWeight > adminPoolTotalWeight-reward.Weight {
			adminRequestError(c, "奖池概率总和超过100%")
			return
		}
		totalWeight += reward.Weight
		if _, exists := uniqueItemIDs[reward.RewardItemID]; !exists {
			uniqueItemIDs[reward.RewardItemID] = struct{}{}
			itemIDs = append(itemIDs, reward.RewardItemID)
		}
	}
	if totalWeight != adminPoolTotalWeight {
		adminRequestError(c, fmt.Sprintf("奖池概率总和必须为100%%，当前为%.4f%%", float64(totalWeight)/10_000))
		return
	}
	adminID := middleware.CurrentAdminID(c)
	err := a.db.Transaction(func(tx *gorm.DB) error {
		var pool model.PrizePool
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id=? AND deleted_at IS NULL", id).First(&pool).Error; err != nil {
			return err
		}
		var items []model.RewardItem
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id IN ? AND status=1 AND deleted_at IS NULL", itemIDs).Find(&items).Error; err != nil {
			return err
		}
		if len(items) != len(uniqueItemIDs) {
			return errAdminPoolRewardItemUnavailable
		}
		itemByID := make(map[uint64]model.RewardItem, len(items))
		for _, item := range items {
			itemByID[item.ID] = item
		}
		if err := tx.Model(&pool).Updates(map[string]any{
			"name": input.Name, "petal_cost_per_draw": input.PetalCostPerDraw,
			"coin_value_per_draw": coinValuePerDraw, "status": input.Status,
		}).Error; err != nil {
			return err
		}
		var maxVersion uint
		if err := tx.Model(&model.PrizePoolVersion{}).Where("prize_pool_id=?", id).
			Select("COALESCE(MAX(version_no),0)").Scan(&maxVersion).Error; err != nil {
			return err
		}
		now := time.Now()
		version := model.PrizePoolVersion{
			PrizePoolID: id, VersionNo: maxVersion + 1, Status: 0,
			EffectiveAt: &now, TotalWeight: adminPoolTotalWeight,
			PublishedBy: &adminID, Remark: input.Remark,
		}
		if err := tx.Create(&version).Error; err != nil {
			return err
		}
		rows := make([]model.PrizePoolReward, 0, len(input.Rewards))
		for index, reward := range input.Rewards {
			item := itemByID[reward.RewardItemID]
			snapshot, marshalErr := json.Marshal(gin.H{
				"item_code": item.ItemCode, "name": item.Name, "item_type": item.ItemType,
				"image_url": item.ImageURL, "animation_url": item.AnimationURL,
				"quantity": reward.Quantity,
			})
			if marshalErr != nil {
				return marshalErr
			}
			rows = append(rows, model.PrizePoolReward{
				VersionID: version.ID, RewardItemID: reward.RewardItemID,
				Quantity: reward.Quantity, Weight: reward.Weight,
				ChoiceGroupCode: reward.ChoiceGroupCode, Snapshot: snapshot, SortNo: index + 1,
			})
		}
		if err := tx.Create(&rows).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.PrizePoolVersion{}).
			Where("prize_pool_id=? AND status=1", id).Update("status", 2).Error; err != nil {
			return err
		}
		return tx.Model(&version).Updates(map[string]any{"status": 1, "effective_at": now}).Error
	})
	if errors.Is(err, errAdminPoolRewardItemUnavailable) {
		adminRequestError(c, "奖励列表包含不存在或已停用的奖品")
		return
	}
	if err != nil {
		adminRecordError(c, err, "奖池不存在")
		return
	}
	config, err := a.loadPoolConfig(id)
	if err != nil {
		writeError(c, err)
		return
	}
	a.recordAdminOperation(c, startedAt, "pool.publish", "prize_pool", strconv.FormatUint(id, 10), gin.H{
		"name": input.Name, "version_no": config.VersionNo, "reward_count": len(input.Rewards),
		"status": input.Status, "petal_cost_per_draw": input.PetalCostPerDraw,
		"coin_value_per_draw": coinValuePerDraw,
	})
	if a.poolConfigs != nil {
		publishedAt := time.Now()
		if config.EffectiveAt != nil {
			publishedAt = *config.EffectiveAt
		}
		a.poolConfigs.Publish(service.PoolConfigUpdate{
			PoolID: id, PoolCode: config.Code,
			PetalCostPerDraw: config.PetalCostPerDraw,
			CoinValuePerDraw: config.CoinValuePerDraw,
			VersionNo:        config.VersionNo, PublishedAt: publishedAt,
		})
	}
	response.Success(c, config)
}

func adminCoinValueForPetalCost(petalCost uint64) (uint64, bool) {
	if petalCost == 0 || petalCost > ^uint64(0)/adminCoinValuePerPetal {
		return 0, false
	}
	return petalCost * adminCoinValuePerPetal, true
}

func (a *AdminController) UpdateActivity(c *gin.Context) {
	startedAt := time.Now()
	id, ok := adminPathID(c)
	if !ok {
		return
	}
	var input adminUpdateActivityInput
	if c.ShouldBindJSON(&input) != nil {
		adminRequestError(c, "活动配置格式不正确")
		return
	}
	input.Name = strings.TrimSpace(input.Name)
	input.Timezone = strings.TrimSpace(input.Timezone)
	if input.Name == "" || utf8.RuneCountInString(input.Name) > 128 || input.Status > 4 || input.Timezone == "" {
		adminRequestError(c, "活动名称、状态或时区无效")
		return
	}
	if input.StartsAt.IsZero() || input.EndsAt.IsZero() || !input.StartsAt.Before(input.EndsAt) {
		adminRequestError(c, "活动开始时间必须早于结束时间")
		return
	}
	if input.LeaderboardFreezesAt.Before(input.StartsAt) || input.LeaderboardFreezesAt.After(input.EndsAt) {
		adminRequestError(c, "榜单冻结时间必须位于活动周期内")
		return
	}
	if input.Status == 2 {
		var overlap int64
		if err := a.db.Model(&model.Activity{}).
			Where("id<>? AND deleted_at IS NULL AND status=2 AND starts_at<? AND ends_at>?", id, input.EndsAt, input.StartsAt).
			Count(&overlap).Error; err != nil {
			writeError(c, err)
			return
		}
		if overlap > 0 {
			response.Error(c, 409, 13005, "当前时间段已有进行中的活动")
			return
		}
	}
	var activity model.Activity
	if err := a.db.Where("id=? AND deleted_at IS NULL", id).First(&activity).Error; err != nil {
		adminRecordError(c, err, "活动不存在")
		return
	}
	if err := a.db.Model(&activity).Updates(map[string]any{
		"name": input.Name, "status": input.Status, "starts_at": input.StartsAt,
		"ends_at": input.EndsAt, "leaderboard_freezes_at": input.LeaderboardFreezesAt,
		"timezone": input.Timezone,
	}).Error; err != nil {
		writeError(c, err)
		return
	}
	var row adminActivityRow
	if err := a.db.Table("activities").
		Select("id,code,name,status,starts_at,ends_at,leaderboard_freezes_at,timezone,updated_at").
		Where("id=?", id).Scan(&row).Error; err != nil {
		writeError(c, err)
		return
	}
	a.recordAdminOperation(c, startedAt, "activity.update", "activity", strconv.FormatUint(id, 10), gin.H{
		"name": input.Name, "status": input.Status,
		"starts_at": input.StartsAt, "ends_at": input.EndsAt,
	})
	response.Success(c, row)
}

func (a *AdminController) ActivityContent(c *gin.Context) {
	id, ok := adminPathID(c)
	if !ok {
		return
	}
	var activity model.Activity
	if err := a.db.Select("id,rules_json").Where("id=? AND deleted_at IS NULL", id).First(&activity).Error; err != nil {
		adminRecordError(c, err, "活动不存在")
		return
	}
	var content model.ActivityContent
	if len(activity.RulesJSON) > 0 {
		if err := json.Unmarshal(activity.RulesJSON, &content); err != nil {
			writeError(c, fmt.Errorf("解析活动文案失败: %w", err))
			return
		}
	}
	response.Success(c, content)
}

func (a *AdminController) UpdateActivityContent(c *gin.Context) {
	startedAt := time.Now()
	id, ok := adminPathID(c)
	if !ok {
		return
	}
	var content model.ActivityContent
	if c.ShouldBindJSON(&content) != nil {
		adminRequestError(c, "活动文案格式不正确")
		return
	}
	if err := validateAdminActivityContent(content); err != nil {
		adminRequestError(c, err.Error())
		return
	}
	rulesJSON, err := json.Marshal(content)
	if err != nil {
		writeError(c, err)
		return
	}
	if len(rulesJSON) > adminActivityContentMaxBytes {
		adminRequestError(c, "活动文案内容过大")
		return
	}
	var activity model.Activity
	if err := a.db.Where("id=? AND deleted_at IS NULL", id).First(&activity).Error; err != nil {
		adminRecordError(c, err, "活动不存在")
		return
	}
	if err := a.db.Model(&activity).Update("rules_json", rulesJSON).Error; err != nil {
		writeError(c, err)
		return
	}
	a.recordAdminOperation(c, startedAt, "activity.content_update", "activity", strconv.FormatUint(id, 10), gin.H{
		"instruction_sections": len(content.Instructions.Sections),
		"day_guide_rows":       len(content.GameGuides.Day),
		"night_guide_rows":     len(content.GameGuides.Night),
		"ranking_sections":     len(content.RankingCustomization.Sections),
	})
	response.Success(c, content)
}

func validateAdminActivityContent(content model.ActivityContent) error {
	if err := validateAdminInstructionContent("活动说明", content.Instructions); err != nil {
		return err
	}
	if err := validateAdminGuideRows("白昼玩法攻略", content.GameGuides.Day); err != nil {
		return err
	}
	if err := validateAdminGuideRows("星夜玩法攻略", content.GameGuides.Night); err != nil {
		return err
	}
	if err := validateAdminInstructionContent("冠名说明", content.RankingCustomization); err != nil {
		return err
	}
	return validateAdminWelfareContent(content.NewRingWelfare)
}

func validateAdminInstructionContent(label string, content model.ActivityInstructionsContent) error {
	if textLength := utf8.RuneCountInString(strings.TrimSpace(content.Title)); textLength == 0 || textLength > 128 {
		return fmt.Errorf("%s标题不能为空且不能超过128个字符", label)
	}
	if len(content.Sections) == 0 || len(content.Sections) > 32 {
		return fmt.Errorf("%s需包含1至32个章节", label)
	}
	for sectionIndex, section := range content.Sections {
		if textLength := utf8.RuneCountInString(strings.TrimSpace(section.Title)); textLength == 0 || textLength > 160 {
			return fmt.Errorf("%s第%d个章节标题无效", label, sectionIndex+1)
		}
		if len(section.Paragraphs) == 0 || len(section.Paragraphs) > 64 {
			return fmt.Errorf("%s第%d个章节需包含1至64个段落", label, sectionIndex+1)
		}
		for paragraphIndex, paragraph := range section.Paragraphs {
			if len(paragraph) == 0 || len(paragraph) > 32 {
				return fmt.Errorf("%s第%d章第%d段的片段数量无效", label, sectionIndex+1, paragraphIndex+1)
			}
			hasText := false
			for _, segment := range paragraph {
				if utf8.RuneCountInString(segment.Text) > 2000 {
					return fmt.Errorf("%s第%d章第%d段内容过长", label, sectionIndex+1, paragraphIndex+1)
				}
				if strings.TrimSpace(segment.Text) != "" {
					hasText = true
				}
			}
			if !hasText {
				return fmt.Errorf("%s第%d章第%d段不能为空", label, sectionIndex+1, paragraphIndex+1)
			}
		}
	}
	link := content.ProbabilityLink
	if utf8.RuneCountInString(link.Text) > 160 || len(link.URL) > 1024 {
		return fmt.Errorf("%s链接内容过长", label)
	}
	if (strings.TrimSpace(link.Text) == "") != (strings.TrimSpace(link.URL) == "") {
		return fmt.Errorf("%s链接文案和地址需同时填写", label)
	}
	if !validAdminResourceURL(strings.TrimSpace(link.URL)) {
		return fmt.Errorf("%s链接地址格式不正确", label)
	}
	return nil
}

func validateAdminGuideRows(label string, rows [][]model.ActivityGuideNode) error {
	if len(rows) == 0 || len(rows) > 12 {
		return fmt.Errorf("%s需包含1至12行内容", label)
	}
	validTypes := map[string]bool{"text": true, "tag": true, "common": true}
	for rowIndex, row := range rows {
		if len(row) == 0 || len(row) > 24 {
			return fmt.Errorf("%s第%d行节点数量无效", label, rowIndex+1)
		}
		for _, node := range row {
			if !validTypes[node.Type] || strings.TrimSpace(node.Content) == "" || utf8.RuneCountInString(node.Content) > 500 {
				return fmt.Errorf("%s第%d行包含无效节点", label, rowIndex+1)
			}
		}
	}
	return nil
}

func validateAdminWelfareContent(content model.NewRingWelfareContent) error {
	hasContent := strings.TrimSpace(content.StoryTitle) != "" || len(content.StoryLines) > 0 ||
		strings.TrimSpace(content.ValueText) != "" || len(content.SelectionSegments) > 0 ||
		len(content.SelectionNames) > 0 || len(content.FirstPublishSegments) > 0 || len(content.FirstPublishCaptions) > 0
	if !hasContent {
		return nil
	}
	if utf8.RuneCountInString(content.StoryTitle) > 160 || utf8.RuneCountInString(content.ValueText) > 1000 ||
		len(content.StoryLines) > 20 || len(content.SelectionSegments) > 40 || len(content.SelectionNames) > 20 ||
		len(content.FirstPublishSegments) > 40 || len(content.FirstPublishCaptions) > 20 {
		return errors.New("新戒福利文案内容过长")
	}
	for _, segments := range [][]model.ActivityStyledSegment{content.SelectionSegments, content.FirstPublishSegments} {
		for _, segment := range segments {
			if segment.Style != "" && segment.Style != "highlight" && segment.Style != "tag" {
				return errors.New("新戒福利文案样式无效")
			}
			if strings.TrimSpace(segment.Text) == "" || utf8.RuneCountInString(segment.Text) > 1000 {
				return errors.New("新戒福利文案片段无效")
			}
		}
	}
	return nil
}

func (a *AdminController) OperationLogs(c *gin.Context) {
	var list []adminOperationLogRow
	query := a.db.Table("admin_operation_logs AS l").
		Select(`l.id AS id,COALESCE(NULLIF(a.display_name,''),a.username) AS admin_name,
			l.method AS method,l.path AS path,l.action AS action,l.target_type AS target_type,
			l.target_id AS target_id,l.response_code AS response_code,l.ip AS ip,
			l.duration_ms AS duration_ms,l.created_at AS created_at`).
		Joins("LEFT JOIN admin_users AS a ON a.id=l.admin_user_id")
	if keyword := strings.TrimSpace(c.Query("q")); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("l.action LIKE ? OR l.target_id LIKE ? OR a.username LIKE ? OR a.display_name LIKE ?", like, like, like, like)
	}
	if action := strings.TrimSpace(c.Query("action")); action != "" {
		query = query.Where("l.action=?", action)
	}
	if err := query.Order("l.id DESC").Limit(500).Scan(&list).Error; err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, list)
}

func (a *AdminController) adminUserByID(id uint64) (*adminUserRow, error) {
	var row adminUserRow
	err := a.db.Table("users AS u").
		Select("u.id AS id,u.user_no AS user_id,u.nickname AS nickname,u.avatar_url AS avatar_url,COALESCE(w.coin_balance,0) AS coin_balance,COALESCE(w.petal_balance,0) AS petal_balance,u.status AS status,u.remark AS remark,u.last_login_at AS last_login_at,u.created_at AS created_at").
		Joins("LEFT JOIN user_wallets AS w ON w.user_id=u.id").
		Where("u.id=? AND u.deleted_at IS NULL", id).Scan(&row).Error
	if err == nil && row.ID == 0 {
		err = gorm.ErrRecordNotFound
	}
	return &row, err
}

func (a *AdminController) loadPoolConfig(id uint64) (*adminPoolConfig, error) {
	var raw struct {
		ID                  uint64     `gorm:"column:id"`
		ActivityID          uint64     `gorm:"column:activity_id"`
		ActivityName        string     `gorm:"column:activity_name"`
		Code                string     `gorm:"column:code"`
		Name                string     `gorm:"column:name"`
		PetalCostPerDraw    uint64     `gorm:"column:petal_cost_per_draw"`
		CoinValuePerDraw    uint64     `gorm:"column:coin_value_per_draw"`
		SupportedDrawCounts []byte     `gorm:"column:supported_draw_counts"`
		Status              uint8      `gorm:"column:status"`
		VersionID           uint64     `gorm:"column:version_id"`
		VersionNo           uint       `gorm:"column:version_no"`
		EffectiveAt         *time.Time `gorm:"column:effective_at"`
		TotalWeight         uint64     `gorm:"column:total_weight"`
	}
	err := a.db.Table("prize_pools AS p").
		Select(`p.id AS id,p.activity_id AS activity_id,a.name AS activity_name,p.code AS code,p.name AS name,
			p.petal_cost_per_draw AS petal_cost_per_draw,p.coin_value_per_draw AS coin_value_per_draw,
			p.supported_draw_counts AS supported_draw_counts,p.status AS status,
			COALESCE(v.id,0) AS version_id,COALESCE(v.version_no,0) AS version_no,
			v.effective_at AS effective_at,COALESCE(v.total_weight,1000000) AS total_weight`).
		Joins("JOIN activities AS a ON a.id=p.activity_id").
		Joins(`LEFT JOIN prize_pool_versions AS v ON v.id=(
			SELECT v2.id FROM prize_pool_versions AS v2
			WHERE v2.prize_pool_id=p.id AND v2.status=1
			ORDER BY v2.version_no DESC LIMIT 1
		)`).
		Where("p.id=? AND p.deleted_at IS NULL", id).Scan(&raw).Error
	if err != nil {
		return nil, err
	}
	if raw.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	config := adminPoolConfig{
		ID: raw.ID, ActivityID: raw.ActivityID, ActivityName: raw.ActivityName,
		Code: raw.Code, Name: raw.Name, PetalCostPerDraw: raw.PetalCostPerDraw,
		CoinValuePerDraw: raw.CoinValuePerDraw, Status: raw.Status,
		VersionID: raw.VersionID, VersionNo: raw.VersionNo,
		EffectiveAt: raw.EffectiveAt, TotalWeight: raw.TotalWeight,
		SupportedDrawCounts: make([]uint, 0), Rewards: make([]adminPoolRewardRow, 0),
	}
	if len(raw.SupportedDrawCounts) > 0 {
		if err := json.Unmarshal(raw.SupportedDrawCounts, &config.SupportedDrawCounts); err != nil {
			return nil, err
		}
	}
	if raw.VersionID > 0 {
		err = a.db.Table("prize_pool_rewards AS pr").
			Select(`pr.id AS id,pr.reward_item_id AS reward_item_id,i.item_code AS item_code,
				i.name AS name,i.item_type AS item_type,i.image_url AS image_url,
				pr.quantity AS quantity,pr.weight AS weight,pr.choice_group_code AS choice_group_code,
				pr.sort_no AS sort_no`).
			Joins("JOIN reward_items AS i ON i.id=pr.reward_item_id").
			Where("pr.version_id=?", raw.VersionID).
			Order("pr.sort_no,pr.id").Scan(&config.Rewards).Error
	}
	return &config, err
}

func normalizeAndValidateExchangeOption(input *adminExchangeOptionInput) bool {
	input.Remark = strings.TrimSpace(input.Remark)
	return input.ActivityID > 0 &&
		input.PetalAmount > 0 && input.PetalAmount <= adminExchangeOptionMaxValue &&
		input.CoinCost > 0 && input.CoinCost <= adminExchangeOptionMaxValue &&
		input.SortNo >= 0 && input.SortNo <= 10000 && input.Status <= 1 &&
		utf8.RuneCountInString(input.Remark) <= 255
}

func adminEnsureActivity(db *gorm.DB, id uint64) error {
	var activity model.Activity
	return db.Select("id").Where("id=? AND deleted_at IS NULL", id).First(&activity).Error
}

func (a *AdminController) adminExchangeOptionByID(id uint64) (*adminExchangeOptionRow, error) {
	var row adminExchangeOptionRow
	err := a.db.Table("exchange_options AS o").
		Select(`o.id AS id,o.activity_id AS activity_id,a.code AS activity_code,a.name AS activity_name,
			o.petal_amount AS petal_amount,o.coin_cost AS coin_cost,o.sort_no AS sort_no,
			o.status AS status,o.remark AS remark,o.created_at AS created_at,o.updated_at AS updated_at`).
		Joins("JOIN activities AS a ON a.id=o.activity_id AND a.deleted_at IS NULL").
		Where("o.id=? AND o.deleted_at IS NULL", id).Scan(&row).Error
	if err == nil && row.ID == 0 {
		err = gorm.ErrRecordNotFound
	}
	return &row, err
}

func normalizeAndValidateRewardItem(input *adminRewardItemInput, requireCode bool) bool {
	input.ItemCode = strings.TrimSpace(input.ItemCode)
	input.Name = strings.TrimSpace(input.Name)
	input.ItemType = strings.TrimSpace(input.ItemType)
	input.ImageURL = strings.TrimSpace(input.ImageURL)
	input.AnimationURL = strings.TrimSpace(input.AnimationURL)
	input.Rarity = strings.TrimSpace(input.Rarity)
	if requireCode && (input.ItemCode == "" || len(input.ItemCode) > 64) {
		return false
	}
	if input.Name == "" || utf8.RuneCountInString(input.Name) > 128 || len(input.ItemType) > 32 {
		return false
	}
	validTypes := map[string]bool{"coin": true, "petal": true, "item": true, "choice": true}
	if !validTypes[input.ItemType] || input.Status > 1 || len(input.Rarity) > 32 {
		return false
	}
	return len(input.ImageURL) <= 512 && len(input.AnimationURL) <= 512 &&
		validAdminResourceURL(input.ImageURL) && validAdminResourceURL(input.AnimationURL)
}

func adminRewardItemFromModel(item model.RewardItem) adminRewardItemRow {
	return adminRewardItemRow{
		ID: item.ID, ItemCode: item.ItemCode, Name: item.Name, ItemType: item.ItemType,
		ImageURL: item.ImageURL, AnimationURL: item.AnimationURL, Rarity: item.Rarity,
		Status: item.Status, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt,
	}
}

func validAdminResourceURL(value string) bool {
	if value == "" || strings.HasPrefix(value, "/") {
		return true
	}
	parsed, err := url.ParseRequestURI(value)
	return err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != ""
}

func adminPathID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		adminRequestError(c, "资源ID无效")
		return 0, false
	}
	return id, true
}

func rewardItemGoingOffline(currentStatus, nextStatus uint8) bool {
	return currentStatus == 1 && nextStatus == 0
}

func onlinePrizePoolRewardReferences(db *gorm.DB, rewardItemID uint64) *gorm.DB {
	return db.Table("prize_pool_rewards AS pr").
		Joins("JOIN prize_pool_versions AS v ON v.id=pr.version_id AND v.status=1").
		Joins("JOIN prize_pools AS p ON p.id=v.prize_pool_id AND p.status=1 AND p.deleted_at IS NULL").
		Where("pr.reward_item_id=?", rewardItemID)
}

func onlinePrizePoolRewardReferenceCount(db *gorm.DB, rewardItemID uint64) (int64, error) {
	var count int64
	err := onlinePrizePoolRewardReferences(db, rewardItemID).Count(&count).Error
	return count, err
}

func adminRequestError(c *gin.Context, message string) {
	response.Error(c, 400, 10001, message)
}

func adminRecordError(c *gin.Context, err error, notFoundMessage string) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Error(c, 404, 13000, notFoundMessage)
		return
	}
	writeError(c, err)
}

func isAdminDuplicateError(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}

func currentAdminActivityID(db *gorm.DB) *uint64 {
	var activity model.Activity
	now := time.Now()
	if err := db.Where("status=2 AND starts_at<=? AND ends_at>? AND deleted_at IS NULL", now, now).First(&activity).Error; err != nil {
		return nil
	}
	id := activity.ID
	return &id
}

func adminMutationRequestID(adminID uint64) string {
	return fmt.Sprintf("ADM%d%d", adminID, time.Now().UnixNano())
}

func (a *AdminController) recordAdminOperation(c *gin.Context, startedAt time.Time, action, targetType, targetID string, body any) {
	requestBody, _ := json.Marshal(body)
	requestID := strings.TrimSpace(c.GetHeader("X-Request-ID"))
	if requestID == "" {
		requestID = adminMutationRequestID(middleware.CurrentAdminID(c))
	}
	path := c.FullPath()
	if path == "" {
		path = c.Request.URL.Path
	}
	duration := time.Since(startedAt).Milliseconds()
	if duration < 0 {
		duration = 0
	}
	log := model.AdminOperationLog{
		AdminUserID: middleware.CurrentAdminID(c), RequestID: truncateAdminString(requestID, 64),
		Method: c.Request.Method, Path: truncateAdminString(path, 255), Action: action,
		TargetType: targetType, TargetID: truncateAdminString(targetID, 64),
		RequestBody: requestBody, ResponseCode: 0, IP: truncateAdminString(c.ClientIP(), 64),
		UserAgent: truncateAdminString(c.Request.UserAgent(), 512), DurationMS: uint(duration),
	}
	_ = a.db.Create(&log).Error
}

func truncateAdminString(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return value[:limit]
}

var (
	errAdminInsufficientBalance         = errors.New("admin adjustment would make balance negative")
	errRewardItemReferencedByOnlinePool = errors.New("reward item is referenced by an online prize pool")
	errAdminPoolRewardItemUnavailable   = errors.New("pool reward item is unavailable")
)
