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

var TextGameBuildingType_Sawmill int = 0 // TODO: Rename
var TextGameBuidingType_CoalMine int = 1
var TextGameBuidingType_Steelworks int = 2
var TextGameBuidingType_GatheringPost int = 3
var TextGameBuidingType_MedicalPost int = 4
var TextGameBuidingType_Infirmary int = 5
var TextGameBuidingType_CareHouse int = 6
var TextGameBuidingType_CookHouse int = 7
var TextGameBuidingType_HuntersHut int = 8
var TextGameBuidingType_HotHouse int = 9
var TextGameBuidingType_Workshop int = 10 // Research
var TextGameBuidingType_Watchtower int = 11
var TextGameBuidingType_BathHouse int = 12
var TextGameBuidingType_ForagersQuarters int = 13
var TextGameBuidingType_FishingHarbour int = 14
var TextGameBuidingType_Docks int = 15
var TextGameBuidingType_ReloadingStation int = 16 // For Docks
var TextGameBuidingType_CharcoalKiln int = 17
var TextGameBuildingType_PublicHouse int = 18
var TextGameBuildingType_TelegraphStation int = 19
var TextGameBuildingType_LabourUnion int = 20
var TextGameBuildingType_Chapel int = 21
var TextGameBuildingType_Temple int = 22
var TextGameBuildingType_TransportDepot int = 23

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
