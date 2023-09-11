package modem

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/tarm/serial"
)

type GSMModem struct {
	ComPort  string
	BaudRate int
	Port     *serial.Port
}

func New(ComPort string, BaudRate int) (modem *GSMModem) {
	modem = &GSMModem{ComPort: ComPort, BaudRate: BaudRate}
	return modem
}

func (m *GSMModem) Connect() (err error) {
	config := &serial.Config{Name: m.ComPort, Baud: m.BaudRate, ReadTimeout: time.Second}
	m.Port, err = serial.OpenPort(config)

	if err == nil {
		m.initModem()
	}

	return err
}

func (m *GSMModem) initModem() {
	m.SendCommand("ATE0\r\n", true) // echo off
	m.SendCommand("AT+CMEE=1\r\n", true) // useful error messages
	m.SendCommand("AT+CPMS=\"SM\"\r\n", true) // useful error messages
	m.SendCommand("AT+CMGF=1\r\n", true) // switch to TEXT mode
}

func (m *GSMModem) Expect(possibilities []string) (string, error) {
	readMax := 0
	for _, possibility := range possibilities {
		length := len(possibility)
		if length > readMax {
			readMax = length
		}
	}

	readMax = readMax + 2; // we need offset for \r\n sent by modem

	var status string = ""
	buf := make([]byte, readMax)

	for i := 0; i < readMax; i++ {
		// ignoring error as EOF raises error on Linux
		n, _ := m.Port.Read(buf)
		if n > 0 {
			status = string(buf[:n])

			for _, possibility := range possibilities {
				if strings.HasSuffix(status, possibility) {
					return status, nil
				}
			}
		}
	}

	return status, errors.New("match not found")
}

func (m *GSMModem) Send(command string) {
	m.Port.Flush()
	_, err := m.Port.Write([]byte(command))
	if err != nil {
		log.Fatal(err)
	}
}

func (m *GSMModem) Read(n int) string {
	var output string = "";
	buf := make([]byte, n)
	for i := 0; i < n; i++ {
		// ignoring error as EOF raises error on Linux
		c, _ := m.Port.Read(buf)
		if c > 0 {
			output = string(buf[:c])
		}
	}

	return output
}

func (m *GSMModem) ReadUntil(terminator string) string {
	var output string = "";
	terminate := false
	
	for !terminate {
		buf := make([]byte, 1)
		// ignoring error as EOF raises error on Linux
		c, _ := m.Port.Read(buf)
		if c > 0 {
			output = output + string(buf[0])
		}

		terminate = strings.HasSuffix(output, terminator)
	}

	output = strings.TrimSuffix(output, terminator)
	return output
}

func (m *GSMModem) SendCommand(command string, waitForOk bool) string {
	m.Send(command)

	if waitForOk {
		output, _ := m.Expect([]string{"OK\r\n", "ERROR\r\n"}) // we will not change api so errors are ignored for now
		return output
	} else {
		return m.Read(1)
	}
}

const CTRL_Z = string(rune(26))

func (m *GSMModem) SMSSend(mobile string, message string) string {
	m.Send("AT+CMGS=\""+mobile+"\"\r")
	m.Expect([]string{"> "})

	// EOM CTRL-Z
	return m.SendCommand(message+CTRL_Z, true)
}

type SMS struct {
	Index string `json:"index"`
	Status string `json:"status"`
	From string `json:"from"`
	Date time.Time `json:"date"`
	Msg string `json:"msg"`
}

func parseSMS(input string) []SMS {
	sections := strings.Split(input, "+CMGL:")
	smsList := []SMS{}

	for i := 1; i < len(sections); i++ {
		lines := strings.Split(strings.TrimSpace(sections[i]), "\n")
		metaData := strings.Split(lines[0], ",")
		index := metaData[0]
		status := strings.Trim(metaData[1], "\" ")
		from := strings.Trim(metaData[2], "\" ")
		dateString := strings.Trim(metaData[4], "\" ") + "," + strings.TrimSuffix(metaData[5], "\"\r")
		date, _ := time.Parse("06/01/02,15:04:05-07", dateString)

		message := ""
		if len(lines) > 1 {
			message = strings.Join(lines[1:], "\n")
		}

		sms := SMS{
			Index:  index,
			Status: status,
			From:   from,
			Date:   date,
			Msg:    message,
		}
		smsList = append(smsList, sms)
	}
	return smsList
}

func (m *GSMModem) SMSReadAll(q string) []SMS {
	m.Send(fmt.Sprintf("AT+CMGL=\"%s\"\r\n", q))
	s := m.ReadUntil("OK\r\n")
	s = strings.TrimPrefix(s, "\r\n")
	s = strings.TrimSuffix(s, "\r\n")
	return parseSMS(s)
}

func (m *GSMModem) SMSDelete(index int) string {
	m.Send(fmt.Sprintf("AT+CMGD=%d\r\n", index))
	output, _ := m.Expect([]string{"OK\r\n", "ERROR\r\n"})
	return output
}