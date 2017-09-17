package game

import (
	"sync"
)

type Villager struct {
	id         int
	dead       bool
	playerName string
	roleName   string
	mutex      *sync.Mutex
	registered bool
	controller *Controller
	isPoisoned bool
}

func CreateVillager(id int, c *Controller) *Villager {
	return &Villager{
		id:         id,
		roleName:   "Villager",
		controller: c,
		mutex:      &sync.Mutex{},
	}
}

func (v *Villager) Die(isPoisoned bool) {
	v.dead = true
	v.isPoisoned = isPoisoned
}

func (v *Villager) SaveLife() {
	v.dead = false
}

func (v *Villager) IsDead() bool {
	return v.dead
}

func (v *Villager) GetRoleName() string {
	return v.roleName
}

func (v *Villager) GetPlayerName() string {
	return v.playerName
}

func (v *Villager) Action() {
	return
}

func (v *Villager) Register(name string) bool {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	if v.registered {
		return false
	}
	v.playerName = name
	v.registered = true
	return true
}

func (v *Villager) IsRegistered() bool {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	return v.registered
}

func (v *Villager) GetActionCode() (bool, []int) {
	return false, nil
}

func (v *Villager) Act(action int, targetId int) (bool, string) {
	return false, "You're not able to use this skill!"
}
