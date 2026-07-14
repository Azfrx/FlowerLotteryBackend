package model

// ActivityContent is the operator-managed copy stored in activities.rules_json.
// UI labels, validation messages, and runtime result copy remain frontend concerns.
type ActivityContent struct {
	Instructions         ActivityInstructionsContent `json:"instructions"`
	GameGuides           ActivityGameGuidesContent   `json:"game_guides"`
	NewRingWelfare       NewRingWelfareContent       `json:"new_ring_welfare"`
	RankingCustomization ActivityInstructionsContent `json:"ranking_customization"`
}

type ActivityInstructionsContent struct {
	Title           string                       `json:"title"`
	Sections        []ActivityInstructionSection `json:"sections"`
	ProbabilityLink ActivityLink                 `json:"probability_link"`
}

type ActivityInstructionSection struct {
	Title      string                  `json:"title"`
	Paragraphs [][]ActivityTextSegment `json:"paragraphs"`
}

type ActivityTextSegment struct {
	Text        string `json:"text"`
	Highlighted bool   `json:"highlighted,omitempty"`
}

type ActivityLink struct {
	Text string `json:"text"`
	URL  string `json:"url"`
}

type ActivityGameGuidesContent struct {
	Day   [][]ActivityGuideNode `json:"day"`
	Night [][]ActivityGuideNode `json:"night"`
}

type ActivityGuideNode struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

type NewRingWelfareContent struct {
	StoryTitle           string                  `json:"story_title"`
	StoryLines           []string                `json:"story_lines"`
	ValueText            string                  `json:"value_text"`
	SelectionSegments    []ActivityStyledSegment `json:"selection_segments"`
	SelectionNames       []string                `json:"selection_names"`
	FirstPublishSegments []ActivityStyledSegment `json:"first_publish_segments"`
	FirstPublishCaptions []string                `json:"first_publish_captions"`
}

type ActivityStyledSegment struct {
	Text  string `json:"text"`
	Style string `json:"style,omitempty"`
}
