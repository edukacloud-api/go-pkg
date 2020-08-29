package api

type Response struct {
	Status  string `valid:"Required" json:"status"`
	Message string `valid:"Required" json:"message"`
}

type DefaultResponse struct {
	Data interface{} `json:"data"`
	Response
}

func CreateResponse(status string, message string, data interface{}) (response DefaultResponse) {
	response = DefaultResponse{
		Response: Response{
			Status:  status,
			Message: message,
		},
		Data: data,
	}

	return
}

