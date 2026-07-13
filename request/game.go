package request

type Lottery struct {
	PoolCode      string `json:"pool_code" binding:"required,oneof=day night"`
	DrawCount     uint   `json:"draw_count" binding:"required,oneof=1 10 30"`
	SkipAnimation bool   `json:"skip_animation"`
	RequestID     string `json:"request_id" binding:"required,max=64"`
}
type Action struct {
	RequestID string `json:"request_id" binding:"required,max=64"`
}
type SelectChest struct {
	CandidateID uint64 `json:"candidate_id" binding:"required"`
	RequestID   string `json:"request_id" binding:"required,max=64"`
}
