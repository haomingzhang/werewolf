package game

import (
	"sync"
	"sync/atomic"
)

type Wizard struct {
	id         int
	dead       bool
	playerName string
	roleName   string
	mutex      *sync.Mutex
	registered bool
	controller *Controller
	isPoisoned bool
	saveUsed   bool
	poisonUsed bool
}

func CreateWizard(id int, c *Controller) *Wizard {
	return &Wizard{
		id:         id,
		roleName:   "Wizard",
		controller: c,
		mutex:      &sync.Mutex{},
	}
}

func (v *Wizard) Die(isPoisoned bool) {
	v.dead = true
	v.isPoisoned = isPoisoned
}

func (v *Wizard) SaveLife() {
	v.dead = false
}

func (v *Wizard) IsDead() bool {
	return v.dead
}

func (v *Wizard) GetRoleName() string {
	return v.roleName
}

func (v *Wizard) GetPlayerName() string {
	return v.playerName
}

func (v *Wizard) Action() {
	return
}

func (v *Wizard) Register(name string) bool {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	if v.registered {
		return false
	}
	v.playerName = name
	v.registered = true
	return true
}

func (v *Wizard) IsRegistered() bool {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	return v.registered
}

func (v *Wizard) GetActionCode() (bool, []int) {
	if atomic.LoadInt32(v.controller.phase) != TurnWizard {
		return false, nil
	}
	ret := []int{}
	if !v.saveUsed {
		ret = append(ret, SkillSave)
	}
	if !v.poisonUsed {
		ret = append(ret, SkillPoison)
	}
	if len(ret) == 0 {
		return false, nil
	}
	ret = append(ret)
	return true, ret
}

func (v *Wizard) Act(action int, targetId int) (bool, string) {
	if atomic.LoadInt32(v.controller.phase) != TurnWizard {
		return false, "Not your turn!"
	}

	switch action {
	case SkillSave:
		if v.saveUsed {
			return false, "Your potion is already Used!"
		}
		v.saveUsed = true
		v.controller.waitChan[TurnWizard] <- -1
	case SkillPoison:
		target := v.controller.Roles[targetId]
		if target.IsDead() {
			return false, "Target is already dead!"
		}
		v.controller.waitChan[TurnWizard] <- targetId
	case SkillDontUse:
		v.controller.waitChan[TurnWizard] <- -2
		return true, "Didn't use any skill!"
	default:
		return false, "You're not able to use this skill!"
	}
	return true, "Successfully use skill!"
}
