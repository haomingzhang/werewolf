package game

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	GetAction = iota
	SkillKill
	SkillSave
	SkillPoison
	SkillVerifyRole
	SkillFire
	SkillProtect
	SkillDontUse
)

var skillName = map[int]string{
	SkillKill:       "Kill",
	SkillSave:       "Save",
	SkillPoison:     "Poison",
	SkillVerifyRole: "VerifyRole",
	SkillProtect:    "Guard",
	SkillFire:       "Fire",
	SkillDontUse:    "Don't_use_skill",
}

const (
	TurnWerewolf = iota
	TurnWizard
	TurnProphet
	TurnDay
	TurnGuard
	TurnNotStarted
	TurnStarted
	TurnGameOver
	TurnNight
	turnNightEnd
)

const (
	NumTurn       = 5
	SleepInterval = 2 * time.Second
)

type Controller struct {
	IsEnd         bool
	VillagerCount int
	GodCount      int
	WerewolfCount int
	ProphetCount  int
	WizardCount   int
	HunterCount   int
	MoronCount    int
	GuardCount    int
	TotalCount    int
	initialized   bool
	started       bool
	Roles         []Role // id -> RoleName
	Passwords     []string
	mutex         *sync.Mutex
	phase         *int32
	waitChan      []chan int
	lastNight     []string
	killedTonight int
}

type Role interface {
	Die(bool)
	GetRoleName() string
	GetPlayerName() string
	Register(name string) bool
	IsRegistered() bool
	GetActionCode() (canAct bool, actionCodes []int)
	Act(action int, target int) (ok bool, message string)
	IsDead() bool
}

func CreateController() *Controller {
	c := &Controller{
		mutex:    &sync.Mutex{},
		phase:    new(int32),
		waitChan: make([]chan int, NumTurn),
	}
	*c.phase = TurnNotStarted
	for i := range c.waitChan {
		c.waitChan[i] = make(chan int)
	}

	return c
}

func (c *Controller) Initialize(sgr *InitGameRequest) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.initialized {
		return false
	}
	// assign vars
	c.VillagerCount = sgr.VillagerCount
	c.GodCount = sgr.GuardCount + sgr.MoronCount + sgr.HunterCount + sgr.ProphetCount + sgr.WizardCount
	c.ProphetCount = sgr.ProphetCount
	c.WizardCount = sgr.WizardCount
	c.HunterCount = sgr.HunterCount
	c.MoronCount = sgr.MoronCount
	c.GuardCount = sgr.GuardCount
	c.WerewolfCount = sgr.WerewolfCount
	c.TotalCount = c.VillagerCount + c.GodCount + c.WerewolfCount

	// assign roles
	c.Roles = make([]Role, c.TotalCount)
	c.Passwords = make([]string, c.TotalCount)
	rand.Seed(time.Now().Unix())
	randIds := rand.Perm(c.TotalCount)
	for i := 0; i < c.TotalCount; i++ {
		switch {
		case i < c.VillagerCount:
			c.Roles[randIds[i]] = CreateVillager(i, c)
		case i < c.VillagerCount+c.WerewolfCount:
			c.Roles[randIds[i]] = CreateWerewolf(i, c)
		case i < c.VillagerCount+c.WerewolfCount+c.ProphetCount:
			c.Roles[randIds[i]] = CreateProphet(i, c)
		case i < c.VillagerCount+c.WerewolfCount+c.ProphetCount+c.WizardCount:
			c.Roles[randIds[i]] = CreateWizard(i, c)
		case i < c.VillagerCount+c.WerewolfCount+c.ProphetCount+c.WizardCount+c.HunterCount:
			c.Roles[randIds[i]] = CreateHunter(i, c)
		case i < c.VillagerCount+c.WerewolfCount+c.ProphetCount+c.WizardCount+c.HunterCount+c.MoronCount:
			c.Roles[randIds[i]] = CreateMoron(i, c)
		case i < c.VillagerCount+c.WerewolfCount+c.ProphetCount+c.WizardCount+c.HunterCount+c.MoronCount+c.GuardCount:
			c.Roles[randIds[i]] = CreateGuard(i, c)
		}
	}

	c.initialized = true
	return true
}

func (c *Controller) isInitialized() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.initialized
}

