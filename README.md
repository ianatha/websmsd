# websmsd

Exposes the SMS capabilities of a GSM modem over a REST API.

Tested only with a Huawei E173 (Firmware ver. 11.126.83.01.486) connected to a
Raspberry Pi 3.

## API

### GET /sms

Returns a list of SMS messages stored on the SIM card.
```json
[
  {
	"index": 0,
	"status": "REC READ",
	"from": "+12025550100",
	"date": "2023-01-01T00:00:00Z",
	"msg": "Hello World!"
  }
]
```

### DELETE /sms/:index

Returns a list of SMS messages stored on the SIM card.
```json
{
	"status": "Message deleted"
}
```