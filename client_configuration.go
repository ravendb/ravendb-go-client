package ravendb

type ClientConfiguration struct {
	Etag                          int                 `json:"Etag"`
	Disabled                      bool                `json:"Disabled"`
	MaxNumberOfRequestsPerSession int                 `json:"MaxNumberOfRequestsPerSession"`
	ReadBalanceBehavior           ReadBalanceBehavior `json:"ReadBalanceBehavior"`
}

/*
    public long getEtag() {
        return etag;
    }

    public void setEtag(long etag) {
        this.etag = etag;
    }

    public boolean isDisabled() {
        return disabled;
    }

    public void setDisabled(boolean disabled) {
        this.disabled = disabled;
    }

    public Integer getMaxNumberOfRequestsPerSession() {
        return maxNumberOfRequestsPerSession;
    }

    public void setMaxNumberOfRequestsPerSession(Integer maxNumberOfRequestsPerSession) {
        this.maxNumberOfRequestsPerSession = maxNumberOfRequestsPerSession;
    }

    public ReadBalanceBehavior getReadBalanceBehavior() {
        return readBalanceBehavior;
    }

    public void setReadBalanceBehavior(ReadBalanceBehavior readBalanceBehavior) {
        this.readBalanceBehavior = readBalanceBehavior;
    }
}
*/
