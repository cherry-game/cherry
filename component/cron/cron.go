package cherryCron

import (
	"fmt"
	"github.com/cherry-game/cherry"
	cherryMini "github.com/cherry-game/cherry/component/mini"
	cherryFacade "github.com/cherry-game/cherry/facade"
	cherryLogger "github.com/cherry-game/cherry/logger"
	"github.com/robfig/cron/v3"
	"time"
)

var _cron = cron.New(
	cron.WithSeconds(),
	cron.WithChain(cron.Recover(&CronLogger{})),
)

type CronLogger struct{}

func (CronLogger) Info(msg string, keysAndValues ...interface{}) {
	cherryLogger.Infow(msg, keysAndValues...)
}

func (CronLogger) Error(err error, _ string, _ ...interface{}) {
	cherryLogger.Error(err)
}

func Init(opts ...cron.Option) {
	if len(opts) < 1 {
		opts = append(opts, cron.WithSeconds())
		opts = append(opts, cron.WithChain(cron.Recover(&CronLogger{})))
	}
	_cron = cron.New(opts...)
}

func RegisterComponent(opts ...cron.Option) {
	Init(opts...)

	mini := cherryMini.New(
		"cron",
		cherryMini.WithInitFunc(func(_ cherryFacade.IApplication) {
			_cron.Start()
			cherryLogger.Info("cron component init.")
		}),
		cherryMini.WithStop(func(_ cherryFacade.IApplication) {
			Stop()
			cherryLogger.Infof("cron component is stopped.")
		}),
	)

	cherry.RegisterComponent(mini)
}

func AddFunc(spec string, cmd func()) (cron.EntryID, error) {
	return _cron.AddJob(spec, cron.FuncJob(cmd))
}

// AddEveryDayFunc 每天的x时x分x秒执行一次(每天1次)
func AddEveryDayFunc(cmd func(), hour, minutes, seconds int) (cron.EntryID, error) {
	spec := fmt.Sprintf("%d %d %d * * ?", seconds, minutes, hour)
	return _cron.AddFunc(spec, cmd)
}

// AddEveryHourFunc 每小时的x分x秒执行一次(每天24次)
func AddEveryHourFunc(cmd func(), minute, second int) (cron.EntryID, error) {
	spec := fmt.Sprintf("%d %d * * * ?", second, minute)
	return _cron.AddFunc(spec, cmd)
}

// AddDurationFunc 每间隔x秒执行一次
func AddDurationFunc(cmd func(), duration time.Duration) (cron.EntryID, error) {
	spec := fmt.Sprintf("@every %ds", int(duration.Seconds()))
	cherryLogger.Info(spec)
	return _cron.AddFunc(spec, cmd)
}

// AddSecondFunc 每x秒执行一次
func AddSecondFunc(cmd func(), t time.Duration) (cron.EntryID, error) {
	spec := fmt.Sprintf("@every %ds", int(t.Seconds()))
	cherryLogger.Info(spec)
	return _cron.AddFunc(spec, cmd)
}

func AddJob(spec string, cmd cron.Job) (cron.EntryID, error) {
	return _cron.AddJob(spec, cmd)
}

func Schedule(schedule cron.Schedule, cmd cron.Job) cron.EntryID {
	return _cron.Schedule(schedule, cmd)
}

func Entries() []cron.Entry {
	return _cron.Entries()
}

func Location() *time.Location {
	return _cron.Location()
}

func Entry(id cron.EntryID) cron.Entry {
	return _cron.Entry(id)
}

func Remove(id cron.EntryID) {
	_cron.Remove(id)
}

func Start() {
	_cron.Start()
}

func Run() {
	_cron.Run()
}

func Stop() {
	_cron.Stop()
}
