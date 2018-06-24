package ravendb

type DatabasePromotionStatus = string

const (
	DatabasePromotionStatus_WAITING_FOR_FIRST_PROMOTION = "WaitingForFirstPromotion"
	DatabasePromotionStatus_NOT_RESPONDING              = "NotResponding"
	DatabasePromotionStatus_INDEX_NOT_UP_TO_DATE        = "IndexNotUpToDate"
	DatabasePromotionStatus_CHANGE_VECTOR_NOT_MERGED    = "ChangeVectorNotMerged"
	DatabasePromotionStatus_WAITING_FOR_RESPONSE        = "WaitingForResponse"
	DatabasePromotionStatus_OK                          = "Ok"
)
