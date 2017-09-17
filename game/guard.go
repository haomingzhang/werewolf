package game

import (
	"sync"
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
}

func (v *Guard) SaveLife() {
	v.dead = false
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

func (v *Guard) Action() {
	return
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
	return false, nil
}

func (v *Guard) Act(action int, targetId int) (bool, string) {
	return false, "You're not able to use this skill!"
}
