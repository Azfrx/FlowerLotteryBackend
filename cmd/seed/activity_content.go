package main

import "flower-lottery-backend/model"

func activityText(text string) model.ActivityTextSegment {
	return model.ActivityTextSegment{Text: text}
}

func activityHighlight(text string) model.ActivityTextSegment {
	return model.ActivityTextSegment{Text: text, Highlighted: true}
}

func guideNode(nodeType, content string) model.ActivityGuideNode {
	return model.ActivityGuideNode{Type: nodeType, Content: content}
}

func styledText(text string) model.ActivityStyledSegment {
	return model.ActivityStyledSegment{Text: text}
}

func styledHighlight(text string) model.ActivityStyledSegment {
	return model.ActivityStyledSegment{Text: text, Style: "highlight"}
}

func styledTag(text string) model.ActivityStyledSegment {
	return model.ActivityStyledSegment{Text: text, Style: "tag"}
}

func activityContentSeed() model.ActivityContent {
	return model.ActivityContent{
		Instructions: model.ActivityInstructionsContent{
			Title: "活动说明",
			Sections: []model.ActivityInstructionSection{
				{
					Title: "一、活动时间",
					Paragraphs: [][]model.ActivityTextSegment{
						{activityText("活动上线至2026年4月14日23:59:59")},
					},
				},
				{
					Title: "二、活动玩法",
					Paragraphs: [][]model.ActivityTextSegment{
						{
							activityText("1、活动期间可选择"),
							activityHighlight("白昼许愿"),
							activityText("或"),
							activityHighlight("星夜许愿"),
							activityText("两种许愿方式，"),
							activityHighlight("星夜许愿必得稀有及以上等级奖励"),
							activityText("。"),
						},
						{activityText("2、每次许愿必得奖励，有概率获得真爱无敌戒指。")},
						{
							activityText("3、每次许愿还有概率点亮花盒中的花朵，"),
							activityHighlight("白昼许愿每进行1次许愿有概率点亮花盒中的1朵花，星夜许愿每进行1次许愿有概率点亮花盒中的多朵花"),
							activityText("。"),
						},
						{
							activityText("4、每次点亮花盒中的"),
							activityHighlight("6朵花"),
							activityText("时，可获得1次开启戒指宝箱召唤戒指的机会，必定获得"),
							activityHighlight("月亮游记戒指、爱意翩跹戒指或真爱无敌戒指自选权之一"),
							activityText("，通过该渠道获得的戒指不会重复获得，每开启3次宝箱后状态重置。"),
						},
						{
							activityText("5、真爱无敌戒指共有三种形态："),
							activityHighlight("经典款、真爱无敌铭文款、爱你万年铭文款"),
							activityText("，在本次活动中通过许愿或集花朵召唤戒指获得真爱无敌戒指选择权后可在以上三种形态中自选1枚领取。"),
						},
						{activityText("6、每点亮18朵花，召唤3次戒指宝箱中的戒指视为完成一轮，之后轮次状态重置。")},
						{
							activityText("7、"),
							activityHighlight("每一轮内累计点亮的花朵数达标"),
							activityText("可领取对应的"),
							activityHighlight("爱意翩跹戒指、全新真爱降临主页特效或月亮游记戒指"),
							activityText("等丰厚奖励。"),
						},
					},
				},
				{
					Title: "三、新戒福利",
					Paragraphs: [][]model.ActivityTextSegment{
						{activityText("活动期间，使用真爱无敌戒指任意一款挂戒，CP双方可获得66天真爱无敌称号，15天挚爱一生戒指墙、15天挚爱一生戒指框，重复挂戒奖励天数可叠加。")},
					},
				},
				{
					Title: "四、活动须知",
					Paragraphs: [][]model.ActivityTextSegment{
						{activityText("1、活动中获得的礼物卡有效期为30天，需尽快使用。")},
						{activityText("2、本活动涉及的相关时间节点皆为：东八区（GMT+8）。")},
						{activityText("3、本活动5折花瓣礼包限购1次，请按正常流程购买，勿用非正常手段多次支付。若出现重复支付，超出购买次数限制的订单将按 1 元 = 100 金币的比例发放对应金币。为保障您的购买体验，避免困扰，后续购买礼包时请注意并遵守购买上限，勿超限制支付。")},
						{activityText("4、本次活动禁止未成年人参加，请用户理性消费、适度参与。用户通过本次活动获得的所有虚拟道具仅限于平台内使用，不得反向兑换为法定货币、现金或其他任何具有交换价值的物品，不得用于任何形式的盈利活动。用户通过本次活动获得的所有虚拟物品均不得被转让、赠送，仅供用户自身使用。若用户出现违反会玩平台规则的行为，我们将采取包括但不限于扣除收益、永久封号等处罚措施。")},
					},
				},
			},
			ProbabilityLink: model.ActivityLink{
				Text: "5、点击查看概率公示",
				URL:  "https://huiwan.wepie.com/probability-preview",
			},
		},
		GameGuides: model.ActivityGameGuidesContent{
			Day: [][]model.ActivityGuideNode{
				{
					guideNode("common", "每次许愿获得"),
					guideNode("tag", "臻品"),
					guideNode("text", "真爱无敌戒指、"),
					guideNode("tag", "极品"),
					guideNode("text", "烟花之恋戒指等奖励之一"),
				},
				{
					guideNode("common", "每次许愿概率"),
					guideNode("text", "点亮1朵花，每次点亮花盒6朵花，可在以下3枚戒指中召唤1枚获得"),
					guideNode("common", "(召唤奖励不会重复获得)"),
				},
			},
			Night: [][]model.ActivityGuideNode{
				{
					guideNode("common", "每次许愿获得"),
					guideNode("tag", "臻品"),
					guideNode("text", "真爱无敌戒指、"),
					guideNode("tag", "臻品"),
					guideNode("text", "爱意翩跹戒指,"),
					guideNode("tag", "臻品"),
					guideNode("text", "闪闪心蝶戒指等奖励之一"),
				},
				{
					guideNode("common", "每次许愿概率"),
					guideNode("text", "点亮1朵花，每次点亮花盒6朵花，可在以下3枚戒指中召唤1枚获得"),
					guideNode("common", "(召唤奖励不会重复获得)"),
				},
			},
		},
		NewRingWelfare: model.NewRingWelfareContent{
			StoryTitle: "戒指物语：",
			StoryLines: []string{
				"戒指物语：True love conquers all",
				"世间万物皆有敌，唯有真爱是无敌。",
			},
			ValueText: "使用后增加被求婚者魅力值59999，守护值59999",
			SelectionSegments: []model.ActivityStyledSegment{
				styledText("在本次活动中通过"),
				styledHighlight("许愿"),
				styledText("或"),
				styledHighlight("集花盒花朵召唤戒指"),
				styledText("获得真爱无敌戒指选择权后可在"),
				styledHighlight("经典款、真爱无敌铭文款、爱你万年铭文款"),
				styledText("中自选1枚领取。"),
			},
			SelectionNames: []string{"经典款", "真爱无敌铭文款", "爱你万年铭文款"},
			FirstPublishSegments: []model.ActivityStyledSegment{
				styledText("在本次活动期间使用以上三款戒指挂戒，挂戒双方均可获得「挚爱一生戒指墙&戒指框」装扮15天及"),
				styledTag("真爱无敌"),
				styledText("称号66天，天数可叠加！"),
			},
			FirstPublishCaptions: []string{"挚爱一生戒指墙装扮", "挚爱一生戒指框装扮"},
		},
		RankingCustomization: model.ActivityInstructionsContent{
			Title: "定制称号说明",
			Sections: []model.ActivityInstructionSection{
				{
					Title: "【定制&修改V8称号】说明：",
					Paragraphs: [][]model.ActivityTextSegment{
						{activityText("1、定制称号的图标、背景色、文字均可定制，文字最多支持10个字符（称号最多5个字，内容不可包含特殊符号及emoji表情，字体不可更改样式及颜色）。1个中文字、1个符号为2个字符，1个英文字母为1个字符；")},
						{activityText("2、称号文字为系统预设字体，不可更改样式及颜色；")},
						{activityText("3、修改V8称号，仅支持在当前已拥有的V8称号基础上，修改原称号的“图标+文字”或“背景+文字”；")},
						{activityText("4、定制称号的动效统一为：溜光动效加图标左右两个光点闪烁动效，不支持其它动效要求；")},
						{activityText("5、定制/修改特权不能转让或售卖，不能涉及广告、谩骂等不良信息。否则，官方有权拒绝违规冠名信息。")},
					},
				},
				{
					Title: "【头像框冠名】说明：",
					Paragraphs: [][]model.ActivityTextSegment{
						{activityText("1、冠名可在指定头像框的指定位置展示第一名玩家本人昵称，成为全服独有的专属款式；冠名头像框的样式、冠名文案颜色/字体/位置均按预览（尚未冠名）的版本进行制作。")},
						{activityText("2、昵称需为近30天内至少使用10天以上的常用昵称，最多8个字符。1个中文字、1个符号为2个字符，1个英文字母为1个字符；")},
						{activityText("3、冠名特权不能转让或售卖，不能涉及广告、谩骂等不良信息。否则，官方有权拒绝违规冠名信息；")},
						{activityText("4、冠名特权需在活动结束后的30天内完成冠名，否则自动失效。")},
					},
				},
			},
		},
	}
}
