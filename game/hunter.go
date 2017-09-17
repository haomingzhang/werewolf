package game

import (
	"sync"
	"sync/atomic"
)

type Hunter struct {
	id         int
	dead       bool
	playerName string
	roleName   string
	mutex      *sync.Mutex
	registered bool
	controller *Controller
	isPoisoned bool
}

func CreateHunter(id int, c *Controller) *Hunter {
	return &Hunter{
		id:         id,
		roleName:   "Hunter",
		controller: c,
		mutex:      &sync.Mutex{},
	}
}

func (v *Hunter) Die(isPoisoned bool) {
	v.dead = true
	v.isPoisoned = isPoisoned
}

func (v *Hunter) SaveLife() {
	v.dead = false
}

func (v *Hunter) IsDead() bool {
	return v.dead
}

func (v *Hunter) GetRoleName() string {
	return v.roleName
}

func (v *Hunter) GetPlayerName() string {
	return v.playerName
}

func (v *Hunter) Action() {
	return
}

func (v *Hunter) Register(name string) bool {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	if v.registered {
		return false
	}
	v.playerName = name
	v.registered = true
	return true
}

func (v *Hunter) IsRegistered() bool {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	return v.registered
}

func (v *Hunter) GetActionCode() (bool, []int) {
	if atomic.LoadInt32(v.controller.phase) != TurnDay {
		return false, nil
	}
	return true, []int{SkillFire}
}

func (v *Hunter) Act(action int, targetId int) (bool, string) {
	if atomic.LoadInt32(v.controller.phase) != TurnDay {
		return false, "Not your turn!"
	}
	if action != SkillFire {
		return false, "You're not able to use this skill!"
	}
	if v.isPoisoned {
		return false, "You're poisoned!"
	}
	if targetId == v.id {
		return false, "You can't fire yourself!"
	}

	// fire somebody
	target := v.controller.Roles[targetId]
	if target.IsDead() {
		return false, "Target is already dead!"
	}
	v.controller.Roles[targetId].Die(false)
	return true, "Fire Succeeded!"
}
