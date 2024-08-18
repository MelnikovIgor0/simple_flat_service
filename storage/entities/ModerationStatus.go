package entities

type ModerationStatus string

const (
	CREATED       ModerationStatus = "created"
	APPROVED                       = "approved"
	DECLINED                       = "declined"
	ON_MODERATION                  = "on_moderation"
)
