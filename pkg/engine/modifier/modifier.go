package modifier

import (
	"github.com/simimpact/srsim/pkg/engine"
	"github.com/simimpact/srsim/pkg/engine/info"
	"github.com/simimpact/srsim/pkg/key"
	"github.com/simimpact/srsim/pkg/model"
)

type ModifierInstance struct {
	name              key.Modifier
	owner             key.TargetID
	source            key.TargetID
	params            map[string]float64
	tickImmediately   bool
	duration          int
	count             int
	maxCount          int
	countAddWhenStack int
	stats             info.PropMap
	debuffRES         info.DebuffRESMap
	renewTurn         int
	manager           *Manager
	listeners         Listeners
	statusType        model.StatusType
	flags             []model.BehaviorFlag
}

func (mgr *Manager) newInstance(owner key.TargetID, mod info.Modifier) *ModifierInstance {
	config := modifierCatalog[mod.Name]
	mi := &ModifierInstance{
		name:              mod.Name,
		owner:             owner,
		source:            mod.Source,
		params:            mod.Params,
		tickImmediately:   mod.TickImmediately,
		duration:          mod.Duration,
		count:             mod.Count,
		maxCount:          mod.MaxCount,
		countAddWhenStack: mod.CountAddWhenStack,
		manager:           mgr,
		listeners:         config.Listeners,
		statusType:        config.StatusType,
		flags:             config.BehaviorFlags,
	}

	if mi.params == nil {
		mi.params = make(map[string]float64)
	}

	// Apply defaults from config as fallback
	if mi.countAddWhenStack == 0 {
		mi.countAddWhenStack = config.CountAddWhenStack
	}
	if mi.maxCount == 0 {
		mi.maxCount = config.MaxCount
	}
	if mi.count == 0 {
		mi.count = config.Count
	}
	if mi.duration == 0 {
		mi.duration = config.Duration
	}

	// default "infinite" cases
	if mi.duration <= 0 {
		mi.duration = -1
	}
	if mi.count <= 0 {
		mi.count = -1
	}
	if mi.maxCount <= 0 {
		mi.maxCount = -1
	}

	// cases that can stack. Need to ensure count is stackable pending what was specified on instance
	if config.Stacking == Replace || config.Stacking == ReplaceBySource || config.Stacking == Merge {
		if mi.count <= 0 && mi.countAddWhenStack > 0 {
			mi.count = mi.countAddWhenStack
		}
	}

	return mi
}

func (mi *ModifierInstance) reset(turnCount int) {
	mi.stats = info.NewPropMap()
	mi.debuffRES = info.NewDebuffRESMap()
	mi.renewTurn = turnCount
}

// Add a property to this modifier instance. Will cause all modifiers attached to the owner of this
// modifier to execute OnPropertyChange listener
func (mi *ModifierInstance) AddProperty(prop model.Property, amt float64) {
	mi.stats.Modify(prop, amt)
	mi.manager.emitPropertyChange(mi.owner, prop)
}

// Add a new debuffRES for the given behavior flag
func (mi *ModifierInstance) AddDebuffRES(flag model.BehaviorFlag, amt float64) {
	mi.debuffRES.Modify(flag, amt)
}

// Get the current value of a property set by this modifier instance
func (mi *ModifierInstance) GetProperty(prop model.Property) float64 {
	return mi.stats[prop]
}

// Get the current value of a debuff res set by this modifier instance
func (mi *ModifierInstance) GetDebuffRES(flags ...model.BehaviorFlag) float64 {
	return mi.debuffRES.GetDebuffRES(flags...)
}

// Name of the modifier this instance represents (what modifier config it is associated with)
func (mi *ModifierInstance) Name() key.Modifier {
	return mi.name
}

// TargetID for who created this modifier instance
func (mi *ModifierInstance) Source() key.TargetID {
	return mi.source
}

// TargetID for who this modifier is currently attached to
func (mi *ModifierInstance) Owner() key.TargetID {
	return mi.owner
}

// Any custom params to be defined that are used by the underlying modifier logic
func (mi *ModifierInstance) Params() map[string]float64 {
	return mi.params
}

// How long before this instance expires. Will be removed when Duration == 0. A negative duration
// means  that this modifier will never expire
func (mi *ModifierInstance) Duration() int {
	return mi.duration
}

// How many stacks are associated with this modifier instance. If count reaches 0, the modifier
// will be removed from the target. A count of -1 means that a count was never specified and/or
// a StackingBehavior that does not support stacking was used.
func (mi *ModifierInstance) Count() int {
	return mi.count
}

// What status type this modifier instance is (copied from the modifier config)
func (mi *ModifierInstance) StatusType() model.StatusType {
	return mi.statusType
}

// What BehaviorFlags exist for this modifier instance (copied from the modifier config)
func (mi *ModifierInstance) BehaviorFlags() []model.BehaviorFlag {
	return mi.flags
}

// Returns a copy of the config associated with this instance (config created via modifier.Register)
func (mi *ModifierInstance) Config() Config {
	return modifierCatalog[mi.name]
}

// Engine access for various operations
func (mi *ModifierInstance) Engine() engine.Engine {
	return mi.manager.engine
}

// Convert modifier instance to model version for read-only access to the instance
func (mi *ModifierInstance) ToModel() info.Modifier {
	props := info.NewPropMap()
	res := info.NewDebuffRESMap()

	props.AddAll(mi.stats)
	res.AddAll(mi.debuffRES)

	return info.Modifier{
		Name:   mi.name,
		Source: mi.source,
		Params: mi.params,
		// Chance: mi.chance,
		Duration:          mi.duration,
		TickImmediately:   mi.tickImmediately,
		Count:             mi.count,
		MaxCount:          mi.maxCount,
		CountAddWhenStack: mi.countAddWhenStack,
		Stats:             props,
		DebuffRES:         res,
	}
}
