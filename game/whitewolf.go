package game

import (
	"sync"
	"sync/atomic"
)

type WhiteWolf struct {
	id         int
	dead       bool
	playerName string
	roleName   string
	mutex      *sync.Mutex
	registered bool
	controller *Controller
	isPoisoned bool
}

func CreateWhiteWolf(id int, c *Controller) *WhiteWolf {
	return &WhiteWolf{
		id:         id,
		roleName:   "WhiteWolf",
		controller: c,
		mutex:      &sync.Mutex{},
	}
}

func (v *WhiteWolf) Die(isPoisoned bool) {
	v.dead = true
	v.isPoisoned = isPoisoned
	v.controller.WhiteWolfCount--
}

func (v *WhiteWolf) IsDead() bool {
	return v.dead
}

func (v *WhiteWolf) GetRoleName() string {
	return v.roleName
}

func (v *WhiteWolf) GetPlayerName() string {
	return v.playerName
}

func (v *WhiteWolf) Register(name string) bool {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	if v.registered {
		return false
	}
	v.playerName = name
	v.registered = true
	return true
}

func (v *WhiteWolf) IsRegistered() bool {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	return v.registered
}

func (v *WhiteWolf) GetActionCode() (bool, []int) {
	if atomic.LoadInt32(v.controller.phase) != TurnWerewolf {
		return false, nil
	}
	return true, []int{SkillKill}
}

func (v *WhiteWolf) Act(action int, targetId int) (bool, string) {
	if atomic.LoadInt32(v.controller.phase) != TurnWerewolf {
		return false, "Not your turn!"
	}
	if action != SkillKill {
		return false, "You're not able to use this skill!"
	}

	// kill somebody
	target := v.controller.Roles[targetId]
	if target.IsDead() {
		return false, "Target is already dead!"
	}
	v.controller.waitChan[TurnWerewolf] <- targetId
	return true, "Kill Succeeded!"
}
