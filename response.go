package main

type Response struct {
	result interface{}
	code   int
}

func (r *Response) Metadata() map[string]interface{} {
	return map[string]interface{}{}
}

func (r *Response) Result() interface{} {
	return r.result
}

func (r *Response) StatusCode() int {
	return r.code
}
