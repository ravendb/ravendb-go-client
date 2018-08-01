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

func (o *MoreLikeThisOptions) getMinimumTermFrequency() *int {
	return o.minimumTermFrequency
}

func (o *MoreLikeThisOptions) setMinimumTermFrequency(minimumTermFrequency int) {
	o.minimumTermFrequency = &minimumTermFrequency
}

func (o *MoreLikeThisOptions) getMaximumQueryTerms() *int {
	return o.maximumQueryTerms
}

func (o *MoreLikeThisOptions) setMaximumQueryTerms(maximumQueryTerms int) {
	o.maximumQueryTerms = &maximumQueryTerms
}

func (o *MoreLikeThisOptions) getMaximumNumberOfTokensParsed() *int {
	return o.maximumNumberOfTokensParsed
}

func (o *MoreLikeThisOptions) setMaximumNumberOfTokensParsed(maximumNumberOfTokensParsed int) {
	o.maximumNumberOfTokensParsed = &maximumNumberOfTokensParsed
}

func (o *MoreLikeThisOptions) getMinimumWordLength() *int {
	return o.minimumWordLength
}

func (o *MoreLikeThisOptions) setMinimumWordLength(minimumWordLength int) {
	o.minimumWordLength = &minimumWordLength
}

func (o *MoreLikeThisOptions) getMaximumWordLength() *int {
	return o.maximumWordLength
}

func (o *MoreLikeThisOptions) setMaximumWordLength(maximumWordLength int) {
	o.maximumWordLength = &maximumWordLength
}

func (o *MoreLikeThisOptions) getMinimumDocumentFrequency() *int {
	return o.minimumDocumentFrequency
}

func (o *MoreLikeThisOptions) setMinimumDocumentFrequency(minimumDocumentFrequency int) {
	o.minimumDocumentFrequency = &minimumDocumentFrequency
}

func (o *MoreLikeThisOptions) getMaximumDocumentFrequency() *int {
	return o.maximumDocumentFrequency
}

func (o *MoreLikeThisOptions) setMaximumDocumentFrequency(maximumDocumentFrequency int) {
	o.maximumDocumentFrequency = &maximumDocumentFrequency
}

func (o *MoreLikeThisOptions) getMaximumDocumentFrequencyPercentage() *int {
	return o.maximumDocumentFrequencyPercentage
}

func (o *MoreLikeThisOptions) setMaximumDocumentFrequencyPercentage(maximumDocumentFrequencyPercentage int) {
	o.maximumDocumentFrequencyPercentage = &maximumDocumentFrequencyPercentage
}

func (o *MoreLikeThisOptions) getBoost() bool {
	return o.boost
}

func (o *MoreLikeThisOptions) setBoost(boost bool) {
	o.boost = boost
}

func (o *MoreLikeThisOptions) getBoostFactor() float32 {
	return o.boostFactor
}

func (o *MoreLikeThisOptions) setBoostFactor(boostFactor float32) {
	o.boostFactor = boostFactor
}

func (o *MoreLikeThisOptions) getStopWordsDocumentId() string {
	return o.stopWordsDocumentId
}

func (o *MoreLikeThisOptions) setStopWordsDocumentId(stopWordsDocumentId string) {
	o.stopWordsDocumentId = stopWordsDocumentId
}

func (o *MoreLikeThisOptions) getFields() []string {
	return o.fields
}

func (o *MoreLikeThisOptions) setFields(fields []string) {
	o.fields = fields
}
