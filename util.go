package ispend

import (
	"encoding/json"
	"io"
)

func SendAPIResp(w io.Writer, data interface{}) error {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = w.Write(dataBytes)
	if err != nil {
		return err
	}
	return nil
}

func SendAPIOKResp(w io.Writer, message string) error {
	apiResp := APIResponse{Status: 200, Message: message}
	return SendAPIResp(w, apiResp)
}

func SendAPIOKRespWithData(w io.Writer, message string, data interface{}) error {
	apiResp := APIResponse{Status: 200, Message: message, Data: data}
	return SendAPIResp(w, apiResp)
}

func SendAPIErrorResp(w io.Writer, message string, status int) error {
	apiErr := APIResponse{Status: status, Message: message}
	return SendAPIResp(w, apiErr)
}
