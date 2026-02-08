package coffeeshop

import (
	"time"
)

type EquipmentType int

const (
	EquipGrinder EquipmentType = iota
	EquipEspressoMachine
	EquipMilkSteamer
	EquipBlender
	EquipWhisk
)

type DrinkType int

const (
	DrinkUnspecified DrinkType = iota
	DrinkEspresso
	DrinkLatte
	DrinkFrappe
	DrinkMatcha
)

type RecipeStep struct {
	Equipment EquipmentType
	Duration  time.Duration
	Semaphore chan struct{}
}

func NewRecipeStep(equipment EquipmentType, duration time.Duration, semaphoreCapacity int) *RecipeStep {
	equip := &RecipeStep{
		Equipment: equipment,
		Duration:  duration,
		Semaphore: make(chan struct{}, semaphoreCapacity),
	}

	return equip
}

var (
	Grinder         = NewRecipeStep(EquipGrinder, 5*time.Millisecond, 1)
	EspressoMachine = NewRecipeStep(EquipEspressoMachine, 8*time.Millisecond, 2)
	MilkSteamer     = NewRecipeStep(EquipMilkSteamer, 15*time.Millisecond, 1)
	Blender         = NewRecipeStep(EquipBlender, 12*time.Millisecond, 1)
	Whisk           = NewRecipeStep(EquipWhisk, 3*time.Millisecond, 2)
)

var Recipes = map[DrinkType][]*RecipeStep{
	DrinkEspresso: {
		Grinder,
		EspressoMachine,
	},
	DrinkLatte: {
		Grinder,
		EspressoMachine,
		MilkSteamer,
	},
	DrinkFrappe: {
		Grinder,
		Blender,
	},
	DrinkMatcha: {
		Grinder,
		MilkSteamer,
		Whisk,
	},
}

type Order struct {
	ID    int64
	Drink DrinkType
}

type StepExecution struct {
	Equipment   EquipmentType
	StartTimeMs int64
	EndTimeMs   int64
}

type OrderResult struct {
	OrderID int64
	Steps   []StepExecution
}
