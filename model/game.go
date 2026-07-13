package model

import "time"

type Activity struct {
	ID                   uint64 `gorm:"primaryKey"`
	Code                 string
	Name                 string
	Status               uint8
	StartsAt             time.Time
	EndsAt               time.Time
	LeaderboardFreezesAt time.Time
	Timezone             string
	RulesJSON            []byte `gorm:"column:rules_json"`
	ResourceJSON         []byte `gorm:"column:resource_json"`
	CreatedAt            time.Time
	UpdatedAt            time.Time
	DeletedAt            *time.Time
}

func (Activity) TableName() string { return "activities" }

type PrizePool struct {
	ID                  uint64 `gorm:"primaryKey"`
	ActivityID          uint64
	Code                string
	Name                string
	PetalCostPerDraw    uint64
	CoinValuePerDraw    uint64
	SupportedDrawCounts []byte
	Status              uint8
	SortNo              int
	CreatedAt           time.Time
	UpdatedAt           time.Time
	DeletedAt           *time.Time
}

func (PrizePool) TableName() string { return "prize_pools" }

type PrizePoolVersion struct {
	ID          uint64 `gorm:"primaryKey"`
	PrizePoolID uint64
	VersionNo   uint
	Status      uint8
	EffectiveAt *time.Time
	TotalWeight uint64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (PrizePoolVersion) TableName() string { return "prize_pool_versions" }

type RewardItem struct {
	ID           uint64 `gorm:"primaryKey"`
	ItemCode     string
	Name         string
	ItemType     string
	ImageURL     string
	AnimationURL string
	Rarity       string
	Status       uint8
	Extra        []byte
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

func (RewardItem) TableName() string { return "reward_items" }

type PrizePoolReward struct {
	ID              uint64 `gorm:"primaryKey"`
	VersionID       uint64
	RewardItemID    uint64
	Quantity        uint64
	Weight          uint64
	ChoiceGroupCode string
	Snapshot        []byte
	SortNo          int
	RewardItem      RewardItem `gorm:"foreignKey:RewardItemID"`
}

func (PrizePoolReward) TableName() string { return "prize_pool_rewards" }

type FlowerLightRule struct {
	ID                  uint64 `gorm:"primaryKey"`
	ActivityID          uint64
	FlowerPosition      uint8
	DayProbabilityPPM   uint
	NightProbabilityPPM uint
	GuaranteeCoinTotal  uint64
	Status              uint8
}

func (FlowerLightRule) TableName() string { return "flower_light_rules" }

type StageRewardRule struct {
	ID              uint64 `gorm:"primaryKey"`
	ActivityID      uint64
	RequiredFlowers uint8
	RewardItemID    uint64
	Quantity        uint64
	Status          uint8
	SortNo          int
	RewardItem      RewardItem `gorm:"foreignKey:RewardItemID"`
}

func (StageRewardRule) TableName() string { return "stage_reward_rules" }

type ChestRewardRule struct {
	ID           uint64 `gorm:"primaryKey"`
	ActivityID   uint64
	ChestNo      uint8
	RewardItemID uint64
	Quantity     uint64
	Weight       uint64
	Status       uint8
	RewardItem   RewardItem `gorm:"foreignKey:RewardItemID"`
}

func (ChestRewardRule) TableName() string { return "chest_reward_rules" }

type UserActivityRound struct {
	ID                  uint64 `gorm:"primaryKey"`
	UserID              uint64
	ActivityID          uint64
	RoundNo             uint
	LitFlowerCount      uint8
	CumulativeCoinValue uint64
	ChestGrantedCount   uint8
	ChestProcessedCount uint8
	Status              uint8
	CompletedAt         *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

func (UserActivityRound) TableName() string { return "user_activity_rounds" }

type LotteryOrder struct {
	ID                    uint64 `gorm:"primaryKey"`
	OrderNo               string
	UserID                uint64
	ActivityID            uint64
	PrizePoolID           uint64
	PoolVersionID         uint64
	RoundID               uint64
	OrderType             string
	RequestedDrawCount    uint
	ExecutedDrawCount     uint
	PetalCost             uint64
	PetalRefund           uint64
	CoinPayment           uint64
	FlowersBefore         uint8
	FlowersAfter          uint8
	LeaderboardScoreAdded uint64
	Status                uint8
	RequestID             string
	ResultSnapshot        []byte
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

func (LotteryOrder) TableName() string { return "lottery_orders" }

type LotteryDraw struct {
	ID                   uint64 `gorm:"primaryKey"`
	LotteryOrderID       uint64
	DrawIndex            uint
	RewardItemID         uint64
	RewardQuantity       uint64
	RewardSnapshot       []byte
	RandomValue          uint64
	FlowerLit            uint8
	FlowerPosition       *uint8
	FlowerRandomValue    *uint
	FlowerProbabilityPPM *uint
	GuaranteeTriggered   uint8
	CreatedAt            time.Time
}

func (LotteryDraw) TableName() string { return "lottery_draws" }

type FlowerLightRecord struct {
	ID                  uint64 `gorm:"primaryKey"`
	UserID              uint64
	ActivityID          uint64
	RoundID             uint64
	LotteryDrawID       uint64
	FlowerPosition      uint8
	TriggerType         string
	CumulativeCoinValue uint64
	CreatedAt           time.Time
}

func (FlowerLightRecord) TableName() string { return "flower_light_records" }

type UserChestOpportunity struct {
	ID                uint64 `gorm:"primaryKey"`
	UserID            uint64
	ActivityID        uint64
	RoundID           uint64
	ChestNo           uint8
	UnlockFlowerCount uint8
	Status            uint8
	OpenedAt          *time.Time
	SelectedAt        *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (UserChestOpportunity) TableName() string { return "user_chest_opportunities" }

type UserChestCandidate struct {
	ID             uint64 `gorm:"primaryKey"`
	OpportunityID  uint64
	RewardItemID   uint64
	Quantity       uint64
	RewardSnapshot []byte
	Selected       uint8
	RewardItem     RewardItem `gorm:"foreignKey:RewardItemID"`
	CreatedAt      time.Time
}

func (UserChestCandidate) TableName() string { return "user_chest_candidates" }

type UserStageRewardClaim struct {
	ID                uint64 `gorm:"primaryKey"`
	UserID            uint64
	ActivityID        uint64
	RoundID           uint64
	StageRewardRuleID uint64
	Status            uint8
	RequestID         string
	ClaimedAt         time.Time
}

func (UserStageRewardClaim) TableName() string { return "user_stage_reward_claims" }

type UserReward struct {
	ID             uint64 `gorm:"primaryKey"`
	UserID         uint64
	ActivityID     uint64
	RewardItemID   uint64
	Quantity       uint64
	SourceType     string
	SourceID       *uint64
	Status         uint8
	RewardSnapshot []byte
	GrantedAt      *time.Time
	ExpiresAt      *time.Time
	Remark         string
	RewardItem     RewardItem `gorm:"foreignKey:RewardItemID"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (UserReward) TableName() string { return "user_rewards" }

type LeaderboardEntry struct {
	ID         uint64 `gorm:"primaryKey"`
	ActivityID uint64
	UserID     uint64
	Score      uint64
	ReachedAt  time.Time
	UpdatedAt  time.Time
	User       User `gorm:"foreignKey:UserID"`
}

func (LeaderboardEntry) TableName() string { return "leaderboard_entries" }
