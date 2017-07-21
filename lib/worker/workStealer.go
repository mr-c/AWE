package worker

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/MG-RAST/AWE/lib/conf"
	"github.com/MG-RAST/AWE/lib/core"
	e "github.com/MG-RAST/AWE/lib/errors"
	"github.com/MG-RAST/AWE/lib/logger"
	"github.com/MG-RAST/AWE/lib/logger/event"
	"github.com/MG-RAST/golib/httpclient"
	"io/ioutil"
	"os"
	"strings"
	"syscall"
	"time"
)

type WorkResponse struct {
	Code int            `bson:"status" json:"status"`
	Data *core.Workunit `bson:"data" json:"data"`
	Errs []string       `bson:"error" json:"error"`
}

type TokenResponse struct {
	Code int      `bson:"status" json:"status"`
	Data string   `bson:"data" json:"data"`
	Errs []string `bson:"error" json:"error"`
}

func workStealer(control chan int) {

	fmt.Printf("workStealer launched, client=%s\n", core.Self.Id)
	logger.Debug(0, fmt.Sprintf("workStealer launched, client=%s\n", core.Self.Id))

	defer fmt.Printf("workStealer exiting...\n")
	retry := 0
	for {
		if core.Service == "proxy" {
			<-core.ProxyWorkChan
		}
		workunit, err := CheckoutWorkunitRemote()
		if err != nil {
			if err.Error() == e.QueueEmpty || err.Error() == e.QueueSuspend || err.Error() == e.NoEligibleWorkunitFound {
				//normal, do nothing
				logger.Debug(3, "(workStealer) client %s received status %s from server %s", core.Self.Id, err.Error(), conf.SERVER_URL)
			} else if err.Error() == e.ClientBusy {
				// client asked for work, but server has not finished processing its last delivered work
				logger.Error("(workStealer) server responds: last work delivered by client not yet processed, retry=%d", retry)
				retry += 1
			} else if err.Error() == e.ClientNotFound {
				logger.Error("(workStealer) server responds: client not found. will wait for heartbeat process to fix this")
				//server may be restarted, waiting for the hearbeater goroutine to try re-register
				//ReRegisterWithSelf(conf.SERVER_URL)
				// this has to be done by the heartbeat
			} else if err.Error() == e.ClientSuspended {
				logger.Error("(workStealer) client suspended, waiting for repair or resume request...")
				//TODO: send out email notice that this client has problem and been suspended
				time.Sleep(2 * time.Minute)
			} else if err.Error() == e.ClientDeleted {
				fmt.Printf("(workStealer) client deleted, exiting...\n")
				os.Exit(1) // TODO is there a better way of exiting ? E.g. in regard of the logger who wants to flush....
			} else {
				//something is wrong, server may be down
				logger.Error("(workStealer) checking out workunit: %s, retry=%d", err.Error(), retry)
				retry += 1
			}
			if core.Service != "proxy" { //proxy: event driven, client: timer driven
				if retry <= 10 {
					logger.Debug(3, "(workStealer) sleep 10 seconds")
					time.Sleep(10 * time.Second)
				} else {
					logger.Debug(3, "(workStealer) sleep 30 seconds")
					time.Sleep(30 * time.Second)
				}
			}
			continue
		} else {
			retry = 0
		}
		logger.Debug(1, "(workStealer) checked out workunit, id="+workunit.Id)
		//log event about work checktout (WC)
		logger.Event(event.WORK_CHECKOUT, "workid="+workunit.Id)

		err = core.Self.Add_work(workunit.Id)
		if err != nil {
			logger.Error("(workStealer) error: %s", err.Error())
			return
		}

		workmap.Set(workunit.Id, ID_WORKSTEALER, "workStealer")

		//hand the work to the next step handler: dataMover
		workstat := core.NewWorkPerf(workunit.Id)
		workstat.Checkout = time.Now().Unix()
		//rawWork := &Mediumwork{
		//	Workunit: wu,
		//	Perfstat: workstat,
		//}
		workunit.WorkPerf = workstat

		//FromStealer <- rawWork // sends to dataMover
		FromStealer <- workunit // sends to dataMover

		//if worker overlap is inhibited, wait until deliverer finishes processing the workunit
		if conf.WORKER_OVERLAP == false && core.Service != "proxy" {
			chanPermit <- true
			// sleep short time to allow server to finish processing last delivered work
			time.Sleep(2 * time.Second)
		}
	}
	control <- ID_WORKSTEALER //we are ending
}

func CheckoutWorkunitRemote() (workunit *core.Workunit, err error) {
	logger.Debug(3, "CheckoutWorkunitRemote()")
	// get available work dir disk space
	var stat syscall.Statfs_t
	syscall.Statfs(conf.WORK_PATH, &stat)
	availableBytes := stat.Bavail * uint64(stat.Bsize)

	response := new(WorkResponse)
	if core.Self == nil {
		err = fmt.Errorf("core.Self == nil")
		return
	}
	targeturl := fmt.Sprintf("%s/work?client=%s&available=%d", conf.SERVER_URL, core.Self.Id, availableBytes)

	var headers httpclient.Header
	if conf.CLIENT_GROUP_TOKEN != "" {
		headers = httpclient.Header{
			"Authorization": []string{"CG_TOKEN " + conf.CLIENT_GROUP_TOKEN},
		}
	}
	logger.Debug(3, fmt.Sprintf("client %s sends a checkout request to %s with available %d", core.Self.Id, conf.SERVER_URL, availableBytes))
	res, err := httpclient.DoTimeout("GET", targeturl, headers, nil, nil, time.Second*0)
	logger.Debug(3, fmt.Sprintf("client %s sent a checkout request to %s with available %d", core.Self.Id, conf.SERVER_URL, availableBytes))
	if err != nil {
		err = fmt.Errorf("error sending checkout request: %s", err.Error())
		return
	}
	defer res.Body.Close()
	jsonstream, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(jsonstream, response); err != nil {
		return
	}
	if len(response.Errs) > 0 {
		return nil, errors.New(strings.Join(response.Errs, ","))
	}
	if response.Code == 200 {
		workunit = response.Data
		if workunit.Info.Auth == true {
			if token, err := FetchDataTokenByWorkId(workunit.Id); err == nil && token != "" {
				workunit.Info.DataToken = token
			} else {
				return workunit, errors.New("need data token but failed to fetch one")
			}
		}
		logger.Debug(3, fmt.Sprintf("client %s got a workunit", core.Self.Id))
		workunit.State = core.WORK_STAT_CHECKOUT
		return workunit, nil
	}
	return
}

func FetchDataTokenByWorkId(workid string) (token string, err error) {
	targeturl := fmt.Sprintf("%s/work/%s?datatoken&client=%s", conf.SERVER_URL, workid, core.Self.Id)
	var headers httpclient.Header
	if conf.CLIENT_GROUP_TOKEN != "" {
		headers = httpclient.Header{
			"Authorization": []string{"CG_TOKEN " + conf.CLIENT_GROUP_TOKEN},
		}
	}
	res, err := httpclient.Get(targeturl, headers, nil, nil)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.Header != nil {
		if _, ok := res.Header["Datatoken"]; ok {
			token = res.Header["Datatoken"][0]
		}
	}
	return
}

func CheckoutTokenByJobId(jobid string) (token string, err error) {
	return
}
