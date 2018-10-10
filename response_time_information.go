package ravendb

import "time"

type ResponseTimeInformation struct {
	totalServerDuration time.Duration
	totalClientDuration time.Duration

	durationBreakdown []ResponseTimeItem
}

func (i *ResponseTimeInformation) computeServerTotal() {
	var total time.Duration
	for _, rti := range i.durationBreakdown {
		total += rti.duration
	}
	i.totalServerDuration = total
}

/*

   public ResponseTimeInformation() {
       totalServerDuration = Duration.ZERO;
       totalClientDuration = Duration.ZERO;
       durationBreakdown = new ArrayList<>();
   }

   public Duration getTotalServerDuration() {
       return totalServerDuration;
   }

   public void setTotalServerDuration(Duration totalServerDuration) {
       this.totalServerDuration = totalServerDuration;
   }

   public Duration getTotalClientDuration() {
       return totalClientDuration;
   }

   public void setTotalClientDuration(Duration totalClientDuration) {
       this.totalClientDuration = totalClientDuration;
   }

   public List<ResponseTimeItem> getDurationBreakdown() {
       return durationBreakdown;
   }

   public void setDurationBreakdown(List<ResponseTimeItem> durationBreakdown) {
       this.durationBreakdown = durationBreakdown;
   }
*/

type ResponseTimeItem struct {
	url      string
	duration time.Duration
}
