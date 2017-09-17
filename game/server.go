package game

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type GameServer struct {
	controller *Controller
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type InitResponse struct {
	Message string `json:"message"`
}

func (g *GameServer) Start() {
	// TODO: An endpoint to end the day
	http.HandleFunc("/init", g.handleInit)
	http.HandleFunc("/start", g.handleStart)
	http.HandleFunc("/health", g.handleHealth)
	http.HandleFunc("/register", g.handleRegister)
	http.HandleFunc("/action", g.handleAction)
	http.HandleFunc("/lastnightinfo", g.handleLastNight)
	http.HandleFunc("/dayend", g.handleDayEnd)

	http.ListenAndServe(":8888", nil)
}

type InitGameRequest struct {
	VillagerCount int `json:"villagerCount"`
	WerewolfCount int `json:"werewolfCount"`
	ProphetCount  int `json:"prophetCount"`
	WizardCount   int `json:"wizardCount"`
	HunterCount   int `json:"hunterCount"`
	MoronCount    int `json:"moronCount"`
	GuardCount    int `json:"guardCount"`
}

type ActionRequest struct {
	Id         int    `json:"id"`
	Password   string `json:"password"`
	ActionCode int    `json:"actionCode"`
	Target     int    `json:"target"`
}

type ActionResponse struct {
	Successful  bool   `json:"successful"`
	ActionCodes []int  `json:"actionCodes"`
	Message     string `json:"message"`
}

type RegisterRequest struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type DayEndRequest struct {
	BanishId int `json:"banishId"`
}

type DayEndResponse struct {
	Successful bool   `json:"successful"`
	Message    string `json:"message"`
}

type RegisterResponse struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	RoleName string `json:"roleName"`
	Code     int    `json:"code"`
}

type LastNightResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type StartGameResponse struct {
	Message string `json:"message"`
}

func (g *GameServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Werewolf Server is healthy! Haoming is healthier!"))
}

func (g *GameServer) handleLastNight(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		g.writeClientError(w, http.StatusBadRequest, "Only GET is supported")
		return
	}
	if !g.controller.isInitialized() {
		g.writeClientError(w, http.StatusForbidden, "Game has not been initialized")
		return
	}
	res := g.controller.GetLastNightInfo()
	resBytes, err := json.Marshal(res)
	if err != nil {
		g.writeServerError(w, err.Error())
		return
	}
	w.Write(resBytes)
}

func (g *GameServer) handleDayEnd(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		g.writeClientError(w, http.StatusBadRequest, "Only POST is supported")
		return
	}
	defer r.Body.Close()
	if !g.controller.isInitialized() {
		g.writeClientError(w, http.StatusForbidden, "Game has not been initialized")
		return
	}
	bodyBytes, err := ioutil.ReadAll(r.Body)
	rr := &DayEndRequest{}
	err = json.Unmarshal(bodyBytes, rr)
	if err != nil {
		g.writeServerError(w, err.Error())
		return
	}

	// validate request
	valid, reason := rr.Validate(g.controller)
	if !valid {
		g.writeClientError(w, http.StatusBadRequest, reason)
		return
	}

	//  banish player
	res := g.controller.BanishPlayer(rr.BanishId)
	resBytes, err := json.Marshal(res)
	if err != nil {
		g.writeServerError(w, err.Error())
		return
	}
	w.Write(resBytes)

}

func (g *GameServer) handleStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		g.writeClientError(w, http.StatusBadRequest, "Only POST is supported")
		return
	}
	if !g.controller.isInitialized() {
		g.writeClientError(w, http.StatusForbidden, "Game has not been initialized")
		return
	}
	if success, msg := g.controller.StartGame(); !success {
		g.writeClientError(w, http.StatusForbidden, msg)
		return
	}
	res := StartGameResponse{
		Message: "Game started",
	}
	// response
	resBytes, err := json.Marshal(res)
	if err != nil {
		g.writeServerError(w, err.Error())
	}
	w.Write(resBytes)
}

func (g *GameServer) handleRegister(w http.ResponseWriter, r *http.Request) {
	// parse request
	if r.Method != "POST" {
		g.writeClientError(w, http.StatusBadRequest, "Only POST is supported")
		return
	}
	defer r.Body.Close()
	if !g.controller.isInitialized() {
		g.writeClientError(w, http.StatusForbidden, "Game has not been initialized")
		return
	}
	bodyBytes, err := ioutil.ReadAll(r.Body)
	rr := &RegisterRequest{}
	err = json.Unmarshal(bodyBytes, rr)
	if err != nil {
		g.writeServerError(w, err.Error())
		return
	}

	// validate request
	valid, reason := rr.Validate(g.controller.TotalCount)
	if !valid {
		g.writeClientError(w, http.StatusBadRequest, reason)
		return
	}

	// send response
	res := g.controller.Register(rr)
	w.WriteHeader(res.Code)
	resBytes, err := json.Marshal(res)
	if err != nil {
		g.writeServerError(w, err.Error())
		return
	}
	w.Write(resBytes)
	if res.Code == http.StatusOK {
		log.Printf("Player %d (%s) registered!", res.Id, res.Name)
	}
}

