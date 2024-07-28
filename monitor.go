package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/xlab/at"
	"github.com/xlab/at/sms"
)

const (
	DeviceCheckInterval = time.Second * 10
)

type State uint8

const (
	NoDeviceState State = iota
	ReadyState
)

type PersistedMessage struct {
	sms.Message

	Id string
}

type Monitor struct {
	Messages []*PersistedMessage
	Ready    bool

	cmdPort      string
	notifyPort   string
	smsStorePath string

	dev          *at.Device
	stateChanged chan State
	checkTimer   *time.Timer
}

func (m *Monitor) DeviceState() *at.DeviceState {
	return m.dev.State
}

func NewMonitor(cmdPort, notifyPort, smsStorePath string) *Monitor {
	result := &Monitor{
		cmdPort:      cmdPort,
		notifyPort:   notifyPort,
		smsStorePath: smsStorePath,
		stateChanged: make(chan State, 10),
	}
	result.loadMessages(result.smsStorePath)
	return result
}

func (m *Monitor) devStop() {
	if m.dev != nil {
		m.dev.Close()
	}
}

func (m *Monitor) Run() (err error) {
	m.checkTimer = time.NewTimer(DeviceCheckInterval)
	defer m.checkTimer.Stop()
	defer m.devStop()

	go func() {
		for {
			<-m.checkTimer.C
			if err := m.openDevice(); err != nil {
				log.Println("failed to open device:", err)
				m.checkTimer.Reset(DeviceCheckInterval)
				continue
			} else {
				m.checkTimer.Stop()
				m.stateChanged <- ReadyState
			}
		}
	}()

	if err := m.openDevice(); err != nil {
		m.stateChanged <- NoDeviceState
	} else {
		m.stateChanged <- ReadyState
		m.checkTimer.Stop()
	}

	go func() {
		for s := range m.stateChanged {
			switch s {
			case NoDeviceState:
				m.Ready = false
				log.Println("Waiting for device")
				m.checkTimer.Reset(DeviceCheckInterval)
			case ReadyState:
				log.Println("Device connected")
				m.Ready = true
				go func() {
					m.dev.Watch()
					m.stateChanged <- NoDeviceState
				}()
				go func() {
					for {
						select {
						case <-m.dev.Closed():
							return
						case msg, ok := <-m.dev.IncomingSms():
							if ok {
								log.Println("Received an SMS")
								m.addMessage(msg)
							}
						}
					}
				}()
			}
		}
	}()

	ginEngine := m.newGinEngine()
	return ginEngine.Run()
}

func (m *Monitor) addMessage(incomingMsg *sms.Message) {
	id := uuid.New().String()
	msg := &PersistedMessage{
		Message: *incomingMsg,
		Id:      id,
	}

	m.Messages = append(m.Messages, msg)

	err := m.storeMessages(m.Messages, m.smsStorePath)
	if err != nil {
		log.Println("Failed to store messages:", err)
	}
}

func (m *Monitor) storeMessages(messages []*PersistedMessage, filename string) error {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(messages); err != nil {
		return err
	}

	return nil
}

func (m *Monitor) deleteMessageWithId(id string) {
	for i, msg := range m.Messages {
		if msg.Id == id {
			m.Messages = append(m.Messages[:i], m.Messages[i+1:]...)
			break
		}
	}
	err := m.storeMessages(m.Messages, m.smsStorePath)
	if err != nil {
		log.Println("Failed to store messages:", err)
	}
}

func (m *Monitor) loadMessages(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Println("Failed to open file:", err)
		return
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&m.Messages); err != nil {
		log.Println("Failed to decode messages:", err)
		return
	}
}

func (m *Monitor) openDevice() (err error) {
	m.dev = &at.Device{
		CommandPort: m.cmdPort,
		NotifyPort:  m.notifyPort,
	}
	if err = m.dev.Open(); err != nil {
		return
	}
	if err = m.dev.Init(at.DeviceE173()); err != nil {
		return
	}
	return
}
