package cabriui

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabridss"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/joule"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type SCabriSyncSpec struct {
	LeftUsers    []string `yaml:"leftUsers"`
	LeftACL      []string `yaml:"leftACL"`
	RightUsers   []string `yaml:"rightUsers"`
	RightACL     []string `yaml:"rightACL"`
	Recursive    bool     `yaml:"recursive"`
	DryRun       bool     `yaml:"dryRun"`
	BiDir        bool     `yaml:"biDir"`
	KeepContent  bool     `yaml:"keepContent"`
	NoCh         bool     `yaml:"noCh"`
	NoACL        bool     `yaml:"noACL"`
	MapACL       []string `yaml:"mapACL"`
	Summary      bool     `yaml:"summary"`
	Verbose      bool     `yaml:"verbose"`
	VerboseLevel int      `yaml:"verboseLevel"`
	LeftTime     string   `yaml:"leftTime"`
	RightTime    string   `yaml:"rightTime"`
	LeftDss      string   `yaml:"leftDss"`
	RightDss     string   `yaml:"rightDss"`
}

type SGitSpec struct {
	RepoUrl         string `yaml:"repoUrl"`
	ClonePath       string `yaml:"clonePath"`
	CloneOptions    string `yaml:"cloneOptions"`
	Branch          string `yaml:"branch"`
	CheckOutOptions string `yaml:"checkOutOptions"`
	PullOptions     string `yaml:"pullOptions"`
}

type SScheduledAction struct {
	Type          string         `yaml:"type"` // currently "cabriSync", "git" or "cmd"
	CabriSyncSpec SCabriSyncSpec `yaml:"cabriSyncSpec"`
	GitSpec       SGitSpec       `yaml:"gitSpec"`
	CmdLine       string         `yaml:"cmdLine"`
}

func (ssa SScheduledAction) String() string {
	if ssa.Type == "cmd" {
		return fmt.Sprintf("command \"%s\"", ssa.CmdLine)
	}
	return fmt.Sprintf("%+v", ssa.Type)
}

type CabriScheduleSpec map[string]struct {
	Period          int                `yaml:"period"` // periodicity of action, if <= 0 no periodic action
	ContinueOnError bool               `yaml:"continueOnError"`
	ExitOnError     bool               `yaml:"exitOnError"`
	Actions         []SScheduledAction `yaml:"actions"`
}

type ScheduleRunStatus struct {
	label     string
	IsRunning bool `json:"isRunning"`
	uow       joule.UnitOfWork
	Count     int    `json:"count"`
	LastTime  int64  `json:"lastTime"`
	LastOut   string `json:"lastOut"`
	LastErr   string `json:"lastErr"`
	LastRunOk bool   `json:"lastRunOk"`
}

type ScheduleConfig struct {
	ctx       context.Context
	cancel    context.CancelFunc
	mux       sync.Mutex
	Spec      CabriScheduleSpec
	isExiting bool
	run       map[string]*ScheduleRunStatus
}

func logSchedule(ctx context.Context, line string) {
	eol := "\n"
	if len(line) > 0 && line[len(line)-1] == '\n' {
		eol = ""
	}
	scheduleErr(ctx, fmt.Sprintf("%s Schedule: %s%s", cabridss.UnixUTC(time.Now().UnixNano()).String(), line, eol))
}

func (srs *ScheduleRunStatus) doRunCommand(sc *ScheduleConfig, action SScheduledAction, cmdLine string, wd string) (stdout, stderr []byte, err error) {
	if action.Type != "cmd" {
		logSchedule(sc.ctx, fmt.Sprintf("%s: running \"%s\" for %s", srs.label, cmdLine, action))
	}
	elems := strings.Split(os.ExpandEnv(cmdLine), " ")
	cmd := exec.CommandContext(sc.ctx, elems[0], elems[1:]...)
	if wd != "" {
		cmd.Dir = wd
	}
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err = cmd.Run()
	stdout = outb.Bytes()
	stderr = errb.Bytes()
	if action.Type != "cmd" {
		if len(stdout) != 0 {
			logSchedule(sc.ctx, fmt.Sprintf("stdout: %s", string(stdout)))
		}
		if len(stderr) != 0 {
			logSchedule(sc.ctx, fmt.Sprintf("stderr: %s", string(stderr)))
		}
	}
	return
}

func (srs *ScheduleRunStatus) doRunGit(sc *ScheduleConfig, action SScheduledAction, gs SGitSpec) (lastCommand string, stdout, stderr []byte, err error) {
	if gs.ClonePath == "" {
		gs.ClonePath = "/data"
	}
	_, err = os.Stat(gs.ClonePath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return
	}
	if err != nil {
		if gs.CloneOptions != "" {
			gs.CloneOptions = " " + gs.CloneOptions
		}
		lastCommand = fmt.Sprintf("git clone%s %s %s", gs.CloneOptions, gs.RepoUrl, gs.ClonePath)
		stdout, stderr, err = srs.doRunCommand(sc, action, lastCommand, "")
		if err != nil {
			return
		}
	}
	if gs.Branch == "" {
		gs.Branch = "main"
	}
	if gs.CheckOutOptions != "" {
		gs.CheckOutOptions = " " + gs.CheckOutOptions
	}
	lastCommand = fmt.Sprintf("git checkout%s %s", gs.CheckOutOptions, gs.Branch)
	stdout, stderr, err = srs.doRunCommand(sc, action, lastCommand, gs.ClonePath)
	if err != nil {
		return
	}
	if gs.PullOptions != "" {
		gs.PullOptions = " " + gs.PullOptions
	} else {
		gs.PullOptions = " --ff-only"
	}
	lastCommand = fmt.Sprintf("git pull%s", gs.PullOptions)
	stdout, stderr, err = srs.doRunCommand(sc, action, lastCommand, gs.ClonePath)
	return
}

