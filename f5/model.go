package f5

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

	client := &http.Client{Transport: tr}

	var jsonStr = []byte(`{"username":"monitoring", "password":"#TrueCom2024#"}`)

	addr := fmt.Sprintf("%s:%s", m.Host, m.Port)
	req, err := http.NewRequest("POST", addr+"/mgmt/shared/authn/login", bytes.NewBuffer(jsonStr))
	if err != nil {
		return "", err
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err := fmt.Errorf("http response error: %v", resp.StatusCode)
		return "", err
	}

	bodyText, err := io.ReadAll(resp.Body)

	err = json.Unmarshal(bodyText, &msg)
	if err != nil {
		fmt.Println("Unable to unmarshal")
		return "", err
	}

	return msg.Token.Token, nil
}

func (m *Model) GetPoolStats(sessionId string) (PoolStats, error) {
	var msg PoolStats

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}

	addr := fmt.Sprintf("%s:%s", m.Host, m.Port)
	req, err := http.NewRequest("GET", addr+"/mgmt/tm/ltm/pool/stats", nil)
	if err != nil {
		return msg, err
	}

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
