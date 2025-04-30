package f5

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"io"
	"net/http"
	"time"
)

type Model struct {
	User string
	Pass string
	Host string
	Port string
}

type PoolStats struct {
	Entries map[string]Entry `json:"entries"`
}

type Entry struct {
	NestedStats NestedStats `json:"nestedStats"`
}

type NestedStats struct {
	Entries Entries `json:"entries"`
}

type Entries struct {
	ActiveMemberCnt         ActiveMemberCnt `json:"activeMemberCnt"`
	AvailableMemberCnt      ActiveMemberCnt `json:"availableMemberCnt"`
	ConnqAllAgeEdm          ActiveMemberCnt `json:"connqAll.ageEdm"`
	ConnqAllAgeEma          ActiveMemberCnt `json:"connqAll.ageEma"`
	ConnqAllAgeHead         ActiveMemberCnt `json:"connqAll.ageHead"`
	ConnqAllAgeMax          ActiveMemberCnt `json:"connqAll.ageMax"`
	ConnqAllDepth           ActiveMemberCnt `json:"connqAll.depth"`
	ConnqAllServiced        ActiveMemberCnt `json:"connqAll.serviced"`
	ConnqAgeEdm             ActiveMemberCnt `json:"connq.ageEdm"`
	ConnqAgeEma             ActiveMemberCnt `json:"connq.ageEma"`
	ConnqAgeHead            ActiveMemberCnt `json:"connq.ageHead"`
	ConnqAgeMax             ActiveMemberCnt `json:"connq.ageMax"`
	ConnqDepth              ActiveMemberCnt `json:"connq.depth"`
	ConnqServiced           ActiveMemberCnt `json:"connq.serviced"`
	CurPriogrp              ActiveMemberCnt `json:"curPriogrp"`
	CurSessions             ActiveMemberCnt `json:"curSessions"`
	HighestPriogrp          ActiveMemberCnt `json:"highestPriogrp"`
	LowestPriogrp           ActiveMemberCnt `json:"lowestPriogrp"`
	MemberCnt               ActiveMemberCnt `json:"memberCnt"`
	MinActiveMembers        ActiveMemberCnt `json:"minActiveMembers"`
	MonitorRule             MonitorRule     `json:"monitorRule"`
	MrMsgIn                 ActiveMemberCnt `json:"mr.msgIn"`
	MrMsgOut                ActiveMemberCnt `json:"mr.msgOut"`
	MrReqIn                 ActiveMemberCnt `json:"mr.reqIn"`
	MrReqOut                ActiveMemberCnt `json:"mr.reqOut"`
	MrRespIn                ActiveMemberCnt `json:"mr.respIn"`
	MrRespOut               ActiveMemberCnt `json:"mr.respOut"`
	TmName                  MonitorRule     `json:"tmName"`
	ServersideBitsIn        ActiveMemberCnt `json:"serverside.bitsIn"`
	ServersideBitsOut       ActiveMemberCnt `json:"serverside.bitsOut"`
	ServersideCurConns      ActiveMemberCnt `json:"serverside.curConns"`
	ServersideMaxConns      ActiveMemberCnt `json:"serverside.maxConns"`
	ServersidePktsIn        ActiveMemberCnt `json:"serverside.pktsIn"`
	ServersidePktsOut       ActiveMemberCnt `json:"serverside.pktsOut"`
	ServersideTotConns      ActiveMemberCnt `json:"serverside.totConns"`
	StatusAvailabilityState MonitorRule     `json:"status.availabilityState"`
	StatusEnabledState      MonitorRule     `json:"status.enabledState"`
	StatusStatusReason      MonitorRule     `json:"status.statusReason"`
	TotRequests             ActiveMemberCnt `json:"totRequests"`
}

type ActiveMemberCnt struct {
	Value int64 `json:"value"`
}

type MonitorRule struct {
	Description string `json:"description"`
}

type F5Token struct {
	Token            Token `json:"token"`
	LastUpdateMicros int64 `json:"lastUpdateMicros"`
}

type Token struct {
	Token            string `json:"token"`
	Timeout          int64  `json:"timeout"`
	StartTime        string `json:"startTime"`
	LastUpdateMicros int64  `json:"lastUpdateMicros"`
	ExpirationMicros int64  `json:"expirationMicros"`
}

func (m *Model) Authenticate() (string, error) {

	var msg F5Token

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   10 * time.Second,
	}

	jsonStr, _ := json.Marshal(map[string]string{"username": m.User, "password": m.Pass, "loginProviderName": "tmos"})
	addr := fmt.Sprintf("%s:%s", m.Host, m.Port)
	req, err := http.NewRequest("POST", "https://"+addr+"/mgmt/shared/authn/login", bytes.NewBuffer(jsonStr))
	if err != nil {
		return "", err
	}

	req.Close = true
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		myError := errors.New("authentication failed")
		return "", myError
	}

	bodyText, err := io.ReadAll(resp.Body)
	err = json.Unmarshal(bodyText, &msg)
	if err != nil {
		return "", err
	}

	return msg.Token.Token, nil
}

func (m *Model) GetPoolStats(sessionId string) (PoolStats, error) {
	var msg PoolStats

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   10 * time.Second,
	}

	addr := fmt.Sprintf("%s:%s", m.Host, m.Port)
	req, err := http.NewRequest("GET", "https://"+addr+"/mgmt/tm/ltm/pool/stats", nil)
	if err != nil {
		return msg, err
	}

	req.Close = true
	req.Header.Add("X-F5-Auth-Token", sessionId)

	resp, err := client.Do(req)
	if err != nil {
		return msg, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err := fmt.Errorf("http response error: %v", resp.StatusCode)
		return msg, err
	}

	bodyText, err := io.ReadAll(resp.Body)

	err = json.Unmarshal(bodyText, &msg)
	if err != nil {

		return msg, err
	}

	return msg, nil
}

func (m *Model) GetSyncStatus(sessionId string) (int, error) {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   10 * time.Second,
	}

	addr := fmt.Sprintf("%s:%s", m.Host, m.Port)
	req, err := http.NewRequest("GET", "https://"+addr+"/mgmt/tm/cm/sync-status", nil)
	if err != nil {
		return 0, err
	}

	req.Close = true
	req.Header.Add("X-F5-Auth-Token", sessionId)

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, err
	}

	bodyText, err := io.ReadAll(resp.Body)

	syncStatus := gjson.Get(string(bodyText), "entries.https://localhost/mgmt/tm/cm/sync-status/0.nestedStats.entries.status.description")

	var status int
	switch syncStatus.String() {
	case "In Sync":
		status = 1
	default:
		status = 0
	}

	return status, nil
}
