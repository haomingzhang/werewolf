package game

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type Prophet struct {
	id         int
	dead       bool
	playerName string
	roleName   string
	mutex      *sync.Mutex
	registered bool
	controller *Controller
	isPoisoned bool
}

func CreateProphet(id int, c *Controller) *Prophet {
	return &Prophet{
		id:         id,
		roleName:   "Prophet",
		controller: c,
		mutex:      &sync.Mutex{},
	}
}

func (v *Prophet) Die(isPoisoned bool) {
	v.dead = true
	v.isPoisoned = isPoisoned
	v.controller.ProphetCount--
}

func (v *Prophet) IsDead() bool {
	return v.dead
}

func (v *Prophet) GetRoleName() string {
	return v.roleName
}

func (v *Prophet) GetPlayerName() string {
	return v.playerName
}

func (v *Prophet) Register(name string) bool {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	if v.registered {
		return false
	}
	v.playerName = name
	v.registered = true
	return true
}

func (v *Prophet) IsRegistered() bool {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	return v.registered
}

func (v *Prophet) GetActionCode() (bool, []int) {
	if atomic.LoadInt32(v.controller.phase) != TurnProphet {
		return false, nil
	}
	return true, []int{SkillVerifyRole}
}

func (v *Prophet) Act(action int, targetId int) (bool, string) {
	if atomic.LoadInt32(v.controller.phase) != TurnProphet {
		return false, "Not your turn!"
	}
	if action != SkillVerifyRole {
		return false, "You're not able to use this skill!"
	}

	// verify role of somebody
	role := v.controller.Roles[targetId]
	roleMsg := "Good"
	if _, ok := role.(*Werewolf); ok {
		roleMsg = "Werewolf"
	}
	if _, ok := role.(*WhiteWolf); ok {
		roleMsg = "Werewolf"
	}
	message := fmt.Sprintf("Player %d (%s) is: %s", targetId+1, role.GetPlayerName(), roleMsg)
	v.controller.waitChan[TurnProphet] <- targetId
	return true, message
}
