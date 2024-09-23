package textgame

import (
	"sync/atomic"
	"time"
)

// All actions happen in segments of 30 minutes?

type TextGameContext struct {
	coal TextGameResource
	wood TextGameResource
	iron TextGameResource
	//stone TextGameResource
	buildings   []TextGameBuilding
	players     []TextGamePlayer
	leader      TextGamePlayerID
	currentTime time.Time // In UTC
}

type TextGameResource struct {
	amount atomic.Int64
}

type TextGameBuildingID int
type TextGameBuildingType int

const (
	TextGameBuildingType_Sawmill TextGameBuildingType = iota // Rename
	TextGameBuidingType_CoalMine
	TextGameBuidingType_Steelworks
	TextGameBuidingType_GatheringPost
	TextGameBuidingType_MedicalPost
	TextGameBuidingType_Infirmary
	TextGameBuidingType_CareHouse
	TextGameBuidingType_CookHouse
	TextGameBuidingType_HuntersHut
	TextGameBuidingType_HotHouse
	TextGameBuidingType_Workshop // Research
	TextGameBuidingType_Watchtower
	TextGameBuidingType_BathHouse
	TextGameBuidingType_ForagersQuarters
	TextGameBuidingType_FishingHarbour
	TextGameBuidingType_Docks
	TextGameBuidingType_ReloadingStation // For Docks
	TextGameBuidingType_CharcoalKiln
	TextGameBuildingType_PublicHouse
	TextGameBuildingType_TelegraphStation
	TextGameBuildingType_LabourUnion
	TextGameBuildingType_Chapel
	TextGameBuildingType_Temple
	TextGameBuildingType_TransportDepot
)


// Storage (Resource Depot)

type TextGameBuilding struct {
	t TextGameBuildingType
}

type TextGamePlayerID int
type TextGamePlayer struct {
	id               TextGamePlayerID
	assignedBuilding TextGameBuildingID
}

func (ctx *TextGameContext) timer() {
	ticker := time.NewTicker(time.Minute * 30)
	for {
		ctx.currentTime = <-ticker.C
	}
}

func (ctx *TextGameContext) IsWorkTime() bool {
	workTime_start := time.Date(ctx.currentTime.Year(), ctx.currentTime.Month(), ctx.currentTime.Day(), 8, 0, 0, 0, ctx.currentTime.Location())
	workTime_end := time.Date(ctx.currentTime.Year(), ctx.currentTime.Month(), ctx.currentTime.Day(), 12+6, 0, 0, 0, ctx.currentTime.Location())
	return ctx.currentTime.Equal(workTime_start) || (ctx.currentTime.After(workTime_start) && ctx.currentTime.Before(workTime_end))
}
func (ctx *TextGameContext) IsFreeTime() bool {
	freeTime_start := time.Date(ctx.currentTime.Year(), ctx.currentTime.Month(), ctx.currentTime.Day(), 12+6, 0, 0, 0, ctx.currentTime.Location())
	midnight := time.Date(ctx.currentTime.Year(), ctx.currentTime.Month(), ctx.currentTime.Day()+1, 0, 0, 0, 0, ctx.currentTime.Location()) // Tomorrow
	return ctx.currentTime.Equal(freeTime_start) || (ctx.currentTime.After(freeTime_start) && ctx.currentTime.Before(midnight))
}
func (ctx *TextGameContext) IsSleepTime() bool {
	midnight := time.Date(ctx.currentTime.Year(), ctx.currentTime.Month(), ctx.currentTime.Day(), 0, 0, 0, 0, ctx.currentTime.Location())
	workTime_start := time.Date(ctx.currentTime.Year(), ctx.currentTime.Month(), ctx.currentTime.Day(), 8, 0, 0, 0, ctx.currentTime.Location())
	return ctx.currentTime.Equal(midnight) || (ctx.currentTime.After(midnight) && ctx.currentTime.Before(workTime_start))
}
