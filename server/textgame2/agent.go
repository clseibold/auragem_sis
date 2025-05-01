package textgame2

type AgentId int

type AgentState uint8 // TODO: Limited size for now

const (
	AgentState_Idle AgentState = iota
	AgentState_Sleep
	AgentState_Work // For kids, either playing, working (for underage work), or at school
	AgentState_Ill
	AgentState_MentallyIll
	AgentState_Injured     // Hospitalized
	AgentState_Diseased    // Hospitalized
	AgentState_MentalBreak // Hospitalized
)

type AgentGender uint8

const (
	AgentGender_Unknown AgentGender = iota
	AgentGender_Male
	AgentGender_Female
	AgentGender_TransMale
	AgentGender_TransFemale
	AgentGender_Max
)

// TODO: Disabled, Senior
type Agent struct {
	state              AgentState
	name               string
	familyID           int  // An ID passed down the generations to detect related family members. Usually the agent id of the first original mother of the family.
	illnessState       bool // TODO
	mentalIllnessState bool // TODO
	age                uint8
	food               uint8 // 0 to 100
	health             uint8 // stress, disease, food, and injury affects health deltas
	stress             uint8 // recreation and work environment affects this.
	gender             AgentGender
	sexualAttraction   [AgentGender_Max]bool // TODO: A bitset would work here to reduce memory

	assignedZone ResourceZoneId
}

// TODO: Sexual Attraction (sexual orientation) rates
// TODO: Gender identity rates

func (a *Agent) ChronicPhysicalIllnessRate() float32 {
	if a.illnessState { // TODO
		return 0
	}

	// Chances of getting sick if not already sick
	switch a.age {
	case 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11:
		return 0.019
	case 12, 13, 14, 15, 16, 17:
		return .074
	case 18, 19, 20, 21, 22, 23, 24:
		return .074
	case 25, 26, 27, 28, 29, 30, 31, 32, 33, 34:
		return .3
	case 35, 36, 37, 38, 39, 40, 41, 42, 43, 44:
		return .4
	case 45, 46, 47, 48, 49, 50, 51, 52, 53, 54:
		return .5
	default: // 55+
		return .8
	}
}

func (a *Agent) MentalIllnessRate() float32 {
	if a.mentalIllnessState { // TODO
		return 0
	}

	// Chances of getting mental illness if don't already have it
	switch a.age {
	case 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11:
		return 0.07
	case 12, 13, 14, 15, 16, 17:
		return .31
	case 18, 19, 20, 21, 22, 23, 24:
		return .29
	case 25, 26, 27, 28, 29, 30, 31, 32, 33, 34:
		return .25
	case 35, 36, 37, 38, 39, 40, 41, 42, 43, 44:
		return .20
	case 45, 46, 47, 48, 49, 50, 51, 52, 53, 54:
		return .18
	default: // 55+
		return .15
	}
}
