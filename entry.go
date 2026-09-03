package loggerstdoutlibrary

// MainEntry merepresentasikan satu record log untuk flow service utama
type MainEntry struct {
	TraceID         string        `json:"traceId"`
	SpanID          string        `json:"spanId"`
	SeverityText    string        `json:"severityText"`   // TRACE,DEBUG,INFO,WARN,ERROR,FATAL
	SeverityNumber  string        `json:"severityNumber"` // 1-6
	Timestamp       string        `json:"timestamp"`
	Event           Event         `json:"event"`
	Body            interface{}   `json:"body"`
	Resource        Resource      `json:"resource"`
	Attributes      Attributes    `json:"attributes"`
	ResponseCode    string        `json:"responseCode"`
	ResponseMessage string        `json:"responseMessage"`
	StartTime       string        `json:"startTime"`
	EndTime         string        `json:"endTime"`
	Duration        string        `json:"duration"`
	NativeCalls     []*NativeCall `json:"nativeCalls"`
}

type Event struct {
	EventName     string `json:"eventName"`
	EventCategory string `json:"eventCategory"`
}

type Resource struct {
	ServiceName      string `json:"serviceName"`
	ServiceVersion   string `json:"serviceVersion"`
	OperationName    string `json:"operationName"`
	OperationVersion string `json:"operationVersion"`
	LocalIP          string `json:"localIp"`
}

type Attributes struct {
	Channel  string `json:"channel"`
	Layer    string `json:"layer"`
	ClientIP string `json:"clientIp"`
}

// NativeCall merepresentasikan satu pemanggilan ke service/eksternal lain
type NativeCall struct {
	Type            string  `json:"type"`
	URL             string  `json:"URL"`
	StartTime       string  `json:"startTime"`
	EndTime         string  `json:"endTime"`
	Duration        float64 `json:"duration"`
	ResponseCode    string  `json:"responseCode"`
	ResponseMessage string  `json:"responseMessage"`
}
