package game

import (
	"sync"
	"sync/atomic"
)

type Guard struct {
	id         int
	dead       bool
	playerName string
	roleName   string
	mutex      *sync.Mutex
	registered bool
	controller *Controller
	isPoisoned bool
}

func CreateGuard(id int, c *Controller) *Guard {
	return &Guard{
		id:         id,
		roleName:   "Guard",
		controller: c,
		mutex:      &sync.Mutex{},
	}
}

func (v *Guard) Die(isPoisoned bool) {
	v.dead = true
	v.isPoisoned = isPoisoned
	v.controller.GuardCount--
}

func (v *Guard) IsDead() bool {
	return v.dead
}

func (v *Guard) GetRoleName() string {
	return v.roleName
}

func (v *Guard) GetPlayerName() string {
	return v.playerName
}

func (v *Guard) Register(name string) bool {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	if v.registered {
		return false
	}
	v.playerName = name
	v.registered = true
	return true
}

func (v *Guard) IsRegistered() bool {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	return v.registered
}

func (v *Guard) GetActionCode() (bool, []int) {
	if atomic.LoadInt32(v.controller.phase) != TurnGuard {
		return false, nil
	}
	return true, []int{SkillProtect, SkillDontUse}
}

func (v *Guard) Act(action int, targetId int) (bool, string) {
	if atomic.LoadInt32(v.controller.phase) != TurnGuard {
		return false, "Not your turn!"
	}
	switch action {
	case SkillProtect:
		// guard somebody
		target := v.controller.Roles[targetId]
		if target.IsDead() {
			return false, "Target is already dead!"
		}
		v.controller.waitChan[TurnGuard] <- targetId
		return true, "Guard Succeeded!"
	case SkillDontUse:
		v.controller.waitChan[TurnGuard] <- -1
		return true, "Didn't use any skill!"
	default:
		return false, "You're not able to use this skill!"
	}
}
