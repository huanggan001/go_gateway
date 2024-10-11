package dto

type PanelGroupDataOutput struct {
	ServiceNum      int64 `json:"serviceNum"`
	AppNum          int64 `json:"appNum"`
	CurrentQPS      int64 `json:"currentQps"`
	TodayRequestNum int64 `json:"todayRequestNum"`
}

type DashBoardServiceStatItemOutput struct {
	Name     string `json:"name"`
	LoadType int    `json:"load_type"`
	Value    int64  `json:"value"`
}

type DashBoardServiceStatOutput struct {
	Legend []string                         `json:"legend"` //服务类型
	Data   []DashBoardServiceStatItemOutput `json:"data"`
}
