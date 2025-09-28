package models

// BaseEvent 基础事件结构（所有平台共同字段）
type BaseEvent struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	CreatedAt string `json:"created_at"`
}

// GitHubEvent GitHub 特定事件结构
type GitHubEvent struct {
	BaseEvent
	Actor struct {
		ID           int64  `json:"id"`
		Login        string `json:"login"`
		DisplayLogin string `json:"display_login"`
		AvatarURL    string `json:"avatar_url"`
		URL          string `json:"url"`
	} `json:"actor"`
	Repo struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"repo"`
	Payload map[string]interface{} `json:"payload"`
	Public  bool                   `json:"public"`
}

// GiteeEvent Gitee 特定事件结构
type GiteeEvent struct {
	BaseEvent
	Actor struct {
		ID          int64  `json:"id"`
		Login       string `json:"login"`
		DisplayName string `json:"display_name"`
		AvatarURL   string `json:"avatar_url"`
		URL         string `json:"url"`
	} `json:"actor"`
	Repo struct {
		ID       int64  `json:"id"`
		FullName string `json:"full_name"`
		HTMLURL  string `json:"html_url"`
	} `json:"repo"`
	Payload map[string]interface{} `json:"payload"`
	Public  bool                   `json:"public"`
}

// UnifiedEvent 统一事件模型（用于内部处理）
type UnifiedEvent struct {
	BaseEvent
	ActorLogin     string                 `json:"actor_login"`
	ActorAvatarURL string                 `json:"actor_avatar_url"`
	RepoName       string                 `json:"repo_name"`
	RepoURL        string                 `json:"repo_url"`
	Payload        map[string]interface{} `json:"payload"`
}

// ToUnifiedEvent 将 GitHubEvent 转换为 UnifiedEvent
func (g *GitHubEvent) ToUnifiedEvent() *UnifiedEvent {
	return &UnifiedEvent{
		BaseEvent:      g.BaseEvent,
		ActorLogin:     g.Actor.Login,
		ActorAvatarURL: g.Actor.AvatarURL,
		RepoName:       g.Repo.Name,
		RepoURL:        g.Repo.URL,
		Payload:        g.Payload,
	}
}

// ToUnifiedEvent 将 GiteeEvent 转换为 UnifiedEvent
func (g *GiteeEvent) ToUnifiedEvent() *UnifiedEvent {
	return &UnifiedEvent{
		BaseEvent:      g.BaseEvent,
		ActorLogin:     g.Actor.Login,
		ActorAvatarURL: g.Actor.AvatarURL,
		RepoName:       g.Repo.FullName,
		RepoURL:        g.Repo.HTMLURL,
		Payload:        g.Payload,
	}
}