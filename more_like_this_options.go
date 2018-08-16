package ravendb

import "math"

const (
	MoreLikeThisOptions_DEFAULT_MAXIMUM_NUMBER_OF_TOKENS_PARSED = 5000
	MoreLikeThisOptions_DEFAULT_MINIMUM_TERM_FREQUENCY          = 2
	MoreLikeThisOptions_DEFAULT_MINIMUM_DOCUMENT_FREQUENCY      = 5
	MoreLikeThisOptions_DEFAULT_MAXIMUM_DOCUMENT_FREQUENCY      = math.MaxInt32
	MoreLikeThisOptions_DEFAULT_BOOST                           = false
	MoreLikeThisOptions_DEFAULT_BOOST_FACTOR                    = 1
	MoreLikeThisOptions_DEFAULT_MINIMUM_WORD_LENGTH             = 0
	MoreLikeThisOptions_DEFAULT_MAXIMUM_WORD_LENGTH             = 0
	MoreLikeThisOptions_DEFAULT_MAXIMUM_QUERY_TERMS             = 25
)

var (
	MoreLikeThisOptions_defaultOptions = &MoreLikeThisOptions{}
)

type MoreLikeThisOptions struct {
	minimumTermFrequency               *int
	maximumQueryTerms                  *int
	maximumNumberOfTokensParsed        *int
	minimumWordLength                  *int
	maximumWordLength                  *int
	minimumDocumentFrequency           *int
	maximumDocumentFrequency           *int
	maximumDocumentFrequencyPercentage *int
	boost                              bool
	boostFactor                        float32
	stopWordsDocumentId                string
	fields                             []string
}

func NewMoreLikeThisOptions() *MoreLikeThisOptions {
	return &MoreLikeThisOptions{}
}

func (o *MoreLikeThisOptions) GetMinimumTermFrequency() *int {
	return o.minimumTermFrequency
}

func (o *MoreLikeThisOptions) SetMinimumTermFrequency(minimumTermFrequency int) {
	o.minimumTermFrequency = &minimumTermFrequency
}

func (o *MoreLikeThisOptions) GetMaximumQueryTerms() *int {
	return o.maximumQueryTerms
}

func (o *MoreLikeThisOptions) SetMaximumQueryTerms(maximumQueryTerms int) {
	o.maximumQueryTerms = &maximumQueryTerms
}

func (o *MoreLikeThisOptions) GetMaximumNumberOfTokensParsed() *int {
	return o.maximumNumberOfTokensParsed
}

func (o *MoreLikeThisOptions) SetMaximumNumberOfTokensParsed(maximumNumberOfTokensParsed int) {
	o.maximumNumberOfTokensParsed = &maximumNumberOfTokensParsed
}

func (o *MoreLikeThisOptions) GetMinimumWordLength() *int {
	return o.minimumWordLength
}

func (o *MoreLikeThisOptions) SetMinimumWordLength(minimumWordLength int) {
	o.minimumWordLength = &minimumWordLength
}

func (o *MoreLikeThisOptions) GetMaximumWordLength() *int {
	return o.maximumWordLength
}

func (o *MoreLikeThisOptions) SetMaximumWordLength(maximumWordLength int) {
	o.maximumWordLength = &maximumWordLength
}

func (o *MoreLikeThisOptions) GetMinimumDocumentFrequency() *int {
	return o.minimumDocumentFrequency
}

func (o *MoreLikeThisOptions) SetMinimumDocumentFrequency(minimumDocumentFrequency int) {
	o.minimumDocumentFrequency = &minimumDocumentFrequency
}

func (o *MoreLikeThisOptions) GetMaximumDocumentFrequency() *int {
	return o.maximumDocumentFrequency
}

func (o *MoreLikeThisOptions) SetMaximumDocumentFrequency(maximumDocumentFrequency int) {
	o.maximumDocumentFrequency = &maximumDocumentFrequency
}

func (o *MoreLikeThisOptions) GetMaximumDocumentFrequencyPercentage() *int {
	return o.maximumDocumentFrequencyPercentage
}

func (o *MoreLikeThisOptions) SetMaximumDocumentFrequencyPercentage(maximumDocumentFrequencyPercentage int) {
	o.maximumDocumentFrequencyPercentage = &maximumDocumentFrequencyPercentage
}

func (o *MoreLikeThisOptions) GetBoost() bool {
	return o.boost
}

func (o *MoreLikeThisOptions) SetBoost(boost bool) {
	o.boost = boost
}

func (o *MoreLikeThisOptions) GetBoostFactor() float32 {
	return o.boostFactor
}

func (o *MoreLikeThisOptions) SetBoostFactor(boostFactor float32) {
	o.boostFactor = boostFactor
}

func (o *MoreLikeThisOptions) GetStopWordsDocumentID() string {
	return o.stopWordsDocumentId
}

func (o *MoreLikeThisOptions) SetStopWordsDocumentID(stopWordsDocumentId string) {
	o.stopWordsDocumentId = stopWordsDocumentId
}

func (o *MoreLikeThisOptions) GetFields() []string {
	return o.fields
}

func (o *MoreLikeThisOptions) SetFields(fields []string) {
	o.fields = fields
}
