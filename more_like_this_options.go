package ravendb

import "math"

const (
	MoreLikeThisOptionsDefaultMaximumNumberOfTokensParsed = 5000
	MoreLikeThisOptionsDefaultMinimumTermFrequency        = 2
	MoreLikeThisOptionsDefaultMinimumDocumentFrequency    = 5
	MoreLikeThisOptionsDefaultMaximumDocumentFrequence    = math.MaxInt32
	MoreLikeThisOptionsDefaultBoost                       = false
	MoreLikeThisOptionsDefaultBoostFactor                 = 1
	MoreLikeThisOptionsDefaultMinimumWordLength           = 0
	MoreLikeThisOptionsDefaultMaximumWordLength           = 0
	MoreLikeThisOptionsDefaultMaximumQueryTerms           = 25
)

var (
	defaultMoreLikeThisOptions = &MoreLikeThisOptions{}
)

type MoreLikeThisOptions struct {
	MinimumTermFrequency               *int     `json:"MinimumTermFrequency"`
	MaximumQueryTerms                  *int     `json:"MaximumQueryTerms"`
	MaximumNumberOfTokensParsed        *int     `json:"MaximumNumberOfTokensParsed"`
	MinimumWordLength                  *int     `json:"MinimumWordLength"`
	MaximumWordLength                  *int     `json:"MaximumWordLength"`
	MinimumDocumentFrequency           *int     `json:"MinimumDocumentFrequency"`
	MaximumDocumentFrequency           *int     `json:"MaximumDocumentFrequency"`
	MaximumDocumentFrequencyPercentage *int     `json:"MaximumDocumentFrequencyPercentage"`
	Boost                              *bool    `json:"Boost"`
	BoostFactor                        *float32 `json:"BoostFactor"`
	StopWordsDocumentID                *string  `json:"StopWordsDocumentId"`
	Fields                             []string `json:"Fields"`
}

func NewMoreLikeThisOptions() *MoreLikeThisOptions {
	return &MoreLikeThisOptions{}
}

func (o *MoreLikeThisOptions) SetMinimumTermFrequency(minimumTermFrequency int) {
	o.MinimumTermFrequency = &minimumTermFrequency
}

func (o *MoreLikeThisOptions) SetMaximumQueryTerms(maximumQueryTerms int) {
	o.MaximumQueryTerms = &maximumQueryTerms
}

func (o *MoreLikeThisOptions) SetMaximumNumberOfTokensParsed(maximumNumberOfTokensParsed int) {
	o.MaximumNumberOfTokensParsed = &maximumNumberOfTokensParsed
}

func (o *MoreLikeThisOptions) SetMinimumWordLength(minimumWordLength int) {
	o.MinimumWordLength = &minimumWordLength
}

func (o *MoreLikeThisOptions) SetMaximumWordLength(maximumWordLength int) {
	o.MaximumWordLength = &maximumWordLength
}

func (o *MoreLikeThisOptions) SetMinimumDocumentFrequency(minimumDocumentFrequency int) {
	o.MinimumDocumentFrequency = &minimumDocumentFrequency
}

func (o *MoreLikeThisOptions) SetMaximumDocumentFrequency(maximumDocumentFrequency int) {
	o.MaximumDocumentFrequency = &maximumDocumentFrequency
}

func (o *MoreLikeThisOptions) SetMaximumDocumentFrequencyPercentage(maximumDocumentFrequencyPercentage int) {
	o.MaximumDocumentFrequencyPercentage = &maximumDocumentFrequencyPercentage
}

func (o *MoreLikeThisOptions) SetBoost(boost bool) {
	o.Boost = &boost
}

func (o *MoreLikeThisOptions) SetBoostFactor(boostFactor float32) {
	o.BoostFactor = &boostFactor
}

func (o *MoreLikeThisOptions) SetStopWordsDocumentID(stopWordsDocumentID string) {
	o.StopWordsDocumentID = &stopWordsDocumentID
}
