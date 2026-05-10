package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func decodeJSON(r *http.Request, dst interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(dst)
}

func pathInt32(r *http.Request, key string) (int32, error) {
	v := r.PathValue(key)
	n, err := strconv.ParseInt(v, 10, 32)
	return int32(n), err
}

func queryInt32(r *http.Request, key string, defaultVal int32) int32 {
	v := r.URL.Query().Get(key)
	if v == "" {
		return defaultVal
	}
	n, err := strconv.ParseInt(v, 10, 32)
	if err != nil {
		return defaultVal
	}
	return int32(n)
}
