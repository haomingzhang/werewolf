package game

import (
	"sync"
	"sync/atomic"
)

type Werewolf struct {
	id         int
	dead       bool
	playerName string
	roleName   string
	mutex      *sync.Mutex
	registered bool
	controller *Controller
	isPoisoned bool
}

func CreateWerewolf(id int, c *Controller) *Werewolf {
	return &Werewolf{
		id:         id,
		roleName:   "Werewolf",
		controller: c,
		mutex:      &sync.Mutex{},
	}
}

func (v *Werewolf) Die(isPoisoned bool) {
	v.dead = true
	v.isPoisoned = isPoisoned
}

func (v *Werewolf) SaveLife() {
	v.dead = false
}

func (v *Werewolf) IsDead() bool {
	return v.dead
}

func (v *Werewolf) GetRoleName() string {
	return v.roleName
}

func (v *Werewolf) GetPlayerName() string {
	return v.playerName
}

func (v *Werewolf) Action() {
	return
}

func (v *Werewolf) Register(name string) bool {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	if v.registered {
		return false
	}
	v.playerName = name
	v.registered = true
	return true
}

func (v *Werewolf) IsRegistered() bool {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	return v.registered
}

func (v *Werewolf) GetActionCode() (bool, []int) {
	if atomic.LoadInt32(v.controller.phase) != TurnWerewolf {
		return false, nil
	}
	return true, []int{SkillKill}
}

func (v *Werewolf) Act(action int, targetId int) (bool, string) {
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
