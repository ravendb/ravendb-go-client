package ravendb

type DatabasePromotionStatus = string

const (
	DatabasePromotionStatusWaitingForFirstPromotion = "WaitingForFirstPromotion"
	DatabasePromotionStatusNotResponding            = "NotResponding"
	DatabasePromotionStatusIndexNotUpToDate         = "IndexNotUpToDate"
	DatabasePromotionStatusChangeVectorNotMerged    = "ChangeVectorNotMerged"
	DatabasePromotionStatusWaitingForResponse       = "WaitingForResponse"
	DatabasePromotionStatusOk                       = "Ok"
)
