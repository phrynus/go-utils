package feishu

// TextContent 文本消息内容
type TextContent struct {
	Text string `json:"text"` // 文本内容
}

// PostContent 富文本消息内容
type PostContent struct {
	Post *Post `json:"post"`
}

// Post 富文本
type Post struct {
	ZhCn *PostDetail `json:"zh_cn,omitempty"` // 中文
	EnUs *PostDetail `json:"en_us,omitempty"` // 英文
}

// PostDetail 富文本详情
type PostDetail struct {
	Title   string       `json:"title"`   // 标题
	Content [][]PostElem `json:"content"` // 内容（二维数组，第一维是段落，第二维是段落内的元素）
}

// PostElem 富文本元素
type PostElem struct {
	Tag      string `json:"tag"`                 // 元素类型：text, a, at, img
	Text     string `json:"text,omitempty"`      // 文本内容
	UnEscape bool   `json:"un_escape,omitempty"` // 是否 unescape 解码（仅用于 text 标签）
	Href     string `json:"href,omitempty"`      // 链接地址
	UserId   string `json:"user_id,omitempty"`   // @用户的open_id 或 user_id，@所有人时填 "all"
	UserName string `json:"user_name,omitempty"` // @用户的姓名
	ImageKey string `json:"image_key,omitempty"` // 图片key
}

// ImageContent 图片消息内容
type ImageContent struct {
	ImageKey string `json:"image_key"` // 图片的key
}

// ShareChatContent 分享群名片消息内容
type ShareChatContent struct {
	ChatId string `json:"chat_id"` // 群聊的chat_id
}

// ResponseMeta 响应操作信息
type ResponseMeta struct {
	Code int    `json:"code"` // 错误码，非0表示失败
	Msg  string `json:"msg"`  // 错误信息
}

// FeiShu 飞书客户端
type FeiShu struct {
	webhookURL string // webhook完整URL
	secret     string // 签名密钥
}

// ============ 富文本元素辅助函数 ============

// NewTextElem 创建文本元素
func NewTextElem(text string) PostElem {
	return PostElem{
		Tag:  "text",
		Text: text,
	}
}

// NewTextElemWithUnescape 创建文本元素（带 unescape）
func NewTextElemWithUnescape(text string, unEscape bool) PostElem {
	return PostElem{
		Tag:      "text",
		Text:     text,
		UnEscape: unEscape,
	}
}

// NewLinkElem 创建超链接元素
func NewLinkElem(text, href string) PostElem {
	return PostElem{
		Tag:  "a",
		Text: text,
		Href: href,
	}
}

// NewAtElem 创建@用户元素
func NewAtElem(userId string) PostElem {
	return PostElem{
		Tag:    "at",
		UserId: userId,
	}
}

// NewAtElemWithName 创建@用户元素（带用户名）
func NewAtElemWithName(userId, userName string) PostElem {
	return PostElem{
		Tag:      "at",
		UserId:   userId,
		UserName: userName,
	}
}

// NewAtAllElem 创建@所有人元素
func NewAtAllElem() PostElem {
	return PostElem{
		Tag:    "at",
		UserId: "all",
	}
}

// NewImageElem 创建图片元素
func NewImageElem(imageKey string) PostElem {
	return PostElem{
		Tag:      "img",
		ImageKey: imageKey,
	}
}