func (g *GameServer) handleAction(w http.ResponseWriter, r *http.Request) {
	// parse request
	if r.Method != "POST" {
		g.writeClientError(w, http.StatusBadRequest, "Only POST is supported")
		return
	}
	defer r.Body.Close()
	bodyBytes, err := ioutil.ReadAll(r.Body)
	req := &ActionRequest{}
	err = json.Unmarshal(bodyBytes, req)
	if err != nil {
		g.writeServerError(w, err.Error())
		return
	}
	// validate request
	valid, reason := req.Validate(g.controller)
	if !valid {
		g.writeClientError(w, http.StatusUnauthorized, reason)
		return
	}

	// sendResponse
	res := g.controller.HandleAction(req.Id, req.ActionCode, req.Target)
	resBytes, err := json.Marshal(res)
	if err != nil {
		g.writeServerError(w, err.Error())
		return
	}
	w.Write(resBytes)
}

func (g *GameServer) handleInit(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		g.writeClientError(w, http.StatusBadRequest, "Only POST is supported")
		return
	}
	defer r.Body.Close()
	bodyBytes, err := ioutil.ReadAll(r.Body)
	sgr := &InitGameRequest{}
	err = json.Unmarshal(bodyBytes, sgr)
	if err != nil {
		g.writeServerError(w, err.Error())
		return
	}

	// validate request
	valid, reason := sgr.Validate()
	if !valid {
		g.writeClientError(w, http.StatusBadRequest, reason)
		return
	}

	// initialize context
	g.controller = CreateController()
	if !g.controller.Initialize(sgr) {
		g.writeClientError(w, http.StatusForbidden, "Game already initialized!")
		return
	}

	// send response
	res := InitResponse{
		Message: "Game successfully initialized!",
	}
	resBytes, err := json.Marshal(res)
	if err != nil {
		g.writeServerError(w, err.Error())
		return
	}
	w.Write(resBytes)
	log.Println("Game successfully initialized!")
}

func (g *GameServer) writeServerError(w http.ResponseWriter, message string) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(message))
	return
}

func (g *GameServer) writeClientError(w http.ResponseWriter, code int, message string) {
	res := ErrorResponse{
		Code:    code,
		Message: message,
	}
	resBytes, err := json.Marshal(res)
	if err != nil {
		g.writeServerError(w, "writeClientError:"+err.Error())
		return
	}
	w.WriteHeader(code)
	w.Write(resBytes)
}

func (s *InitGameRequest) Validate() (bool, string) {
	valid := true
	reason := []string{}
	if s.VillagerCount <= 0 {
		valid = false
		reason = append(reason, "VillagerCount")
	}
	if s.WerewolfCount <= 0 {
		valid = false
		reason = append(reason, "WerewolfCount")
	}
	if s.ProphetCount < 0 || s.ProphetCount > 1 {
		valid = false
		reason = append(reason, "ProphetCount")
	}
	if s.WizardCount < 0 || s.WizardCount > 1 {
		valid = false
		reason = append(reason, "WizardCount")
	}
	if s.HunterCount < 0 || s.HunterCount > 1 {
		valid = false
		reason = append(reason, "HunterCount")
	}
	if s.MoronCount < 0 || s.MoronCount > 1 {
		valid = false
		reason = append(reason, "MoronCount")
	}
	if s.GuardCount < 0 || s.GuardCount > 1 {
		valid = false
		reason = append(reason, "GuardCount")
	}
	return valid, strings.Join(reason, " && ")
}

func (r *RegisterRequest) Validate(totalNum int) (bool, string) {
	if r.Id < 0 || r.Id >= totalNum {
		return false, "Invalid id"
	}
	return true, ""
}

func (r *ActionRequest) Validate(c *Controller) (bool, string) {
	if r.Id < 0 || r.Id >= c.TotalCount {
		return false, "Invalid id"
	}
	if c.Passwords[r.Id] != r.Password {
		return false, "Wrong Password"
	}
	return true, ""
}

func (r *DayEndRequest) Validate(c *Controller) (bool, string) {
	if r.BanishId < 0 || r.BanishId >= c.TotalCount {
		return false, "Invalid id"
	}
	return true, ""
}
