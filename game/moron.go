package game

import (
	"sync"
)

type Moron struct {
	id         int
	dead       bool
	playerName string
	roleName   string
	mutex      *sync.Mutex
	registered bool
	controller *Controller
	isPoisoned bool
}

func CreateMoron(id int, c *Controller) *Moron {
	return &Moron{
		id:         id,
		roleName:   "Moron",
		controller: c,
		mutex:      &sync.Mutex{},
	}
}

func (v *Moron) Die(isPoisoned bool) {
	v.dead = true
	v.isPoisoned = isPoisoned
}

func (v *Moron) SaveLife() {
	v.dead = false
}

func (v *Moron) IsDead() bool {
	return v.dead
}

func (v *Moron) GetRoleName() string {
	return v.roleName
}

func (v *Moron) GetPlayerName() string {
	return v.playerName
}

func (v *Moron) Action() {
	return
}

func (v *Moron) Register(name string) bool {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	if v.registered {
		return false
	}
	v.playerName = name
	v.registered = true
	return true
}

func (v *Moron) IsRegistered() bool {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	return v.registered
}

func (v *Moron) GetActionCode() (bool, []int) {
	return false, nil
}

func (v *Moron) Act(action int, targetId int) (bool, string) {
	return false, "You're not able to use this skill!"
}
