package payload

type Payload struct {
	Category      string `json:"category"`
	Event         string `json:"event"`
	LogDate       string `json:"log_date"`
	Hour          string `json:"hour"`
	PlatID        int32  `json:"plat_id"`
	AppID         int32  `json:"app_id"`
	Ts            int64  `json:"ts"` // ms
	SchemaVersion int32  `json:"schema_version"`
}
