package models

// Message 消息结构（用于 Issue 中的消息列表）
type Message struct {
	Sender    string `json:"sender"`    // 发送方标识（taskPathName 或 "system"）
	Timestamp int64  `json:"timestamp"` // 时间戳（毫秒）
	Content   string `json:"content"`   // 消息内容
	Read      bool   `json:"read"`      // 是否已读
}

// Issue 问答/问题模型
type Issue struct {
	ID               string `json:"id" gorm:"primaryKey;type:text"`
	FromTaskPathName string `json:"fromTaskPathName" gorm:"not null;type:text;index"`
	ToTaskPathName   string `json:"toTaskPathName" gorm:"not null;type:text;index"`
	Type             string `json:"type" gorm:"not null;type:text;index"` // contract/test/other
	Title            string `json:"title" gorm:"not null;type:text"`
	Status           string `json:"status" gorm:"not null;default:'pending';type:text;index"` // pending/replied/resolved
	Messages         string `json:"messages" gorm:"type:text"` // JSON array of Message
	CreatedAt        int64  `json:"createdAt" gorm:"not null"`
	UpdatedAt        int64  `json:"updatedAt" gorm:"not null"`
	Version          int    `json:"version" gorm:"not null;default:1"`
	SyncStatus       string `json:"syncStatus" gorm:"not null;default:'SYNCED';type:text"`
}

// TableName 指定表名
func (Issue) TableName() string {
	return "issues"
}
