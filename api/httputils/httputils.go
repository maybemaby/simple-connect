package httputils

import (
	"encoding/json"
	"net/http"
	"time"

	"connectrpc.com/connect"
)

func SetCookie(r connect.AnyResponse, cookie http.Cookie) {
	r.Header().Add("Set-Cookie", cookie.String())
}

func GetCookie(r connect.AnyRequest, name string) (*http.Cookie, error) {
	cookieString := r.Header().Get("Cookie")

	cookies, err := http.ParseCookie(cookieString)

	if err != nil {
		return nil, err
	}

	var cookie *http.Cookie

	for _, c := range cookies {
		if c.Name == name {
			cookie = c
			break
		}
	}

	if cookie == nil {
		return nil, http.ErrNoCookie
	}

	return cookie, nil
}

func DeleteCookie(r connect.AnyResponse, name string) {
	deletedCookie := http.Cookie{
		Name:    name,
		Value:   "",
		Path:    "/",
		Expires: time.Unix(0, 0),
	}

	r.Header().Add("Set-Cookie", deletedCookie.String())
}

func ReadJSON[T any](r *http.Request, target T) error {
	dec := json.NewDecoder(r.Body)
	return dec.Decode(target)
}

func WriteJSON[T any](w http.ResponseWriter, r *http.Request, data T) error {
	enc := json.NewEncoder(w)

	err := enc.Encode(data)

	if err != nil {
		return err
	}

	r.Header.Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}