func (srs *ScheduleRunStatus) doRun(sc *ScheduleConfig, action SScheduledAction) (lastCommand string, stdout, stderr []byte, err error) {
	if action.Type == "git" {
		lastCommand, stdout, stderr, err = srs.doRunGit(sc, action, action.GitSpec)
	} else if action.Type == "cmd" {
		stdout, stderr, err = srs.doRunCommand(sc, action, action.CmdLine, "")
	} else {
		return "", nil, nil, fmt.Errorf("action type %s is not (yet) implemented", action.Type)
	}
	sc.mux.Lock()
	srs.IsRunning = false
	sc.mux.Unlock()
	return
}

func (srs *ScheduleRunStatus) run(sc *ScheduleConfig) (isRunning bool, err error) {
	sc.mux.Lock()
	if srs.IsRunning {
		sc.mux.Unlock()
		return true, nil
	}
	srs.IsRunning = true
	sc.mux.Unlock()
	go func() {
		srs.Count++
		srs.LastTime = time.Now().UnixNano()
		for _, action := range sc.Spec[srs.label].Actions {
			logSchedule(sc.ctx, fmt.Sprintf("%s %s running", srs.label, action))
			lc, so, se, err := srs.doRun(sc, action)
			if err != nil && lc != "" {
				logSchedule(sc.ctx, fmt.Sprintf("error on command \"%s\" in action %s\n", lc, srs.label))
			}
			if action.Type == "cmd" && len(so) != 0 {
				logSchedule(sc.ctx, fmt.Sprintf("stdout: %s", string(so)))
			}
			if action.Type == "cmd" && len(se) != 0 {
				logSchedule(sc.ctx, fmt.Sprintf("stderr: %s", string(se)))
			}
			if err != nil {
				if sc.Spec[srs.label].ExitOnError {
					scheduleErr(sc.ctx, fmt.Sprintf("ScheduleRunStatus.run: exiting on: %v", err))
					sc.mux.Lock()
					sc.isExiting = true
					sc.mux.Unlock()
					sc.cancel()
				}
				if !sc.Spec[srs.label].ContinueOnError {
					logSchedule(sc.ctx, fmt.Sprintf("command error %v, stopping action %s\n", err, srs.label))
					break
				}
				logSchedule(sc.ctx, fmt.Sprintf("command error %v, continue action %s\n", err, srs.label))
			}
		}
	}()
	return false, nil
}

func NewServerErr(where string, err error) error {
	return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("in %s: %v", where, err))
}

func sSchedGet(c echo.Context) error {
	label := ""
	if err := echo.PathParamsBinder(c).String("label", &label).BindError(); err != nil {
		return NewServerErr("sRestGet", err)
	}
	sc := cabridss.GetCustomConfig(c).(*ScheduleConfig)
	if label == "" {
		return c.JSON(http.StatusOK, &sc)
	}
	srs, ok := sc.run[label]
	if !ok {
		return NewServerErr("sRestGet", fmt.Errorf("sSchedGet, no such scheduled entry: %s", label))
	}
	return c.JSON(http.StatusOK, &srs)
}

func sSchedPut(c echo.Context) error {
	label := ""
	if err := echo.PathParamsBinder(c).String("label", &label).BindError(); err != nil {
		return NewServerErr("sSchedPut", err)
	}
	sc := cabridss.GetCustomConfig(c).(*ScheduleConfig)
	srs, ok := sc.run[label]
	if !ok {
		return NewServerErr("sRestGet", fmt.Errorf("sSchedGet, no such scheduled entry: %s", label))
	}
	srs.run(sc)
	return c.JSON(http.StatusOK, &srs)
}

func SchedServerConfigurator(e *echo.Echo, root string, configs map[string]interface{}) error {
	e.GET(root, sSchedGet)
	e.PUT(root+":label", sSchedPut)
	e.GET(root+":label", sSchedGet)
	return nil
}

func Schedule(sc *ScheduleConfig) (time.Duration, error) {
	gNext := time.Duration(10) * time.Second
	now := time.Now().UnixNano()
	for label, entry := range sc.Spec {
		if entry.Period <= 0 {
			continue
		}
		srs := sc.run[label]
		if srs.LastTime+int64(entry.Period)*1e9 <= now {
			isRunning, _ := srs.run(sc)
			if !isRunning {
				continue
			}
		}
		nextNs := srs.LastTime + int64(entry.Period)*1e9 - now
		if nextNs < int64(gNext) {
			gNext = time.Duration(nextNs)
		}
	}
	if int64(gNext) < 1e9 {
		gNext = time.Second
	}
	//println(fmt.Sprintf("next %d", gNext/1e9))
	//for _, srs := range sc.run {
	//	println(fmt.Sprintf("\t%+v", *srs))
	//}
	return gNext, nil
}