func (c *Controller) Register(request *RegisterRequest) *RegisterResponse {
	role := c.Roles[request.Id]
	res := &RegisterResponse{
		Name:     request.Name,
		Id:       request.Id,
		RoleName: role.GetRoleName(),
	}
	if role.Register(request.Name) {
		res.Code = http.StatusOK
		c.Passwords[request.Id] = request.Password
	} else {
		res.Code = http.StatusAlreadyReported
	}
	return res
}

func (c *Controller) StartGame() (bool, string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.started {
		return false, "Game already started!"
	}

	// verify all players have registered
	for _, r := range c.Roles {
		if !r.IsRegistered() {
			return false, "Somebody has not registered!"
		}
	}

	// print message
	log.Println("Game Started:")
	for i, r := range c.Roles {
		log.Printf("Player	%d	Name:	%s\n", i+1, r.GetPlayerName())
	}

	// start game
	go c.beginNight(1)
	c.started = true
	atomic.StoreInt32(c.phase, TurnStarted)
	return true, ""
}

func (c *Controller) HandleAction(id int, action int, target int) *ActionResponse {
	res := &ActionResponse{}
	switch action {
	case GetAction:
		res.Successful, res.ActionCodes = c.Roles[id].GetActionCode()
		if !res.Successful {
			res.Message = "You can't use skill now!"
			return res
		}
		// dead info
		if isInSlice(SkillSave, res.ActionCodes) {
			res.Message = fmt.Sprintf("Player id=%d is killed tonight.", c.killedTonight+1)
		}
	default:
		res.Successful, res.Message = c.Roles[id].Act(action, target)
	}
	for _, code := range res.ActionCodes {
		res.ActionName = append(res.ActionName, skillName[code])
	}
	return res
}

func isInSlice(t int, s []int) bool {
	for _, n := range s {
		if t == n {
			return true
		}
	}
	return false
}

func (c *Controller) GameIsEnd() bool {
	if c.IsEnd {
		return true
	}

	var leftWerewolf, leftVillager, leftGod int

	for _, role := range c.Roles {
		if w, ok := role.(*Werewolf); ok && !w.IsDead() {
			leftWerewolf++
		} else if v, ok := role.(*Villager); ok && !v.IsDead() {
			leftVillager++
		} else {
			leftGod++
		}
	}

	if leftVillager == 0 || leftVillager == 0 || leftGod == 0 {
		c.IsEnd = true
		return true
	}

	return false
}

func (c *Controller) BanishPlayer(id int) *DayEndResponse {
	if atomic.LoadInt32(c.phase) != TurnDay {
		return &DayEndResponse{
			Successful: false,
			Message:    "You can only banish player during the day!",
		}
	}

	if c.Roles[id].IsDead() {
		return &DayEndResponse{
			Successful: false,
			Message:    fmt.Sprintf("Error: Player %d is already dead!", id),
		}
	}

	c.waitChan[TurnDay] <- id
	return &DayEndResponse{
		Successful: true,
		Message:    fmt.Sprintf("Successfully banished player %d", id),
	}
}

func (c *Controller) GetLastNightInfo() *LastNightResponse {
	if atomic.LoadInt32(c.phase) != TurnDay {
		return &LastNightResponse{
			Code:    http.StatusForbidden,
			Message: "You can only get last night info during the day!",
		}
	}
	return &LastNightResponse{
		Code:    http.StatusOK,
		Message: "Players who died last night: " + strings.Join(c.lastNight, ","),
	}
}

func (c *Controller) beginDay(day int) {
	//TODO: sync here instead of sleeping
	c.sleepAndPlayAudio(TurnDay)
	// Check game over
	if c.GameIsEnd() {
		atomic.StoreInt32(c.phase, TurnGameOver)
		log.Print("Game Over!")
		c.sleepAndPlayAudio(TurnGameOver)
		return
	}

	// day
	atomic.StoreInt32(c.phase, TurnDay)
	deadId := <-c.waitChan[TurnDay]

	// end the day
	c.Roles[deadId].Die(false)
	go c.beginNight(day + 1)
}

