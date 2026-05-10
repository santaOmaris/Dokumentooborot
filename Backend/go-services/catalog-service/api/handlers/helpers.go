package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func decodeJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func pathInt32(r *http.Request, key string) (int32, error) {
	v := r.PathValue(key)
	n, err := strconv.ParseInt(v, 10, 32)
	return int32(n), err
}
