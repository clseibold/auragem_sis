package textgame2

type BuildingType uint8 // uint8 max is 255, uint16 max is 65535

const (
	BuildingType_Garden BuildingType = iota
	BuildingType_Sawmill
	BuidingType_Steelworks
	BuidingType_CoalMine
	BuidingType_GatheringPost
	BuidingType_MedicalPost
	BuidingType_Infirmary
	BuidingType_CareHouse
	BuidingType_CookHouse
	BuidingType_HuntersHut
	BuidingType_HotHouse
	BuidingType_Workshop // Research
	BuidingType_Watchtower
	BuidingType_BathHouse
	BuidingType_ForagersQuarters
	BuidingType_FishingHarbour
	BuidingType_Docks
	BuidingType_ReloadingStation // For Docks
	BuidingType_CharcoalKiln
	BuildingType_PublicHouse
	BuildingType_TelegraphStation
	BuildingType_LabourUnion
	BuildingType_Chapel
	BuildingType_Temple
	BuildingType_TransportDepot
	BuildingType_Max
)

type Building struct {
	t            BuildingType
	upgradeLevel uint
}

// How many ticks per resource collection cycle for each building type
var TicksPerCollectionCycle = [BuildingType_Max]int{
	BuildingType_Garden: 1 * 60 * 60 / InGameSecondsPerTick, // 1 in-game hour (igh)
}