func (c *Controller) beginNight(day int) {
	//TODO: sync here instead of sleeping
	c.sleepAndPlayAudio(TurnNight)
	// Check game over
	if c.GameIsEnd() {
		atomic.StoreInt32(c.phase, TurnGameOver)
		log.Print("Game Over!")
		c.sleepAndPlayAudio(TurnGameOver)
		return
	}

	// reset night info
	c.lastNight = make([]string, 0)

	// Werewolf
	atomic.StoreInt32(c.phase, TurnWerewolf)
	c.sleepAndPlayAudio(TurnWerewolf)
	killedId := <-c.waitChan[TurnWerewolf]
	saved := false
	c.killedTonight = killedId

	// Guard
	/*
		if c.GuardCount>0 {
			c.sleepAndPlayAudio(TurnGuard)
			guardId := <-c.waitChan[TurnGuard]
		}
	*/

	// Wizard
	atomic.StoreInt32(c.phase, TurnWizard)
	c.sleepAndPlayAudio(TurnWizard)
	targetId := -2
	if c.WizardCount > 0 {
		targetId = <-c.waitChan[TurnWizard]
		if targetId == -1 { // save
			saved = true
		} else if targetId == -2 {
			// do not use skill
		} else { // poison
			defer c.Roles[targetId].Die(true)
		}
	} else {
		time.Sleep(SleepInterval)
	}
	// Prophet
	atomic.StoreInt32(c.phase, TurnProphet)
	c.sleepAndPlayAudio(TurnProphet)
	if c.ProphetCount > 0 {
		<-c.waitChan[TurnProphet]
	} else {
		time.Sleep(SleepInterval)
	}
	// end the night
	if !saved {
		c.Roles[killedId].Die(false)
		c.lastNight = append(c.lastNight, strconv.Itoa(killedId+1))
	}
	if targetId >= 0 && targetId != killedId {
		c.lastNight = append(c.lastNight, strconv.Itoa(targetId+1))
	}
	c.sleepAndPlayAudio(turnNightEnd)
	go c.beginDay(day)
}

func (c *Controller) sleepAndPlayAudio(turn int) {
	time.Sleep(SleepInterval)
	var cmd *exec.Cmd
	var err error
	switch turn {
	case TurnNight:
		cmd = exec.Command("/bin/bash", "-c", "afplay ./audio/closeEyes.mpg")
		err = cmd.Run()
		if err != nil {
			log.Fatal(err.Error())
		}
	case TurnWerewolf:
		cmd = exec.Command("/bin/bash", "-c", "afplay ./audio/werewolf.mpg")
		err = cmd.Run()
		if err != nil {
			log.Fatal(err.Error())
		}
	case TurnWizard:
		cmd = exec.Command("/bin/bash", "-c", "afplay ./audio/werewolfEnd.mpg")
		err = cmd.Run()
		if err != nil {
			log.Fatal(err.Error())
		}
		time.Sleep(SleepInterval)
		cmd = exec.Command("/bin/bash", "-c", "afplay ./audio/wizard.mpg")
		err = cmd.Run()
		if err != nil {
			log.Fatal(err.Error())
		}
	case TurnProphet:
		cmd = exec.Command("/bin/bash", "-c", "afplay ./audio/wizardEnd.mpg")
		err = cmd.Run()
		if err != nil {
			log.Fatal(err.Error())
		}
		time.Sleep(SleepInterval)
		cmd = exec.Command("/bin/bash", "-c", "afplay ./audio/prophet.mpg")
		err = cmd.Run()
		if err != nil {
			log.Fatal(err.Error())
		}
	case turnNightEnd:
		cmd = exec.Command("/bin/bash", "-c", "afplay ./audio/prophetEnd.mpg")
		err = cmd.Run()
		if err != nil {
			log.Fatal(err.Error())
		}
	case TurnDay:
		cmd = exec.Command("/bin/bash", "-c", "afplay ./audio/day.mpg")
		err = cmd.Run()
		if err != nil {
			log.Fatal(err.Error())
		}
	case TurnGameOver:
		cmd = exec.Command("/bin/bash", "-c", "afplay ./audio/gameOver.mpg")
		err = cmd.Run()
		if err != nil {
			log.Fatal(err.Error())
		}
	case TurnGuard:
		cmd = exec.Command("/bin/bash", "-c", "afplay ./audio/guard.mpg")
		err = cmd.Run()
		if err != nil {
			log.Fatal(err.Error())
		}
	}

}
