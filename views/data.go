package views

const (
	AlertLevelError   = "danger"
	AlertLevelWarning = "warning"
	AlertLevelInfo    = "info"
	AlertLevelSuccess = "success"

	AlertMessageGeneric = "Something went wrong. Please try again and contact us if problem persists."
)

type Alert struct {
	Level   string
	Message string
}

type Data struct {
	Alert *Alert
	Yield interface{}
}

func (data *Data) SetAlert(err error) {
	if publicErr, ok := err.(PublicError); ok {
		data.Alert = &Alert{
			Level:   AlertLevelError,
			Message: publicErr.Public(),
		}
	} else {
		data.Alert = &Alert{
			Level:   AlertLevelError,
			Message: AlertMessageGeneric,
		}
	}
}

func (data *Data) AlertError(msg string) {
	data.Alert = &Alert{
		Level:   AlertLevelError,
		Message: msg,
	}
}

type PublicError interface {
	error
	Public() string
}
