package types

type FeedBackInfo struct {
	HonestLoseRate   float64
	AttackerLoseRate float64
}

type FeedBack struct {
	Uid  string
	Info FeedBackInfo
}

type FeedBacker interface {
	GetFeedBack(uid string) (FeedBackInfo, error)
}
